// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"fmt"
	"log"
	"sort"
	"strings"

	ecs20140526Client "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type Artifact struct {
	// A map of regions to alicloud image IDs.
	AlicloudImages map[string]string

	// BuilderId is the unique ID for the builder that created this alicloud image
	BuilderIdValue string

	// Alcloud connection for performing API stuff.
	Client *ClientWrapper
}

func (a *Artifact) BuilderId() string {
	return a.BuilderIdValue
}

func (*Artifact) Files() []string {
	// We have no files
	return nil
}

func (a *Artifact) Id() string {
	parts := make([]string, 0, len(a.AlicloudImages))
	for region, ecsImageId := range a.AlicloudImages {
		parts = append(parts, fmt.Sprintf("%s:%s", region, ecsImageId))
	}

	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func (a *Artifact) String() string {
	alicloudImageStrings := make([]string, 0, len(a.AlicloudImages))
	for region, id := range a.AlicloudImages {
		single := fmt.Sprintf("%s: %s", region, id)
		alicloudImageStrings = append(alicloudImageStrings, single)
	}

	sort.Strings(alicloudImageStrings)
	return fmt.Sprintf("Alicloud images were created:\n\n%s", strings.Join(alicloudImageStrings, "\n"))
}

func (a *Artifact) State(name string) interface{} {
	switch name {
	case "atlas.artifact.metadata":
		return a.stateAtlasMetadata()
	default:
		return nil
	}
}

func (a *Artifact) Destroy() error {
	errors := make([]error, 0)

	copyingImages := make(map[string]string, len(a.AlicloudImages))
	sourceImage := make(map[string]*ecs20140526Client.DescribeImagesResponseBodyImagesImage, 1)
	for regionId, imageId := range a.AlicloudImages {
		describeImagesRequest := &ecs20140526Client.DescribeImagesRequest{
			RegionId: tea.String(regionId),
			ImageId:  tea.String(imageId),
			Status:   tea.String(ImageStatusQueried),
		}
		imagesResponse, err := a.Client.DescribeImages(describeImagesRequest)
		if err != nil {
			errors = append(errors, err)
		}

		images := imagesResponse.Body.Images.Image
		if len(images) == 0 {
			err := fmt.Errorf("Error retrieving details for alicloud image(%s), no alicloud images found", imageId)
			errors = append(errors, err)
			continue
		}

		image := images[0]
		if tea.BoolValue(image.IsCopied) && tea.StringValue(image.Status) != ImageStatusAvailable {
			copyingImages[regionId] = imageId
		} else {
			sourceImage[regionId] = image
		}
	}

	for regionId, imageId := range copyingImages {
		log.Printf("Cancel copying alicloud image (%s) from region (%s)", imageId, regionId)

		errs := a.unsharedAccountsOnImages(regionId, imageId)
		if errs != nil {
			errors = append(errors, errs...)
		}

		cancelImageCopyRequest := ecs.CreateCancelCopyImageRequest()
		cancelImageCopyRequest.RegionId = regionId
		cancelImageCopyRequest.ImageId = imageId
		if _, err := a.Client.CancelCopyImage(cancelImageCopyRequest); err != nil {
			errors = append(errors, err)
		}
	}

	for regionId, image := range sourceImage {
		imageId := tea.StringValue(image.ImageId)
		log.Printf("Delete alicloud image (%s) from region (%s)", imageId, regionId)

		errs := a.unsharedAccountsOnImages(regionId, imageId)
		if errs != nil {
			errors = append(errors, errs...)
		}

		deleteImageRequest := ecs.CreateDeleteImageRequest()
		deleteImageRequest.RegionId = regionId
		deleteImageRequest.ImageId = imageId
		if _, err := a.Client.DeleteImage(deleteImageRequest); err != nil {
			errors = append(errors, err)
		}

		//Delete the snapshot of this images
		for _, diskDevices := range image.DiskDeviceMappings.DiskDeviceMapping {
			deleteSnapshotRequest := ecs.CreateDeleteSnapshotRequest()
			deleteSnapshotRequest.SnapshotId = tea.StringValue(diskDevices.SnapshotId)
			_, err := a.Client.DeleteSnapshot(deleteSnapshotRequest)
			if err != nil {
				errors = append(errors, err)
			}
		}
	}

	if len(errors) > 0 {
		if len(errors) == 1 {
			return errors[0]
		} else {
			return &packersdk.MultiError{Errors: errors}
		}
	}

	return nil
}

func (a *Artifact) unsharedAccountsOnImages(regionId string, imageId string) []error {
	var errors []error

	describeImageShareRequest := ecs.CreateDescribeImageSharePermissionRequest()
	describeImageShareRequest.RegionId = regionId
	describeImageShareRequest.ImageId = imageId
	imageShareResponse, err := a.Client.DescribeImageSharePermission(describeImageShareRequest)
	if err != nil {
		errors = append(errors, err)
		return errors
	}

	accountsNumber := len(imageShareResponse.Accounts.Account)
	if accountsNumber > 0 {
		accounts := make([]string, accountsNumber)
		for index, account := range imageShareResponse.Accounts.Account {
			accounts[index] = account.AliyunId
		}

		modifyImageShareRequest := ecs.CreateModifyImageSharePermissionRequest()
		modifyImageShareRequest.RegionId = regionId
		modifyImageShareRequest.ImageId = imageId
		modifyImageShareRequest.RemoveAccount = &accounts
		_, err := a.Client.ModifyImageSharePermission(modifyImageShareRequest)
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func (a *Artifact) stateAtlasMetadata() interface{} {
	metadata := make(map[string]string)
	for region, imageId := range a.AlicloudImages {
		k := fmt.Sprintf("region.%s", region)
		metadata[k] = imageId
	}

	return metadata
}
