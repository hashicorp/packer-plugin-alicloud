package ecs

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepStopBuilder struct {
	Message string
}

func (s *stepStopBuilder) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say(fmt.Sprintf("Image build stopped because: %s", s.Message))

	return multistep.ActionHalt
}

func (s *stepStopBuilder) Cleanup(multistep.StateBag) {
	// No cleanup...
}
