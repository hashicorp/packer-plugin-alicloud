// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/template/config"
)

func TestStepRegionCopyAlicloudImage_buildCopyImageRequest_sameRegionKMSKeyId(t *testing.T) {
	step := &stepRegionCopyAlicloudImage{
		RegionId:                      "cn-hangzhou",
		AlicloudImageDestinationNames: []string{"test-image"},
	}
	cfg := &Config{}
	cfg.AlicloudImageConfig.AlicloudImageName = "test-image"
	cfg.AlicloudImageConfig.ImageEncrypted = config.TrileanFromBool(true)
	cfg.AlicloudImageConfig.KMSKeyId = "same-region-kms-key"

	req := step.buildCopyImageRequest(0, "cn-hangzhou", cfg, "m-source", 1, -1)

	if req.DestinationRegionId != "cn-hangzhou" {
		t.Fatalf("expected destination region cn-hangzhou, got %s", req.DestinationRegionId)
	}
	if req.DestinationImageName != "test-image" {
		t.Fatalf("expected destination image name test-image, got %s", req.DestinationImageName)
	}
	if !req.Encrypted.HasValue() {
		t.Fatal("expected encrypted to be set")
	}
	encrypted, err := req.Encrypted.GetValue()
	if err != nil || !encrypted {
		t.Fatalf("expected encrypted to be true, got %v, err: %v", req.Encrypted, err)
	}
	if req.KMSKeyId != "same-region-kms-key" {
		t.Fatalf("expected KMS key id same-region-kms-key, got %s", req.KMSKeyId)
	}
}

func TestStepRegionCopyAlicloudImage_buildCopyImageRequest_sameRegionDefaultKMS(t *testing.T) {
	step := &stepRegionCopyAlicloudImage{RegionId: "cn-hangzhou"}
	cfg := &Config{}
	cfg.AlicloudImageConfig.AlicloudImageName = "test-image"
	cfg.AlicloudImageConfig.ImageEncrypted = config.TrileanFromBool(true)

	req := step.buildCopyImageRequest(0, "cn-hangzhou", cfg, "m-source", 0, -1)

	if !req.Encrypted.HasValue() {
		t.Fatal("expected encrypted to be set")
	}
	encrypted, err := req.Encrypted.GetValue()
	if err != nil || !encrypted {
		t.Fatalf("expected encrypted to be true, got %v, err: %v", req.Encrypted, err)
	}
	if req.KMSKeyId != "" {
		t.Fatalf("expected default KMS key id, got %s", req.KMSKeyId)
	}
}

func TestStepRegionCopyAlicloudImage_buildCopyImageRequest_crossRegionKMSKeyIds(t *testing.T) {
	step := &stepRegionCopyAlicloudImage{
		RegionId:                      "cn-hangzhou",
		AlicloudImageDestinationNames: []string{"test-image-sh", "test-image-bj"},
	}
	cfg := &Config{}
	cfg.AlicloudImageConfig.AlicloudImageName = "test-image"
	cfg.AlicloudImageConfig.ImageEncrypted = config.TrileanFromBool(true)
	cfg.AlicloudImageConfig.ImageCopyKMSKeyIds = []string{"key-sh", "key-bj"}

	req0 := step.buildCopyImageRequest(0, "cn-shanghai", cfg, "m-source", 2, 0)
	if req0.DestinationImageName != "test-image-sh" {
		t.Fatalf("expected destination image name test-image-sh, got %s", req0.DestinationImageName)
	}
	if req0.KMSKeyId != "key-sh" {
		t.Fatalf("expected KMS key id key-sh, got %s", req0.KMSKeyId)
	}

	req1 := step.buildCopyImageRequest(1, "cn-beijing", cfg, "m-source", 2, 1)
	if req1.DestinationImageName != "test-image-bj" {
		t.Fatalf("expected destination image name test-image-bj, got %s", req1.DestinationImageName)
	}
	if req1.KMSKeyId != "key-bj" {
		t.Fatalf("expected KMS key id key-bj, got %s", req1.KMSKeyId)
	}
}

func TestStepRegionCopyAlicloudImage_buildCopyImageRequest_crossRegionKMSKeyIdsShorterThanRegions(t *testing.T) {
	step := &stepRegionCopyAlicloudImage{RegionId: "cn-hangzhou"}
	cfg := &Config{}
	cfg.AlicloudImageConfig.AlicloudImageName = "test-image"
	cfg.AlicloudImageConfig.ImageEncrypted = config.TrileanFromBool(true)
	cfg.AlicloudImageConfig.AlicloudImageDestinationRegions = []string{"cn-shanghai", "cn-beijing"}
	cfg.AlicloudImageConfig.ImageCopyKMSKeyIds = []string{"key-sh"}

	req0 := step.buildCopyImageRequest(0, "cn-shanghai", cfg, "m-source", 0, 0)
	if req0.KMSKeyId != "key-sh" {
		t.Fatalf("expected KMS key id key-sh, got %s", req0.KMSKeyId)
	}

	req1 := step.buildCopyImageRequest(1, "cn-beijing", cfg, "m-source", 0, 1)
	if req1.KMSKeyId != "" {
		t.Fatalf("expected default KMS key id for missing entry, got %s", req1.KMSKeyId)
	}
}

func TestStepRegionCopyAlicloudImage_buildCopyImageRequest_crossRegionEmptyKMSKeyIdEntry(t *testing.T) {
	step := &stepRegionCopyAlicloudImage{RegionId: "cn-hangzhou"}
	cfg := &Config{}
	cfg.AlicloudImageConfig.AlicloudImageName = "test-image"
	cfg.AlicloudImageConfig.ImageEncrypted = config.TrileanFromBool(true)
	cfg.AlicloudImageConfig.AlicloudImageDestinationRegions = []string{"cn-shanghai", "cn-beijing"}
	cfg.AlicloudImageConfig.ImageCopyKMSKeyIds = []string{"", "key-bj"}

	req0 := step.buildCopyImageRequest(0, "cn-shanghai", cfg, "m-source", 0, 0)
	if req0.KMSKeyId != "" {
		t.Fatalf("expected default KMS key id for empty entry, got %s", req0.KMSKeyId)
	}

	req1 := step.buildCopyImageRequest(1, "cn-beijing", cfg, "m-source", 0, 1)
	if req1.KMSKeyId != "key-bj" {
		t.Fatalf("expected KMS key id key-bj, got %s", req1.KMSKeyId)
	}
}

func TestStepRegionCopyAlicloudImage_buildCopyImageRequest_noEncryption(t *testing.T) {
	step := &stepRegionCopyAlicloudImage{RegionId: "cn-hangzhou"}
	cfg := &Config{}
	cfg.AlicloudImageConfig.AlicloudImageName = "test-image"

	req := step.buildCopyImageRequest(0, "cn-hangzhou", cfg, "m-source", 0, -1)

	if req.Encrypted.HasValue() {
		t.Fatalf("expected encrypted not to be set, got %v", req.Encrypted)
	}
	if req.KMSKeyId != "" {
		t.Fatalf("expected empty KMS key id, got %s", req.KMSKeyId)
	}
}
