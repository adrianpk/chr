package app

import (
	"forge.adrianpk.com/chc/internal/core"
)

func agentWorkLoopEventSink(projectDir string, record core.LoopRunRecord) core.AgentEventSink {
	config, err := core.LoadWorkspaceConfig(projectDir)
	if err != nil {
		return nil
	}

	plan, err := core.PlanRuntime(core.RuntimePlanOptions{
		ProjectDir: projectDir,
		Config:     config.Runtime,
		RunID:      record.ID,
	})
	if err != nil {
		return nil
	}

	return core.NewFileAgentEventSink(core.AgentEventStreamPath(plan.RunArtifactDir))
}

func agentWorkExecutionEventSink(record core.ExecutionRecord) core.AgentEventSink {
	return core.NewFileAgentEventSink(core.AgentEventStreamPath(record.Artifacts.RunDir))
}
