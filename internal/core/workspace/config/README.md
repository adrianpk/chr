# Project configuration images and admission

Config owns the complete parsed project configuration image. App operations,
adapters and workflow operations consume it directly. Runtime, notification,
Web, evidence, harness and work settings retain their existing value owners and
encodings; workflow/state does not contain a copy of this aggregate.

Load observes the existing config.yml path and rejects a retained workflow
upgrade before reading policy. Parse keeps the same bounded parser, legacy and
version-one defaults, explicit-key tracking, validation order and errors. Default
selects legacy defaults without activating a project. Revision identifies the
complete raw image, including comments and unknown extension fields. Explicit
reports key presence without exposing the mutable parser map.

ResolveWorkflowConfiguration, EffectiveWorkflowConfiguration and WorkflowValue
retain the original effective values and project/version/legacy source labels.
Work state authority remains distinct from task status and execution mode.
WorkIntegrationConfig selects publication independently of review policy; its
publication strings and configured bounds remain unchanged.

ChangeHarnessSelections retains the configuration owner's cross-process lock,
authorization and preparation callbacks, exact revision and request identity,
bounded journal/checksum, image replacement and immutable receipt replay. A
pending selection intent fences workflow upgrades until its original request is
recovered. Lost-response replay returns the retained snapshot without redoing
preparation or rewriting a newer selection image.

The workspace parent owns path resolution and the configuration lock. Workflow
store upgrades, record transitions, worker admission and execution/publication
policy keep their operation owners. This package cannot import the root core or
workflow coordinator; the workspace parent cannot import this child.
