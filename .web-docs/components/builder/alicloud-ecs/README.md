Type: `alicloud-ecs`
Artifact BuilderId: `alibaba.alicloud`

The `alicloud-ecs` Packer builder plugin provide the capability to build
customized images based on an existing base images.

## Configuration Reference

The following configuration options are available for building Alicloud images.
In addition to the options listed here, a
[communicator](/packer/docs/templates/legacy_json_templates/communicator) can be configured for this
builder.

### Required:

<!-- Code generated from the comments of the AlicloudAccessConfig struct in builder/ecs/access_config.go; DO NOT EDIT MANUALLY -->

- `access_key` (string) - Alicloud access key must be provided unless `profile` is set, but it can
  also be sourced from the `ALICLOUD_ACCESS_KEY` environment variable.

- `secret_key` (string) - Alicloud secret key must be provided unless `profile` is set, but it can
  also be sourced from the `ALICLOUD_SECRET_KEY` environment variable.

- `region` (string) - Alicloud region must be provided unless `profile` is set, but it can
  also be sourced from the `ALICLOUD_REGION` environment variable.

- `ram_role_name` (string) - Alicloud RamRole must be provided for EcsRamRole mode unless `profile` is set.

- `ram_role_arn` (string) - Alicloud RamRoleArn must be provided for RamRoleArn mode unless `profile` is set.

- `ram_session_name` (string) - Alicloud RamSessionName must be provided for RamRoleArn mode unless `profile` is set.

<!-- End of code generated from the comments of the AlicloudAccessConfig struct in builder/ecs/access_config.go; -->


<!-- Code generated from the comments of the RunConfig struct in builder/ecs/run_config.go; DO NOT EDIT MANUALLY -->

- `instance_type` (string) - Type of the instance. For values, see [Instance Type
  Table](https://www.alibabacloud.com/help/doc-detail/25378.htm?spm=a3c0i.o25499en.a3.9.14a36ac8iYqKRA).
  You can also obtain the latest instance type table by invoking the
  [Querying Instance Type
  Table](https://intl.aliyun.com/help/doc-detail/25620.htm?spm=a3c0i.o25499en.a3.6.Dr1bik)
  interface.

- `source_image` (string) - This is the base image id which you want to
  create your customized images.

- `image_family` (string) - The name of the image family. Customer can set this parameter to choose the latest available custom image from
  the specified image family to create the instance.

<!-- End of code generated from the comments of the RunConfig struct in builder/ecs/run_config.go; -->


<!-- Code generated from the comments of the AlicloudImageConfig struct in builder/ecs/image_config.go; DO NOT EDIT MANUALLY -->

- `image_name` (string) - The name of the user-defined image, [2, 128] English or Chinese
  characters. It must begin with an uppercase/lowercase letter or a
  Chinese character, and may contain numbers, `_` or `-`. It cannot begin
  with `http://` or `https://`.

<!-- End of code generated from the comments of the AlicloudImageConfig struct in builder/ecs/image_config.go; -->


### Optional:

<!-- Code generated from the comments of the AlicloudAccessConfig struct in builder/ecs/access_config.go; DO NOT EDIT MANUALLY -->

- `skip_region_validation` (bool) - The region validation can be skipped if this value is true, the default
  value is false.

- `skip_image_validation` (bool) - The image validation can be skipped if this value is true, the default
  value is false.

- `profile` (string) - Alicloud profile must be set unless `access_key` is set; it can also be
  sourced from the `ALICLOUD_PROFILE` environment variable.

- `shared_credentials_file` (string) - Alicloud shared credentials file path. If this file exists, access and
  secret keys will be read from this file.

- `security_token` (string) - STS access token, can be set through template or by exporting as
  environment variable such as `export SECURITY_TOKEN=value`.

- `custom_endpoint_ecs` (string) - This option is useful if you use a cloud provider whose API is
  compatible with aliyun ECS. Specify another endpoint with this option.

<!-- End of code generated from the comments of the AlicloudAccessConfig struct in builder/ecs/access_config.go; -->


<!-- Code generated from the comments of the AlicloudDiskDevices struct in builder/ecs/image_config.go; DO NOT EDIT MANUALLY -->

- `system_disk_mapping` (AlicloudDiskDevice) - Image disk mapping for the system disk.
  See the [disk device configuration](#disk-devices-configuration) section
  for more information on options.
  Usage example:
  
  ```json
  "builders": [{
    "type":"alicloud-ecs",
    "system_disk_mapping": {
      "disk_size": 50,
      "disk_name": "mydisk"
    },
    ...
  }
  ```

- `image_disk_mappings` ([]AlicloudDiskDevice) - Add one or more data disks to the image.
  See the [disk device configuration](#disk-devices-configuration) section
  for more information on options.
  Usage example:
  
  ```json
   "builders": [{
     "type":"alicloud-ecs",
     "image_disk_mappings": [
       {
         "disk_snapshot_id": "someid",
         "disk_device": "dev/xvdb"
       }
     ],
     ...
   }
   ```

<!-- End of code generated from the comments of the AlicloudDiskDevices struct in builder/ecs/image_config.go; -->


<!-- Code generated from the comments of the RunConfig struct in builder/ecs/run_config.go; DO NOT EDIT MANUALLY -->

- `associate_public_ip_address` (bool) - Associate Public Ip Address

- `zone_id` (string) - ID of the zone to which the disk belongs.

- `io_optimized` (boolean) - Whether an ECS instance is I/O optimized or not. If this option is not
  provided, the value will be determined by product API according to what
  `instance_type` is used.

- `description` (string) - Description

- `force_stop_instance` (bool) - Whether to force shutdown upon device
  restart. The default value is `false`.
  
  If it is set to `false`, the system is shut down normally; if it is set to
  `true`, the system is forced to shut down.

- `disable_stop_instance` (bool) - If this option is set to true, Packer
  will not stop the instance for you, and you need to make sure the instance
  will be stopped in the final provisioner command. Otherwise, Packer will
  timeout while waiting the instance to be stopped. This option is provided
  for some specific scenarios that you want to stop the instance by yourself.
  E.g., Sysprep a windows which may shutdown the instance within its command.
  The default value is false.

- `ecs_ram_role_name` (string) - Ram Role to apply when launching the instance.

- `run_tags` (map[string]string) - Key/value pair tags to apply to the instance that is *launched*
  to create the image.

- `security_group_id` (string) - ID of the security group to which a newly
  created instance belongs. Mutual access is allowed between instances in one
  security group. If not specified, the newly created instance will be added
  to the default security group. If the default group doesn’t exist, or the
  number of instances in it has reached the maximum limit, a new security
  group will be created automatically.

- `security_group_name` (string) - The security group name. The default value
  is blank. [2, 128] English or Chinese characters, must begin with an
  uppercase/lowercase letter or Chinese character. Can contain numbers, .,
  _ or -. It cannot begin with `http://` or `https://`.

- `security_enhancement_strategy` (string) - Specifies whether to enable security hardening. Valid values:
  Active: enables security hardening. This value is applicable only to public images.
  Deactive: does not enable security hardening. This value is applicable to all image types.

- `user_data` (string) - User data to apply when launching the instance. Note
  that you need to be careful about escaping characters due to the templates
  being JSON. It is often more convenient to use user_data_file, instead.
  Packer will not automatically wait for a user script to finish before
  shutting down the instance this must be handled in a provisioner.

- `user_data_file` (string) - Path to a file that will be used for the user
  data when launching the instance.

- `vpc_id` (string) - VPC ID allocated by the system.

- `vpc_name` (string) - The VPC name. The default value is blank. [2, 128]
  English or Chinese characters, must begin with an uppercase/lowercase
  letter or Chinese character. Can contain numbers, _ and -. The disk
  description will appear on the console. Cannot begin with `http://` or
  `https://`.

- `vpc_cidr_block` (string) - Value options: 192.168.0.0/16 and
  172.16.0.0/16. When not specified, the default value is 172.16.0.0/16.

- `vswitch_id` (string) - The ID of the VSwitch to be used.

- `vswitch_name` (string) - The ID of the VSwitch to be used.

- `eip_id` (string) - The ID of the EIP to be used as public ip for the instance

- `instance_name` (string) - Display name of the instance, which is a string of 2 to 128 Chinese or
  English characters. It must begin with an uppercase/lowercase letter or
  a Chinese character and can contain numerals, `.`, `_`, or `-`. The
  instance name is displayed on the Alibaba Cloud console. If this
  parameter is not specified, the default value is InstanceId of the
  instance. It cannot begin with `http://` or `https://`.

- `internet_charge_type` (string) - Internet charge type, which can be
  `PayByTraffic` or `PayByBandwidth`. Optional values:
  -   `PayByBandwidth`
  -   `PayByTraffic`
  
  If this parameter is not specified, the default value is `PayByBandwidth`.
  For the regions out of China, currently only support `PayByTraffic`, you
  must set it manfully.

- `internet_max_bandwidth_out` (int) - Maximum outgoing bandwidth to the
  public network, measured in Mbps (Mega bits per second).
  
  Value range:
  -   `PayByBandwidth`: \[0, 100\]. If this parameter is not specified, API
      automatically sets it to 0 Mbps.
  -   `PayByTraffic`: \[1, 100\]. If this parameter is not specified, an
      error is returned.

- `wait_snapshot_ready_timeout` (int) - Timeout of creating snapshot(s).
  The default timeout is 3600 seconds if this option is not set or is set
  to 0. For those disks containing lots of data, it may require a higher
  timeout value.

- `wait_copying_image_ready_timeout` (int) - Timeout of copying image.
  The default timeout is 3600 seconds if this option is not set or is set
  to 0.

- `ssh_private_ip` (bool) - If this value is true, packer will connect to
  the ECS created through private ip instead of allocating a public ip or an
  EIP. The default value is false.

- `skip_create_image` (bool) - If true, Packer will not create a final image. Defaults to `false`.

<!-- End of code generated from the comments of the RunConfig struct in builder/ecs/run_config.go; -->


<!-- Code generated from the comments of the AlicloudImageConfig struct in builder/ecs/image_config.go; DO NOT EDIT MANUALLY -->

- `image_version` (string) - The version number of the image, with a length limit of 1 to 40 English
  characters.

- `image_description` (string) - The description of the image, with a length limit of 0 to 256
  characters. Leaving it blank means null, which is the default value. It
  cannot begin with `http://` or `https://`.

- `resource_group_id` (string) - The ID of the resource group to which to assign the custom image.
  If you do not specify this parameter, the image is assigned to the default resource group.

- `image_share_account` ([]string) - The IDs of to-be-added Aliyun accounts to which the image is shared. The
  number of accounts is 1 to 10. If number of accounts is greater than 10,
  this parameter is ignored.

- `image_unshare_account` ([]string) - Alicloud Image UN Share Accounts

- `image_copy_regions` ([]string) - Copy to the destination regionIds.

- `image_copy_names` ([]string) - The name of the destination image, [2, 128] English or Chinese
  characters. It must begin with an uppercase/lowercase letter or a
  Chinese character, and may contain numbers, _ or -. It cannot begin with
  `http://` or `https://`.

- `image_encrypted` (boolean) - Whether or not to encrypt the target images,            including those
  copied if image_copy_regions is specified. If this option is set to
  true, a temporary image will be created from the provisioned instance in
  the main region and an encrypted copy will be generated in the same
  region. By default, Packer will keep the encryption setting to what it
  was in the source image.

- `image_force_delete` (bool) - If this value is true, when the target image names including those
  copied are duplicated with existing images, it will delete the existing
  images and then create the target images, otherwise, the creation will
  fail. The default value is false. Check `image_name` and
  `image_copy_names` options for names of target images. If
  [-force](/packer/docs/commands/build#force) option is provided in `build`
  command, this option can be omitted and taken as true.

- `image_force_delete_snapshots` (bool) - If this value is true, when delete the duplicated existing images, the
  source snapshots of those images will be delete either. If
  [-force](/packer/docs/commands/build#force) option is provided in `build`
  command, this option can be omitted and taken as true.

- `image_force_delete_instances` (bool) - Alicloud Image Force Delete Instances

- `image_ignore_data_disks` (bool) - If this value is true, the image created will not include any snapshot
  of data disks. This option would be useful for any circumstance that
  default data disks with instance types are not concerned. The default
  value is false.

- `tags` (map[string]string) - Key/value pair tags applied to the destination image and relevant
  snapshots.

- `tag` ([]{key string, value string}) - Same as [`tags`](#tags) but defined as a singular repeatable block
  containing a `key` and a `value` field. In HCL2 mode the
  [`dynamic_block`](/packer/docs/templates/hcl_templates/expressions#dynamic-blocks)
  will allow you to create those programatically.

- `target_image_family` (string) - The image family of the user-defined image, [2, 128] English or Chinese
  characters. It must begin with an uppercase/lowercase letter or a
  Chinese character, and may contain numbers, `_` or `-`. It cannot begin
  with `aliyun`, `acs:`, `http://` or `https://`.

- `boot_mode` (string) - The boot mode of the user-defined image, it should to be one of 'BIOS', 'UEFI' or 'UEFI-Preferred'.

- `kms_key_copy_ids` ([]string) - Copy to the destination KMS key ID array

- `kms_key_id` (string) - The source image KMS key ID used to encrypt the disk.

<!-- End of code generated from the comments of the AlicloudImageConfig struct in builder/ecs/image_config.go; -->


<!-- Code generated from the comments of the SSHTemporaryKeyPair struct in communicator/config.go; DO NOT EDIT MANUALLY -->

- `temporary_key_pair_type` (string) - `dsa` | `ecdsa` | `ed25519` | `rsa` ( the default )
  
  Specifies the type of key to create. The possible values are 'dsa',
  'ecdsa', 'ed25519', or 'rsa'.
  
  NOTE: DSA is deprecated and no longer recognized as secure, please
  consider other alternatives like RSA or ED25519.

- `temporary_key_pair_bits` (int) - Specifies the number of bits in the key to create. For RSA keys, the
  minimum size is 1024 bits and the default is 4096 bits. Generally, 3072
  bits is considered sufficient. DSA keys must be exactly 1024 bits as
  specified by FIPS 186-2. For ECDSA keys, bits determines the key length
  by selecting from one of three elliptic curve sizes: 256, 384 or 521
  bits. Attempting to use bit lengths other than these three values for
  ECDSA keys will fail. Ed25519 keys have a fixed length and bits will be
  ignored.
  
  NOTE: DSA is deprecated and no longer recognized as secure as specified
  by FIPS 186-5, please consider other alternatives like RSA or ED25519.

<!-- End of code generated from the comments of the SSHTemporaryKeyPair struct in communicator/config.go; -->


- `ssh_keypair_name` (string) - If specified, this is the key that will be used for SSH with the
  machine. The key must match a key pair name loaded up into the remote.
  By default, this is blank, and Packer will generate a temporary keypair
  unless [`ssh_password`](#ssh_password) is used.
  [`ssh_private_key_file`](#ssh_private_key_file) or
  [`ssh_agent_auth`](#ssh_agent_auth) must be specified when
  [`ssh_keypair_name`](#ssh_keypair_name) is utilized.


- `ssh_private_key_file` (string) - Path to a PEM encoded private key file to use to authenticate with SSH.
  The `~` can be used in path and will be expanded to the home directory
  of current user.


- `ssh_agent_auth` (bool) - If true, the local SSH agent will be used to authenticate connections to
  the source instance. No temporary keypair will be created, and the
  values of [`ssh_password`](#ssh_password) and
  [`ssh_private_key_file`](#ssh_private_key_file) will be ignored. The
  environment variable `SSH_AUTH_SOCK` must be set for this option to work
  properly.


### Alicloud RAM permission

Finally the plugin should gain a set of Alicloud RAM permission to call Alicloud API.

The following policy document provides the minimal set permissions necessary for the Alicloud plugin to work:

```json
{
  "Version": "1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecs:AttachKeyPair",
        "ecs:CreateKeyPair",
        "ecs:DeleteKeyPairs",
        "ecs:DetachKeyPair",
        "ecs:DescribeKeyPairs",
        "ecs:DescribeDisks",
        "ecs:ImportKeyPair",
        "ecs:CreateSecurityGroup",
        "ecs:AuthorizeSecurityGroup",
        "ecs:AuthorizeSecurityGroupEgress",
        "ecs:DescribeSecurityGroups",
        "ecs:DeleteSecurityGroup",
        "ecs:CopyImage",
        "ecs:CancelCopyImage",
        "ecs:CreateImage",
        "ecs:DescribeImages",
        "ecs:DescribeImageFromFamily",
        "ecs:DeleteImage",
        "ecs:ModifyImageAttribute",
        "ecs:DescribeImageSharePermission",
        "ecs:ModifyImageSharePermission",
        "ecs:DescribeInstances",
        "ecs:StartInstance",
        "ecs:StopInstance",
        "ecs:CreateInstance",
        "ecs:DeleteInstance",
        "ecs:RunInstances",
        "ecs:RebootInstance",
        "ecs:RenewInstance",
        "ecs:CreateSnapshot",
        "ecs:DeleteSnapshot",
        "ecs:DescribeSnapshots",
        "ecs:TagResources",
        "ecs:UntagResources",
        "ecs:AllocatePublicIpAddress",
        "ecs:AddTags",
        "vpc:DescribeVpcs",
        "vpc:CreateVpc",
        "vpc:DeleteVpc",
        "vpc:DescribeVSwitches",
        "vpc:CreateVSwitch",
        "vpc:DeleteVSwitch",
        "vpc:AllocateEipAddress",
        "vpc:AssociateEipAddress",
        "vpc:UnassociateEipAddress",
        "vpc:ReleaseEipAddress",
        "vpc:DescribeEipAddresses"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
```

# Disk Devices Configuration

<!-- Code generated from the comments of the AlicloudDiskDevice struct in builder/ecs/image_config.go; DO NOT EDIT MANUALLY -->

- `disk_name` (string) - The value of disk name is blank by default. [2,
  128] English or Chinese characters, must begin with an
  uppercase/lowercase letter or Chinese character. Can contain numbers,
  ., _ and -. The disk name will appear on the console. It cannot
  begin with `http://` or `https://`.

- `disk_category` (string) - Category of the system disk. Optional values are:
      -   cloud - general cloud disk
      -   cloud_efficiency - efficiency cloud disk
      -   cloud_ssd - cloud SSD
      -   cloud_essd - cloud ESSD

- `disk_size` (int) - Size of the system disk, measured in GiB. Value
  range: [20, 500]. The specified value must be equal to or greater
  than max{20, ImageSize}. Default value: max{40, ImageSize}.

- `disk_snapshot_id` (string) - Snapshots are used to create the data
  disk After this parameter is specified, Size is ignored. The actual
  size of the created disk is the size of the specified snapshot.
  This field is only used in the ECSImagesDiskMappings option, not
  the ECSSystemDiskMapping option.

- `disk_description` (string) - The value of disk description is blank by
  default. [2, 256] characters. The disk description will appear on the
  console. It cannot begin with `http://` or `https://`.

- `disk_delete_with_instance` (bool) - Whether or not the disk is
  released along with the instance:

- `disk_device` (string) - Device information of the related instance:
  such as /dev/xvdb It is null unless the Status is In_use.

- `disk_encrypted` (boolean) - Whether or not to encrypt the data disk.
  If this option is set to true, the data disk will be encryped and
  corresponding snapshot in the target image will also be encrypted. By
  default, if this is an extra data disk, Packer will not encrypt the
  data disk. Otherwise, Packer will keep the encryption setting to what
  it was in the source image. Please refer to Introduction of ECS disk
  encryption for more details.

<!-- End of code generated from the comments of the AlicloudDiskDevice struct in builder/ecs/image_config.go; -->


## Basic Example

Here is a basic example for Alicloud.

**JSON**

```json
{
  "variables": {
    "access_key": "{{env `ALICLOUD_ACCESS_KEY`}}",
    "secret_key": "{{env `ALICLOUD_SECRET_KEY`}}"
  },
  "builders": [
    {
      "type": "alicloud-ecs",
      "access_key": "{{user `access_key`}}",
      "secret_key": "{{user `secret_key`}}",
      "region": "cn-beijing",
      "image_name": "packer_test2",
      "source_image": "centos_7_04_64_20G_alibase_201701015.vhd",
      "ssh_username": "root",
      "instance_type": "ecs.n1.tiny",
      "io_optimized": "true",
      "internet_charge_type": "PayByTraffic",
      "image_force_delete": "true"
      "run_tags": {
        "Managed by": "Packer",
        "Built by": "Packer"
      }
    }
  ],
  "provisioners": [
    {
      "type": "shell",
      "inline": ["sleep 30", "yum install redis.x86_64 -y"]
    }
  ]
}
```

**HCL2**

```hcl
variable "access_key" {
  type = string
}

variable "secret_key" {
  type = string
}

source "alicloud-ecs" "basic-example" {
      access_key = var.access_key
      secret_key = var.secret_key
      region = "cn-beijing"
      image_name = "packer_test2"
      source_image = "centos_7_04_64_20G_alibase_201701015.vhd"
      ssh_username = "root"
      instance_type = "ecs.n1.tiny"
      io_optimized = true
      internet_charge_type = "PayByTraffic"
      image_force_delete = true
      run_tags  = {
        "Built by"   = "Packer"
        "Managed by" = "Packer"
      }
}

build {
  sources = ["sources.alicloud-ecs.basic-example"]

  provisioner "shell" {
    inline = [
      "sleep 30", "yum install redis.x86_64 -y",
    ]
  }
}
```


~> Note: Images can become deprecated after a while; run
`aliyun ecs DescribeImages` to find one that exists.

~> Note: Since WinRM is closed by default in the system image. If you are
planning to use Windows as the base image, you need enable it by userdata in
order to connect to the instance, check
[alicloud_windows.json](https://github.com/hashicorp/packer-plugin-alicloud/tree/main/builder/examples/basic/alicloud_windows.json)
and
[winrm_enable_userdata.ps1](https://github.com/hashicorp/packer-plugin-alicloud/tree/main/builder/examples/basic/winrm_enable_userdata.ps1)
for details.

See the
[examples/alicloud](https://github.com/hashicorp/packer-plugin-alicloud/tree/main/builder/examples)
folder in the Packer project for more examples.
