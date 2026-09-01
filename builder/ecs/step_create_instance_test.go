// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"bytes"
	"testing"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
)

func testStepCreateAlicloudInstanceState(networkType InstanceNetWork) multistep.StateBag {
	state := new(multistep.BasicStateBag)
	state.Put("ui", &packersdk.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	state.Put("client", &ClientWrapper{})
	state.Put("source_image", &ecs.Image{ImageId: "m-test-image"})
	state.Put("securitygroupid", "sg-test")
	state.Put("networktype", networkType)
	state.Put("vswitchid", "vsw-test")
	return state
}

func TestStepCreateAlicloudInstance_buildRunInstancesRequest_Classic(t *testing.T) {
	step := &stepCreateAlicloudInstance{
		RegionId:     "cn-beijing",
		InstanceType: "ecs.n1.tiny",
		IOOptimized:  config.TrileanFromBool(true),
	}
	state := testStepCreateAlicloudInstanceState(InstanceNetworkClassic)
	cfg := &Config{}
	cfg.AlicloudImageConfig.ECSSystemDiskMapping = AlicloudDiskDevice{
		DiskName:  "system_disk",
		DiskSize:  60,
		Encrypted: config.TrileanFromBool(true),
		KMSKeyId:  "shared-kms-key",
	}
	cfg.AlicloudImageConfig.ECSImagesDiskMappings = []AlicloudDiskDevice{
		{
			DiskName:  "data_disk1",
			DiskSize:  100,
			Encrypted: config.TrileanFromBool(true),
			KMSKeyId:  "shared-kms-key",
		},
	}
	cfg.Comm.SSHKeyPairName = "test-keypair"
	state.Put("config", cfg)

	req, err := step.buildRunInstancesRequest(state)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if req.ImageId != "m-test-image" {
		t.Fatalf("expected image id m-test-image, got %s", req.ImageId)
	}
	if req.SecurityGroupId != "sg-test" {
		t.Fatalf("expected security group sg-test, got %s", req.SecurityGroupId)
	}
	if req.KeyPairName != "test-keypair" {
		t.Fatalf("expected key pair test-keypair, got %s", req.KeyPairName)
	}
	if req.InternetChargeType != "PayByTraffic" {
		t.Fatalf("expected internet charge type PayByTraffic, got %s", req.InternetChargeType)
	}
	if req.InternetMaxBandwidthOut != "5" {
		t.Fatalf("expected internet max bandwidth out 5, got %s", req.InternetMaxBandwidthOut)
	}
	if req.SystemDiskDiskName != "system_disk" || req.SystemDiskSize != "60" {
		t.Fatalf("system disk fields not set properly, got name=%s size=%s", req.SystemDiskDiskName, req.SystemDiskSize)
	}
	if req.SystemDisk.Encrypted != "true" {
		t.Fatalf("expected system disk encrypted true, got %s", req.SystemDisk.Encrypted)
	}
	if req.SystemDisk.KMSKeyId != "shared-kms-key" {
		t.Fatalf("expected system disk kms key shared-kms-key, got %s", req.SystemDisk.KMSKeyId)
	}
	if req.DataDisk == nil || len(*req.DataDisk) != 1 {
		t.Fatalf("expected 1 data disk, got %v", req.DataDisk)
	}
	dataDisk := (*req.DataDisk)[0]
	if dataDisk.Encrypted != "true" || dataDisk.KMSKeyId != "shared-kms-key" {
		t.Fatalf("expected data disk encrypted=true kms=shared-kms-key, got encrypted=%s kms=%s", dataDisk.Encrypted, dataDisk.KMSKeyId)
	}
}

func TestStepCreateAlicloudInstance_buildRunInstancesRequest_VpcNoEncryption(t *testing.T) {
	step := &stepCreateAlicloudInstance{
		RegionId:     "cn-hangzhou",
		InstanceType: "ecs.g7.large",
		UserData:     "#!/bin/bash\necho hello",
		IOOptimized:  config.TrileanFromBool(true),
	}
	state := testStepCreateAlicloudInstanceState(InstanceNetworkVpc)
	cfg := &Config{}
	cfg.AlicloudImageConfig.ECSSystemDiskMapping = AlicloudDiskDevice{
		DiskName: "system_disk",
		DiskSize: 40,
	}
	cfg.AlicloudImageConfig.ECSImagesDiskMappings = []AlicloudDiskDevice{
		{DiskName: "data_disk1", DiskSize: 50},
	}
	state.Put("config", cfg)

	req, err := step.buildRunInstancesRequest(state)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if req.VSwitchId != "vsw-test" {
		t.Fatalf("expected vswitch vsw-test, got %s", req.VSwitchId)
	}
	if req.UserData == "" {
		t.Fatal("expected user data to be set in VPC mode")
	}
	if req.SystemDisk.Encrypted != "" {
		t.Fatalf("expected no system disk encryption, got %s", req.SystemDisk.Encrypted)
	}
	if req.SystemDisk.KMSKeyId != "" {
		t.Fatalf("expected no system disk kms key, got %s", req.SystemDisk.KMSKeyId)
	}
	dataDisk := (*req.DataDisk)[0]
	if dataDisk.Encrypted != "" || dataDisk.KMSKeyId != "" {
		t.Fatalf("expected no data disk encryption, got encrypted=%s kms=%s", dataDisk.Encrypted, dataDisk.KMSKeyId)
	}
}
