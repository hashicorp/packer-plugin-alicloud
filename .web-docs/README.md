The Alicloud plugin can be used to create or import custom images on the Alibaba Cloud platform.

### Installation

To install this plugin, copy and paste this code into your Packer configuration .
Then, run [`packer init`](https://www.packer.io/docs/commands/init).

```hcl
packer {
  required_plugins {
    alicloud = {
      source  = "github.com/hashicorp/alicloud"
      version = "~> 1"
    }
  }
}
```

Alternatively, you can use `packer plugins install` to manage installation of this plugin.

```sh
$ packer plugins install github.com/hashicorp/alicloud
```

### Components

#### Builders
- [alicloud-ecs](/packer/integrations/hashicorp/alicloud/latest/components/builder/alicloud-ecs) - Provides the capability to build customized images based on an existing base image.

#### Post-Processors
- [alicloud-import](/packer/integrations/hashicorp/alicloud/latest/components/post-processor/alicloud-import) - Takes a RAW or VHD artifact from various builders and imports it to an Alicloud ECS Image.
