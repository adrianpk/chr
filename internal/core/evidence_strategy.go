package core

// EvidenceStrategyStatus defines a package boundary type.
type EvidenceStrategyStatus string

// EvidenceStrategyDiagnosticCode defines a package boundary type.
type EvidenceStrategyDiagnosticCode string

const (
	// EvidenceStrategyStatusMissing defines a package boundary constant.
	EvidenceStrategyStatusMissing EvidenceStrategyStatus = "missing"
	// EvidenceStrategyStatusPlanned defines a package boundary constant.
	EvidenceStrategyStatusPlanned EvidenceStrategyStatus = "planned"
	// EvidenceStrategyStatusBlocked defines a package boundary constant.
	EvidenceStrategyStatusBlocked EvidenceStrategyStatus = "blocked"
	// EvidenceStrategyStatusComplete defines a package boundary constant.
	EvidenceStrategyStatusComplete EvidenceStrategyStatus = "complete"
)

const (
	// EvidenceStrategyDiagnosticMissing defines a package boundary constant.
	EvidenceStrategyDiagnosticMissing EvidenceStrategyDiagnosticCode = "evidence_strategy_missing"
	// EvidenceStrategyDiagnosticMalformed defines a package boundary constant.
	EvidenceStrategyDiagnosticMalformed EvidenceStrategyDiagnosticCode = "evidence_strategy_malformed"
	// EvidenceStrategyDiagnosticBlocked defines a package boundary constant.
	EvidenceStrategyDiagnosticBlocked EvidenceStrategyDiagnosticCode = "evidence_strategy_blocked"
	// EvidenceStrategyDiagnosticGherkinRequiredMissing defines a package boundary constant.
	EvidenceStrategyDiagnosticGherkinRequiredMissing EvidenceStrategyDiagnosticCode = "gherkin_required_missing"
	// EvidenceStrategyDiagnosticTestsRequiredMissing defines a package boundary constant.
	EvidenceStrategyDiagnosticTestsRequiredMissing EvidenceStrategyDiagnosticCode = "tests_required_missing"
	// EvidenceStrategyDiagnosticValidationMissing defines a package boundary constant.
	EvidenceStrategyDiagnosticValidationMissing EvidenceStrategyDiagnosticCode = "validation_missing"
	// EvidenceStrategyDiagnosticValidationFailed defines a package boundary constant.
	EvidenceStrategyDiagnosticValidationFailed EvidenceStrategyDiagnosticCode = "validation_failed"
	// EvidenceStrategyDiagnosticRuntimeUnavailable defines a package boundary constant.
	EvidenceStrategyDiagnosticRuntimeUnavailable EvidenceStrategyDiagnosticCode = "runtime_unavailable"
	// EvidenceStrategyDiagnosticCodexRuntimeUnavailable defines a package boundary constant.
	EvidenceStrategyDiagnosticCodexRuntimeUnavailable EvidenceStrategyDiagnosticCode = "codex_runtime_unavailable"
	// EvidenceStrategyDiagnosticSummaryMissing defines a package boundary constant.
	EvidenceStrategyDiagnosticSummaryMissing EvidenceStrategyDiagnosticCode = "evidence_summary_missing"
)
