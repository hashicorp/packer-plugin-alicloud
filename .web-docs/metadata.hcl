# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# For full specification on the configuration of this file visit:
# https://github.com/hashicorp/integration-template#metadata-configuration
integration {
  name = "Alicloud"
  description = "The Alicloud plugin can be used with HashiCorp Packer to create custom images on the Alibaba Cloud platform"
  identifier = "packer/hashicorp/alicloud"
  component {
    type = "builder"
    name = "Alicloud ECS"
    slug = "alicloud-ecs"
  }
  component {
    type = "post-processor"
    name = "Alicloud Import"
    slug = "alicloud-import"
  }
}
