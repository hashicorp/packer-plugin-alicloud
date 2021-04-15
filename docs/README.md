# Alicloud Plugins

<!--
  Include a short overview about the plugin.

  This document is a great location for creating a table of contents for each
  of the components the plugin may provide. This document should load automatically
  when navigating to the docs directory for a plugin.

-->

The Alicloud plugin is intended as a starting point for creating Packer plugins, containing:

- [alicloud-ecs builder](/docs/builders/alicloud-ecs.mdx) - provides the capability to build customized images based on an existing base image.

- [alicloud-import post-processor](/docs/post-processors/alicloud-import.mdx) - Takes a RAW or VHD artifact from various builders and imports it to an Alicloud ECS Image.