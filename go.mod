module packer-plugin-alicloud

go 1.16

require (
	github.com/hashicorp/hcl/v2 v2.9.1
	github.com/hashicorp/packer-plugin-sdk v0.1.2
	github.com/zclconf/go-cty v1.8.1
)

replace github.com/hashicorp/packer-plugin-alicloud/builder/alicloud/ecs => /Users/mmarsh/Projects/packer-plugin-alicloud

