// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/hashicorp/packer-plugin-sdk/uuid"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	confighelper "github.com/hashicorp/packer-plugin-sdk/template/config"
)

type stepCreateAlicloudInstance struct {
	IOOptimized                 confighelper.Trilean
	InstanceType                string
	UserData                    string
	UserDataFile                string
	RamRoleName                 string
	Tags                        map[string]string
	RegionId                    string
	InternetChargeType          string
	InternetMaxBandwidthOut     int
	InstanceName                string
	ZoneId                      string
	SecurityEnhancementStrategy string
	AlicloudImageFamily         string
	instance                    *ecs.Instance
}

var createInstanceRetryErrors = []string{
	"IdempotentProcessing",
}

var deleteInstanceRetryErrors = []string{
	"IncorrectInstanceStatus.Initializing",
}

func (s *stepCreateAlicloudInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Creating instance...")
	runInstanceRequest, err := s.buildCreateInstanceRequest(state)
	if err != nil {
		return halt(state, err, "")
	}

	runInstancesResponse, err := client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			return client.RunInstances(runInstanceRequest)
		},
		EvalFunc: client.EvalCouldRetryResponse(createInstanceRetryErrors, EvalRetryErrorType),
	})

	if err != nil {
		return halt(state, err, "Error creating instance")
	}

	instanceId := runInstancesResponse.(*ecs.RunInstancesResponse).InstanceIdSets.InstanceIdSet[0]

	_, err = client.WaitForInstanceStatus(s.RegionId, instanceId, InstanceStatusRunning)
	if err != nil {
		return halt(state, err, "Error waiting create instance")
	}

	describeInstancesRequest := ecs.CreateDescribeInstancesRequest()
	describeInstancesRequest.InstanceIds = fmt.Sprintf("[\"%s\"]", instanceId)
	instances, err := client.DescribeInstances(describeInstancesRequest)
	if err != nil {
		return halt(state, err, "")
	}
	status := instances.Instances.Instance[0].Status
	if status == InstanceStatusRunning {
		stopInstanceRequest := ecs.CreateStopInstanceRequest()
		stopInstanceRequest.InstanceId = instanceId
		if _, err := client.StopInstance(stopInstanceRequest); err != nil {
			return halt(state, err, "Error stopping instance")
		}

		ui.Say(fmt.Sprintf("Stoping instance: %s", instanceId))

		_, err = client.WaitForInstanceStatus(s.RegionId, instanceId, InstanceStatusStopped)
		if err != nil {
			return halt(state, err, "Timeout waiting for instance to stop")
		}
	}

	ui.Message(fmt.Sprintf("Created instance: %s", instanceId))
	s.instance = &instances.Instances.Instance[0]
	state.Put("instance", s.instance)
	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put("instance_id", instanceId)

	return multistep.ActionContinue
}

func (s *stepCreateAlicloudInstance) Cleanup(state multistep.StateBag) {
	if s.instance == nil {
		return
	}
	cleanUpMessage(state, "instance")

	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packersdk.Ui)

	_, err := client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			request := ecs.CreateDeleteInstanceRequest()
			request.InstanceId = s.instance.InstanceId
			request.Force = requests.NewBoolean(true)
			return client.DeleteInstance(request)
		},
		EvalFunc:   client.EvalCouldRetryResponse(deleteInstanceRetryErrors, EvalRetryErrorType),
		RetryTimes: shortRetryTimes,
	})

	if err != nil {
		ui.Say(fmt.Sprintf("Failed to clean up instance %s: %s", s.instance.InstanceId, err))
	}
}

func (s *stepCreateAlicloudInstance) buildCreateInstanceRequest(state multistep.StateBag) (*ecs.RunInstancesRequest, error) {
	request := ecs.CreateRunInstancesRequest()
	request.ClientToken = uuid.TimeOrderedUUID()
	request.RegionId = s.RegionId
	request.InstanceType = s.InstanceType
	request.InstanceName = s.InstanceName
	request.RamRoleName = s.RamRoleName
	request.Tag = buildCreateInstanceTags(s.Tags)
	request.ZoneId = s.ZoneId
	request.SecurityEnhancementStrategy = s.SecurityEnhancementStrategy
	if s.AlicloudImageFamily != "" {
		request.ImageFamily = s.AlicloudImageFamily
	} else {
		sourceImage := state.Get("source_image").(*ecs.Image)
		request.ImageId = sourceImage.ImageId
	}
	securityGroupId := state.Get("securitygroupid").(string)
	request.SecurityGroupId = securityGroupId

	networkType := state.Get("networktype").(InstanceNetWork)
	if networkType == InstanceNetworkVpc {
		vswitchId := state.Get("vswitchid").(string)
		request.VSwitchId = vswitchId

		userData, err := s.getUserData(state)
		if err != nil {
			return nil, err
		}

		request.UserData = userData
	} else {
		if s.InternetChargeType == "" {
			s.InternetChargeType = "PayByTraffic"
		}

		if s.InternetMaxBandwidthOut == 0 {
			s.InternetMaxBandwidthOut = 5
		}
	}
	request.InternetChargeType = s.InternetChargeType
	request.InternetMaxBandwidthOut = requests.Integer(convertNumber(s.InternetMaxBandwidthOut))

	if s.IOOptimized.True() {
		request.IoOptimized = IOOptimizedOptimized
	} else if s.IOOptimized.False() {
		request.IoOptimized = IOOptimizedNone
	}

	config := state.Get("config").(*Config)
	password := config.Comm.SSHPassword
	if password == "" && config.Comm.WinRMPassword != "" {
		password = config.Comm.WinRMPassword
	}
	request.Password = password

	systemDisk := config.AlicloudImageConfig.ECSSystemDiskMapping
	request.SystemDiskDiskName = systemDisk.DiskName
	request.SystemDiskCategory = systemDisk.DiskCategory
	request.SystemDiskSize = strconv.Itoa(systemDisk.DiskSize)
	request.SystemDiskDescription = systemDisk.Description

	var runInstancesSystemDisk ecs.RunInstancesSystemDisk
	if systemDisk.Encrypted != confighelper.TriUnset {
		runInstancesSystemDisk.Encrypted = strconv.FormatBool(systemDisk.Encrypted.True())
	}
	request.SystemDisk = runInstancesSystemDisk

	imageDisks := config.AlicloudImageConfig.ECSImagesDiskMappings
	var dataDisks []ecs.RunInstancesDataDisk
	for _, imageDisk := range imageDisks {
		var dataDisk ecs.RunInstancesDataDisk
		dataDisk.DiskName = imageDisk.DiskName
		dataDisk.Category = imageDisk.DiskCategory
		dataDisk.Size = convertNumber(imageDisk.DiskSize)
		dataDisk.SnapshotId = imageDisk.SnapshotId
		dataDisk.Description = imageDisk.Description
		dataDisk.DeleteWithInstance = strconv.FormatBool(imageDisk.DeleteWithInstance)
		dataDisk.Device = imageDisk.Device
		if imageDisk.Encrypted != confighelper.TriUnset {
			dataDisk.Encrypted = strconv.FormatBool(imageDisk.Encrypted.True())
		}

		dataDisks = append(dataDisks, dataDisk)
	}
	request.DataDisk = &dataDisks

	return request, nil
}

func (s *stepCreateAlicloudInstance) getUserData(state multistep.StateBag) (string, error) {
	userData := s.UserData

	if s.UserDataFile != "" {
		data, err := ioutil.ReadFile(s.UserDataFile)
		if err != nil {
			return "", err
		}

		userData = string(data)
	}

	if userData != "" {
		userData = base64.StdEncoding.EncodeToString([]byte(userData))
	}

	return userData, nil

}

func buildCreateInstanceTags(tags map[string]string) *[]ecs.RunInstancesTag {
	var ecsTags []ecs.RunInstancesTag

	for k, v := range tags {
		ecsTags = append(ecsTags, ecs.RunInstancesTag{Key: k, Value: v})
	}

	return &ecsTags
}
