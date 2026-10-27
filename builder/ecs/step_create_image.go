// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/random"

	ecs20140526Client "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
)

type stepCreateAlicloudImage struct {
	AlicloudImageIgnoreDataDisks bool
	WaitSnapshotReadyTimeout     int
	Tags                         map[string]string
	EnableImageInstanceAccess    bool
	image                        *ecs20140526Client.DescribeImagesResponseBodyImagesImage
}

var createImageRetryErrors = []string{
	"IdempotentProcessing",
}

func (s *stepCreateAlicloudImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packersdk.Ui)

	tempImageName := config.AlicloudImageName
	if config.ImageEncrypted.True() {
		tempImageName = fmt.Sprintf("packer_%s", random.AlphaNum(7))
		ui.Say(fmt.Sprintf("Creating temporary image for encryption: %s", tempImageName))
	} else {
		ui.Say(fmt.Sprintf("Creating image: %s", tempImageName))
	}

	createImageRequest := s.buildCreateImageRequest(state, tempImageName)
	createImageResponse, err := client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			return client.CreateImage(createImageRequest)
		},
		EvalFunc: client.EvalCouldRetryResponse(createImageRetryErrors, EvalRetryErrorType),
	})

	if err != nil {
		return halt(state, err, "Error creating image")
	}

	imageId := createImageResponse.(*ecs.CreateImageResponse).ImageId

	// The source instance can be released right after CreateImage succeeds,
	// regardless of whether the image is usable or fully available.
	s.deleteSourceInstance(state)

	var imagesResponse *ecs20140526Client.DescribeImagesResponse
	if s.EnableImageInstanceAccess {
		imagesResponse, err = client.WaitForImageUsable(config.AlicloudRegion, imageId, time.Duration(s.WaitSnapshotReadyTimeout)*time.Second)
	} else {
		imagesResponse, err = client.WaitForImageStatus(config.AlicloudRegion, imageId, ImageStatusAvailable, time.Duration(s.WaitSnapshotReadyTimeout)*time.Second)
	}
	if err != nil {
		return halt(state, err, "Timeout waiting for image to be created")
	}

	// save image first for cleaning up if timeout
	images := imagesResponse.Body.Images.Image
	if len(images) == 0 {
		return halt(state, err, "Unable to find created image")
	}
	s.image = images[0]

	// The images response is from the last DescribeImages poll of the wait
	// above, i.e. the image state observed when the wait condition was met.
	logImageState(ui, "Image state (wait finished)", s.image)

	var snapshotIds []string
	for _, device := range s.image.DiskDeviceMappings.DiskDeviceMapping {
		snapshotIds = append(snapshotIds, tea.StringValue(device.SnapshotId))
	}

	state.Put("alicloudimage", imageId)
	state.Put("alicloudsnapshots", snapshotIds)

	alicloudImages := make(map[string]string)
	alicloudImages[config.AlicloudRegion] = tea.StringValue(s.image.ImageId)
	state.Put("alicloudimages", alicloudImages)

	return multistep.ActionContinue
}

// logImageState prints one line describing the image's status, progress and
// usable fields.
func logImageState(ui packersdk.Ui, prefix string, image *ecs20140526Client.DescribeImagesResponseBodyImagesImage) {
	if image == nil {
		return
	}
	ui.Say(fmt.Sprintf("%s: image=%s, status=%s, progress=%s, usable=%t",
		prefix,
		tea.StringValue(image.ImageId),
		tea.StringValue(image.Status),
		tea.StringValue(image.Progress),
		tea.BoolValue(image.Usable)))
}

func (s *stepCreateAlicloudImage) Cleanup(state multistep.StateBag) {
	if s.image == nil {
		return
	}

	config := state.Get("config").(*Config)
	encryptedSet := config.ImageEncrypted.True()

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if !cancelled && !halted && !encryptedSet {
		return
	}

	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packersdk.Ui)

	if !cancelled && !halted && encryptedSet {
		ui.Say(fmt.Sprintf("Deleting temporary image %s(%s) and related snapshots after finishing encryption...", tea.StringValue(s.image.ImageId), tea.StringValue(s.image.ImageName)))
	} else {
		ui.Say("Deleting the image and related snapshots because of cancellation or error...")
	}

	deleteImageRequest := ecs.CreateDeleteImageRequest()
	deleteImageRequest.RegionId = config.AlicloudRegion
	deleteImageRequest.ImageId = tea.StringValue(s.image.ImageId)
	if _, err := client.DeleteImage(deleteImageRequest); err != nil {
		ui.Error(fmt.Sprintf("Error deleting image, it may still be around: %s", err))
		return
	}

	//Delete the snapshot of this image
	for _, diskDevices := range s.image.DiskDeviceMappings.DiskDeviceMapping {
		deleteSnapshotRequest := ecs.CreateDeleteSnapshotRequest()
		deleteSnapshotRequest.SnapshotId = tea.StringValue(diskDevices.SnapshotId)
		if _, err := client.DeleteSnapshot(deleteSnapshotRequest); err != nil {
			ui.Error(fmt.Sprintf("Error deleting snapshot, it may still be around: %s", err))
			return
		}
	}
}

func (s *stepCreateAlicloudImage) deleteSourceInstance(state multistep.StateBag) {
	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packersdk.Ui)

	if instanceRaw, ok := state.GetOk("instance"); ok {
		instance := instanceRaw.(*ecs.Instance)
		ui.Say(fmt.Sprintf("Deleting source instance %s after CreateImage succeeded...", instance.InstanceId))

		request := ecs.CreateDeleteInstanceRequest()
		request.InstanceId = instance.InstanceId
		request.Force = requests.NewBoolean(true)

		// The instance may still be initializing, in which case
		// DeleteInstance fails with IncorrectInstanceStatus.Initializing.
		// Retry for up to 2 minutes before falling back to the end-of-build
		// instance cleanup.
		_, err := client.WaitForExpected(&WaitForExpectArgs{
			RequestFunc: func() (responses.AcsResponse, error) {
				return client.DeleteInstance(request)
			},
			EvalFunc:     client.EvalCouldRetryResponse(deleteInstanceRetryErrors, EvalRetryErrorType),
			RetryTimeout: 2 * time.Minute,
		})
		if err != nil {
			ui.Error(fmt.Sprintf("Error deleting source instance %s after CreateImage: %s", instance.InstanceId, err))
			return
		}
		state.Put("instance_deleted_early", true)
	}
}

func (s *stepCreateAlicloudImage) buildCreateImageRequest(state multistep.StateBag, imageName string) *ecs.CreateImageRequest {
	config := state.Get("config").(*Config)

	request := ecs.CreateCreateImageRequest()
	request.ClientToken = uuid.TimeOrderedUUID()
	request.RegionId = config.AlicloudRegion
	request.ImageName = imageName
	request.ImageVersion = config.AlicloudImageVersion
	request.Description = config.AlicloudImageDescription
	request.ResourceGroupId = config.AlicloudResourceGroupId
	request.ImageFamily = config.AlicloudTargetImageFamily
	request.BootMode = config.AlicloudBootMode

	if s.AlicloudImageIgnoreDataDisks {
		snapshotId := state.Get("alicloudsnapshot").(string)
		request.SnapshotId = snapshotId
	} else {
		instance := state.Get("instance").(*ecs.Instance)
		request.InstanceId = instance.InstanceId
	}

	if len(s.Tags) != 0 {
		var tags []ecs.CreateImageTag
		for key, value := range s.Tags {
			var tag ecs.CreateImageTag
			tag.Key = key
			tag.Value = value
			tags = append(tags, tag)
		}
		request.Tag = &tags
	}

	return request
}
