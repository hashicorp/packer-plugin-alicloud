// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepCheckAlicloudImageFamily struct {
	ImageFamily string
}

func (s *stepCheckAlicloudImageFamily) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ClientWrapper)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	describeImagesFromFamilyRequest := ecs.CreateDescribeImageFromFamilyRequest()
	describeImagesFromFamilyRequest.RegionId = config.AlicloudRegion
	describeImagesFromFamilyRequest.ImageFamily = config.AlicloudImageFamily

	imagesResponse, err := client.DescribeImageFromFamily(describeImagesFromFamilyRequest)
	if err != nil {
		return halt(state, err, "Error querying alicloud image by image family")
	}

	imageId := imagesResponse.Image.ImageId

	if imageId == "" {
		err := fmt.Errorf("No alicloud image was found matching image family: %s", config.AlicloudImageFamily)
		return halt(state, err, "")
	}

	ui.Message(fmt.Sprintf("Found lastest image: %s by image family: %s", imageId, config.AlicloudImageFamily))

	return multistep.ActionContinue
}

func (s *stepCheckAlicloudImageFamily) Cleanup(multistep.StateBag) {}
