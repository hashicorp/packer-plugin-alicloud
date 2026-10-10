// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"fmt"
	"time"

	ecs20140526Client "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"
)

// DescribeImages uses the v7 ECS SDK to query images.
func (c *ClientWrapper) DescribeImages(request *ecs20140526Client.DescribeImagesRequest) (*ecs20140526Client.DescribeImagesResponse, error) {
	if c.v7Client == nil {
		return nil, fmt.Errorf("v7 ECS client is not initialized")
	}
	return c.v7Client.DescribeImages(request)
}

// WaitForImageStatus polls DescribeImages via the v7 SDK until the image reaches
// the expected status or the timeout expires.
func (c *ClientWrapper) WaitForImageStatus(regionId string, imageId string, expectedStatus string, timeout time.Duration) (*ecs20140526Client.DescribeImagesResponse, error) {
	if timeout <= 0 {
		timeout = defaultRetryTimes * defaultRetryInterval
	}

	interval := defaultRetryInterval
	deadline := time.Now().Add(timeout)

	var lastResponse *ecs20140526Client.DescribeImagesResponse
	var lastError error

	for {
		request := &ecs20140526Client.DescribeImagesRequest{
			RegionId: tea.String(regionId),
			ImageId:  tea.String(imageId),
			Status:   tea.String(ImageStatusQueried),
		}

		response, err := c.DescribeImages(request)
		lastResponse = response
		lastError = err

		if err == nil && response != nil && response.Body != nil && response.Body.Images != nil {
			for _, image := range response.Body.Images.Image {
				if tea.StringValue(image.Status) == expectedStatus {
					return response, nil
				}
			}
		}

		if time.Now().After(deadline) {
			break
		}
		time.Sleep(interval)
	}

	if lastError == nil {
		lastError = fmt.Errorf("<no error>")
	}
	return lastResponse, fmt.Errorf("evaluate failed after %d seconds timeout with %d seconds retry interval: %s", int(timeout.Seconds()), int(interval.Seconds()), lastError)
}

// WaitForImageUsable polls DescribeImages via the v7 SDK until the image's
// Usable field becomes true or the timeout expires.
func (c *ClientWrapper) WaitForImageUsable(regionId string, imageId string, timeout time.Duration) (*ecs20140526Client.DescribeImagesResponse, error) {
	if timeout <= 0 {
		timeout = defaultRetryTimes * defaultRetryInterval
	}

	interval := defaultRetryInterval
	deadline := time.Now().Add(timeout)

	var lastResponse *ecs20140526Client.DescribeImagesResponse
	var lastError error

	for {
		request := &ecs20140526Client.DescribeImagesRequest{
			RegionId: tea.String(regionId),
			ImageId:  tea.String(imageId),
			Status:   tea.String(ImageStatusQueried),
		}

		response, err := c.DescribeImages(request)
		lastResponse = response
		lastError = err

		if err == nil && response != nil && response.Body != nil && response.Body.Images != nil {
			for _, image := range response.Body.Images.Image {
				if tea.BoolValue(image.Usable) {
					return response, nil
				}
			}
		}

		if time.Now().After(deadline) {
			break
		}
		time.Sleep(interval)
	}

	if lastError == nil {
		lastError = fmt.Errorf("<no error>")
	}
	return lastResponse, fmt.Errorf("evaluate failed after %d seconds timeout with %d seconds retry interval: %s", int(timeout.Seconds()), int(interval.Seconds()), lastError)
}
