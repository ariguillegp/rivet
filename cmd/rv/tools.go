package main

import "github.com/ariguillegp/rivet/internal/core"

func installedAgentTools() []string {
	supported := core.SupportedTools()
	installed := make([]string, 0, len(supported))
	for _, tool := range supported {
		if tool == core.ToolNone {
			continue
		}
		if _, err := doctorLookPath(tool); err == nil {
			installed = append(installed, tool)
		}
	}
	return installed
}

func availableToolsForUI() []string {
	installed := installedAgentTools()
	return append(installed, core.ToolNone)
}
