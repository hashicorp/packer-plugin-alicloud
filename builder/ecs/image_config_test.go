// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"testing"
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
