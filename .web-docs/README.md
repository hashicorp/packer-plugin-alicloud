# Alicloud Plugins

<!--
  Include a short overview about the plugin.

  This document is a great location for creating a table of contents for each
  of the components the plugin may provide. This document should load automatically
  when navigating to the docs directory for a plugin.

-->

The Alicloud plugin can be used to create or import custom images on on the Alibaba Cloud platform.

### Components

- [alicloud-ecs builder](/packer/integrations/hashicorp/alicloud-ecs) - Provides the capability to build customized images based on an existing base image.

- [alicloud-import post-processor](/packer/integrations/hashicorp/alicloud-import) - Takes a RAW or VHD artifact from various builders and imports it to an Alicloud ECS Image.
