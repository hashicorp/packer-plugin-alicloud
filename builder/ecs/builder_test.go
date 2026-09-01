// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"reflect"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	helperconfig "github.com/hashicorp/packer-plugin-sdk/template/config"
)

func testBuilderConfig() map[string]interface{} {
	return map[string]interface{}{
		"access_key":          "foo",
		"secret_key":          "bar",
		"source_image":        "foo",
		"instance_type":       "ecs.n1.tiny",
		"region":              "cn-beijing",
		"custom_endpoint_ecs": "ecs.aliyuncs.com",
		"ssh_username":        "root",
		"image_name":          "foo",
		"io_optimized":        true,
	}
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packersdk.Builder); !ok {
		t.Fatalf("Builder should be a builder")
	}
}

func TestBuilder_Prepare_BadType(t *testing.T) {
	b := &Builder{}
	c := map[string]interface{}{
		"access_key": []string{},
	}

	_, warnings, err := b.Prepare(c)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatalf("prepare should fail")
	}
}

func TestBuilderPrepare_ECSImageName(t *testing.T) {
	var b Builder
	config := testBuilderConfig()

	// Test good
	config["image_name"] = "ecs.n1.tiny"
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test bad
	config["ecs_image_name"] = "foo {{"
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test bad
	delete(config, "image_name")
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_InvalidKey(t *testing.T) {
	var b Builder
	config := testBuilderConfig()

	// Add a random key
	config["i_should_not_be_valid"] = true
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_Devices(t *testing.T) {
	var b Builder
	config := testBuilderConfig()
	config["system_disk_mapping"] = map[string]interface{}{
		"disk_category":    "cloud",
		"disk_description": "system disk",
		"disk_name":        "system_disk",
		"disk_size":        60,
	}
	config["image_disk_mappings"] = []map[string]interface{}{
		{
			"disk_category":             "cloud_efficiency",
			"disk_name":                 "data_disk1",
			"disk_size":                 100,
			"disk_snapshot_id":          "s-1",
			"disk_description":          "data disk1",
			"disk_device":               "/dev/xvdb",
			"disk_delete_with_instance": false,
		},
		{
			"disk_name":   "data_disk2",
			"disk_device": "/dev/xvdc",
		},
	}
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	expected := AlicloudDiskDevice{
		DiskCategory: "cloud",
		Description:  "system disk",
		DiskName:     "system_disk",
		DiskSize:     60,
		Encrypted:    helperconfig.TriUnset,
	}
	if !reflect.DeepEqual(b.config.ECSSystemDiskMapping, expected) {
		t.Fatalf("system disk is not set properly, actual: %v; expected: %v", b.config.ECSSystemDiskMapping, expected)
	}
	if !reflect.DeepEqual(b.config.ECSImagesDiskMappings, []AlicloudDiskDevice{
		{
			DiskCategory:       "cloud_efficiency",
			DiskName:           "data_disk1",
			DiskSize:           100,
			SnapshotId:         "s-1",
			Description:        "data disk1",
			Device:             "/dev/xvdb",
			DeleteWithInstance: false,
		},
		{
			DiskName: "data_disk2",
			Device:   "/dev/xvdc",
		},
	}) {
		t.Fatalf("data disks are not set properly, actual: %#v", b.config.ECSImagesDiskMappings)
	}
}

func TestBuilderPrepare_DevicesEncrypted(t *testing.T) {
	var b Builder
	config := testBuilderConfig()
	config["system_disk_mapping"] = map[string]interface{}{
		"disk_encrypted":  true,
		"disk_kms_key_id": "shared-kms-key",
	}
	config["image_disk_mappings"] = []map[string]interface{}{
		{
			"disk_name":       "data_disk1",
			"disk_size":       100,
			"disk_encrypted":  true,
			"disk_kms_key_id": "shared-kms-key",
		},
	}
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	expectedSystem := AlicloudDiskDevice{
		Encrypted: helperconfig.TriTrue,
		KMSKeyId:  "shared-kms-key",
	}
	if !reflect.DeepEqual(b.config.ECSSystemDiskMapping, expectedSystem) {
		t.Fatalf("system disk encryption is not set properly, actual: %v; expected: %v", b.config.ECSSystemDiskMapping, expectedSystem)
	}

	expectedData := []AlicloudDiskDevice{
		{
			DiskName:  "data_disk1",
			DiskSize:  100,
			Encrypted: helperconfig.TriTrue,
			KMSKeyId:  "shared-kms-key",
		},
	}
	if !reflect.DeepEqual(b.config.ECSImagesDiskMappings, expectedData) {
		t.Fatalf("data disk encryption is not set properly, actual: %v; expected: %v", b.config.ECSImagesDiskMappings, expectedData)
	}
}

func TestBuilderPrepare_DevicesEncryptedDifferentKMSKeyIds(t *testing.T) {
	var b Builder
	config := testBuilderConfig()
	config["system_disk_mapping"] = map[string]interface{}{
		"disk_encrypted":  true,
		"disk_kms_key_id": "system-kms-key",
	}
	config["image_disk_mappings"] = []map[string]interface{}{
		{
			"disk_name":       "data_disk1",
			"disk_size":       100,
			"disk_encrypted":  true,
			"disk_kms_key_id": "data-kms-key",
		},
	}
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error when encrypted disks use different disk_kms_key_id values: %s", err)
	}

	expectedSystem := AlicloudDiskDevice{
		Encrypted: helperconfig.TriTrue,
		KMSKeyId:  "system-kms-key",
	}
	if !reflect.DeepEqual(b.config.ECSSystemDiskMapping, expectedSystem) {
		t.Fatalf("system disk encryption is not set properly, actual: %v; expected: %v", b.config.ECSSystemDiskMapping, expectedSystem)
	}

	expectedData := []AlicloudDiskDevice{
		{
			DiskName:  "data_disk1",
			DiskSize:  100,
			Encrypted: helperconfig.TriTrue,
			KMSKeyId:  "data-kms-key",
		},
	}
	if !reflect.DeepEqual(b.config.ECSImagesDiskMappings, expectedData) {
		t.Fatalf("data disk encryption is not set properly, actual: %v; expected: %v", b.config.ECSImagesDiskMappings, expectedData)
	}
}

func TestBuilderPrepare_IgnoreDataDisks(t *testing.T) {
	var b Builder
	config := testBuilderConfig()

	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.AlicloudImageIgnoreDataDisks != false {
		t.Fatalf("image_ignore_data_disks is not set properly, expect: %t, actual: %t", false, b.config.AlicloudImageIgnoreDataDisks)
	}

	config["image_ignore_data_disks"] = "false"
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.AlicloudImageIgnoreDataDisks != false {
		t.Fatalf("image_ignore_data_disks is not set properly, expect: %t, actual: %t", false, b.config.AlicloudImageIgnoreDataDisks)
	}

	config["image_ignore_data_disks"] = "true"
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.AlicloudImageIgnoreDataDisks != true {
		t.Fatalf("image_ignore_data_disks is not set properly, expect: %t, actual: %t", true, b.config.AlicloudImageIgnoreDataDisks)
	}
}

func TestBuilderPrepare_WaitSnapshotReadyTimeout(t *testing.T) {
	var b Builder
	config := testBuilderConfig()

	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.WaitSnapshotReadyTimeout != 0 {
		t.Fatalf("wait_snapshot_ready_timeout is not set properly, expect: %d, actual: %d", 0, b.config.WaitSnapshotReadyTimeout)
	}
	if b.getSnapshotReadyTimeout() != ALICLOUD_DEFAULT_LONG_TIMEOUT {
		t.Fatalf("default timeout is not set properly, expect: %d, actual: %d", ALICLOUD_DEFAULT_LONG_TIMEOUT, b.getSnapshotReadyTimeout())
	}

	config["wait_snapshot_ready_timeout"] = ALICLOUD_DEFAULT_TIMEOUT
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.WaitSnapshotReadyTimeout != ALICLOUD_DEFAULT_TIMEOUT {
		t.Fatalf("wait_snapshot_ready_timeout is not set properly, expect: %d, actual: %d", ALICLOUD_DEFAULT_TIMEOUT, b.config.WaitSnapshotReadyTimeout)
	}

	if b.getSnapshotReadyTimeout() != ALICLOUD_DEFAULT_TIMEOUT {
		t.Fatalf("default timeout is not set properly, expect: %d, actual: %d", ALICLOUD_DEFAULT_TIMEOUT, b.getSnapshotReadyTimeout())
	}
}
