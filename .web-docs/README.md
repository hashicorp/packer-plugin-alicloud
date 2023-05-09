The Alicloud plugin can be used to create or import custom images on on the Alibaba Cloud platform.

### Installation
Packer 1.7.0 and later

```hcl
packer {
  required_plugins {
    alicloud = {
      version = ">= 1.0.0"
      source = "github.com/hashicorp/alicloud"
    }
  }
}
```

### Components

#### Builders
- [alicloud-ecs](/packer/integrations/hashicorp/alicloud/latest/components/alicloud-ecs) - Provides the capability to build customized images based on an existing base image.

#### Post-Processors
- [alicloud-import](/packer/integrations/hashicorp/alicloud/latest/components/alicloud-import) - Takes a RAW or VHD artifact from various builders and imports it to an Alicloud ECS Image.
