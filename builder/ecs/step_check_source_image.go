// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"

	ecs20140526Client "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepCheckAlicloudSourceImage struct {
	SourceECSImageId string
}

func (s *stepCheckAlicloudSourceImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ClientWrapper)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	describeImagesRequest := &ecs20140526Client.DescribeImagesRequest{
		RegionId: tea.String(config.AlicloudRegion),
		ImageId:  tea.String(config.AlicloudSourceImage),
	}
	if config.AlicloudSkipImageValidation {
		describeImagesRequest.ShowExpired = tea.Bool(true)
	}
	imagesResponse, err := client.DescribeImages(describeImagesRequest)
	if err != nil {
		return halt(state, err, "Error querying alicloud image")
	}

	images := imagesResponse.Body.Images.Image

	// Describe marketplace image
	describeImagesRequest.ImageOwnerAlias = tea.String("marketplace")
	marketImagesResponse, err := client.DescribeImages(describeImagesRequest)
	if err != nil {
		return halt(state, err, "Error querying alicloud marketplace image")
	}

	marketImages := marketImagesResponse.Body.Images.Image
	if len(marketImages) > 0 {
		images = append(images, marketImages...)
	}

	if len(images) == 0 {
		err := fmt.Errorf("No alicloud image was found matching filters: %v", config.AlicloudSourceImage)
		return halt(state, err, "")
	}

	ui.Message(fmt.Sprintf("Found image ID: %s", tea.StringValue(images[0].ImageId)))

	state.Put("source_image", images[0])
	return multistep.ActionContinue
}

func (s *stepCheckAlicloudSourceImage) Cleanup(multistep.StateBag) {}
