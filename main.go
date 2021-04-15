package main

import (
	"fmt"
	"os"

	ecsbuilder "github.com/hashicorp/packer-plugin-alicloud/builder/ecs"
	importpp "github.com/hashicorp/packer-plugin-alicloud/post-processor/alicloud-import"
	version "github.com/hashicorp/packer-plugin-alicloud/version"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder("ecs", new(ecsbuilder.Builder))
	pps.RegisterPostProcessor("import", new(importpp.PostProcessor))
	pps.SetVersion(version.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
