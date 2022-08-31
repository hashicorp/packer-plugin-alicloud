package ecs

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/hashicorp/packer-plugin-sdk/acctest/testutils"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
)

const defaultTestRegion = "cn-beijing"

type TestCheckFunc func(*exec.Cmd, string) error

func TestBuilderAcc_validateRegion(t *testing.T) {
	t.Parallel()

	if skip := testAccPreCheck(t); skip == true {
		return
	}

	access := &AlicloudAccessConfig{AlicloudRegion: "cn-beijing"}
	err := access.Config()
	if err != nil {
		t.Fatalf("init AlicloudAccessConfig failed: %s", err)
	}

	err = access.ValidateRegion("cn-hangzhou")
	if err != nil {
		t.Fatalf("Expect pass with valid region id but failed: %s", err)
	}

	err = access.ValidateRegion("invalidRegionId")
	if err == nil {
		t.Fatal("Expect failure due to invalid region id but passed")
	}
}

func TestBuilderAcc_basic(t *testing.T) {
	t.Parallel()
	testAccPreCheck(t)
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-alicloud-builder-basic",
		Type:     "alicloud-ecs",
		Template: testBuilderAccBasic,
	})
}

const testBuilderAccBasic = `
{	"builders": [{
		"type": "alicloud-ecs",
		"region": "cn-beijing",
		"instance_type": "ecs.n1.tiny",
		"source_image":"ubuntu_18_04_64_20G_alibase_20190509.vhd",
		"io_optimized":"true",
		"ssh_username":"root",
		"image_name": "packer-test-basic_{{timestamp}}"
	}]
}`

func TestBuilderAcc_withDiskSettings(t *testing.T) {
	t.Parallel()
	outputfilename := fmt.Sprintf("test-manifest-%s.json", uuid.TimeOrderedUUID())
	if skip := testAccPreCheck(t); skip == true {
		return
	}
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-alicloud-builder-with-disk-settings",
		Type:     "alicloud-ecs",
		Template: fmt.Sprintf(testBuilderAccWithDiskSettings, outputfilename),
		Check:    checkImageDisksSettings(outputfilename),
		Teardown: func() error {
			return os.Remove(outputfilename)
		},
	})
}

const testBuilderAccWithDiskSettings = `
{	"builders": [{
		"type": "alicloud-ecs",
		"region": "cn-beijing",
		"instance_type": "ecs.n1.tiny",
		"source_image":"ubuntu_18_04_64_20G_alibase_20190509.vhd",
		"io_optimized":"true",
		"ssh_username":"root",
		"image_name": "packer-test-withDiskSettings_{{timestamp}}",
		"system_disk_mapping": {
			"disk_size": 60
		},
		"image_disk_mappings": [
			{
				"disk_name": "datadisk1",
				"disk_size": 25,
				"disk_delete_with_instance": true
			},
			{
				"disk_name": "datadisk2",
				"disk_size": 25,
				"disk_delete_with_instance": true
			}
		]
	}],
	"post-processors": [{
		"type": "manifest",
		"output": "%s"
	}]
}`

func checkImageDisksSettings(manifestfilepath string) TestCheckFunc {
	return func(buildCommand *exec.Cmd, logfile string) error {
		regionArtifacts, err := loadArtifacts(manifestfilepath)
		if err != nil {
			return fmt.Errorf("could not load manifest: %s", err)
		}

		if len(regionArtifacts) > 1 {
			return fmt.Errorf("more than 1 artifact")
		}

		imageId := regionArtifacts[defaultTestRegion]

		// describe the image, get block devices with a snapshot
		client, _ := testAliyunClient()

		describeImagesRequest := ecs.CreateDescribeImagesRequest()
		describeImagesRequest.RegionId = defaultTestRegion
		describeImagesRequest.ImageId = imageId
		imagesResponse, err := client.DescribeImages(describeImagesRequest)
		if err != nil {
			return fmt.Errorf("describe images failed due to %s", err)
		}

		if len(imagesResponse.Images.Image) == 0 {
			return fmt.Errorf("image %s generated can not be found", imageId)
		}

		image := imagesResponse.Images.Image[0]
		if image.Size != 60 {
			return fmt.Errorf("the size of image %s should be equal to 60G but got %dG", imageId, image.Size)
		}
		if len(image.DiskDeviceMappings.DiskDeviceMapping) != 3 {
			return fmt.Errorf("image %s should contains 3 disks", imageId)
		}

		var snapshotIds []string
		for _, mapping := range image.DiskDeviceMappings.DiskDeviceMapping {
			if mapping.Type == DiskTypeSystem {
				if mapping.Size != "60" {
					return fmt.Errorf("the system snapshot size of image %s should be equal to 60G but got %sG", imageId, mapping.Size)
				}
			} else {
				if mapping.Size != "25" {
					return fmt.Errorf("the data disk size of image %s should be equal to 25G but got %sG", imageId, mapping.Size)
				}

				snapshotIds = append(snapshotIds, mapping.SnapshotId)
			}
		}

		data, _ := json.Marshal(snapshotIds)

		describeSnapshotRequest := ecs.CreateDescribeSnapshotsRequest()
		describeSnapshotRequest.RegionId = defaultTestRegion
		describeSnapshotRequest.SnapshotIds = string(data)
		describeSnapshotsResponse, err := client.DescribeSnapshots(describeSnapshotRequest)
		if err != nil {
			return fmt.Errorf("describe data snapshots failed due to %s", err)
		}
		if len(describeSnapshotsResponse.Snapshots.Snapshot) != 2 {
			return fmt.Errorf("expect %d data snapshots but got %d", len(snapshotIds), len(describeSnapshotsResponse.Snapshots.Snapshot))
		}

		var dataDiskIds []string
		for _, snapshot := range describeSnapshotsResponse.Snapshots.Snapshot {
			dataDiskIds = append(dataDiskIds, snapshot.SourceDiskId)
		}
		data, _ = json.Marshal(dataDiskIds)

		describeDisksRequest := ecs.CreateDescribeDisksRequest()
		describeDisksRequest.RegionId = defaultTestRegion
		describeDisksRequest.DiskIds = string(data)
		describeDisksResponse, err := client.DescribeDisks(describeDisksRequest)
		if err != nil {
			return fmt.Errorf("describe snapshots failed due to %s", err)
		}
		if len(describeDisksResponse.Disks.Disk) != 0 {
			return fmt.Errorf("data disks should be deleted but %d left", len(describeDisksResponse.Disks.Disk))
		}

		return nil
	}
}

func TestBuilderAcc_withIgnoreDataDisks(t *testing.T) {
	t.Parallel()
	if skip := testAccPreCheck(t); skip == true {
		return
	}
	outputfilename := fmt.Sprintf("test-manifest-%s.json", uuid.TimeOrderedUUID())
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-alicloud-builder-with-ignore-data-disks",
		Type:     "alicloud-ecs",
		Template: fmt.Sprintf(testBuilderAccIgnoreDataDisks, outputfilename),
		Check:    checkIgnoreDataDisks(outputfilename),
		Teardown: func() error {
			return os.Remove(outputfilename)
		},
	})
}

const testBuilderAccIgnoreDataDisks = `
{	"builders": [{
		"type": "alicloud-ecs",
		"region": "cn-beijing",
		"instance_type": "ecs.gn5-c8g1.2xlarge",
		"source_image":"ubuntu_18_04_64_20G_alibase_20190509.vhd",
		"io_optimized":"true",
		"ssh_username":"root",
		"image_name": "packer-test-ignoreDataDisks_{{timestamp}}",
		"image_ignore_data_disks": true
	}],
	"post-processors": [{
		"type": "manifest",
		"output": "%s"
	}]
}`

func checkIgnoreDataDisks(manifestfilepath string) TestCheckFunc {
	return func(buildCommand *exec.Cmd, logfile string) error {
		regionArtifacts, err := loadArtifacts(manifestfilepath)
		if err != nil {
			return fmt.Errorf("could not load manifest: %s", err)
		}

		if len(regionArtifacts) > 1 {
			return fmt.Errorf("more than 1 artifact")
		}

		imageId := regionArtifacts[defaultTestRegion]

		// describe the image, get block devices with a snapshot
		client, _ := testAliyunClient()

		describeImagesRequest := ecs.CreateDescribeImagesRequest()
		describeImagesRequest.RegionId = defaultTestRegion
		describeImagesRequest.ImageId = imageId
		imagesResponse, err := client.DescribeImages(describeImagesRequest)
		if err != nil {
			return fmt.Errorf("describe images failed due to %s", err)
		}

		if len(imagesResponse.Images.Image) == 0 {
			return fmt.Errorf("image %s generated can not be found", imageId)
		}

		image := imagesResponse.Images.Image[0]
		if len(image.DiskDeviceMappings.DiskDeviceMapping) != 1 {
			return fmt.Errorf("image %s should only contain one disks", imageId)
		}

		return nil
	}
}

func TestBuilderAcc_windows(t *testing.T) {
	t.Parallel()
	if skip := testAccPreCheck(t); skip == true {
		return
	}

	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-alicloud-builder-windows",
		Type:     "alicloud-ecs",
		Template: testBuilderAccWindows,
	})
}

const testBuilderAccWindows = `
{	"builders": [{
		"type": "alicloud-ecs",
		"region": "cn-beijing",
		"instance_type": "ecs.n1.tiny",
		"source_image":"winsvr_64_dtcC_1809_en-us_40G_alibase_20190318.vhd",
		"io_optimized":"true",
		"communicator": "winrm",
		"winrm_port": 5985,
		"winrm_username": "Administrator",
		"winrm_password": "Test1234",
		"image_name": "packer-test-windows_{{timestamp}}",
		"user_data_file": "../../../examples/alicloud/basic/winrm_enable_userdata.ps1"
	}]
}`

func TestBuilderAcc_regionCopy(t *testing.T) {
	t.Parallel()
	if skip := testAccPreCheck(t); skip == true {
		return
	}
	outputfilename := fmt.Sprintf("test-manifest-%s.json", uuid.TimeOrderedUUID())
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-alicloud-builder-region-copy",
		Type:     "alicloud",
		Template: fmt.Sprintf(testBuilderAccRegionCopy, outputfilename),
		Check:    checkRegionCopy([]string{"cn-hangzhou", "cn-shenzhen"}, outputfilename),
		Teardown: func() error {
			return os.Remove(outputfilename)
		},
	})
}

const testBuilderAccRegionCopy = `
{
	"builders": [{
		"type": "alicloud-ecs",
		"region": "cn-beijing",
		"instance_type": "ecs.n1.tiny",
		"source_image":"ubuntu_18_04_64_20G_alibase_20190509.vhd",
		"io_optimized":"true",
		"ssh_username":"root",
		"image_name": "packer-test-regionCopy_{{timestamp}}",
		"image_copy_regions": ["cn-hangzhou", "cn-shenzhen"],
		"image_copy_names": ["packer-copy-test-hz_{{timestamp}}", "packer-copy-test-sz_{{timestamp}}"]
	}],
	"post-processors": [{
		"type": "manifest",
		"output": "%s"
	}]
}
`

func checkRegionCopy(regions []string, manifestfilepath string) TestCheckFunc {
	return func(buildCommand *exec.Cmd, logfile string) error {
		regionArtifacts, err := loadArtifacts(manifestfilepath)
		if err != nil {
			return fmt.Errorf("could not load manifest: %s", err)
		}

		// Verify that we copied to only the regions given
		regionSet := make(map[string]struct{})
		for _, r := range regions {
			regionSet[r] = struct{}{}
		}

		for r := range regionArtifacts {
			if r == "cn-beijing" {
				delete(regionSet, r)
				continue
			}

			if _, ok := regionSet[r]; !ok {
				return fmt.Errorf("region %s is not the target region but found in artifacts", r)
			}

			delete(regionSet, r)
		}

		if len(regionSet) > 0 {
			return fmt.Errorf("following region(s) should be the copying targets but corresponding artifact(s) not found: %#v", regionSet)
		}

		client, _ := testAliyunClient()
		for regionId, imageId := range regionArtifacts {
			describeImagesRequest := ecs.CreateDescribeImagesRequest()
			describeImagesRequest.RegionId = regionId
			describeImagesRequest.ImageId = imageId
			describeImagesRequest.Status = ImageStatusQueried
			describeImagesResponse, err := client.DescribeImages(describeImagesRequest)
			if err != nil {
				return fmt.Errorf("describe generated image %s failed due to %s", imageId, err)
			}
			if len(describeImagesResponse.Images.Image) == 0 {
				return fmt.Errorf("image %s in artifacts can not be found", imageId)
			}

			image := describeImagesResponse.Images.Image[0]
			if image.IsCopied && regionId == "cn-hangzhou" && !strings.HasPrefix(image.ImageName, "packer-copy-test-hz") {
				return fmt.Errorf("the name of image %s in artifacts should begin with %s but got %s", imageId, "packer-copy-test-hz", image.ImageName)
			}
			if image.IsCopied && regionId == "cn-shenzhen" && !strings.HasPrefix(image.ImageName, "packer-copy-test-sz") {
				return fmt.Errorf("the name of image %s in artifacts should begin with %s but got %s", imageId, "packer-copy-test-sz", image.ImageName)
			}
		}

		return nil
	}
}

func TestBuilderAcc_forceDelete(t *testing.T) {
	t.Parallel()
	if skip := testAccPreCheck(t); skip == true {
		return
	}
	// Build the same alicloud image twice, with ecs_image_force_delete on the second run
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-alicloud-builder-force-delete",
		Type:     "alicloud-ecs",
		Template: buildForceDeregisterConfig("false", "delete"),
	})

	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-alicloud-builder-force-delete-2",
		Type:     "alicloud-ecs",
		Template: buildForceDeregisterConfig("true", "delete"),
	})
}

func buildForceDeregisterConfig(val, name string) string {
	return fmt.Sprintf(testBuilderAccForceDelete, val, name)
}

const testBuilderAccForceDelete = `
{
	"builders": [{
		"type": "alicloud-ecs",
		"region": "cn-beijing",
		"instance_type": "ecs.n1.tiny",
		"source_image":"ubuntu_18_04_64_20G_alibase_20190509.vhd",
		"io_optimized":"true",
		"ssh_username":"root",
		"image_force_delete": "%s",
		"image_name": "packer-test-forceDelete_%s"
	}]
}
`

func TestBuilderAcc_ECSImageSharing(t *testing.T) {
	t.Parallel()
	if skip := testAccPreCheck(t); skip == true {
		return
	}
	outputfilename := fmt.Sprintf("test-manifest-%s.json", uuid.TimeOrderedUUID())
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-alicloud-builder-image-sharing",
		Type:     "alicloud-ecs",
		Template: fmt.Sprintf(testBuilderAccSharing, outputfilename),
		Check:    checkECSImageSharing("1309208528360047", outputfilename),
		Teardown: func() error {
			return os.Remove(outputfilename)
		},
	})
}

const testBuilderAccSharing = `
{
	"builders": [{
		"type": "alicloud-ecs",
		"region": "cn-beijing",
		"instance_type": "ecs.n1.tiny",
		"source_image":"ubuntu_18_04_64_20G_alibase_20190509.vhd",
		"io_optimized":"true",
		"ssh_username":"root",
		"image_name": "packer-test-ECSImageSharing_{{timestamp}}",
		"image_share_account":["1309208528360047"]
	}],
	"post-processors": [{
		"type": "manifest",
		"output": "%s"
	}]
}
`

func checkECSImageSharing(uid string, manifestfilepath string) TestCheckFunc {
	return func(buildCommand *exec.Cmd, logfile string) error {
		regionArtifacts, err := loadArtifacts(manifestfilepath)
		if err != nil {
			return fmt.Errorf("could not load manifest: %s", err)
		}

		// describe the image, get block devices with a snapshot
		client, _ := testAliyunClient()

		describeImageShareRequest := ecs.CreateDescribeImageSharePermissionRequest()
		describeImageShareRequest.RegionId = "cn-beijing"
		describeImageShareRequest.ImageId = regionArtifacts["cn-beijing"]
		imageShareResponse, err := client.DescribeImageSharePermission(describeImageShareRequest)

		if err != nil {
			return fmt.Errorf("Error retrieving Image Attributes for ECS Image Artifact (%#v) "+
				"in ECS Image Sharing Test: %s", regionArtifacts["cn-beijing"], err)
		}

		if len(imageShareResponse.Accounts.Account) != 1 && imageShareResponse.Accounts.Account[0].AliyunId != uid {
			return fmt.Errorf("share account is incorrect %d", len(imageShareResponse.Accounts.Account))
		}

		return nil
	}
}

func TestBuilderAcc_forceDeleteSnapshot(t *testing.T) {
	t.Parallel()
	if skip := testAccPreCheck(t); skip == true {
		return
	}
	destImageName := "delete"

	// Build the same alicloud image name twice, with force_delete_snapshot on the second run
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-alicloud-builder-force-delete-snapshot",
		Type:     "alicloud-ecs",
		Template: buildForceDeleteSnapshotConfig("false", destImageName),
	})

	// Get image data by image image name
	client, _ := testAliyunClient()

	describeImagesRequest := ecs.CreateDescribeImagesRequest()
	describeImagesRequest.RegionId = "cn-beijing"
	describeImagesRequest.ImageName = "packer-test-" + destImageName
	images, _ := client.DescribeImages(describeImagesRequest)

	image := images.Images.Image[0]

	// Get snapshot ids for image
	snapshotIds := []string{}
	for _, device := range image.DiskDeviceMappings.DiskDeviceMapping {
		if device.Device != "" && device.SnapshotId != "" {
			snapshotIds = append(snapshotIds, device.SnapshotId)
		}
	}

	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-alicloud-builder-force-delete-snapshot-true",
		Type:     "alicloud-ecs",
		Template: buildForceDeleteSnapshotConfig("true", destImageName),
		Check:    checkSnapshotsDeleted(snapshotIds),
	})
}

func buildForceDeleteSnapshotConfig(val, name string) string {
	return fmt.Sprintf(testBuilderAccForceDeleteSnapshot, val, val, name)
}

const testBuilderAccForceDeleteSnapshot = `
{
	"builders": [{
		"type": "alicloud-ecs",
		"region": "cn-beijing",
		"instance_type": "ecs.n1.tiny",
		"source_image":"ubuntu_18_04_64_20G_alibase_20190509.vhd",
		"io_optimized":"true",
		"ssh_username":"root",
		"image_force_delete_snapshots": "%s",
		"image_force_delete": "%s",
		"image_name": "packer-test-%s"
	}]
}
`

func checkSnapshotsDeleted(snapshotIds []string) TestCheckFunc {
	return func(buildCommand *exec.Cmd, logfile string) error {
		// Verify the snapshots are gone
		client, _ := testAliyunClient()
		data, err := json.Marshal(snapshotIds)
		if err != nil {
			return fmt.Errorf("Marshal snapshotIds array failed %v", err)
		}

		describeSnapshotsRequest := ecs.CreateDescribeSnapshotsRequest()
		describeSnapshotsRequest.RegionId = "cn-beijing"
		describeSnapshotsRequest.SnapshotIds = string(data)
		snapshotResp, err := client.DescribeSnapshots(describeSnapshotsRequest)
		if err != nil {
			return fmt.Errorf("Query snapshot failed %v", err)
		}
		snapshots := snapshotResp.Snapshots.Snapshot
		if len(snapshots) > 0 {
			return fmt.Errorf("Snapshots weren't successfully deleted by " +
				"`ecs_image_force_delete_snapshots`")
		}
		return nil
	}
}

func TestBuilderAcc_imageTags(t *testing.T) {
	t.Parallel()
	if skip := testAccPreCheck(t); skip == true {
		return
	}
	outputfilename := fmt.Sprintf("test-manifest-%s.json", uuid.TimeOrderedUUID())
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-alicloud-builder-image-tags",
		Type:     "alicloud-ecs",
		Template: fmt.Sprintf(testBuilderAccImageTags, outputfilename),
		Check:    checkImageTags(outputfilename),
	})
}

const testBuilderAccImageTags = `
{	"builders": [{
		"type": "alicloud-ecs",
		"region": "cn-beijing",
		"instance_type": "ecs.n1.tiny",
		"source_image":"ubuntu_18_04_64_20G_alibase_20190509.vhd",
		"ssh_username": "root",
		"io_optimized":"true",
		"image_name": "packer-test-imageTags_{{timestamp}}",
		"tags": {
			"TagKey1": "TagValue1",
			"TagKey2": "TagValue2"
       }
	}],
	"post-processors": [{
		"type": "manifest",
		"output": "%s"
	}]
}`

func checkImageTags(manifestfilepath string) TestCheckFunc {
	return func(buildCommand *exec.Cmd, logfile string) error {
		regionArtifacts, err := loadArtifacts(manifestfilepath)
		if err != nil {
			return fmt.Errorf("could not load manifest: %s", err)
		}

		if len(regionArtifacts) > 1 {
			return fmt.Errorf("more than 1 artifact")
		}

		imageId := regionArtifacts[defaultTestRegion]

		// describe the image, get block devices with a snapshot
		client, _ := testAliyunClient()

		describeImageTagsRequest := ecs.CreateDescribeTagsRequest()
		describeImageTagsRequest.RegionId = defaultTestRegion
		describeImageTagsRequest.ResourceType = TagResourceImage
		describeImageTagsRequest.ResourceId = imageId
		imageTagsResponse, err := client.DescribeTags(describeImageTagsRequest)
		if err != nil {
			return fmt.Errorf("Error retrieving Image Attributes for ECS Image Artifact (%#v) "+
				"in ECS Image Tags Test: %s", regionArtifacts, err)
		}

		if len(imageTagsResponse.Tags.Tag) != 2 {
			return fmt.Errorf("expect 2 tags set on image %s but got %d", imageId, len(imageTagsResponse.Tags.Tag))
		}

		for _, tag := range imageTagsResponse.Tags.Tag {
			if tag.TagKey != "TagKey1" && tag.TagKey != "TagKey2" {
				return fmt.Errorf("tags on image %s should be within the list of TagKey1 and TagKey2 but got %s", imageId, tag.TagKey)
			}

			if tag.TagKey == "TagKey1" && tag.TagValue != "TagValue1" {
				return fmt.Errorf("the value for tag %s on image %s should be TagValue1 but got %s", tag.TagKey, imageId, tag.TagValue)
			} else if tag.TagKey == "TagKey2" && tag.TagValue != "TagValue2" {
				return fmt.Errorf("the value for tag %s on image %s should be TagValue2 but got %s", tag.TagKey, imageId, tag.TagValue)
			}
		}

		describeImagesRequest := ecs.CreateDescribeImagesRequest()
		describeImagesRequest.RegionId = defaultTestRegion
		describeImagesRequest.ImageId = imageId
		imagesResponse, err := client.DescribeImages(describeImagesRequest)
		if err != nil {
			return fmt.Errorf("describe images failed due to %s", err)
		}

		if len(imagesResponse.Images.Image) == 0 {
			return fmt.Errorf("image %s generated can not be found", imageId)
		}

		image := imagesResponse.Images.Image[0]
		for _, mapping := range image.DiskDeviceMappings.DiskDeviceMapping {
			describeSnapshotTagsRequest := ecs.CreateDescribeTagsRequest()
			describeSnapshotTagsRequest.RegionId = defaultTestRegion
			describeSnapshotTagsRequest.ResourceType = TagResourceSnapshot
			describeSnapshotTagsRequest.ResourceId = mapping.SnapshotId
			snapshotTagsResponse, err := client.DescribeTags(describeSnapshotTagsRequest)
			if err != nil {
				return fmt.Errorf("failed to get snapshot tags due to %s", err)
			}

			if len(snapshotTagsResponse.Tags.Tag) != 2 {
				return fmt.Errorf("expect 2 tags set on snapshot %s but got %d", mapping.SnapshotId, len(snapshotTagsResponse.Tags.Tag))
			}

			for _, tag := range snapshotTagsResponse.Tags.Tag {
				if tag.TagKey != "TagKey1" && tag.TagKey != "TagKey2" {
					return fmt.Errorf("tags on snapshot %s should be within the list of TagKey1 and TagKey2 but got %s", mapping.SnapshotId, tag.TagKey)
				}

				if tag.TagKey == "TagKey1" && tag.TagValue != "TagValue1" {
					return fmt.Errorf("the value for tag %s on snapshot %s should be TagValue1 but got %s", tag.TagKey, mapping.SnapshotId, tag.TagValue)
				} else if tag.TagKey == "TagKey2" && tag.TagValue != "TagValue2" {
					return fmt.Errorf("the value for tag %s on snapshot %s should be TagValue2 but got %s", tag.TagKey, mapping.SnapshotId, tag.TagValue)
				}
			}
		}

		return nil
	}
}

func TestBuilderAcc_dataDiskEncrypted(t *testing.T) {
	t.Parallel()
	if skip := testAccPreCheck(t); skip == true {
		return
	}
	outputfilename := fmt.Sprintf("test-manifest-%s.json", uuid.TimeOrderedUUID())
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-alicloud-builder-data-disk-encrypted",
		Type:     "alicloud-ecs",
		Template: fmt.Sprintf(testBuilderAccDataDiskEncrypted, outputfilename),
		Check:    checkDataDiskEncrypted(outputfilename),
	})
}

const testBuilderAccDataDiskEncrypted = `
{	"builders": [{
		"type": "alicloud-ecs",
		"region": "cn-beijing",
		"instance_type": "ecs.n1.tiny",
		"source_image":"ubuntu_18_04_64_20G_alibase_20190509.vhd",
		"io_optimized":"true",
		"ssh_username":"root",
		"image_name": "packer-test-dataDiskEncrypted_{{timestamp}}",
		"image_disk_mappings": [
			{
				"disk_name": "data_disk1",
				"disk_size": 25,
				"disk_encrypted": true,
				"disk_delete_with_instance": true
			},
			{
				"disk_name": "data_disk2",
				"disk_size": 35,
				"disk_encrypted": false,
				"disk_delete_with_instance": true
			},
			{
				"disk_name": "data_disk3",
				"disk_size": 45,
				"disk_delete_with_instance": true
			}
		]
	}],
	"post-processors": [{
		"type": "manifest",
		"output": "%s"
	}]
}`

func checkDataDiskEncrypted(manifestfilepath string) TestCheckFunc {
	return func(buildCommand *exec.Cmd, logfile string) error {
		regionArtifacts, err := loadArtifacts(manifestfilepath)
		if err != nil {
			return fmt.Errorf("could not load manifest: %s", err)
		}

		if len(regionArtifacts) > 1 {
			return fmt.Errorf("more than 1 artifact")
		}

		imageId := regionArtifacts[defaultTestRegion]

		// describe the image, get block devices with a snapshot
		client, _ := testAliyunClient()

		describeImagesRequest := ecs.CreateDescribeImagesRequest()
		describeImagesRequest.RegionId = defaultTestRegion
		describeImagesRequest.ImageId = imageId
		imagesResponse, err := client.DescribeImages(describeImagesRequest)
		if err != nil {
			return fmt.Errorf("describe images failed due to %s", err)
		}

		if len(imagesResponse.Images.Image) == 0 {
			return fmt.Errorf("image %s generated can not be found", imageId)
		}
		image := imagesResponse.Images.Image[0]

		var snapshotIds []string
		for _, mapping := range image.DiskDeviceMappings.DiskDeviceMapping {
			snapshotIds = append(snapshotIds, mapping.SnapshotId)
		}

		data, _ := json.Marshal(snapshotIds)

		describeSnapshotRequest := ecs.CreateDescribeSnapshotsRequest()
		describeSnapshotRequest.RegionId = defaultTestRegion
		describeSnapshotRequest.SnapshotIds = string(data)
		describeSnapshotsResponse, err := client.DescribeSnapshots(describeSnapshotRequest)
		if err != nil {
			return fmt.Errorf("describe data snapshots failed due to %s", err)
		}
		if len(describeSnapshotsResponse.Snapshots.Snapshot) != 4 {
			return fmt.Errorf("expect %d data snapshots but got %d", len(snapshotIds), len(describeSnapshotsResponse.Snapshots.Snapshot))
		}
		snapshots := describeSnapshotsResponse.Snapshots.Snapshot
		for _, snapshot := range snapshots {
			if snapshot.SourceDiskType == DiskTypeSystem {
				if snapshot.Encrypted != false {
					return fmt.Errorf("the system snapshot expected to be non-encrypted but got true")
				}

				continue
			}

			if snapshot.SourceDiskSize == "25" && snapshot.Encrypted != true {
				return fmt.Errorf("the first snapshot expected to be encrypted but got false")
			}

			if snapshot.SourceDiskSize == "35" && snapshot.Encrypted != false {
				return fmt.Errorf("the second snapshot expected to be non-encrypted but got true")
			}

			if snapshot.SourceDiskSize == "45" && snapshot.Encrypted != false {
				return fmt.Errorf("the third snapshot expected to be non-encrypted but got true")
			}
		}
		return nil
	}
}

func TestBuilderAcc_systemDiskEncrypted(t *testing.T) {
	t.Parallel()
	if skip := testAccPreCheck(t); skip == true {
		return
	}
	outputfilename := fmt.Sprintf("test-manifest-%s.json", uuid.TimeOrderedUUID())
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-alicloud-builder-system-disk-encrypted",
		Type:     "alicloud-ecs",
		Template: fmt.Sprintf(testBuilderAccSystemDiskEncrypted, outputfilename),
		Check:    checkSystemDiskEncrypted(outputfilename),
	})
}

const testBuilderAccSystemDiskEncrypted = `
{
	"builders": [{
		"type": "alicloud-ecs",
		"region": "cn-beijing",
		"instance_type": "ecs.n1.tiny",
		"source_image":"ubuntu_18_04_64_20G_alibase_20190509.vhd",
		"io_optimized":"true",
		"ssh_username":"root",
		"image_name": "packer-test_{{timestamp}}",
		"image_encrypted": "true"
	}],
	"post-processors": [{
		"type": "manifest",
		"output": "%s"
	}]
}`

func checkSystemDiskEncrypted(manifestfilepath string) TestCheckFunc {
	return func(buildCommand *exec.Cmd, logfile string) error {
		regionArtifacts, err := loadArtifacts(manifestfilepath)
		if err != nil {
			return fmt.Errorf("could not load manifest: %s", err)
		}

		if len(regionArtifacts) > 1 {
			return fmt.Errorf("more than 1 artifact")
		}

		imageId := regionArtifacts[defaultTestRegion]
		// describe the image, get block devices with a snapshot
		client, _ := testAliyunClient()

		describeImagesRequest := ecs.CreateDescribeImagesRequest()
		describeImagesRequest.RegionId = defaultTestRegion
		describeImagesRequest.ImageId = imageId
		describeImagesRequest.Status = ImageStatusQueried
		imagesResponse, err := client.DescribeImages(describeImagesRequest)
		if err != nil {
			return fmt.Errorf("describe images failed due to %s", err)
		}

		if len(imagesResponse.Images.Image) == 0 {
			return fmt.Errorf("image %s generated can not be found", imageId)
		}

		image := imagesResponse.Images.Image[0]
		if image.IsCopied == false {
			return fmt.Errorf("image %s generated expexted to be copied but false", image.ImageId)
		}

		describeSnapshotRequest := ecs.CreateDescribeSnapshotsRequest()
		describeSnapshotRequest.RegionId = defaultTestRegion
		describeSnapshotRequest.SnapshotIds = fmt.Sprintf("[\"%s\"]", image.DiskDeviceMappings.DiskDeviceMapping[0].SnapshotId)
		describeSnapshotsResponse, err := client.DescribeSnapshots(describeSnapshotRequest)
		if err != nil {
			return fmt.Errorf("describe system snapshots failed due to %s", err)
		}
		snapshots := describeSnapshotsResponse.Snapshots.Snapshot[0]

		if snapshots.Encrypted != true {
			return fmt.Errorf("system snapshot of image %s expected to be encrypted but got false", imageId)
		}

		return nil
	}
}

func testAccPreCheck(t *testing.T) bool {
	if os.Getenv(acctest.TestEnvVar) == "" {
		t.Skipf("Acceptance tests skipped unless env '%s' set", acctest.TestEnvVar)
		return true
	}

	if v := os.Getenv("ALICLOUD_ACCESS_KEY"); v == "" {
		t.Fatal("ALICLOUD_ACCESS_KEY must be set for acceptance tests")
		return true
	}

	if v := os.Getenv("ALICLOUD_SECRET_KEY"); v == "" {
		t.Fatal("ALICLOUD_SECRET_KEY must be set for acceptance tests")
		return true
	}
	return false
}

func testAliyunClient() (*ClientWrapper, error) {
	access := &AlicloudAccessConfig{AlicloudRegion: "cn-beijing"}
	err := access.Config()
	if err != nil {
		return nil, err
	}
	client, err := access.Client()
	if err != nil {
		return nil, err
	}

	return client, nil
}

// loads the manifest file, and reads the artifact ids from the manifest,
// returning a map[region]artifact_id
// To make sure this works, you must make sure the manifestfilepath is unique.
func loadArtifacts(manifestfilepath string) (map[string]string, error) {
	regionArtifacts := map[string]string{}
	// load manifest, read artifact names, and check them
	manifest, err := testutils.GetArtifact(manifestfilepath)
	if err != nil {
		return regionArtifacts, fmt.Errorf("Error loading artifacts from manifest")
	}

	regionArtifactsString := manifest.Builds[0].ArtifactId
	regionStrings := strings.Split(regionArtifactsString, ",")
	for _, rstring := range regionStrings {
		artifact := strings.Split(rstring, ":")
		if len(artifact) != 2 {
			return regionArtifacts, fmt.Errorf("Malformed manifest: artifact is %s", rstring)
		}
		regionArtifacts[artifact[0]] = artifact[1]
	}

	return regionArtifacts, nil
}
