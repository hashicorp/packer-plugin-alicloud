// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"os"
	"testing"
)

func testAlicloudAccessConfig() *AlicloudAccessConfig {
	return &AlicloudAccessConfig{
		AlicloudAccessKey: "ak",
		AlicloudSecretKey: "acs",
	}

}

func TestAlicloudAccessConfigPrepareRegion(t *testing.T) {
	c := testAlicloudAccessConfig()

	if v := os.Getenv("ALICLOUD_REGION"); v != "" {
		os.Unsetenv("ALICLOUD_REGION")
		defer os.Setenv("ALICLOUD_REGION", v)
	}
	c.AlicloudRegion = ""
	if err := c.Prepare(nil); err == nil {
		t.Fatalf("should have err")
	}

	c.AlicloudRegion = "cn-beijing"
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	os.Setenv("ALICLOUD_REGION", "cn-hangzhou")
	c.AlicloudRegion = ""
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	// store access key and reset it after tests.
	if v := os.Getenv("ALICLOUD_ACCESS_KEY"); v != "" {
		os.Unsetenv("ALICLOUD_ACCESS_KEY")
		defer os.Setenv("ALICLOUD_ACCESS_KEY", v)
	}
	// store profile and reset it after tests.
	if v := os.Getenv("ALICLOUD_PROFILE"); v != "" {
		os.Unsetenv("ALICLOUD_PROFILE")
		defer os.Setenv("ALICLOUD_PROFILE", v)
	}
	c.AlicloudAccessKey = ""
	if err := c.Prepare(nil); err == nil {
		t.Fatalf("should have err")
	}

	c.AlicloudProfile = "default"
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.AlicloudProfile = ""
	os.Setenv("ALICLOUD_PROFILE", "default")
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.Protocol = "http"
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.Protocol = "https"
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.Protocol = "htt"
	if err := c.Prepare(nil); err == nil {
		t.Fatalf("should have err")
	}

	c.AlicloudSkipValidation = false
}

func TestAlicloudAccessConfigClientRamRoleArn(t *testing.T) {
	newConfig := func() *AlicloudAccessConfig {
		c := testAlicloudAccessConfig()
		c.AlicloudRegion = "cn-beijing"
		return c
	}

	c := newConfig()
	c.AlicloudRamRoleArn = "acs:ram::1234567890123456:role/packer"
	if _, err := c.Client(); err == nil {
		t.Fatalf("should have err when ram_session_name is missing")
	}

	c = newConfig()
	c.AlicloudRamSessionName = "packer"
	if _, err := c.Client(); err == nil {
		t.Fatalf("should have err when ram_role_arn is missing")
	}

	c = newConfig()
	c.AlicloudRamRoleArn = "acs:ram::1234567890123456:role/packer"
	c.AlicloudRamSessionName = "packer"
	if _, err := c.Client(); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}
}
