// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	confighelper "github.com/hashicorp/packer-plugin-sdk/template/config"
)

type stepRegionCopyAlicloudImage struct {
	AlicloudImageDestinationRegions []string
	AlicloudImageDestinationNames   []string
	RegionId                        string
	WaitCopyingImageReadyTimeout    int
}

func (s *stepRegionCopyAlicloudImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)

	// When the image is configured to be encrypted, we must include the source
	// region itself in the destination list. CopyImage within the same region is
	// the API operation that produces an encrypted copy of the original image,
	// so skipping the source region would leave no encrypted image behind.
	if config.ImageEncrypted != confighelper.TriUnset {
		s.AlicloudImageDestinationRegions = append(s.AlicloudImageDestinationRegions, s.RegionId)
		s.AlicloudImageDestinationNames = append(s.AlicloudImageDestinationNames, config.AlicloudImageName)
	}

	if len(s.AlicloudImageDestinationRegions) == 0 {
		return multistep.ActionContinue
	}

	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packersdk.Ui)

	srcImageId := state.Get("alicloudimage").(string)
	alicloudImages := state.Get("alicloudimages").(map[string]string)
	numberOfName := len(s.AlicloudImageDestinationNames)

	ui.Say(fmt.Sprintf("Coping image %s from %s...", srcImageId, s.RegionId))
	crossRegionIndex := -1
	for index, destinationRegion := range s.AlicloudImageDestinationRegions {
		// Normally there is no reason to copy an image to the same region it
		// already lives in. However, when image encryption is enabled, CopyImage
		// must still be invoked for the source region so that an encrypted copy
		// of the image is created. Do not simplify this to a plain same-region
		// skip without considering the encryption use case.
		if destinationRegion == s.RegionId && !config.ImageEncrypted.True() {
			continue
		}

		if destinationRegion != s.RegionId {
			crossRegionIndex++
		}

		copyImageRequest := s.buildCopyImageRequest(index, destinationRegion, config, srcImageId, numberOfName, crossRegionIndex)

		imageResponse, err := client.CopyImage(copyImageRequest)
		if err != nil {
			return halt(state, err, "Error copying images")
		}

		alicloudImages[destinationRegion] = imageResponse.ImageId
		ui.Message(fmt.Sprintf("Copy image from %s(%s) to %s(%s)", s.RegionId, srcImageId, destinationRegion, imageResponse.ImageId))
	}

	if config.ImageEncrypted != confighelper.TriUnset {
		if _, err := client.WaitForImageStatus(s.RegionId, alicloudImages[s.RegionId], ImageStatusAvailable, time.Duration(s.WaitCopyingImageReadyTimeout)*time.Second); err != nil {
			return halt(state, err, fmt.Sprintf("Timeout waiting image %s finish copying", alicloudImages[s.RegionId]))
		}
	}

	return multistep.ActionContinue
}

func (s *stepRegionCopyAlicloudImage) buildCopyImageRequest(index int, destinationRegion string, config *Config, srcImageId string, numberOfName int, crossRegionIndex int) *ecs.CopyImageRequest {
	// Leave the destination image name empty by default so that ECS auto-generates
	// a unique name. Reusing config.AlicloudImageName here would force every
	// destination region to share the same name, causing conflicts on re-runs
	// unless image_force_delete is enabled.
	ecsImageName := ""
	if numberOfName > 0 && index < numberOfName {
		ecsImageName = s.AlicloudImageDestinationNames[index]
	}

	copyImageRequest := ecs.CreateCopyImageRequest()
	copyImageRequest.RegionId = s.RegionId
	copyImageRequest.ImageId = srcImageId
	copyImageRequest.DestinationRegionId = destinationRegion
	copyImageRequest.DestinationImageName = ecsImageName
	copyImageRequest.ResourceGroupId = config.AlicloudResourceGroupId
	if config.ImageEncrypted.True() {
		copyImageRequest.Encrypted = requests.NewBoolean(true)
		if destinationRegion == s.RegionId {
			if config.KMSKeyId != "" {
				copyImageRequest.KMSKeyId = config.KMSKeyId
			}
		} else {
			if crossRegionIndex < len(config.ImageCopyKMSKeyIds) && config.ImageCopyKMSKeyIds[crossRegionIndex] != "" {
				copyImageRequest.KMSKeyId = config.ImageCopyKMSKeyIds[crossRegionIndex]
			}
		}
	}

	return copyImageRequest
}

func (s *stepRegionCopyAlicloudImage) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if !cancelled && !halted {
		return
	}

	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Stopping copy image because cancellation or error...")

	client := state.Get("client").(*ClientWrapper)
	alicloudImages := state.Get("alicloudimages").(map[string]string)
	srcImageId := state.Get("alicloudimage").(string)

	for copiedRegionId, copiedImageId := range alicloudImages {
		if copiedImageId == srcImageId {
			continue
		}

		cancelCopyImageRequest := ecs.CreateCancelCopyImageRequest()
		cancelCopyImageRequest.RegionId = copiedRegionId
		cancelCopyImageRequest.ImageId = copiedImageId
		if _, err := client.CancelCopyImage(cancelCopyImageRequest); err != nil {

			ui.Error(fmt.Sprintf("Error cancelling copy image: %v", err))
		}
	}
}
