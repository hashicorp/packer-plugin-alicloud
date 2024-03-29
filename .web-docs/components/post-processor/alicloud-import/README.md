Type: `alicloud-import`
Artifact BuilderId: `packer.post-processor.alicloud-import`

The Packer Alicloud Import post-processor takes a RAW or VHD artifact from
various builders and imports it to an Alicloud ECS Image.

## How Does it Work?

The import process operates by making a temporary copy of the RAW or VHD to an
OSS bucket, and calling an import task in ECS on the RAW or VHD file. Once
completed, an Alicloud ECS Image is returned. The temporary RAW or VHD copy in
OSS can be discarded after the import is complete.

## Configuration

There are some configuration options available for the post-processor. There
are two categories: required and optional parameters.

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


<!-- Code generated from the comments of the AlicloudImageConfig struct in builder/ecs/image_config.go; DO NOT EDIT MANUALLY -->

- `image_name` (string) - The name of the user-defined image, [2, 128] English or Chinese
  characters. It must begin with an uppercase/lowercase letter or a
  Chinese character, and may contain numbers, `_` or `-`. It cannot begin
  with `http://` or `https://`.

<!-- End of code generated from the comments of the AlicloudImageConfig struct in builder/ecs/image_config.go; -->


<!-- Code generated from the comments of the Config struct in post-processor/alicloud-import/post-processor.go; DO NOT EDIT MANUALLY -->

- `oss_bucket_name` (string) - The name of the OSS bucket where the RAW or VHD file will be copied to
  for import. If the Bucket doesn't exist, the post-process will create it for
  you.

- `image_os_type` (string) - Type of the OS, like linux/windows

- `image_platform` (string) - Platform such as `CentOS`

- `image_architecture` (string) - Platform type of the image system: `i386` or `x86_64`

- `format` (string) - The format of the image for import, now alicloud only support RAW and
  VHD.

<!-- End of code generated from the comments of the Config struct in post-processor/alicloud-import/post-processor.go; -->


### Optional:

- `keep_input_artifact` (boolean) - if true, do not delete the RAW or VHD
  disk image after importing it to the cloud. Defaults to false.

<!-- Code generated from the comments of the Config struct in post-processor/alicloud-import/post-processor.go; DO NOT EDIT MANUALLY -->

- `oss_key_name` (string) - The name of the object key in `oss_bucket_name` where the RAW or VHD
  file will be copied to for import. This is treated as a [template
  engine](/packer/docs/templates/legacy_json_templates/engine), and you may access any of the variables
  stored in the generated data using the [build](/packer/docs/templates/legacy_json_templates/engine)
  template function.

- `skip_clean` (bool) - Whether we should skip removing the RAW or VHD file uploaded to OSS
  after the import process has completed. `true` means that we should
  leave it in the OSS bucket, `false` means to clean it out. Defaults to
  `false`.

- `image_system_size` (string) - Size of the system disk, in GB, values
   range:
    - cloud - 5 \~ 2000
    - cloud_efficiency - 20 \~ 2048
    - cloud_ssd - 20 \~ 2048
    - cloud_essd - 20 \~ 2048

<!-- End of code generated from the comments of the Config struct in post-processor/alicloud-import/post-processor.go; -->


## Basic Example

Here is a basic example. This assumes that the builder has produced a RAW
artifact. The user must have the role `AliyunECSImageImportDefaultRole` with
`AliyunECSImageImportRolePolicy`, post-process will automatically configure the
role and policy for you if you have the privilege, otherwise, you have to ask
the administrator configure for you in advance.

```json
"post-processors":[
    {
      "access_key":"{{user `access_key`}}",
      "secret_key":"{{user `secret_key`}}",
      "type":"alicloud-import",
      "oss_bucket_name": "packer",
      "image_name": "packer_import",
      "image_os_type": "linux",
      "image_platform": "CentOS",
      "image_architecture": "x86_64",
      "image_system_size": "40",
      "region":"cn-beijing"
    }
  ]
```

This will take the RAW generated by a builder and upload it to OSS. In this
case, an existing bucket called `packer` in the `cn-beijing` region will be
where the copy is placed.

Once uploaded, the import process will start, creating an Alicloud ECS image in
the `cn-beijing` region with the name you specified in template file.
