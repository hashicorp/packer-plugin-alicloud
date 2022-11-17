package ecs

import (
	"context"
	"fmt"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepStopBuilder struct {
	StopReason string
}

func (s *stepStopBuilder) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say(fmt.Sprintf("build image stopped because: %s", s.StopReason))

	return multistep.ActionHalt
}

func (s *stepStopBuilder) Cleanup(multistep.StateBag) {
	// No cleanup...
}
