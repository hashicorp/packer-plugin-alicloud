The Alicloud plugin can be used to create or import custom images on on the Alibaba Cloud platform.

### Example Usage
Packer 1.7.0 and later

---

```
packer {
  required_plugins {
    amazon = {
      version = ">= 1.0.0"
      source = "github.com/hashicorp/amazon"
    }
  }
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
}
```

### Available Components

**Builders**
- [alicloud-ecs](/packer/integrations/hashicorp/alicloud-ecs) - Provides the capability to build customized images based on an existing base image.

**Post-Processors**
- [alicloud-import](/packer/integrations/hashicorp/alicloud-import) - Takes a RAW or VHD artifact from various builders and imports it to an Alicloud ECS Image.
