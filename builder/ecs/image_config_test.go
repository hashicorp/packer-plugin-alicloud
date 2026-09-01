// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/template/config"
)

func testAlicloudImageConfig() *AlicloudImageConfig {
	return &AlicloudImageConfig{
		AlicloudImageName: "foo",
	}
}

func TestECSImageConfigPrepare_name(t *testing.T) {
	c := testAlicloudImageConfig()
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.AlicloudImageName = ""
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}
}

func TestAMIConfigPrepare_regions(t *testing.T) {
	c := testAlicloudImageConfig()
	c.AlicloudImageDestinationRegions = nil
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.AlicloudImageDestinationRegions = []string{"cn-beijing", "cn-hangzhou", "eu-central-1"}
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("bad: %s", err)
	}

	c.AlicloudImageDestinationRegions = nil
	if err := c.Prepare(nil); err != nil {
		t.Fatal("shouldn't have error")
	}
}

func TestECSImageConfigPrepare_imageTags(t *testing.T) {
	c := testAlicloudImageConfig()
	c.AlicloudImageTags = map[string]string{
		"TagKey1": "TagValue1",
		"TagKey2": "TagValue2",
	}
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}
	if len(c.AlicloudImageTags) != 2 || c.AlicloudImageTags["TagKey1"] != "TagValue1" ||
		c.AlicloudImageTags["TagKey2"] != "TagValue2" {
		t.Fatalf("invalid value, expected: %s, actual: %s", map[string]string{
			"TagKey1": "TagValue1",
			"TagKey2": "TagValue2",
		}, c.AlicloudImageTags)
	}
}

func TestECSImageConfigPrepare_targetImageFamily(t *testing.T) {
	c := testAlicloudImageConfig()

	// 1 character
	c.AlicloudTargetImageFamily = "a"
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}

	// 129 characters
	c.AlicloudTargetImageFamily = "abcdefghijklmnopqrs1abcdefghijklmnopqrs2abcdefghijklmnopqrs3abcdefghijklmnopqrs4abcdefghijklmnopqrs5abcdefghijklmnopqrs6123456789"
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}

	// invalid character
	c.AlicloudTargetImageFamily = "abc%&"
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}

	// begin with invalid character
	c.AlicloudTargetImageFamily = ":abc"
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}

	// start with acs:
	c.AlicloudTargetImageFamily = "acs:"
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}

	// start with aliyun
	c.AlicloudTargetImageFamily = "aliyun"
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}

	// start with http://
	c.AlicloudTargetImageFamily = "http://"
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}

	// start with https://
	c.AlicloudTargetImageFamily = "https://"
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}

	// success, begin with Chinese character， and contain :, -, _, .
	c.AlicloudTargetImageFamily = "啊:-_5s是u.ccess"
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	// success, begin with English character
	c.AlicloudTargetImageFamily = "a啊:-_5s是u.ccess"
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}
}

func TestECSImageConfigPrepare_bootMode(t *testing.T) {
	c := testAlicloudImageConfig()

	// invalid
	c.AlicloudBootMode = "boot"
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}

	// UEFI
	c.AlicloudBootMode = "UEFI"
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	// BIOS
	c.AlicloudBootMode = "BIOS"
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	// UEFI-Preferred
	c.AlicloudBootMode = "UEFI-Preferred"
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}
}

func TestECSImageConfigPrepare_imageCopyKMSKeyIds(t *testing.T) {
	c := testAlicloudImageConfig()
	c.ImageEncrypted = config.TrileanFromBool(true)
	c.AlicloudImageDestinationRegions = []string{"cn-beijing", "cn-hangzhou"}
	c.ImageCopyKMSKeyIds = []string{"", "key-2"}
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err with empty placeholder and matching length: %s", err)
	}

	c = testAlicloudImageConfig()
	c.ImageEncrypted = config.TrileanFromBool(true)
	c.AlicloudImageDestinationRegions = []string{"cn-beijing", "cn-hangzhou", "cn-shanghai"}
	c.ImageCopyKMSKeyIds = []string{"key-1", "key-2", "key-3"}
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err with matching length: %s", err)
	}

	c = testAlicloudImageConfig()
	c.ImageEncrypted = config.TrileanFromBool(true)
	c.AlicloudImageDestinationRegions = []string{"cn-beijing", "cn-hangzhou"}
	c.ImageCopyKMSKeyIds = []string{"key-1"}
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err when image_copy_kms_ids is shorter than image_copy_regions: %s", err)
	}

	c = testAlicloudImageConfig()
	c.AlicloudImageDestinationRegions = []string{"cn-beijing"}
	c.ImageCopyKMSKeyIds = []string{"key-1"}
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have err when image_copy_kms_ids is specified but image_encrypted is not true")
	}

	c = testAlicloudImageConfig()
	c.ImageEncrypted = config.TrileanFromBool(false)
	c.AlicloudImageDestinationRegions = []string{"cn-beijing"}
	c.ImageCopyKMSKeyIds = []string{"key-1"}
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have err when image_copy_kms_ids is specified but image_encrypted is false")
	}
}

func TestECSImageConfigPrepare_kmsKeyIdRequiresEncrypted(t *testing.T) {
	c := testAlicloudImageConfig()
	c.ImageEncrypted = config.TrileanFromBool(true)
	c.KMSKeyId = "same-region-kms-key"
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err when kms_key_id is specified with image_encrypted=true: %s", err)
	}

	c = testAlicloudImageConfig()
	c.KMSKeyId = "same-region-kms-key"
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have err when kms_key_id is specified but image_encrypted is not set")
	}

	c = testAlicloudImageConfig()
	c.ImageEncrypted = config.TrileanFromBool(false)
	c.KMSKeyId = "same-region-kms-key"
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have err when kms_key_id is specified but image_encrypted is false")
	}
}

func TestECSImageConfigPrepare_diskKMSKeyIdRequiresEncrypted(t *testing.T) {
	c := testAlicloudImageConfig()
	c.ECSSystemDiskMapping = AlicloudDiskDevice{
		Encrypted: config.TrileanFromBool(true),
		KMSKeyId:  "system-kms-key",
	}
	c.ECSImagesDiskMappings = []AlicloudDiskDevice{
		{Encrypted: config.TrileanFromBool(true), KMSKeyId: "data-kms-key-1"},
		{Encrypted: config.TrileanFromBool(true), KMSKeyId: "data-kms-key-2"},
	}
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("should allow different disk_kms_key_id values when encrypted: %s", err)
	}

	c = testAlicloudImageConfig()
	c.ECSSystemDiskMapping = AlicloudDiskDevice{
		Encrypted: config.TrileanFromBool(true),
	}
	c.ECSImagesDiskMappings = []AlicloudDiskDevice{
		{Encrypted: config.TrileanFromBool(true)},
	}
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err when encrypted disks omit KMS key ID: %s", err)
	}

	c = testAlicloudImageConfig()
	c.ECSSystemDiskMapping = AlicloudDiskDevice{
		KMSKeyId: "system-kms-key",
	}
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error when disk_kms_key_id is set but disk_encrypted is not true")
	}

	c = testAlicloudImageConfig()
	c.ECSImagesDiskMappings = []AlicloudDiskDevice{
		{Encrypted: config.TrileanFromBool(false), KMSKeyId: "data-kms-key"},
	}
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error when disk_kms_key_id is set but disk_encrypted is false")
	}
}
