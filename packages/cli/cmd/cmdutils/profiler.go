package cmdutils

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/pkg/profile"
)

func CollectProfile(enabled bool) func() {
	if !enabled {
		return func() {}
	}
	fmt.Println(color.YellowString("⚠️ CPU Profiling enabled"))
	var stop = profile.Start(profile.ProfilePath("."), profile.CPUProfile, profile.Quiet)
	return func() {
		stop.Stop()
		fmt.Println(color.YellowString("⚠️ CPU Profiling output - cpu.pprof"))
	}
}
