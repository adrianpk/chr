# Code Quality Review Detections

Baseline: `28c3db33` on `dev`, 2026-10-26.

Procedure: deterministic `go/ast` scan of every non-test `.go` file under
`cmd/` and `internal/`. Every function declaration and interface method was
inspected. The lists below are complete for this baseline. New code can add
entries until the enforcement in the report is active.

Parameter counts include `ctx` and exclude the method receiver.

## A. Functions With More Than 5 Parameters

Total: 121.

| Params | Location | Function |
| --- | --- | --- |
| 10 | `internal/core/telegram_work_projection.go:151` | `insertTelegramWorkNotice` |
| 10 | `internal/core/workflow_merger_driver.go:43` | `executeWorkflowMergerDriver` |
| 9 | `internal/app/ticket.go:984` | `failTicketLoopExecution` |
| 9 | `internal/core/telegram_document_http.go:23` | `TelegramReceiverAPI.SendDocument` |
| 9 | `internal/core/telegram_operation_transition.go:58` | `finishTelegramOperation` |
| 9 | `internal/core/telegram_receiver_action_http.go:35` | `TelegramReceiverAPI.sendAction` |
| 9 | `internal/core/workflow_execution_command.go:110` | `runWorkflowSupervisedProcess` |
| 9 | `internal/core/workflow_merger_claim_persist.go:14` | `WorkflowStore.persistWorkflowMergerClaim` |
| 9 | `internal/core/workflow_preparation_result.go:119` | `WorkflowStore.checkpointWorkflowPreparationFailure` |
| 8 | `internal/app/ticket.go:1020` | `finishTicketLoopExecution` |
| 8 | `internal/core/telegram_authoring_finish.go:22` | `TelegramReceiverStore.FinishTelegramAuthoring` |
| 8 | `internal/core/telegram_document_outbox.go:71` | `retainTelegramDocument` |
| 8 | `internal/core/telegram_guided_projection.go:92` | `bindTelegramGuidedNoticeDocument` |
| 8 | `internal/core/telegram_operation_transition.go:43` | `TelegramReceiverStore.FinishTelegramOperation` |
| 8 | `internal/core/telegram_receiver_action_http.go:107` | `TelegramReceiverAPI.SendActionPresentation` |
| 8 | `internal/core/workflow_closing_command.go:54` | `WorkflowStore.waitWorkflowClosingCommand` |
| 8 | `internal/core/workflow_merger_acceptance.go:82` | `WorkflowStore.commitWorkflowMergerRepair` |
| 8 | `internal/core/workflow_merger_acceptance_regenerate.go:41` | `WorkflowStore.runWorkflowMergerRegeneration` |
| 8 | `internal/core/workflow_preparation_claim.go:181` | `WorkflowStore.prepareWorkflowIntegrationGit` |
| 7 | `internal/app/service_harness_selection.go:64` | `resolveConfiguredHarnessSelection` |
| 7 | `internal/app/service_work_observation_legacy.go:80` | `service.observeLegacyContext` |
| 7 | `internal/app/service_work_observation_legacy.go:98` | `service.observeLegacyActivity` |
| 7 | `internal/app/ticket.go:854` | `runAndPrintTicketLoopExecution` |
| 7 | `internal/core/notification_telegram_delivery.go:29` | `sqliteNotificationEventRepository.FinishTelegramNotification` |
| 7 | `internal/core/quiz_openai_model.go:312` | `OpenAIQuizModelAdapter.recordOpenAIQuizResponse` |
| 7 | `internal/core/telegram_document_outbox.go:49` | `bindTelegramDocumentImage` |
| 7 | `internal/core/telegram_document_result.go:10` | `queueTelegramOperationResult` |
| 7 | `internal/core/telegram_guided_finish.go:18` | `TelegramReceiverStore.FinishTelegramGuided` |
| 7 | `internal/core/telegram_guided_projection.go:14` | `TelegramReceiverStore.ProjectTelegramGuided` |
| 7 | `internal/core/telegram_receiver_inbox.go:50` | `retainTelegramBatchUpdates` |
| 7 | `internal/core/telegram_receiver_outbox.go:57` | `queueTelegramReferencedResponse` |
| 7 | `internal/core/workflow_materialization_build.go:37` | `WorkflowStore.buildMaterializationForSet` |
| 7 | `internal/core/workflow_merger_acceptance_regenerate.go:11` | `WorkflowStore.regenerateWorkflowMerger` |
| 7 | `internal/core/workflow_merger_admission.go:33` | `WorkflowStore.checkWorkflowMergerAdmissionInputs` |
| 7 | `internal/core/workflow_merger_claim.go:34` | `WorkflowStore.claimWorkflowMerger` |
| 7 | `internal/core/workflow_merger_repair_verify.go:44` | `workflowIntegrationReader.verifyMergerTreeRepair` |
| 7 | `internal/core/workflow_preparation_claim.go:137` | `WorkflowStore.claimWorkflowIntegrationValidation` |
| 7 | `internal/core/workflow_preparation_result.go:82` | `WorkflowStore.finishWorkflowIntegrationValidation` |
| 7 | `internal/core/workflow_review_result.go:10` | `WorkflowStore.finishWorkflowReview` |
| 7 | `internal/core/workflow_review_result.go:40` | `WorkflowStore.acceptWorkflowReview` |
| 7 | `internal/core/workflow_submission_native_store.go:80` | `WorkflowStore.persistWorkflowNativeIntent` |
| 7 | `internal/tui/documents.go:73` | `appendDocumentView` |
| 7 | `internal/tui/shell.go:993` | `renderCompactTicketColumns` |
| 6 | `internal/app/agent_work_events.go:198` | `emitReviewStartedAgentEvent` |
| 6 | `internal/app/app.go:156` | `loadTUIOptions` |
| 6 | `internal/app/app.go:39` | `Run` |
| 6 | `internal/app/author.go:528` | `runAuthorContextualQuizNew` |
| 6 | `internal/app/author_assisted_dump.go:18` | `tryAuthorAssistedDump` |
| 6 | `internal/app/daemon.go:244` | `runDaemonForegroundLoop` |
| 6 | `internal/app/daemon.go:298` | `runDaemonForegroundIterations` |
| 6 | `internal/app/gate.go:232` | `buildGateDevMainReport` |
| 6 | `internal/app/gate.go:296` | `writeGateFailedEvidenceReport` |
| 6 | `internal/app/gate.go:314` | `writeGateFailedValidationRunReport` |
| 6 | `internal/app/notify.go:340` | `runNotifyTelegramTestDispatch` |
| 6 | `internal/app/service_harness_selection.go:84` | `prepareHarnessSelection` |
| 6 | `internal/app/telegram_confirmation_document.go:16` | `telegramReceiver.offerConfirmationDocument` |
| 6 | `internal/app/telegram_document_results.go:28` | `telegramDocumentForTarget` |
| 6 | `internal/app/telegram_harness_action.go:125` | `telegramReceiver.finishHarnessAction` |
| 6 | `internal/app/telegram_harness_prepare.go:19` | `telegramReceiver.prepareHarnessChange` |
| 6 | `internal/app/telegram_harness_prepare.go:43` | `telegramReceiver.reserveHarnessChange` |
| 6 | `internal/app/telegram_notification_dispatch.go:70` | `sendTelegramNotificationWithFallback` |
| 6 | `internal/app/telegram_prepared_work_dispatch.go:73` | `telegramReceiver.preparedWorkInput` |
| 6 | `internal/app/ticket.go:1054` | `executeLoopTicket` |
| 6 | `internal/app/ticket.go:1608` | `buildTicketStartPreflightForTicketWithOptions` |
| 6 | `internal/app/ticket.go:1616` | `buildLegacyTicketStartPreflight` |
| 6 | `internal/app/ticket.go:1718` | `prepareTicketStartWorktree` |
| 6 | `internal/app/ticket.go:1867` | `runTicketStartExecution` |
| 6 | `internal/app/ticket.go:924` | `runTicketLoopExecution` |
| 6 | `internal/core/execution_runner.go:333` | `ExecutionRunner.emitExecutionAgentEvent` |
| 6 | `internal/core/guided_use_case_progress.go:211` | `sealGuidedStage` |
| 6 | `internal/core/harness_generation_operation.go:124` | `prepareHarnessGenerationOperation` |
| 6 | `internal/core/harness_generation_operation.go:60` | `FreezeHarnessGeneration` |
| 6 | `internal/core/harness_generation_operation.go:99` | `freezeHarnessGenerationLocked` |
| 6 | `internal/core/plan_order_publication.go:14` | `preparePlanOrderChanges` |
| 6 | `internal/core/proving.go:379` | `runProvingRalphLoopAttempt` |
| 6 | `internal/core/remote_command_record.go:46` | `NewRejectedRemoteCommandRecord` |
| 6 | `internal/core/remote_command_record_repository.go:234` | `finishScannedRemoteCommandRecord` |
| 6 | `internal/core/telegram_action_presentation.go:105` | `telegramActionManifestValid` |
| 6 | `internal/core/telegram_document_http.go:72` | `telegramDocumentMultipartDestination` |
| 6 | `internal/core/telegram_document_outbox.go:24` | `queueTelegramDocument` |
| 6 | `internal/core/telegram_help_outbox.go:13` | `TelegramReceiverStore.QueueFormattedReadResponse` |
| 6 | `internal/core/telegram_presentation_outbox.go:26` | `queueTelegramPresentation` |
| 6 | `internal/core/telegram_receiver_action_http.go:20` | `TelegramReceiverAPI.SendAction` |
| 6 | `internal/core/telegram_receiver_http.go:120` | `TelegramReceiverAPI.callBody` |
| 6 | `internal/core/telegram_receiver_http.go:128` | `TelegramReceiverAPI.postBody` |
| 6 | `internal/core/telegram_receiver_inbox.go:75` | `retainTelegramUpdate` |
| 6 | `internal/core/telegram_receiver_outbox.go:38` | `TelegramReceiverStore.QueueReferencedResponse` |
| 6 | `internal/core/telegram_receiver_outbox.go:53` | `queueTelegramResponse` |
| 6 | `internal/core/telegram_receiver_outbox.go:61` | `queueTelegramResponseParts` |
| 6 | `internal/core/telegram_work_projection.go:102` | `queueTelegramWorkObservation` |
| 6 | `internal/core/telegram_work_projection.go:39` | `projectTelegramWork` |
| 6 | `internal/core/telegram_work_status.go:77` | `TelegramReceiverStore.QueueTelegramWorkStatus` |
| 6 | `internal/core/ticket_index.go:490` | `assignIndexedTicketTimes` |
| 6 | `internal/core/ticket_review_command.go:143` | `ticketReviewCommandPlan` |
| 6 | `internal/core/workflow_decision_control.go:64` | `WorkflowStore.applyWorkflowDecisionControl` |
| 6 | `internal/core/workflow_execution_prepare.go:50` | `WorkflowStore.prepareWorkflowAttempt` |
| 6 | `internal/core/workflow_frontier_observe.go:38` | `WorkflowStore.workflowFrontierReceiptProof` |
| 6 | `internal/core/workflow_frontier_observe.go:57` | `observeWorkflowFrontierProof` |
| 6 | `internal/core/workflow_integration_decision_answer.go:123` | `WorkflowStore.checkWorkflowFreshAnswer` |
| 6 | `internal/core/workflow_integration_entry_store.go:107` | `WorkflowStore.admitWorkflowIntegrationUnit` |
| 6 | `internal/core/workflow_integration_replay.go:149` | `validWorkflowReplayMapping` |
| 6 | `internal/core/workflow_integration_snapshot.go:118` | `WorkflowStore.appendWorkflowIntegrationRunEntries` |
| 6 | `internal/core/workflow_invocation.go:397` | `WorkflowStore.createWorkflowInvocation` |
| 6 | `internal/core/workflow_materialization_build.go:102` | `WorkflowStore.materializationTask` |
| 6 | `internal/core/workflow_materialization_build.go:71` | `WorkflowStore.buildMaterializationTasks` |
| 6 | `internal/core/workflow_merger_acceptance.go:44` | `WorkflowStore.acceptWorkflowMerger` |
| 6 | `internal/core/workflow_merger_acceptance_observe.go:9` | `WorkflowStore.observeWorkflowMergerRepair` |
| 6 | `internal/core/workflow_merger_acceptance_tree.go:79` | `workflowIntegrationReader.checkMergerObservedPaths` |
| 6 | `internal/core/workflow_merger_finish.go:10` | `WorkflowStore.finishWorkflowMerger` |
| 6 | `internal/core/workflow_preparation_replay.go:26` | `workflowIntegrationReader.attributedPreparationCommit` |
| 6 | `internal/core/workflow_preparation_replay.go:38` | `workflowIntegrationReader.preparationObject` |
| 6 | `internal/core/workflow_preparation_result.go:134` | `WorkflowStore.retainWorkflowPreparationFailureRepair` |
| 6 | `internal/core/workflow_preparation_result.go:169` | `WorkflowStore.checkWorkflowPreparationCompletion` |
| 6 | `internal/core/workflow_preparation_result.go:8` | `WorkflowStore.commitWorkflowPreparationInput` |
| 6 | `internal/core/workflow_review_prepare.go:119` | `WorkflowStore.claimWorkflowReview` |
| 6 | `internal/core/workflow_review_prepare.go:98` | `WorkflowStore.prepareWorkflowReviewPolicy` |
| 6 | `internal/core/workflow_submission_native_store.go:262` | `checkWorkflowNativeSubmissionCallback` |
| 6 | `internal/core/workflow_submission_native_store.go:55` | `WorkflowStore.authorizeWorkflowNativeIntent` |
| 6 | `internal/core/workspace_init.go:189` | `workspace.initWorkspaceFiles` |
| 6 | `internal/harness/process/exchange.go:43` | `exchangeCommand` |
| 6 | `internal/tui/observability.go:81` | `NewObservabilityView` |

## B. Functions With Exactly 5 Parameters

Total: 230. These meet the hard limit but exceed the preferred
target. Treat them as a watch list; fix them when the file is touched.

| Location | Function |
| --- | --- |
| `internal/app/agent_work_events.go:181` | `emitCycleRehydratedAgentEvent` |
| `internal/app/agent_work_events.go:210` | `emitReviewFinishedAgentEvent` |
| `internal/app/agent_work_events.go:234` | `emitRepairFinishedAgentEvent` |
| `internal/app/agent_work_events.go:247` | `emitMergeStartedAgentEvent` |
| `internal/app/agent_work_events.go:259` | `emitMergeFinishedAgentEvent` |
| `internal/app/app.go:450` | `loadTUIAgentEvents` |
| `internal/app/author.go:1038` | `resolveAuthoringSessionForPromotion` |
| `internal/app/gate.go:170` | `runGateDevMainForProject` |
| `internal/app/notification_emit.go:109` | `emitWorkflowNotification` |
| `internal/app/notification_emit.go:48` | `emitLoopRunNotification` |
| `internal/app/service_acceptance_feature.go:148` | `service.acceptanceFeatureNonGeneratedResult` |
| `internal/app/service_acceptance_feature.go:168` | `service.acceptanceFeatureRoundPath` |
| `internal/app/service_acceptance_feature.go:269` | `service.writeAcceptanceFeatureRoundNotes` |
| `internal/app/service_dump.go:229` | `renderRawDumpMarkdown` |
| `internal/app/service_dump_curation_openai.go:298` | `appendDumpCurationFileContextSection` |
| `internal/app/service_dump_curation_openai.go:454` | `openAIDumpCurationProvider.recordResponse` |
| `internal/app/service_harness_generation.go:243` | `service.materializeBoundPlanTracking` |
| `internal/app/service_plan.go:379` | `service.generatePlanSliceDetail` |
| `internal/app/service_plan.go:797` | `service.newPlanningModelRequest` |
| `internal/app/service_quiz.go:728` | `appendQuizFileContextSection` |
| `internal/app/service_work.go:734` | `parseAgenticReviewTicket` |
| `internal/app/service_work.go:743` | `writeAgenticReviewNonBlockingComments` |
| `internal/app/service_work.go:800` | `service.failAgentWorkIntegrationCycles` |
| `internal/app/service_work_authority.go:111` | `service.startNativeAgentWork` |
| `internal/app/service_workflow_admission.go:20` | `service.admitBoundWorkExtent` |
| `internal/app/service_workflow_controls.go:123` | `service.continueWorkExtent` |
| `internal/app/service_workflow_selection.go:94` | `service.StartBoundWorkExtent` |
| `internal/app/telegram.go:127` | `storeTelegramSetupToken` |
| `internal/app/telegram_authoring_input.go:43` | `telegramReceiver.authoringOwnerInput` |
| `internal/app/telegram_control_action.go:59` | `telegramReceiver.completeControlAction` |
| `internal/app/telegram_control_action.go:91` | `telegramReceiver.finishControlAction` |
| `internal/app/telegram_control_dispatch.go:130` | `telegramReceiver.respondNativeStatus` |
| `internal/app/telegram_control_service.go:42` | `service.ApplyTelegramControl` |
| `internal/app/telegram_doctor.go:67` | `checkTelegramConnection` |
| `internal/app/telegram_document_read.go:11` | `telegramReceiver.finishOriginalDocument` |
| `internal/app/telegram_guided_ack.go:16` | `telegramReceiver.finishOvertakenGuided` |
| `internal/app/telegram_guided_apply.go:61` | `telegramReceiver.finishTelegramGuided` |
| `internal/app/telegram_guided_dispatch.go:49` | `telegramReceiver.routeTelegramGuided` |
| `internal/app/telegram_guided_dispatch.go:99` | `telegramReceiver.reserveTelegramGuided` |
| `internal/app/telegram_notification_dispatch.go:12` | `createTelegramNotification` |
| `internal/app/telegram_notification_dispatch.go:34` | `dispatchTelegramNotification` |
| `internal/app/telegram_prepared_work_action.go:42` | `telegramReceiver.dispatchPreparedWorkAction` |
| `internal/app/telegram_prepared_work_apply.go:52` | `telegramReceiver.applyPreparedWork` |
| `internal/app/telegram_record_apply.go:114` | `telegramReceiver.finishRecordFailure` |
| `internal/app/telegram_work_failure.go:28` | `telegramReceiver.finishWorkFailure` |
| `internal/app/telegram_work_service.go:21` | `service.ResumeTelegramWork` |
| `internal/app/ticket.go:1332` | `applyTicketResult` |
| `internal/app/ticket.go:1407` | `refreshAppliedTicketIndex` |
| `internal/app/ticket.go:1596` | `buildTicketStartPreflightForTicket` |
| `internal/app/ticket.go:1679` | `prepareTicketStartWorkspace` |
| `internal/app/ticket.go:1943` | `printTicketStartResult` |
| `internal/app/ticket.go:686` | `selectTicketLoopTickets` |
| `internal/app/ticket.go:876` | `printTicketLoopPlan` |
| `internal/app/web_guided_result.go:61` | `guidedWebValidation` |
| `internal/app/work_workflow_cli.go:254` | `performWorkflowControl` |
| `internal/app/work_workflow_integration_cli.go:76` | `performWorkflowOperatorControl` |
| `internal/app/work_workflow_integration_cli.go:87` | `performWorkflowIntegrationControl` |
| `internal/app/work_workflow_render.go:36` | `workWorkflowCLI.status` |
| `internal/core/agent_run.go:62` | `recordAgentRunResponseEvent` |
| `internal/core/agent_run.go:89` | `RecordAgentRunResponseEvent` |
| `internal/core/assisted_answer_curation.go:242` | `updateAssistedAnswerCurationSession` |
| `internal/core/assisted_quiz_loop.go:185` | `assistedQuizModelRequest` |
| `internal/core/assisted_quiz_loop.go:318` | `assistedQuizStopTurn` |
| `internal/core/assisted_quiz_loop.go:352` | `assistedQuizTurnFromModelResponse` |
| `internal/core/authoring_contextual_quiz.go:184` | `NewAuthoringQuizGenerationRequest` |
| `internal/core/authoring_continue.go:39` | `fileAuthoringSessionRepository.continueAuthoring` |
| `internal/core/authoring_dump.go:12` | `fileAuthoringSessionRepository.prepareAuthoringDump` |
| `internal/core/authoring_dump.go:128` | `authoringDumpPreparation` |
| `internal/core/authoring_dump.go:172` | `authoringDumpBinding` |
| `internal/core/authoring_dump_snapshot.go:109` | `appendAuthoringDumpObservation` |
| `internal/core/authoring_prepare.go:186` | `fileAuthoringSessionRepository.prepareAuthoringOutput` |
| `internal/core/authoring_prepare.go:217` | `fileAuthoringSessionRepository.prepareAuthoringOutputChanges` |
| `internal/core/authoring_prepare.go:47` | `fileAuthoringSessionRepository.prepareAuthoringContent` |
| `internal/core/config_workflow_parse.go:117` | `guardWorkflowConfigRoot` |
| `internal/core/config_workflow_parse.go:133` | `guardWorkflowConfigScalar` |
| `internal/core/config_workflow_parse.go:97` | `guardWorkflowConfigLine` |
| `internal/core/context_gograph.go:201` | `runGoGraphBuild` |
| `internal/core/context_gograph.go:216` | `runGoGraphQueries` |
| `internal/core/context_gograph.go:241` | `symbolGoGraphQueries` |
| `internal/core/execution_runner.go:189` | `ExecutionRunner.runHarness` |
| `internal/core/gate_config.go:159` | `setGateConfigValue` |
| `internal/core/gate_report.go:117` | `NewGateReport` |
| `internal/core/gate_report.go:148` | `NewFailedGateReport` |
| `internal/core/guided_recovery.go:19` | `RecoverGuidedHostPage` |
| `internal/core/guided_use_case_git.go:42` | `fileAuthoringSessionRepository.observeGuidedCheckpoint` |
| `internal/core/harness_invocation.go:158` | `harnessWorkerCommand` |
| `internal/core/harness_invocation.go:86` | `harnessRunCommand` |
| `internal/core/notification_telegram_read.go:124` | `sqliteNotificationEventRepository.readTelegramNotificationDocument` |
| `internal/core/plan_order_publication.go:114` | `fileAuthoringSessionRepository.planCurrentTrackerChanges` |
| `internal/core/plan_order_publication.go:83` | `fileAuthoringSessionRepository.planRetainedImages` |
| `internal/core/project_dump_artifact.go:93` | `projectDumpArtifact` |
| `internal/core/quiz_action_commit.go:91` | `fileAuthoringSessionRepository.commitQuizTransition` |
| `internal/core/review_provider_diagnostics.go:177` | `ReviewProviderDiagnostic.addProviderChecks` |
| `internal/core/review_provider_diagnostics.go:258` | `ReviewProviderDiagnostic.checkProviderPullRequestRead` |
| `internal/core/review_workflow_native_capability.go:137` | `RemoteReviewAdapter.RecheckWorkflowNativeCapability` |
| `internal/core/review_workflow_native_capability.go:77` | `RemoteReviewAdapter.finishWorkflowNativeCapability` |
| `internal/core/review_workflow_native_policy.go:69` | `workflowNativeProtectionReasons` |
| `internal/core/review_workflow_pull_request.go:280` | `workflowRemoteAPIArgs` |
| `internal/core/review_workflow_pull_request.go:91` | `RemoteReviewAdapter.readWorkflowPRPage` |
| `internal/core/spec_generation_publication.go:105` | `prepareGeneratedSpecDocument` |
| `internal/core/spec_generation_publication.go:52` | `publishSpecContent` |
| `internal/core/telegram_action_finish.go:114` | `finishTelegramActionResult` |
| `internal/core/telegram_action_finish.go:142` | `finishTelegramActionResponse` |
| `internal/core/telegram_action_finish.go:191` | `queueTelegramActionResultResponse` |
| `internal/core/telegram_action_finish.go:78` | `finishTelegramActionRequests` |
| `internal/core/telegram_action_finish.go:94` | `finishTelegramActionState` |
| `internal/core/telegram_action_presentation.go:10` | `queueTelegramActionView` |
| `internal/core/telegram_action_presentation.go:78` | `validateTelegramActionPresentation` |
| `internal/core/telegram_authoring_finish.go:53` | `updateTelegramAuthoringConversation` |
| `internal/core/telegram_authoring_work_request.go:62` | `projectTelegramAuthoringWorkPhase` |
| `internal/core/telegram_conversation_question.go:37` | `TelegramReceiverStore.BindTelegramConversationQuestion` |
| `internal/core/telegram_conversation_reply.go:9` | `retainTelegramReply` |
| `internal/core/telegram_document_http.go:68` | `telegramDocumentMultipart` |
| `internal/core/telegram_document_outbox.go:13` | `TelegramReceiverStore.QueueDocumentResponse` |
| `internal/core/telegram_guided_finish.go:45` | `updateTelegramGuidedBinding` |
| `internal/core/telegram_guided_finish.go:89` | `projectTelegramGuidedAction` |
| `internal/core/telegram_harness_action.go:124` | `finishTelegramHarnessConversation` |
| `internal/core/telegram_harness_action.go:15` | `TelegramReceiverStore.ReserveTelegramHarnessOperation` |
| `internal/core/telegram_intake_reserve.go:11` | `rejectTelegramIntakeOverflow` |
| `internal/core/telegram_interpretation.go:24` | `TelegramReceiverStore.ResolveTelegramInterpretation` |
| `internal/core/telegram_notification_delivery.go:44` | `TelegramReceiverAPI.sendNotificationDocument` |
| `internal/core/telegram_notification_delivery.go:58` | `TelegramReceiverAPI.sendNotificationText` |
| `internal/core/telegram_operation_model.go:25` | `TelegramReceiverStore.ResolveTelegramOperationModel` |
| `internal/core/telegram_operation_request.go:183` | `reservePreparedTelegramOperation` |
| `internal/core/telegram_presentation_outbox.go:15` | `TelegramReceiverStore.QueuePresentationResponse` |
| `internal/core/telegram_receiver_action_http.go:31` | `TelegramReceiverAPI.SendPresentation` |
| `internal/core/telegram_receiver_http.go:111` | `TelegramReceiverAPI.call` |
| `internal/core/telegram_receiver_outbox.go:217` | `TelegramReceiverStore.deliverySent` |
| `internal/core/telegram_receiver_outbox.go:33` | `TelegramReceiverStore.QueueReadResponse` |
| `internal/core/telegram_transport_retry.go:23` | `RetryLocalTelegramTransport` |
| `internal/core/telegram_work_binding.go:44` | `projectTelegramWorkBinding` |
| `internal/core/telegram_work_binding.go:56` | `projectTelegramAdmissionBinding` |
| `internal/core/telegram_work_binding.go:80` | `projectTelegramControlBinding` |
| `internal/core/telegram_work_projection.go:25` | `TelegramReceiverStore.ProjectTelegramWork` |
| `internal/core/telegram_work_target.go:22` | `TelegramReceiverStore.resolveTelegramActionTarget` |
| `internal/core/ticket_document_create.go:77` | `createNativeTicketDocument` |
| `internal/core/ticket_edit.go:223` | `writeNativeTicketEdit` |
| `internal/core/ticket_parser.go:197` | `indexedTicketFromMetadata` |
| `internal/core/work_selection_preflight.go:486` | `workBranchPolicyRemoteBlockReason` |
| `internal/core/workflow_agent_command.go:123` | `WorkflowStore.workflowAgentSettlement` |
| `internal/core/workflow_agent_command.go:88` | `WorkflowStore.settleWorkflowAgentCall` |
| `internal/core/workflow_amendment_build.go:111` | `WorkflowStore.buildWorkflowAmendedTask` |
| `internal/core/workflow_amendment_build.go:135` | `WorkflowStore.newWorkflowAmendmentTask` |
| `internal/core/workflow_amendment_persist.go:98` | `WorkflowStore.workflowAmendmentInvocationFiles` |
| `internal/core/workflow_candidate_repair_admission.go:27` | `WorkflowStore.checkWorkflowCandidateRepairAdmission` |
| `internal/core/workflow_candidate_repair_retain.go:5` | `WorkflowStore.retainWorkflowCandidateRepair` |
| `internal/core/workflow_checkpoint.go:20` | `WorkflowStore.checkpointWorkflowAttempt` |
| `internal/core/workflow_closing_begin.go:83` | `WorkflowStore.prepareWorkflowClosing` |
| `internal/core/workflow_closing_command.go:77` | `WorkflowStore.registerWorkflowClosingProcess` |
| `internal/core/workflow_closing_finish.go:10` | `WorkflowStore.finishWorkflowClosing` |
| `internal/core/workflow_closing_proof.go:89` | `WorkflowStore.observeWorkflowClosingSlice` |
| `internal/core/workflow_closing_proof.go:97` | `WorkflowStore.observeWorkflowClosingLegacySlice` |
| `internal/core/workflow_continuation.go:83` | `WorkflowStore.finishWorkflowContinuation` |
| `internal/core/workflow_continuation_steps.go:8` | `WorkflowStore.continueWorkflowSteps` |
| `internal/core/workflow_control_authority.go:6` | `WorkflowStore.StopWorkflowGuarded` |
| `internal/core/workflow_execution_prepare.go:26` | `WorkflowStore.prepareWorkflowTaskHarness` |
| `internal/core/workflow_execution_result.go:11` | `WorkflowStore.finishWorkflowTask` |
| `internal/core/workflow_execution_reuse.go:124` | `WorkflowStore.workflowRetainedResultMatches` |
| `internal/core/workflow_execution_reuse.go:87` | `WorkflowStore.checkWorkflowReusableTaskPrepared` |
| `internal/core/workflow_frontier_observe.go:8` | `WorkflowStore.observeWorkflowFrontier` |
| `internal/core/workflow_integration.go:72` | `WorkflowStore.VerifyWorkflowSliceIntegration` |
| `internal/core/workflow_integration.go:94` | `WorkflowStore.verifyWorkflowSliceIntegration` |
| `internal/core/workflow_integration_budget_reserve.go:46` | `WorkflowStore.checkWorkflowIntegrationReservationAdmission` |
| `internal/core/workflow_integration_decision_answer.go:10` | `WorkflowStore.answerWorkflowDecision` |
| `internal/core/workflow_integration_decision_answer.go:63` | `WorkflowStore.checkWorkflowDecisionCurrent` |
| `internal/core/workflow_integration_decision_store.go:241` | `WorkflowStore.workflowDecisionRequestRetry` |
| `internal/core/workflow_integration_decision_store.go:74` | `WorkflowStore.checkWorkflowDecisionAdmission` |
| `internal/core/workflow_integration_entry_transition.go:220` | `WorkflowStore.checkWorkflowIntegrationFinalization` |
| `internal/core/workflow_integration_status.go:106` | `appendWorkflowIntegrationAttemptOwner` |
| `internal/core/workflow_invocation.go:304` | `WorkflowStore.startNewWorkflow` |
| `internal/core/workflow_invocation.go:444` | `newWorkflowInvocation` |
| `internal/core/workflow_lifecycle_control.go:58` | `WorkflowStore.acceptWorkflowLifecycleControl` |
| `internal/core/workflow_lifecycle_control.go:95` | `WorkflowStore.applyWorkflowLifecycleControl` |
| `internal/core/workflow_materialization.go:78` | `WorkflowStore.materializeNewPlan` |
| `internal/core/workflow_merger_acceptance.go:26` | `WorkflowStore.ExecuteWorkflowMergerRepair` |
| `internal/core/workflow_merger_acceptance_regenerate.go:84` | `WorkflowStore.runWorkflowMergerProcedure` |
| `internal/core/workflow_merger_acceptance_tree.go:163` | `workflowIntegrationReader.checkMergerCopiedSide` |
| `internal/core/workflow_merger_admission.go:11` | `WorkflowStore.workflowMergerAdmission` |
| `internal/core/workflow_merger_driver.go:19` | `runWorkflowMergerDriver` |
| `internal/core/workflow_merger_driver.go:90` | `exchangeWorkflowMerger` |
| `internal/core/workflow_merger_execution.go:18` | `WorkflowStore.executeWorkflowMerger` |
| `internal/core/workflow_merger_execution.go:69` | `WorkflowStore.runWorkflowMerger` |
| `internal/core/workflow_merger_finish.go:46` | `WorkflowStore.persistWorkflowMergerFinish` |
| `internal/core/workflow_merger_repair_plan.go:11` | `PlanWorkflowMergerRepair` |
| `internal/core/workflow_merger_repair_regeneration.go:8` | `importWorkflowMergerRepairTrees` |
| `internal/core/workflow_merger_repair_validate.go:95` | `validateWorkflowRepairPosition` |
| `internal/core/workflow_merger_repair_verify.go:9` | `workflowIntegrationReader.expectedIntegrationTree` |
| `internal/core/workflow_native_finalization.go:74` | `WorkflowStore.commitWorkflowNativeFanout` |
| `internal/core/workflow_native_observation.go:54` | `WorkflowStore.observeWorkflowNativeSubmission` |
| `internal/core/workflow_native_observation_store.go:93` | `WorkflowStore.recordWorkflowNativeCanonicalObservation` |
| `internal/core/workflow_native_reconcile_git.go:76` | `observeWorkflowNativeGit` |
| `internal/core/workflow_native_stop.go:44` | `WorkflowStore.retainWorkflowNativeStop` |
| `internal/core/workflow_preparation_claim.go:265` | `WorkflowStore.workflowPreparationQuery` |
| `internal/core/workflow_preparation_claim.go:87` | `WorkflowStore.cachedWorkflowPreparation` |
| `internal/core/workflow_preparation_drift.go:38` | `WorkflowStore.reprepareWorkflowNativeInputs` |
| `internal/core/workflow_preparation_drift.go:65` | `workflowNativePreparationChanged` |
| `internal/core/workflow_preparation_drift_replace.go:5` | `WorkflowStore.replaceWorkflowNativePreparation` |
| `internal/core/workflow_preparation_native.go:165` | `WorkflowStore.cachedWorkflowPreparationCapability` |
| `internal/core/workflow_process_record.go:12` | `WorkflowStore.registerWorkflowProcess` |
| `internal/core/workflow_process_record.go:53` | `WorkflowStore.finishWorkflowProcess` |
| `internal/core/workflow_process_supervisor.go:44` | `workflowSupervisedCommand` |
| `internal/core/workflow_publication_native.go:82` | `WorkflowStore.continueWorkflowNativePhase` |
| `internal/core/workflow_publication_reconcile.go:47` | `WorkflowStore.reconcileWorkflowPublishedSubmission` |
| `internal/core/workflow_review.go:129` | `WorkflowStore.ApproveWorkflowSlice` |
| `internal/core/workflow_review.go:82` | `WorkflowStore.reviewWorkflowSlice` |
| `internal/core/workflow_review_inputs.go:146` | `WorkflowStore.approveWorkflowSlice` |
| `internal/core/workflow_review_native.go:29` | `WorkflowStore.bindWorkflowNativeReview` |
| `internal/core/workflow_review_native.go:37` | `WorkflowStore.bindWorkflowNativeReviewCapability` |
| `internal/core/workflow_review_persist.go:46` | `WorkflowStore.persistWorkflowReviewDecision` |
| `internal/core/workflow_review_prepare.go:11` | `WorkflowStore.prepareWorkflowReview` |
| `internal/core/workflow_review_validate.go:166` | `WorkflowStore.readWorkflowReviewEvidence` |
| `internal/core/workflow_slice_result.go:10` | `WorkflowStore.finishWorkflowSliceValidation` |
| `internal/core/workflow_source_git.go:46` | `workflowIntegrationReader.pushWorkflowSource` |
| `internal/core/workflow_stop.go:20` | `WorkflowStore.stopWorkflowChecked` |
| `internal/core/workflow_submission_observe.go:83` | `observeWorkflowSubmittedDelivery` |
| `internal/core/workflow_submission_prepare.go:103` | `WorkflowStore.observeWorkflowSubmissionRetry` |
| `internal/core/workflow_submission_prepare.go:78` | `WorkflowStore.checkWorkflowSubmissionRetry` |
| `internal/core/workflow_submission_response.go:8` | `WorkflowStore.retainWorkflowSubmissionResponse` |
| `internal/core/workflow_upgrade_config.go:120` | `replaceWorkflowUpgradeValue` |
| `internal/core/workspace_context.go:328` | `missingWorkspaceContextSource` |
| `internal/core/workspace_init.go:224` | `workspace.initRootFiles` |
| `internal/core/workspace_semantic_digest.go:223` | `NewWorkspaceSemanticDigest` |
| `internal/harness/pi/session.go:62` | `session.exchange` |
| `internal/harness/pi/worker_connection.go:22` | `exchangeWorkerInput` |
| `internal/harness/pi/worker_resources.go:171` | `WithWorkerCredential` |
| `internal/harness/pi/worker_resources.go:253` | `publishRefreshedCredential` |
| `internal/harness/pi/worker_resources.go:26` | `ResolveWorkerResources` |
| `internal/harness/pi/worker_resources.go:266` | `atomicWorkerCredential` |
| `internal/tui/shell.go:1345` | `renderPanel` |
| `internal/tui/shell.go:1358` | `Model.renderDetailPanel` |

## C. Result Shapes Outside the Policy

Allowed shapes: no result, one result, `(T, error)` and `(T, bool)`.
Total outside the policy: 420.

### Shape Summary

| Count | Shape |
| --- | --- |
| 32 | `(tea.Model, tea.Cmd)` |
| 16 | `(string, string, error)` |
| 11 | `(string, *TelegramConversation, error)` |
| 10 | `(WorkflowSubmission, bool, error)` |
| 9 | `(WorkReadinessAppliedRepair, bool, error)` |
| 6 | `(*WorkflowClaim, workflowPreparedTask, error)` |
| 6 | `(tea.Model, tea.Cmd, bool)` |
| 5 | `([]ContextDocument, []ContextSourceInventoryEntry, error)` |
| 5 | `(WorkflowInvocation, bool, error)` |
| 5 | `(TelegramConversation, TelegramOperationTarget, error)` |
| 5 | `(string, string)` |
| 5 | `(int, int)` |
| 4 | `(core.HarnessGenerationBinding, core.HarnessGenerationLaunch, error)` |
| 4 | `(core.TelegramOperationTarget, bool, error)` |
| 4 | `(core.TelegramConversationInput, bool, error)` |
| 4 | `(webui.WriteActionResponse, bool, error)` |
| 4 | `(string, bool, error)` |
| 4 | `(WorkflowMaterializationResult, bool, error)` |
| 4 | `(TelegramActionAnswer, bool, error)` |
| 4 | `(json.RawMessage, *WorkflowMergerUsage, error)` |
| 3 | `(core.GeneratedPlan, []string, []string, error)` |
| 3 | `(ticketLoopExecutionResult, []core.IndexedTicket, int, error)` |
| 3 | `(WorkflowWorkStatus, bool, error)` |
| 3 | `([]byte, string, error)` |
| 3 | `(TelegramProcessingLane, bool, error)` |
| 3 | `(int, bool, error)` |
| 3 | `(workflowInvocationFile, *WorkflowStore, error)` |
| 3 | `(*WorkflowRecord, bool, error)` |
| 3 | `(WorkflowTrackerStatus, []byte, error)` |
| 2 | `([]byte, []WorkflowNativeUnsupported, string, error)` |
| 2 | `(service, commandProjectContext, error)` |
| 2 | `(service, context.Context, error)` |
| 2 | `(*core.HarnessBinding, []string, error)` |
| 2 | `(string, string, string)` |
| 2 | `(telegramAuthoringCommand, bool, error)` |
| 2 | `(telegramControlCommand, bool, error)` |
| 2 | `(telegramExplanationCommand, bool, error)` |
| 2 | `(telegramRecordCommand, bool, error)` |
| 2 | `(telegramWorkCommand, bool, error)` |
| 2 | `(ContextDocument, ContextSourceInventoryEntry, error)` |
| 2 | `(AuthoringArtifactRef, []workflowFileChange, error)` |
| 2 | `(ArchivedPlanPaths, bool, error)` |
| 2 | `([]workflowFileChange, ProjectDumpMutationResult, error)` |
| 2 | `(*os.File, *os.File, error)` |
| 2 | `(TelegramNotificationDelivery, bool, error)` |
| 2 | `([]byte, []byte, error)` |
| 2 | `(PlanPublicationResult, bool, error)` |
| 2 | `(string, string, bool)` |
| 2 | `([]byte, openAIQuizExecutionResult, error)` |
| 2 | `(ReviewProviderReference, ReviewProviderTargetConfig, error)` |
| 2 | `(string, []byte, error)` |
| 2 | `(TelegramOperationRequest, bool, error)` |
| 2 | `(TelegramConversation, bool, error)` |
| 2 | `(*WorkflowMergerRepairState, []WorkflowEvidenceRef, error)` |
| 2 | `(WorkflowProcess, bool, error)` |
| 2 | `(workflowInvocationFile, WorkflowRecord, error)` |
| 2 | `(WorkflowSliceDeliveryProof, string, error)` |
| 2 | `(int, string, error)` |
| 2 | `(WorkflowControlResult, bool, error)` |
| 2 | `(WorkflowIntegrationEntry, WorkflowIntegrationBudget, error)` |
| 2 | `(WorkflowMutationResult, bool, error)` |
| 2 | `(WorkflowSubmission, WorkflowGitPublication, error)` |
| 2 | `(workflowSliceObservation, WorkflowReviewInput, error)` |
| 2 | `(workflowUpgradeJournal, []byte, error)` |
| 2 | `(core.HarnessReadiness, string)` |
| 1 | `(string, string, string, string, string, string, error)` |
| 1 | `(*WorkflowClaim, workflowPreparedTask, WorkflowPreparationInput, int, *WorkflowIntegrationCandidate, error)` |
| 1 | `([]core.QuizModelContextSection, core.QuizRuntimeState, []core.QuizEntry, core.GuidedStageCheckpoint, error)` |
| 1 | `(core.GeneratedPlanOutline, []string, []string, []string, error)` |
| 1 | `(string, string, string, string, bool)` |
| 1 | `(string, []GateGitCommit, []GateGitChangedFile, string, error)` |
| 1 | `(*time.Time, *time.Time, *time.Time, int, error)` |
| 1 | `(core.WorkspaceGraph, core.WorkspaceIndex, []core.WorkspaceSemanticDigestState, error)` |
| 1 | `(core.UserEnvironment, commandProjectContext, core.WorkspaceConfig, error)` |
| 1 | `([]string, string, string, error)` |
| 1 | `(core.PlanSourceObservation, core.QuizRuntimeState, string, error)` |
| 1 | `(core.GeneratedPlanSliceDetail, string, string, error)` |
| 1 | `(core.IndexedTicket, ticketStartPreflight, func(), error)` |
| 1 | `(core.IndexedTicket, core.ExecutionRecord, bool, error)` |
| 1 | `(*core.WorkflowStore, core.WorkflowPublicationProvider, core.WorkflowInvocation, error)` |
| 1 | `(string, string, string, bool)` |
| 1 | `(string, string, string, error)` |
| 1 | `([]byte, []byte, []workflowFileChange, error)` |
| 1 | `(ProvingRalphLoopAttempt, *ProvingValidationResult, string, string)` |
| 1 | `(string, string, int, error)` |
| 1 | `(OpsLayout, string, bool, error)` |
| 1 | `(string, string, time.Time, error)` |
| 1 | `(workflowPreparedTask, workflowFileChange, workflowFileChange, error)` |
| 1 | `(WorkflowIntegrationEntry, WorkflowIntegrationBudget, string, error)` |
| 1 | `(workflowInvocationFile, *WorkflowStore, WorkflowNativePublisherProfile, error)` |
| 1 | `(WorkflowSubmission, WorkflowNativePublicationObservation, WorkflowIntegrationBudgetGuard, error)` |
| 1 | `(WorkflowSubmission, WorkflowGitPublication, WorkflowNativeSubmissionUpdate, error)` |
| 1 | `(map[string]any, workflowFileChange, bool, error)` |
| 1 | `(*WorkflowClaim, workflowPreparedTask, WorkflowPreparationInput, error)` |
| 1 | `(workflowSliceObservation, WorkflowReviewInput, WorkflowPublicationCapability, error)` |
| 1 | `(*WorkflowClaim, workflowPreparedTask, WorkflowReviewBinding, error)` |
| 1 | `(*WorkflowClaim, workflowPreparedTask, *WorkflowSliceCandidate, error)` |
| 1 | `(WorkflowTrackerStatus, []byte, WorkflowTrackerStatus, error)` |
| 1 | `([]byte, []byte, string, error)` |
| 1 | `(string, context.Context, context.CancelFunc, error)` |
| 1 | `(int, int, int)` |
| 1 | `(core.AuthoringQuizQuestionProvider, context.Context, error)` |
| 1 | `(core.PlanSourceObservation, core.QuizRuntimeState, error)` |
| 1 | `(core.QuizModelRequest, string, error)` |
| 1 | `(core.QuizRuntimeState, core.GuidedStageCheckpoint, error)` |
| 1 | `(PlanView, bool, error)` |
| 1 | `(core.GeneratedWorkflowPlan, core.OpsLayout, error)` |
| 1 | `(PlanOperationResult, bool, error)` |
| 1 | `(agentWorkIntegrationCycleState, bool, error)` |
| 1 | `(core.IndexedTicket, bool, error)` |
| 1 | `(*core.WorkflowStore, core.WorkflowPublicationGuardProvider, error)` |
| 1 | `(*core.WorkflowStore, core.WorkflowPullRequestProvider, error)` |
| 1 | `(core.WorkflowSelector, core.WorkflowSelectionGuard, error)` |
| 1 | `(core.TelegramConversationInput, bool, string)` |
| 1 | `(telegramGuidedInput, core.TelegramReceiverConfig, error)` |
| 1 | `(core.TelegramGuidedProgress, json.RawMessage, error)` |
| 1 | `(telegramHarnessChoices, bool, error)` |
| 1 | `([]WorkflowWorkStatus, bool, string)` |
| 1 | `(ConversationIntentContext, string, error)` |
| 1 | `(core.ExecutionRecord, core.ResultApplication, error)` |
| 1 | `(*core.EvidenceStrategyChecklist, string, error)` |
| 1 | `(core.WorkspaceConfig, string, error)` |
| 1 | `(core.IndexedTicket, []core.IndexedTicket, error)` |
| 1 | `(commandProjectContext, core.TicketIndex, error)` |
| 1 | `([]core.IndexedTicket, int, error)` |
| 1 | `([]tui.HarnessRetainedView, bool, string)` |
| 1 | `(AcceptanceFeaturePaths, AcceptanceFeatureCheckResult, error)` |
| 1 | `(AuthoringArtifactRef, string, error)` |
| 1 | `(AssistedTranscriptSummary, []ContextSourceInventoryEntry, error)` |
| 1 | `(AssistedArtifactSynthesisResult, []workflowFileChange, error)` |
| 1 | `(dumpMutationBinding, dumpMutationTarget, error)` |
| 1 | `(AssistedArtifactSynthesisResult, workflowFileChange, error)` |
| 1 | `(AssistedArtifactSynthesisResult, bool, error)` |
| 1 | `(AssistedModelResponse, AssistedModelInvocationPaths, error)` |
| 1 | `([]workflowFileChange, []AssistedTranscriptEvent, error)` |
| 1 | `(AuthoringContinuationRequest, string, error)` |
| 1 | `(AuthoringContinuationResult, bool, error)` |
| 1 | `(CaseMutationRequest, string, error)` |
| 1 | `(CaseMutationResult, bool, error)` |
| 1 | `(WorkflowExecutionPolicy, bool, error)` |
| 1 | `(string, ContextSourceInventoryEntry, error)` |
| 1 | `(*ExecutionResult, ContextSourceInventoryEntry, error)` |
| 1 | `(IndexedTicket, ContextDocument, error)` |
| 1 | `(dumpMutationBinding, bool, error)` |
| 1 | `(EvidenceStrategyChecklistItem, bool, error)` |
| 1 | `(GateGitRef, GateGitRef, error)` |
| 1 | `(HarnessExecutableRef, []string, error)` |
| 1 | `(string, HarnessGenerationBinding, bool)` |
| 1 | `(HarnessSelectionSnapshot, bool, error)` |
| 1 | `(TelegramPresentation, bool, error)` |
| 1 | `(PlanRevisionBinding, PlanOrderDocument, error)` |
| 1 | `(ProjectDumpArtifact, int, error)` |
| 1 | `(ProjectDumpMutationResult, bool, error)` |
| 1 | `(string, ProvingObjective, error)` |
| 1 | `(QuizActionResult, bool, error)` |
| 1 | `(ReportListEntry, bool, error)` |
| 1 | `(*url.URL, []string, error)` |
| 1 | `(SpecDocument, string, error)` |
| 1 | `(SpecMutationResult, bool, error)` |
| 1 | `(TelegramReceivedUpdate, bool, error)` |
| 1 | `(TelegramInterpretation, bool, error)` |
| 1 | `(TelegramOperationModel, bool, error)` |
| 1 | `(string, bool, time.Duration)` |
| 1 | `([]TelegramWorkReference, []string, error)` |
| 1 | `(string, int, bool)` |
| 1 | `(IndexedTicket, TicketReconciliationBlocker, bool)` |
| 1 | `(int, int, error)` |
| 1 | `([]WorkflowRecord, []string, error)` |
| 1 | `(WorkflowMaterializationResult, []workflowFileChange, error)` |
| 1 | `(WorkflowRecord, []workflowInvocationFile, error)` |
| 1 | `(WorkflowSlice, int, error)` |
| 1 | `(WorkflowIntegrationProof, []WorkflowEvidenceRef, error)` |
| 1 | `(*workflowClosingClaim, WorkflowSetClosureResult, error)` |
| 1 | `(WorkflowControlMutationRequest, string, error)` |
| 1 | `(workflowPreparedTask, workflowFileChange, error)` |
| 1 | `([]WorkflowInputImage, map[string]string, error)` |
| 1 | `(WorkflowPreviewTarget, WorkflowIntegrationProof, error)` |
| 1 | `(WorkflowStartRequest, bool, error)` |
| 1 | `([]WorkflowInvocation, int, error)` |
| 1 | `(workflowSliceObservation, WorkflowSliceCandidate, error)` |
| 1 | `(int64, int64, error)` |
| 1 | `(WorkflowMergerRepairPlan, bool, error)` |
| 1 | `(WorkflowIntegrationDecisions, bool, error)` |
| 1 | `([]string, string, error)` |
| 1 | `(WorkflowIntegrationSelection, string, error)` |
| 1 | `([]WorkflowIntegrationSelection, map[string]int, error)` |
| 1 | `(workflowInvocationFile, bool, error)` |
| 1 | `([]IndexedTicket, int, error)` |
| 1 | `(WorkflowTaskResult, bool, error)` |
| 1 | `([]WorkflowRecord, map[string]string, error)` |
| 1 | `(WorkflowLane, string, error)` |
| 1 | `(WorkflowRecord, []workflowFileChange, error)` |
| 1 | `(WorkflowMutationRequest, string, error)` |
| 1 | `(WorkflowSubmission, WorkflowGitPublicationResult, error)` |
| 1 | `(workflowSubmissionJournal, WorkflowIntegrationEntry, error)` |
| 1 | `(bool, WorkflowRecord, error)` |
| 1 | `(WorkflowSubmission, WorkflowIntegrationBudgetGuard, error)` |
| 1 | `(WorkflowNativePublicationObservation, WorkflowEvidenceRef, error)` |
| 1 | `(WorkflowTrackerStatus, bool, error)` |
| 1 | `(WorkflowPreparationInput, WorkflowLane, error)` |
| 1 | `(workflowFileChange, WorkflowEvidenceRef, error)` |
| 1 | `(*workflowCommandOutput, *workflowCommandOutput, error)` |
| 1 | `(WorkflowProcess, int, error)` |
| 1 | `(workflowSubmissionJournal, WorkflowPullRequestSnapshot, error)` |
| 1 | `(WorkflowIntegrationEntry, WorkflowIntegrationBudgetGuard, error)` |
| 1 | `(WorkflowResultValidation, int, error)` |
| 1 | `(WorkflowTaskInput, int, error)` |
| 1 | `(WorkflowPullRequestSnapshot, string, error)` |
| 1 | `(WorkflowReviewInput, WorkflowReviewDecision, error)` |
| 1 | `(WorkflowRecord, []WorkflowRecord, error)` |
| 1 | `(workflowUpgradeJournal, bool, error)` |
| 1 | `(WorkspaceIndex, bool, error)` |
| 1 | `(entry, bool, error)` |
| 1 | `(WorkBranchPolicy, WorkWorkspaceConfig, error)` |
| 1 | `(harness.AgentCommand, string, error)` |
| 1 | `(*harness.RuntimeMount, bool, error)` |
| 1 | `(*os.File, bool, error)` |
| 1 | `([]byte, bool, error)` |
| 1 | `(PageData, string, error)` |
| 1 | `(context.Context, func() error)` |
| 1 | `(core.NotificationEventType, core.NotificationSeverity)` |
| 1 | `(string, core.QuizMode)` |
| 1 | `(<-chan error, <-chan struct{})` |
| 1 | `(*core.EvidenceStrategyChecklist, string)` |
| 1 | `(AgentEvent, *AgentEventMalformedLine)` |
| 1 | `([]ExecutionReportChangedFile, ExecutionReportChangedFilesSummary)` |
| 1 | `(map[string]gateExecutionEvidence, []GateMissingAgentEvidence)` |
| 1 | `(*HarnessRunUsage, string)` |
| 1 | `([]LoopReportExecution, LoopReportExecutionSummary)` |
| 1 | `(map[string]string, string)` |
| 1 | `([]string, int)` |
| 1 | `([]TicketReviewRef, int)` |
| 1 | `([]string, []string)` |
| 1 | `(*EvidenceStrategyChecklist, string)` |
| 1 | `(int, string)` |
| 1 | `(int64, int64)` |
| 1 | `(*exec.Cmd, *workflowCommandGuard)` |
| 1 | `(context.Context, context.CancelFunc)` |
| 1 | `(string, []string)` |

### Full List

| Results | Location | Function | Shape |
| --- | --- | --- | --- |
| 7 | `internal/core/runtime_plan.go:186` | `runtimeArtifactPaths` | `(string, string, string, string, string, string, error)` |
| 6 | `internal/core/workflow_preparation_claim.go:12` | `WorkflowStore.prepareWorkflowIntegrationValidation` | `(*WorkflowClaim, workflowPreparedTask, WorkflowPreparationInput, int, *WorkflowIntegrationCandidate, error)` |
| 5 | `internal/app/service_guided_spec.go:153` | `service.guidedSpecInputs` | `([]core.QuizModelContextSection, core.QuizRuntimeState, []core.QuizEntry, core.GuidedStageCheckpoint, error)` |
| 5 | `internal/app/service_plan.go:343` | `service.generatePlanOutline` | `(core.GeneratedPlanOutline, []string, []string, []string, error)` |
| 5 | `internal/core/gate_config.go:114` | `parseGateConfigLine` | `(string, string, string, string, bool)` |
| 5 | `internal/core/gate_git_evidence.go:133` | `collectGateGitRange` | `(string, []GateGitCommit, []GateGitChangedFile, string, error)` |
| 5 | `internal/core/ticket_parser.go:260` | `parseIndexedTicketMetadataTimes` | `(*time.Time, *time.Time, *time.Time, int, error)` |
| 4 | `internal/app/app.go:278` | `loadTUIWorkspaceState` | `(core.WorkspaceGraph, core.WorkspaceIndex, []core.WorkspaceSemanticDigestState, error)` |
| 4 | `internal/app/notify.go:265` | `loadNotifyProjectConfig` | `(core.UserEnvironment, commandProjectContext, core.WorkspaceConfig, error)` |
| 4 | `internal/app/service_harness_worker.go:59` | `service.selectedWorkerCommand` | `([]string, string, string, error)` |
| 4 | `internal/app/service_plan.go:209` | `service.preparePlanGeneration` | `(core.PlanSourceObservation, core.QuizRuntimeState, string, error)` |
| 4 | `internal/app/service_plan.go:268` | `service.generatePlanWithConfiguredBackend` | `(core.GeneratedPlan, []string, []string, error)` |
| 4 | `internal/app/service_plan.go:276` | `service.generateSinglePassPlan` | `(core.GeneratedPlan, []string, []string, error)` |
| 4 | `internal/app/service_plan.go:308` | `service.generateStagedPlan` | `(core.GeneratedPlan, []string, []string, error)` |
| 4 | `internal/app/service_plan.go:379` | `service.generatePlanSliceDetail` | `(core.GeneratedPlanSliceDetail, string, string, error)` |
| 4 | `internal/app/service_work.go:1046` | `service.finishAgentWorkIntegrationCycles` | `(ticketLoopExecutionResult, []core.IndexedTicket, int, error)` |
| 4 | `internal/app/service_work.go:165` | `service.prepareAgentWorkTicket` | `(core.IndexedTicket, ticketStartPreflight, func(), error)` |
| 4 | `internal/app/service_work.go:324` | `service.runAgentWorkIntegrationCycles` | `(ticketLoopExecutionResult, []core.IndexedTicket, int, error)` |
| 4 | `internal/app/service_work.go:485` | `service.repairAgentWorkCycleTicket` | `(core.IndexedTicket, core.ExecutionRecord, bool, error)` |
| 4 | `internal/app/service_work.go:800` | `service.failAgentWorkIntegrationCycles` | `(ticketLoopExecutionResult, []core.IndexedTicket, int, error)` |
| 4 | `internal/app/service_workflow_submission.go:49` | `service.workflowSubmissionDependencies` | `(*core.WorkflowStore, core.WorkflowPublicationProvider, core.WorkflowInvocation, error)` |
| 4 | `internal/core/config.go:500` | `parseWorkspaceConfigLine` | `(string, string, string, bool)` |
| 4 | `internal/core/context_agent_protocols.go:158` | `normalizeAgentProtocolContextOptions` | `(string, string, string, error)` |
| 4 | `internal/core/dump_mutation_commit.go:83` | `fileAuthoringSessionRepository.dumpMutationBefore` | `([]byte, []byte, []workflowFileChange, error)` |
| 4 | `internal/core/proving.go:379` | `runProvingRalphLoopAttempt` | `(ProvingRalphLoopAttempt, *ProvingValidationResult, string, string)` |
| 4 | `internal/core/remote_command.go:261` | `remoteCommandFlagValue` | `(string, string, int, error)` |
| 4 | `internal/core/review_workflow_native_capability.go:167` | `RemoteReviewAdapter.readWorkflowNativePolicy` | `([]byte, []WorkflowNativeUnsupported, string, error)` |
| 4 | `internal/core/review_workflow_native_policy.go:25` | `workflowNativeRemotePolicy` | `([]byte, []WorkflowNativeUnsupported, string, error)` |
| 4 | `internal/core/runtime_plan.go:279` | `runtimeOpsLayout` | `(OpsLayout, string, bool, error)` |
| 4 | `internal/core/work_selection_preflight.go:151` | `normalizeAgentOperationLockOptions` | `(string, string, time.Time, error)` |
| 4 | `internal/core/workflow_agent_reserve.go:127` | `WorkflowStore.encodeWorkflowAgentInput` | `(workflowPreparedTask, workflowFileChange, workflowFileChange, error)` |
| 4 | `internal/core/workflow_merger_admission.go:11` | `WorkflowStore.workflowMergerAdmission` | `(WorkflowIntegrationEntry, WorkflowIntegrationBudget, string, error)` |
| 4 | `internal/core/workflow_native_host_policy.go:40` | `WorkflowNativeReviewAdapter.nativePublisher` | `(workflowInvocationFile, *WorkflowStore, WorkflowNativePublisherProfile, error)` |
| 4 | `internal/core/workflow_native_observation.go:54` | `WorkflowStore.observeWorkflowNativeSubmission` | `(WorkflowSubmission, WorkflowNativePublicationObservation, WorkflowIntegrationBudgetGuard, error)` |
| 4 | `internal/core/workflow_native_reconcile_git.go:32` | `WorkflowStore.workflowNativeReconciliationInput` | `(WorkflowSubmission, WorkflowGitPublication, WorkflowNativeSubmissionUpdate, error)` |
| 4 | `internal/core/workflow_plan_order_publish.go:159` | `WorkflowStore.currentPlanOrderMetadata` | `(map[string]any, workflowFileChange, bool, error)` |
| 4 | `internal/core/workflow_preparation_claim.go:137` | `WorkflowStore.claimWorkflowIntegrationValidation` | `(*WorkflowClaim, workflowPreparedTask, WorkflowPreparationInput, error)` |
| 4 | `internal/core/workflow_preparation_claim.go:49` | `WorkflowStore.workflowPreparationPreflight` | `(workflowSliceObservation, WorkflowReviewInput, WorkflowPublicationCapability, error)` |
| 4 | `internal/core/workflow_review_prepare.go:11` | `WorkflowStore.prepareWorkflowReview` | `(*WorkflowClaim, workflowPreparedTask, WorkflowReviewBinding, error)` |
| 4 | `internal/core/workflow_slice_prepare.go:10` | `WorkflowStore.prepareWorkflowSliceValidation` | `(*WorkflowClaim, workflowPreparedTask, *WorkflowSliceCandidate, error)` |
| 4 | `internal/core/workflow_tracker.go:174` | `WorkflowStore.prepareWorkflowTrackerImage` | `(WorkflowTrackerStatus, []byte, WorkflowTrackerStatus, error)` |
| 4 | `internal/harness/pi/resources.go:227` | `loadCallResources` | `([]byte, []byte, string, error)` |
| 4 | `internal/harness/pi/worker_connection.go:75` | `ReadWorkerInput` | `(string, context.Context, context.CancelFunc, error)` |
| 3 | `internal/app/author.go:1312` | `assistedEventCounts` | `(int, int, int)` |
| 3 | `internal/app/author_generation.go:11` | `contextualQuizGenerationProvider` | `(core.AuthoringQuizQuestionProvider, context.Context, error)` |
| 3 | `internal/app/metadata_cli.go:169` | `serviceFromCommand` | `(service, commandProjectContext, error)` |
| 3 | `internal/app/service_conversation_generation.go:48` | `service.bindIntentGeneration` | `(service, context.Context, error)` |
| 3 | `internal/app/service_dump_curation_openai.go:368` | `dumpCurationOpenAIResponseText` | `(string, string, error)` |
| 3 | `internal/app/service_guided_plan_context.go:11` | `service.guidedPlanSources` | `(core.PlanSourceObservation, core.QuizRuntimeState, error)` |
| 3 | `internal/app/service_guided_spec.go:129` | `service.guidedSpecModelRequest` | `(core.QuizModelRequest, string, error)` |
| 3 | `internal/app/service_guided_spec.go:192` | `service.guidedCommittedQuiz` | `(core.QuizRuntimeState, core.GuidedStageCheckpoint, error)` |
| 3 | `internal/app/service_guided_work_readiness.go:9` | `service.prepareGuidedWorkflowHarness` | `(*core.HarnessBinding, []string, error)` |
| 3 | `internal/app/service_harness_generation.go:111` | `prepareLegacyAPIGeneration` | `(core.HarnessGenerationBinding, core.HarnessGenerationLaunch, error)` |
| 3 | `internal/app/service_harness_generation.go:122` | `prepareCodexGenerationBinding` | `(core.HarnessGenerationBinding, core.HarnessGenerationLaunch, error)` |
| 3 | `internal/app/service_harness_generation.go:18` | `service.bindGeneration` | `(service, context.Context, error)` |
| 3 | `internal/app/service_harness_generation.go:219` | `preparePiGenerationBinding` | `(core.HarnessGenerationBinding, core.HarnessGenerationLaunch, error)` |
| 3 | `internal/app/service_harness_generation.go:60` | `service.prepareGenerationBinding` | `(core.HarnessGenerationBinding, core.HarnessGenerationLaunch, error)` |
| 3 | `internal/app/service_harness_worker.go:13` | `service.prepareWorkflowHarness` | `(*core.HarnessBinding, []string, error)` |
| 3 | `internal/app/service_plan.go:132` | `service.existingGeneratedPlan` | `(PlanView, bool, error)` |
| 3 | `internal/app/service_plan_delivery.go:101` | `service.planDeliveryInput` | `(core.GeneratedWorkflowPlan, core.OpsLayout, error)` |
| 3 | `internal/app/service_plan_replay.go:10` | `service.replayPlanPublication` | `(PlanOperationResult, bool, error)` |
| 3 | `internal/app/service_spec.go:478` | `summarizeWorkspaceMarkdownDocument` | `(string, string, string)` |
| 3 | `internal/app/service_work.go:385` | `service.runAgentWorkIntegrationCycle` | `(agentWorkIntegrationCycleState, bool, error)` |
| 3 | `internal/app/service_work.go:532` | `service.reviewAgentWorkCycleTicket` | `(core.IndexedTicket, bool, error)` |
| 3 | `internal/app/service_workflow_control_read.go:11` | `service.ReadWorkControlReceipt` | `(WorkflowWorkStatus, bool, error)` |
| 3 | `internal/app/service_workflow_preparation.go:40` | `service.workflowPreparationDependencies` | `(*core.WorkflowStore, core.WorkflowPublicationGuardProvider, error)` |
| 3 | `internal/app/service_workflow_review.go:39` | `service.workflowReviewDependencies` | `(*core.WorkflowStore, core.WorkflowPullRequestProvider, error)` |
| 3 | `internal/app/service_workflow_selection.go:55` | `service.selectedWorkflowIdentity` | `(core.WorkflowSelector, core.WorkflowSelectionGuard, error)` |
| 3 | `internal/app/telegram_authoring_command.go:31` | `parseTelegramAuthoring` | `(telegramAuthoringCommand, bool, error)` |
| 3 | `internal/app/telegram_authoring_command.go:76` | `telegramIntentAuthoring` | `(telegramAuthoringCommand, bool, error)` |
| 3 | `internal/app/telegram_control_command.go:48` | `parseTelegramControl` | `(telegramControlCommand, bool, error)` |
| 3 | `internal/app/telegram_control_command.go:91` | `telegramIntentControl` | `(telegramControlCommand, bool, error)` |
| 3 | `internal/app/telegram_control_intent.go:26` | `telegramReceiver.intentControlTarget` | `(core.TelegramOperationTarget, bool, error)` |
| 3 | `internal/app/telegram_control_intent.go:58` | `telegramReceiver.intentNativeWorkTarget` | `(core.TelegramOperationTarget, bool, error)` |
| 3 | `internal/app/telegram_control_service.go:11` | `telegramControlService.ReadTelegramControlReceipt (interface method)` | `(WorkflowWorkStatus, bool, error)` |
| 3 | `internal/app/telegram_control_service.go:16` | `service.ReadTelegramControlReceipt` | `(WorkflowWorkStatus, bool, error)` |
| 3 | `internal/app/telegram_conversation.go:33` | `telegramConversationInput` | `(core.TelegramConversationInput, bool, string)` |
| 3 | `internal/app/telegram_explanation_command.go:27` | `parseTelegramExplanation` | `(telegramExplanationCommand, bool, error)` |
| 3 | `internal/app/telegram_explanation_command.go:51` | `telegramIntentExplanation` | `(telegramExplanationCommand, bool, error)` |
| 3 | `internal/app/telegram_guided_action.go:108` | `telegramReceiver.prepareGuidedAction` | `(telegramGuidedInput, core.TelegramReceiverConfig, error)` |
| 3 | `internal/app/telegram_guided_apply.go:87` | `telegramGuidedResult` | `(core.TelegramGuidedProgress, json.RawMessage, error)` |
| 3 | `internal/app/telegram_guided_input.go:68` | `telegramGuidedTextAction` | `(string, string, error)` |
| 3 | `internal/app/telegram_guided_input.go:87` | `telegramGuidedStageAction` | `(string, string, error)` |
| 3 | `internal/app/telegram_guided_selection.go:98` | `telegramReceiver.resolveTelegramGuidedInput` | `(core.TelegramConversationInput, bool, error)` |
| 3 | `internal/app/telegram_harness_command.go:16` | `parseTelegramHarness` | `(telegramHarnessChoices, bool, error)` |
| 3 | `internal/app/telegram_harness_service.go:55` | `service.telegramHarnessRetained` | `([]WorkflowWorkStatus, bool, string)` |
| 3 | `internal/app/telegram_intent_dispatch.go:129` | `telegramReceiver.intentContext` | `(ConversationIntentContext, string, error)` |
| 3 | `internal/app/telegram_intent_guard.go:124` | `telegramReceiver.intentPreparationTarget` | `(core.TelegramOperationTarget, bool, error)` |
| 3 | `internal/app/telegram_intent_guard.go:44` | `telegramReceiver.intentTarget` | `(core.TelegramOperationTarget, bool, error)` |
| 3 | `internal/app/telegram_intent_registry.go:118` | `telegramIntentRecord` | `(telegramRecordCommand, bool, error)` |
| 3 | `internal/app/telegram_intent_registry.go:45` | `telegramIntentConversation` | `(core.TelegramConversationInput, bool, error)` |
| 3 | `internal/app/telegram_prepared_work_command.go:21` | `parseTelegramPreparedWork` | `(core.TelegramConversationInput, bool, error)` |
| 3 | `internal/app/telegram_prepared_work_command.go:37` | `telegramIntentPreparedWork` | `(core.TelegramConversationInput, bool, error)` |
| 3 | `internal/app/telegram_record_command.go:30` | `parseTelegramRecord` | `(telegramRecordCommand, bool, error)` |
| 3 | `internal/app/telegram_work_command.go:26` | `parseTelegramWork` | `(telegramWorkCommand, bool, error)` |
| 3 | `internal/app/telegram_work_command.go:61` | `telegramIntentWork` | `(telegramWorkCommand, bool, error)` |
| 3 | `internal/app/ticket.go:1054` | `executeLoopTicket` | `(core.ExecutionRecord, core.ResultApplication, error)` |
| 3 | `internal/app/ticket.go:1513` | `ticketExecutionReportEvidenceChecklist` | `(*core.EvidenceStrategyChecklist, string, error)` |
| 3 | `internal/app/ticket.go:2238` | `ticketMergeCommandConfig` | `(core.WorkspaceConfig, string, error)` |
| 3 | `internal/app/ticket.go:2259` | `ticketMergeCommandTickets` | `(core.IndexedTicket, []core.IndexedTicket, error)` |
| 3 | `internal/app/ticket.go:2847` | `openTicketIndexForCommand` | `(commandProjectContext, core.TicketIndex, error)` |
| 3 | `internal/app/ticket.go:686` | `selectTicketLoopTickets` | `([]core.IndexedTicket, int, error)` |
| 3 | `internal/app/tui_adapters_harness.go:42` | `tuiHarnessSelectionExecutor.retainedHarnessBindings` | `([]tui.HarnessRetainedView, bool, string)` |
| 3 | `internal/app/web_actions.go:101` | `webViewProvider.writeGenerationAction` | `(webui.WriteActionResponse, bool, error)` |
| 3 | `internal/app/web_actions.go:136` | `webViewProvider.writeWorkAction` | `(webui.WriteActionResponse, bool, error)` |
| 3 | `internal/app/web_actions.go:60` | `webViewProvider.writeDocumentAction` | `(webui.WriteActionResponse, bool, error)` |
| 3 | `internal/app/web_ticket.go:12` | `webViewProvider.writeTicketAction` | `(webui.WriteActionResponse, bool, error)` |
| 3 | `internal/app/work.go:220` | `workDoctorTicketContext` | `(string, string, error)` |
| 3 | `internal/app/work.go:380` | `workServiceFromCommand` | `(service, commandProjectContext, error)` |
| 3 | `internal/core/acceptance_feature.go:233` | `WriteAcceptanceFeature` | `(AcceptanceFeaturePaths, AcceptanceFeatureCheckResult, error)` |
| 3 | `internal/core/assisted_answer_curation.go:210` | `writeAssistedAnswerCurationArtifact` | `(AuthoringArtifactRef, string, error)` |
| 3 | `internal/core/assisted_context.go:131` | `loadAuthoringSelectedDocuments` | `([]ContextDocument, []ContextSourceInventoryEntry, error)` |
| 3 | `internal/core/assisted_context.go:164` | `loadAuthoringSelectedDocument` | `(ContextDocument, ContextSourceInventoryEntry, error)` |
| 3 | `internal/core/assisted_context.go:216` | `summarizeAssistedTranscript` | `(AssistedTranscriptSummary, []ContextSourceInventoryEntry, error)` |
| 3 | `internal/core/assisted_dump.go:114` | `fileAuthoringSessionRepository.prepareAssistedDump` | `(AssistedArtifactSynthesisResult, []workflowFileChange, error)` |
| 3 | `internal/core/assisted_dump.go:156` | `fileAuthoringSessionRepository.assistedDumpPreparation` | `(dumpMutationBinding, dumpMutationTarget, error)` |
| 3 | `internal/core/assisted_dump.go:186` | `assistedDumpResult` | `(AssistedArtifactSynthesisResult, workflowFileChange, error)` |
| 3 | `internal/core/assisted_dump_receipt.go:31` | `fileAuthoringSessionRepository.assistedDumpReceipt` | `(AssistedArtifactSynthesisResult, bool, error)` |
| 3 | `internal/core/assisted_quiz_loop.go:204` | `invokeAndRecordAssistedQuizModel` | `(AssistedModelResponse, AssistedModelInvocationPaths, error)` |
| 3 | `internal/core/authoring_artifact_layout.go:131` | `authoringArtifactPath` | `(string, string, error)` |
| 3 | `internal/core/authoring_dump.go:12` | `fileAuthoringSessionRepository.prepareAuthoringDump` | `(AuthoringArtifactRef, []workflowFileChange, error)` |
| 3 | `internal/core/authoring_prepare.go:217` | `fileAuthoringSessionRepository.prepareAuthoringOutputChanges` | `(AuthoringArtifactRef, []workflowFileChange, error)` |
| 3 | `internal/core/authoring_prepare.go:47` | `fileAuthoringSessionRepository.prepareAuthoringContent` | `([]workflowFileChange, []AssistedTranscriptEvent, error)` |
| 3 | `internal/core/authoring_request.go:13` | `normalizeAuthoringRequest` | `(AuthoringContinuationRequest, string, error)` |
| 3 | `internal/core/authoring_request.go:150` | `fileAuthoringSessionRepository.readAuthoringReceipt` | `(AuthoringContinuationResult, bool, error)` |
| 3 | `internal/core/authoring_writer_owner.go:26` | `ArchiveGeneratedPlan` | `(ArchivedPlanPaths, bool, error)` |
| 3 | `internal/core/case_mutation.go:73` | `CaseMetadataService.normalizeCaseMutation` | `(CaseMutationRequest, string, error)` |
| 3 | `internal/core/case_mutation_journal.go:188` | `CaseMetadataService.readCaseMutationReceipt` | `(CaseMutationResult, bool, error)` |
| 3 | `internal/core/config_workflow_execution.go:9` | `WorkflowStore.retainedWorkflowExecutionPolicy` | `(WorkflowExecutionPolicy, bool, error)` |
| 3 | `internal/core/context_agent_protocols.go:224` | `loadApplicableAgentFiles` | `([]ContextDocument, []ContextSourceInventoryEntry, error)` |
| 3 | `internal/core/context_agent_protocols.go:310` | `loadChrProtocolFiles` | `([]ContextDocument, []ContextSourceInventoryEntry, error)` |
| 3 | `internal/core/context_agent_protocols.go:371` | `loadProtocolCandidateFiles` | `([]ContextDocument, []ContextSourceInventoryEntry, error)` |
| 3 | `internal/core/context_previous_attempt.go:118` | `loadPreviousAttemptValidation` | `(string, ContextSourceInventoryEntry, error)` |
| 3 | `internal/core/context_previous_attempt.go:144` | `loadPreviousAttemptResult` | `(*ExecutionResult, ContextSourceInventoryEntry, error)` |
| 3 | `internal/core/context_project_documents.go:126` | `normalizeProjectDocumentContextOptions` | `(string, string, error)` |
| 3 | `internal/core/context_project_documents.go:154` | `loadContextTicket` | `(IndexedTicket, ContextDocument, error)` |
| 3 | `internal/core/context_project_documents.go:182` | `loadReferencedProjectDocuments` | `([]ContextDocument, []ContextSourceInventoryEntry, error)` |
| 3 | `internal/core/dump_mutation_binding.go:51` | `fileAuthoringSessionRepository.readDumpMutationBinding` | `(dumpMutationBinding, bool, error)` |
| 3 | `internal/core/dump_mutation_commit.go:45` | `fileAuthoringSessionRepository.dumpMutationChanges` | `([]workflowFileChange, ProjectDumpMutationResult, error)` |
| 3 | `internal/core/evidence_strategy_checklist.go:124` | `evidenceStrategyChecklistBlock` | `(string, bool, error)` |
| 3 | `internal/core/evidence_strategy_checklist.go:175` | `parseEvidenceStrategyChecklistLine` | `(EvidenceStrategyChecklistItem, bool, error)` |
| 3 | `internal/core/execution_runner.go:170` | `openExecutionOutputArtifacts` | `(*os.File, *os.File, error)` |
| 3 | `internal/core/gate_git_evidence.go:116` | `collectGateGitRefs` | `(GateGitRef, GateGitRef, error)` |
| 3 | `internal/core/guided_control_admission.go:57` | `WorkflowStore.lookupGuidedAdmission` | `(WorkflowInvocation, bool, error)` |
| 3 | `internal/core/guided_control_admission.go:85` | `findGuidedControlInvocation` | `(WorkflowInvocation, bool, error)` |
| 3 | `internal/core/harness_binding_executable.go:20` | `FreezeHarnessExecutable` | `(HarnessExecutableRef, []string, error)` |
| 3 | `internal/core/harness_generation_context.go:21` | `HarnessGenerationSelection` | `(string, HarnessGenerationBinding, bool)` |
| 3 | `internal/core/harness_selection_journal.go:151` | `readHarnessSelectionReceipt` | `(HarnessSelectionSnapshot, bool, error)` |
| 3 | `internal/core/migration.go:206` | `parseMigrationName` | `(string, string, error)` |
| 3 | `internal/core/migration.go:217` | `splitMigrationSQL` | `(string, string, error)` |
| 3 | `internal/core/notification_event_repository.go:28` | `NotificationEventRepository.ReadTelegramNotification (interface method)` | `(TelegramNotificationDelivery, bool, error)` |
| 3 | `internal/core/notification_telegram_read.go:33` | `sqliteNotificationEventRepository.ReadTelegramNotification` | `(TelegramNotificationDelivery, bool, error)` |
| 3 | `internal/core/notification_telegram_read.go:74` | `decodeTelegramNotificationManifest` | `(TelegramPresentation, bool, error)` |
| 3 | `internal/core/ops_named_artifact.go:10` | `NamedOpsArtifactBody` | `(string, string, error)` |
| 3 | `internal/core/plan_delivery_resume.go:12` | `WorkflowStore.ResumePlanDelivery` | `(WorkflowMaterializationResult, bool, error)` |
| 3 | `internal/core/plan_delivery_resume.go:27` | `WorkflowStore.resumePlanDelivery` | `(WorkflowMaterializationResult, bool, error)` |
| 3 | `internal/core/plan_generation.go:135` | `archiveGeneratedPlan` | `(ArchivedPlanPaths, bool, error)` |
| 3 | `internal/core/plan_generation.go:271` | `scopedPlanArtifactBody` | `(string, string, error)` |
| 3 | `internal/core/plan_order.go:119` | `planOrderBytes` | `([]byte, []byte, error)` |
| 3 | `internal/core/plan_order_publication.go:61` | `newPlanRevision` | `(PlanRevisionBinding, PlanOrderDocument, error)` |
| 3 | `internal/core/plan_publication_receipt.go:37` | `LookupPlanPublication` | `(PlanPublicationResult, bool, error)` |
| 3 | `internal/core/plan_publication_receipt.go:68` | `fileAuthoringSessionRepository.planReceipt` | `(PlanPublicationResult, bool, error)` |
| 3 | `internal/core/project_dump_artifact.go:12` | `fileAuthoringSessionRepository.readProjectDumpArtifact` | `(ProjectDumpArtifact, int, error)` |
| 3 | `internal/core/project_dump_composition.go:13` | `fileAuthoringSessionRepository.projectDumpImages` | `([]byte, []byte, error)` |
| 3 | `internal/core/project_dump_composition.go:132` | `fileAuthoringSessionRepository.projectDumpReconciliationChanges` | `([]workflowFileChange, ProjectDumpMutationResult, error)` |
| 3 | `internal/core/project_dump_receipt.go:40` | `fileAuthoringSessionRepository.projectDumpReceipt` | `(ProjectDumpMutationResult, bool, error)` |
| 3 | `internal/core/project_metadata.go:272` | `splitMarkdownFrontmatter` | `(string, string, error)` |
| 3 | `internal/core/project_metadata_fields.go:76` | `projectDocumentField` | `(string, string, bool)` |
| 3 | `internal/core/proving.go:769` | `loadProvingObjectiveForProject` | `(string, ProvingObjective, error)` |
| 3 | `internal/core/quiz_action_receipt.go:49` | `fileAuthoringSessionRepository.quizReceipt` | `(QuizActionResult, bool, error)` |
| 3 | `internal/core/quiz_generation_binding.go:10` | `OpenAIQuizModelAdapter.invokeBoundQuizAPI` | `([]byte, openAIQuizExecutionResult, error)` |
| 3 | `internal/core/quiz_openai_model.go:133` | `OpenAIQuizModelAdapter.invokeOpenAI` | `([]byte, openAIQuizExecutionResult, error)` |
| 3 | `internal/core/report_listing.go:125` | `reportListEntryFromPath` | `(ReportListEntry, bool, error)` |
| 3 | `internal/core/review_provider_diagnostics.go:340` | `parseReviewProviderURL` | `(*url.URL, []string, error)` |
| 3 | `internal/core/review_remote_adapter.go:212` | `RemoteReviewAdapter.resolve` | `(ReviewProviderReference, ReviewProviderTargetConfig, error)` |
| 3 | `internal/core/review_workflow_pull_request.go:258` | `RemoteReviewAdapter.resolveWorkflowRepository` | `(ReviewProviderReference, ReviewProviderTargetConfig, error)` |
| 3 | `internal/core/spec_document.go:124` | `splitSpecDocument` | `(string, []byte, error)` |
| 3 | `internal/core/spec_generation_publication.go:105` | `prepareGeneratedSpecDocument` | `(SpecDocument, string, error)` |
| 3 | `internal/core/spec_mutation_content.go:73` | `fileAuthoringSessionRepository.specReceipt` | `(SpecMutationResult, bool, error)` |
| 3 | `internal/core/telegram_action_answer.go:50` | `TelegramReceiverStore.ResolveTelegramActionAnswer` | `(TelegramActionAnswer, bool, error)` |
| 3 | `internal/core/telegram_action_answer.go:77` | `checkTelegramActionAnswer` | `(TelegramActionAnswer, bool, error)` |
| 3 | `internal/core/telegram_action_answer.go:85` | `TelegramReceiverStore.telegramReplyActionAnswer` | `(TelegramActionAnswer, bool, error)` |
| 3 | `internal/core/telegram_action_finish.go:170` | `telegramActionResponseSource` | `(TelegramReceivedUpdate, bool, error)` |
| 3 | `internal/core/telegram_authoring_target.go:11` | `TelegramReceiverStore.ResolveTelegramAuthoringTarget` | `(TelegramConversation, TelegramOperationTarget, error)` |
| 3 | `internal/core/telegram_authoring_work_request.go:37` | `TelegramReceiverStore.LatestTelegramAuthoringWork` | `(TelegramOperationRequest, bool, error)` |
| 3 | `internal/core/telegram_conversation.go:149` | `telegramConversationTransaction.apply` | `(string, *TelegramConversation, error)` |
| 3 | `internal/core/telegram_conversation.go:275` | `telegramConversationTransaction.peerSelection` | `(string, string, error)` |
| 3 | `internal/core/telegram_conversation.go:285` | `telegramConversationTransaction.selectProject` | `(string, *TelegramConversation, error)` |
| 3 | `internal/core/telegram_conversation.go:306` | `telegramConversationTransaction.start` | `(string, *TelegramConversation, error)` |
| 3 | `internal/core/telegram_conversation.go:351` | `telegramConversationTransaction.insertConversation` | `(string, *TelegramConversation, error)` |
| 3 | `internal/core/telegram_conversation_capture.go:10` | `telegramConversationTransaction.applyCapture` | `(string, *TelegramConversation, error)` |
| 3 | `internal/core/telegram_conversation_capture.go:39` | `telegramConversationTransaction.applyCaptureAction` | `(string, *TelegramConversation, error)` |
| 3 | `internal/core/telegram_conversation_capture.go:50` | `telegramConversationTransaction.appendCapture` | `(string, *TelegramConversation, error)` |
| 3 | `internal/core/telegram_conversation_capture.go:86` | `telegramConversationTransaction.finishCapture` | `(string, *TelegramConversation, error)` |
| 3 | `internal/core/telegram_conversation_capture.go:95` | `telegramConversationTransaction.cancelCapture` | `(string, *TelegramConversation, error)` |
| 3 | `internal/core/telegram_conversation_reference.go:11` | `TelegramReceiverStore.TelegramConversationReference` | `(TelegramConversation, bool, error)` |
| 3 | `internal/core/telegram_conversation_target.go:183` | `telegramConversationTransaction.focus` | `(string, *TelegramConversation, error)` |
| 3 | `internal/core/telegram_conversation_target.go:207` | `telegramConversationTransaction.list` | `(string, *TelegramConversation, error)` |
| 3 | `internal/core/telegram_document_http.go:68` | `telegramDocumentMultipart` | `([]byte, string, error)` |
| 3 | `internal/core/telegram_document_http.go:72` | `telegramDocumentMultipartDestination` | `([]byte, string, error)` |
| 3 | `internal/core/telegram_guided_binding.go:112` | `TelegramReceiverStore.ResolveTelegramGuidedTarget` | `(TelegramConversation, TelegramOperationTarget, error)` |
| 3 | `internal/core/telegram_guided_binding.go:16` | `TelegramReceiverStore.ReserveTelegramGuidedConversation` | `(TelegramConversation, TelegramOperationTarget, error)` |
| 3 | `internal/core/telegram_guided_binding.go:22` | `TelegramReceiverStore.BindTelegramGuidedConversation` | `(TelegramConversation, TelegramOperationTarget, error)` |
| 3 | `internal/core/telegram_guided_binding.go:36` | `TelegramReceiverStore.reserveTelegramGuidedAssociation` | `(TelegramConversation, TelegramOperationTarget, error)` |
| 3 | `internal/core/telegram_guided_projection.go:58` | `telegramGuidedProjectionReady` | `(TelegramConversation, bool, error)` |
| 3 | `internal/core/telegram_interpretation.go:84` | `TelegramReceiverStore.reserveTelegramInterpretation` | `(TelegramInterpretation, bool, error)` |
| 3 | `internal/core/telegram_operation_model.go:71` | `TelegramReceiverStore.reserveTelegramOperationModel` | `(TelegramOperationModel, bool, error)` |
| 3 | `internal/core/telegram_operation_request.go:37` | `TelegramReceiverStore.TelegramOperation` | `(TelegramOperationRequest, bool, error)` |
| 3 | `internal/core/telegram_processing_lane.go:121` | `telegramUnboundReplyLane` | `(TelegramProcessingLane, bool, error)` |
| 3 | `internal/core/telegram_processing_lane.go:136` | `telegramInputActionToken` | `(string, bool, error)` |
| 3 | `internal/core/telegram_processing_lane.go:35` | `telegramInputLane` | `(TelegramProcessingLane, bool, error)` |
| 3 | `internal/core/telegram_processing_lane.go:85` | `telegramActionInputLane` | `(TelegramProcessingLane, bool, error)` |
| 3 | `internal/core/telegram_receiver_config_yaml.go:101` | `telegramConfigParser.fieldHeader` | `(string, string, error)` |
| 3 | `internal/core/telegram_receiver_store.go:243` | `telegramFailureState` | `(string, bool, time.Duration)` |
| 3 | `internal/core/telegram_reply_resolution.go:41` | `TelegramReceiverStore.telegramUnboundReply` | `(TelegramActionAnswer, bool, error)` |
| 3 | `internal/core/telegram_work_reference.go:46` | `scanTelegramWorkReferences` | `([]TelegramWorkReference, []string, error)` |
| 3 | `internal/core/ticket_loop_query.go:338` | `splitTicketKey` | `(string, int, bool)` |
| 3 | `internal/core/ticket_parser.go:116` | `ticketFrontmatterField` | `(string, string, bool)` |
| 3 | `internal/core/ticket_reconciliation.go:135` | `ticketReconciler.parseProjectTicket` | `(IndexedTicket, TicketReconciliationBlocker, bool)` |
| 3 | `internal/core/ticket_repository.go:201` | `fileTicketRepository.ticketNumber` | `(int, bool, error)` |
| 3 | `internal/core/ticket_review_metadata.go:210` | `ticketFrontmatterBounds` | `(int, int, error)` |
| 3 | `internal/core/work_readiness_service.go:1176` | `OSWorkReadinessGit.GitStateInProgress` | `(string, bool, error)` |
| 3 | `internal/core/work_readiness_service.go:319` | `applySafeRepair` | `(WorkReadinessAppliedRepair, bool, error)` |
| 3 | `internal/core/work_readiness_service.go:342` | `applyRuntimeArtifactsRepair` | `(WorkReadinessAppliedRepair, bool, error)` |
| 3 | `internal/core/work_readiness_service.go:360` | `applyFetchRemoteRepair` | `(WorkReadinessAppliedRepair, bool, error)` |
| 3 | `internal/core/work_readiness_service.go:372` | `applyFastForwardRepair` | `(WorkReadinessAppliedRepair, bool, error)` |
| 3 | `internal/core/work_readiness_service.go:388` | `applySwitchBranchRepair` | `(WorkReadinessAppliedRepair, bool, error)` |
| 3 | `internal/core/work_readiness_service.go:404` | `applyCreateBranchRepair` | `(WorkReadinessAppliedRepair, bool, error)` |
| 3 | `internal/core/work_readiness_service.go:42` | `WorkReadinessGit.GitStateInProgress (interface method)` | `(string, bool, error)` |
| 3 | `internal/core/work_readiness_service.go:420` | `applyCreateWorktreeRootRepair` | `(WorkReadinessAppliedRepair, bool, error)` |
| 3 | `internal/core/work_readiness_service.go:435` | `applyPruneWorktreeMetadataRepair` | `(WorkReadinessAppliedRepair, bool, error)` |
| 3 | `internal/core/work_readiness_service.go:455` | `applyRemoveOrphanedWorktreeRepair` | `(WorkReadinessAppliedRepair, bool, error)` |
| 3 | `internal/core/workflow_adoption.go:204` | `WorkflowStore.resumePlanAdoption` | `(WorkflowMaterializationResult, bool, error)` |
| 3 | `internal/core/workflow_adoption_records.go:59` | `WorkflowStore.prepareLegacyWorkflowTasks` | `([]WorkflowRecord, []string, error)` |
| 3 | `internal/core/workflow_amendment_build.go:21` | `WorkflowStore.buildWorkflowAmendment` | `(WorkflowMaterializationResult, []workflowFileChange, error)` |
| 3 | `internal/core/workflow_amendment_guard.go:9` | `WorkflowStore.observeWorkflowAmendment` | `(WorkflowRecord, []workflowInvocationFile, error)` |
| 3 | `internal/core/workflow_amendment_slices.go:56` | `buildWorkflowAmendedSlice` | `(WorkflowSlice, int, error)` |
| 3 | `internal/core/workflow_candidate_repair_gate.go:59` | `WorkflowStore.prepareWorkflowCandidateRepairGit` | `(WorkflowIntegrationProof, []WorkflowEvidenceRef, error)` |
| 3 | `internal/core/workflow_candidate_repair_preparation.go:15` | `WorkflowStore.acceptedWorkflowCandidateRepair` | `(*WorkflowMergerRepairState, []WorkflowEvidenceRef, error)` |
| 3 | `internal/core/workflow_candidate_repair_preparation.go:40` | `WorkflowStore.readWorkflowAcceptedRepair` | `(*WorkflowMergerRepairState, []WorkflowEvidenceRef, error)` |
| 3 | `internal/core/workflow_capacity_process.go:50` | `workflowCapacityProcess` | `(WorkflowProcess, bool, error)` |
| 3 | `internal/core/workflow_capacity_process.go:59` | `observeWorkflowCapacityProcess` | `(WorkflowProcess, bool, error)` |
| 3 | `internal/core/workflow_capacity_slot.go:14` | `WorkflowStore.acquireWorkflowWorkerSlot` | `(*os.File, *os.File, error)` |
| 3 | `internal/core/workflow_claim.go:308` | `WorkflowStore.checkWorkflowClaim` | `(workflowInvocationFile, WorkflowRecord, error)` |
| 3 | `internal/core/workflow_claim.go:312` | `WorkflowStore.checkWorkflowCredential` | `(workflowInvocationFile, WorkflowRecord, error)` |
| 3 | `internal/core/workflow_claim_read.go:8` | `WorkflowStore.locateWorkflowClaim` | `(workflowInvocationFile, *WorkflowStore, error)` |
| 3 | `internal/core/workflow_closing_begin.go:12` | `WorkflowStore.beginWorkflowClosing` | `(*workflowClosingClaim, WorkflowSetClosureResult, error)` |
| 3 | `internal/core/workflow_closing_native.go:5` | `WorkflowStore.workflowClosingNativeDeliveryProof` | `(WorkflowSliceDeliveryProof, string, error)` |
| 3 | `internal/core/workflow_closing_proof.go:63` | `WorkflowStore.workflowClosingDeliveryProof` | `(WorkflowSliceDeliveryProof, string, error)` |
| 3 | `internal/core/workflow_continuation_preparation.go:22` | `WorkflowStore.retryWorkflowContinuationPreparation` | `(int, string, error)` |
| 3 | `internal/core/workflow_continuation_steps.go:53` | `WorkflowStore.readWorkflowContinuation` | `(WorkflowInvocation, bool, error)` |
| 3 | `internal/core/workflow_continuation_steps.go:8` | `WorkflowStore.continueWorkflowSteps` | `(int, string, error)` |
| 3 | `internal/core/workflow_control_read.go:7` | `WorkflowStore.ReadWorkflowControlReceipt` | `(WorkflowControlResult, bool, error)` |
| 3 | `internal/core/workflow_control_receipt.go:62` | `WorkflowStore.readWorkflowControlReceipt` | `(WorkflowControlResult, bool, error)` |
| 3 | `internal/core/workflow_control_request.go:46` | `freezeWorkflowControlMutation` | `(WorkflowControlMutationRequest, string, error)` |
| 3 | `internal/core/workflow_entry.go:12` | `WorkflowStore.FindWorkflowEntry` | `(WorkflowInvocation, bool, error)` |
| 3 | `internal/core/workflow_entry.go:68` | `WorkflowStore.matchWorkflowEntry` | `(WorkflowInvocation, bool, error)` |
| 3 | `internal/core/workflow_execution_prepare.go:137` | `encodeWorkflowTaskInput` | `(workflowPreparedTask, workflowFileChange, error)` |
| 3 | `internal/core/workflow_execution_prerequisites.go:8` | `WorkflowStore.workflowPrerequisiteImages` | `([]WorkflowInputImage, map[string]string, error)` |
| 3 | `internal/core/workflow_frontier_observe.go:8` | `WorkflowStore.observeWorkflowFrontier` | `(WorkflowPreviewTarget, WorkflowIntegrationProof, error)` |
| 3 | `internal/core/workflow_harness_admission.go:43` | `WorkflowStore.retainedWorkflowHarnessRequest` | `(WorkflowStartRequest, bool, error)` |
| 3 | `internal/core/workflow_inspection.go:86` | `WorkflowStore.InspectWorkflowPage` | `([]WorkflowInvocation, int, error)` |
| 3 | `internal/core/workflow_integration.go:138` | `WorkflowStore.workflowDeliveryCandidate` | `(workflowSliceObservation, WorkflowSliceCandidate, error)` |
| 3 | `internal/core/workflow_integration_budget_store.go:192` | `WorkflowStore.currentWorkflowIntegrationBudget` | `(WorkflowIntegrationEntry, WorkflowIntegrationBudget, error)` |
| 3 | `internal/core/workflow_integration_budget_validate.go:157` | `workflowIntegrationTokenTotals` | `(int64, int64, error)` |
| 3 | `internal/core/workflow_integration_decision_replay.go:22` | `workflowMergerAnsweredPlan` | `(WorkflowMergerRepairPlan, bool, error)` |
| 3 | `internal/core/workflow_integration_decision_store.go:241` | `WorkflowStore.workflowDecisionRequestRetry` | `(WorkflowIntegrationDecisions, bool, error)` |
| 3 | `internal/core/workflow_integration_grant.go:8` | `WorkflowStore.workflowGrantBudget` | `(WorkflowIntegrationEntry, WorkflowIntegrationBudget, error)` |
| 3 | `internal/core/workflow_integration_replay_verify.go:23` | `workflowIntegrationReader.rebaseResults` | `([]string, string, error)` |
| 3 | `internal/core/workflow_integration_schedule_capacity.go:117` | `WorkflowStore.workflowPreparedTargetCount` | `(int, bool, error)` |
| 3 | `internal/core/workflow_integration_schedule_capacity.go:151` | `WorkflowStore.workflowLegacyPreparedCount` | `(int, bool, error)` |
| 3 | `internal/core/workflow_integration_schedule_capacity.go:68` | `WorkflowStore.workflowIntegrationPreparationTarget` | `(string, string, error)` |
| 3 | `internal/core/workflow_integration_schedule_select.go:46` | `workflowIntegrationScheduler.selectTask` | `(WorkflowIntegrationSelection, string, error)` |
| 3 | `internal/core/workflow_integration_schedule_select.go:8` | `workflowIntegrationScheduler.selectTasks` | `([]WorkflowIntegrationSelection, map[string]int, error)` |
| 3 | `internal/core/workflow_invocation.go:374` | `WorkflowStore.findWorkflowInvocation` | `(workflowInvocationFile, bool, error)` |
| 3 | `internal/core/workflow_invocation_read.go:58` | `WorkflowStore.locateWorkflowInvocation` | `(workflowInvocationFile, *WorkflowStore, error)` |
| 3 | `internal/core/workflow_invocation_read.go:62` | `WorkflowStore.locateWorkflowInvocationPrepared` | `(workflowInvocationFile, *WorkflowStore, error)` |
| 3 | `internal/core/workflow_lane.go:133` | `WorkflowStore.workflowLaneCandidate` | `(string, string, error)` |
| 3 | `internal/core/workflow_legacy_write.go:52` | `FilterLegacyWorkTickets` | `([]IndexedTicket, int, error)` |
| 3 | `internal/core/workflow_manual_reuse.go:8` | `WorkflowStore.reuseWorkflowManualResult` | `(WorkflowTaskResult, bool, error)` |
| 3 | `internal/core/workflow_markdown.go:14` | `splitWorkflowMarkdown` | `(string, []byte, error)` |
| 3 | `internal/core/workflow_materialization_build.go:71` | `WorkflowStore.buildMaterializationTasks` | `([]WorkflowRecord, map[string]string, error)` |
| 3 | `internal/core/workflow_materialization_receipt.go:52` | `WorkflowStore.resumeMaterialization` | `(WorkflowMaterializationResult, bool, error)` |
| 3 | `internal/core/workflow_merger_broker.go:100` | `workflowMergerTerminal` | `(json.RawMessage, *WorkflowMergerUsage, error)` |
| 3 | `internal/core/workflow_merger_broker.go:125` | `writeWorkflowMergerTerminal` | `(json.RawMessage, *WorkflowMergerUsage, error)` |
| 3 | `internal/core/workflow_merger_broker.go:44` | `serveWorkflowMergerBroker` | `(json.RawMessage, *WorkflowMergerUsage, error)` |
| 3 | `internal/core/workflow_merger_claim.go:34` | `WorkflowStore.claimWorkflowMerger` | `(*WorkflowClaim, workflowPreparedTask, error)` |
| 3 | `internal/core/workflow_merger_claim.go:9` | `WorkflowStore.prepareWorkflowMerger` | `(*WorkflowClaim, workflowPreparedTask, error)` |
| 3 | `internal/core/workflow_merger_claim_persist.go:14` | `WorkflowStore.persistWorkflowMergerClaim` | `(*WorkflowClaim, workflowPreparedTask, error)` |
| 3 | `internal/core/workflow_merger_driver.go:90` | `exchangeWorkflowMerger` | `(json.RawMessage, *WorkflowMergerUsage, error)` |
| 3 | `internal/core/workflow_merger_workspace.go:15` | `WorkflowStore.workflowMergerWorkspace` | `(WorkflowLane, string, error)` |
| 3 | `internal/core/workflow_mutation.go:136` | `WorkflowStore.readMutationReceipt` | `(WorkflowMutationResult, bool, error)` |
| 3 | `internal/core/workflow_mutation.go:286` | `WorkflowStore.prepareMutation` | `(WorkflowRecord, []workflowFileChange, error)` |
| 3 | `internal/core/workflow_mutation.go:88` | `WorkflowStore.normalizeMutation` | `(WorkflowMutationRequest, string, error)` |
| 3 | `internal/core/workflow_mutation_reserve.go:8` | `WorkflowStore.ReserveRecordMutation` | `(WorkflowMutationResult, bool, error)` |
| 3 | `internal/core/workflow_native_dispatch.go:11` | `WorkflowStore.PublishWorkflowNativeSubmission` | `(WorkflowSubmission, WorkflowGitPublicationResult, error)` |
| 3 | `internal/core/workflow_native_dispatch.go:42` | `WorkflowStore.prepareWorkflowNativeDispatch` | `(WorkflowSubmission, WorkflowGitPublication, error)` |
| 3 | `internal/core/workflow_native_dispatch.go:68` | `WorkflowStore.authorizeWorkflowNativeDispatch` | `(WorkflowSubmission, WorkflowGitPublication, error)` |
| 3 | `internal/core/workflow_native_dispatch_evidence.go:150` | `WorkflowStore.readWorkflowOwnedDispatch` | `(workflowSubmissionJournal, WorkflowIntegrationEntry, error)` |
| 3 | `internal/core/workflow_native_finalization.go:11` | `WorkflowStore.finalizeWorkflowNativeUnit` | `(bool, WorkflowRecord, error)` |
| 3 | `internal/core/workflow_native_observation_store.go:60` | `WorkflowStore.retainWorkflowNativeCanonicalObservation` | `(WorkflowSubmission, WorkflowIntegrationBudgetGuard, error)` |
| 3 | `internal/core/workflow_native_receipt_delivery.go:94` | `WorkflowStore.workflowNativeReceiptObservation` | `(WorkflowNativePublicationObservation, WorkflowEvidenceRef, error)` |
| 3 | `internal/core/workflow_plan_order_context.go:100` | `WorkflowStore.retainedOrderRegistration` | `(WorkflowTrackerStatus, bool, error)` |
| 3 | `internal/core/workflow_preparation_claim.go:181` | `WorkflowStore.prepareWorkflowIntegrationGit` | `(WorkflowPreparationInput, WorkflowLane, error)` |
| 3 | `internal/core/workflow_preparation_native_evidence.go:100` | `workflowNativePreparationEvidenceFile` | `(workflowFileChange, WorkflowEvidenceRef, error)` |
| 3 | `internal/core/workflow_preparation_signing.go:171` | `workflowPreparationSignaturePayload` | `(string, string, error)` |
| 3 | `internal/core/workflow_process_record.go:104` | `WorkflowStore.workflowProcessOutput` | `(*workflowCommandOutput, *workflowCommandOutput, error)` |
| 3 | `internal/core/workflow_process_record.go:12` | `WorkflowStore.registerWorkflowProcess` | `(WorkflowProcess, int, error)` |
| 3 | `internal/core/workflow_publication.go:115` | `WorkflowStore.reconcileWorkflowPublicationCheckpoint` | `(*WorkflowRecord, bool, error)` |
| 3 | `internal/core/workflow_publication.go:145` | `WorkflowStore.finalizeWorkflowObservedPublication` | `(*WorkflowRecord, bool, error)` |
| 3 | `internal/core/workflow_publication.go:212` | `WorkflowStore.readWorkflowPublicationCheckpoint` | `(workflowSubmissionJournal, WorkflowPullRequestSnapshot, error)` |
| 3 | `internal/core/workflow_publication.go:97` | `WorkflowStore.observeWorkflowPublicationCheckpoint` | `(*WorkflowRecord, bool, error)` |
| 3 | `internal/core/workflow_publication_native.go:62` | `WorkflowStore.workflowNativeContinuationOwner` | `(WorkflowIntegrationEntry, WorkflowIntegrationBudgetGuard, error)` |
| 3 | `internal/core/workflow_result_validation.go:54` | `WorkflowStore.workflowResultValidation` | `(WorkflowResultValidation, int, error)` |
| 3 | `internal/core/workflow_result_validation.go:90` | `WorkflowStore.workflowResultTask` | `(WorkflowTaskInput, int, error)` |
| 3 | `internal/core/workflow_review_inputs.go:10` | `WorkflowStore.workflowReviewInputs` | `(workflowSliceObservation, WorkflowReviewInput, error)` |
| 3 | `internal/core/workflow_review_inputs.go:36` | `WorkflowStore.workflowReviewSourceInputs` | `(workflowSliceObservation, WorkflowReviewInput, error)` |
| 3 | `internal/core/workflow_review_inputs.go:74` | `observeWorkflowReviewPR` | `(WorkflowPullRequestSnapshot, string, error)` |
| 3 | `internal/core/workflow_review_persist.go:147` | `WorkflowStore.readWorkflowReview` | `(WorkflowReviewInput, WorkflowReviewDecision, error)` |
| 3 | `internal/core/workflow_review_prepare.go:119` | `WorkflowStore.claimWorkflowReview` | `(*WorkflowClaim, workflowPreparedTask, error)` |
| 3 | `internal/core/workflow_review_prepare.go:98` | `WorkflowStore.prepareWorkflowReviewPolicy` | `(*WorkflowClaim, workflowPreparedTask, error)` |
| 3 | `internal/core/workflow_slice_prepare.go:46` | `WorkflowStore.claimWorkflowSliceValidation` | `(*WorkflowClaim, workflowPreparedTask, error)` |
| 3 | `internal/core/workflow_submission_native.go:92` | `WorkflowStore.BeginWorkflowNativeSubmission` | `(WorkflowSubmission, bool, error)` |
| 3 | `internal/core/workflow_submission_native_store.go:42` | `WorkflowStore.retryWorkflowNativeSubmission` | `(WorkflowSubmission, bool, error)` |
| 3 | `internal/core/workflow_submission_native_store.go:55` | `WorkflowStore.authorizeWorkflowNativeIntent` | `(WorkflowSubmission, bool, error)` |
| 3 | `internal/core/workflow_submission_native_store.go:8` | `WorkflowStore.beginWorkflowNativeSubmission` | `(WorkflowSubmission, bool, error)` |
| 3 | `internal/core/workflow_submission_native_store.go:80` | `WorkflowStore.persistWorkflowNativeIntent` | `(WorkflowSubmission, bool, error)` |
| 3 | `internal/core/workflow_submission_prepare.go:103` | `WorkflowStore.observeWorkflowSubmissionRetry` | `(WorkflowSubmission, bool, error)` |
| 3 | `internal/core/workflow_submission_prepare.go:12` | `WorkflowStore.beginWorkflowSubmission` | `(WorkflowSubmission, bool, error)` |
| 3 | `internal/core/workflow_submission_prepare.go:151` | `WorkflowStore.persistWorkflowSubmissionIntent` | `(WorkflowSubmission, bool, error)` |
| 3 | `internal/core/workflow_submission_prepare.go:31` | `WorkflowStore.prepareWorkflowSubmission` | `(WorkflowSubmission, bool, error)` |
| 3 | `internal/core/workflow_submission_prepare.go:78` | `WorkflowStore.checkWorkflowSubmissionRetry` | `(WorkflowSubmission, bool, error)` |
| 3 | `internal/core/workflow_tracker_image.go:11` | `WorkflowStore.workflowTrackerImage` | `(WorkflowTrackerStatus, []byte, error)` |
| 3 | `internal/core/workflow_tracker_image.go:43` | `workflowTrackerMembers` | `(WorkflowRecord, []WorkflowRecord, error)` |
| 3 | `internal/core/workflow_tracker_order.go:32` | `WorkflowStore.versionedWorkflowTrackerImage` | `(WorkflowTrackerStatus, []byte, error)` |
| 3 | `internal/core/workflow_tracker_order.go:43` | `WorkflowStore.workflowOrderTrackerImage` | `(WorkflowTrackerStatus, []byte, error)` |
| 3 | `internal/core/workflow_upgrade_journal.go:163` | `WorkflowStore.retainedWorkflowUpgrade` | `(workflowUpgradeJournal, bool, error)` |
| 3 | `internal/core/workflow_upgrade_journal.go:65` | `WorkflowStore.prepareWorkflowUpgrade` | `(workflowUpgradeJournal, []byte, error)` |
| 3 | `internal/core/workflow_upgrade_journal.go:97` | `WorkflowStore.newWorkflowUpgradeJournal` | `(workflowUpgradeJournal, []byte, error)` |
| 3 | `internal/core/workspace_context.go:272` | `loadWorkspaceContextNode` | `(ContextDocument, ContextSourceInventoryEntry, error)` |
| 3 | `internal/core/workspace_index_cache.go:102` | `readWorkspaceIndexCache` | `(WorkspaceIndex, bool, error)` |
| 3 | `internal/core/workspace_init.go:296` | `workspace.profileFile` | `(entry, bool, error)` |
| 3 | `internal/core/worktree_plan.go:88` | `normalizedWorktreePlanConfig` | `(WorkBranchPolicy, WorkWorkspaceConfig, error)` |
| 3 | `internal/harness/pi/resources.go:77` | `selectedCredential` | `([]byte, string, error)` |
| 3 | `internal/harness/pi/worker_container.go:105` | `prepareWorkerContainer` | `(harness.AgentCommand, string, error)` |
| 3 | `internal/harness/pi/worker_container.go:215` | `workerModelMount` | `(*harness.RuntimeMount, bool, error)` |
| 3 | `internal/harness/pi/worker_resources.go:334` | `openStableCredentialPublication` | `(*os.File, bool, error)` |
| 3 | `internal/harness/pi/worker_resources.go:61` | `workerCredential` | `([]byte, bool, error)` |
| 3 | `internal/tui/agent_work_refresh.go:112` | `Model.updateAgentWorkRefreshMessage` | `(tea.Model, tea.Cmd, bool)` |
| 3 | `internal/tui/documents.go:354` | `summarizeMarkdownDocument` | `(string, string, string)` |
| 3 | `internal/tui/guided_input.go:12` | `Model.updateGuidedOrHarnessKey` | `(tea.Model, tea.Cmd, bool)` |
| 3 | `internal/tui/model.go:330` | `Model.updateSystemMessage` | `(tea.Model, tea.Cmd, bool)` |
| 3 | `internal/tui/model.go:368` | `Model.updateActionFinishedMessage` | `(tea.Model, tea.Cmd, bool)` |
| 3 | `internal/tui/model.go:496` | `Model.updateActiveModeKey` | `(tea.Model, tea.Cmd, bool)` |
| 3 | `internal/tui/model.go:549` | `Model.updateActiveQuizModeKey` | `(tea.Model, tea.Cmd, bool)` |
| 3 | `internal/webui/ticket_editor.go:57` | `handler.ticketWriteData` | `(PageData, string, error)` |
| 2 | `internal/app/daemon_workflow.go:168` | `watchDaemonHeartbeat` | `(context.Context, func() error)` |
| 2 | `internal/app/notification_emit.go:96` | `loopRunNotificationType` | `(core.NotificationEventType, core.NotificationSeverity)` |
| 2 | `internal/app/service_guided_quiz.go:210` | `guidedQuizNativeInput` | `(string, core.QuizMode)` |
| 2 | `internal/app/service_harness_selection.go:146` | `piProfileReadiness` | `(core.HarnessReadiness, string)` |
| 2 | `internal/app/service_harness_selection.go:173` | `piWorkerReadiness` | `(core.HarnessReadiness, string)` |
| 2 | `internal/app/telegram_conversation.go:63` | `telegramCommandWord` | `(string, string)` |
| 2 | `internal/app/telegram_lifecycle.go:92` | `startTelegramReceiverHosts` | `(<-chan error, <-chan struct{})` |
| 2 | `internal/app/ticket.go:1524` | `ticketExecutionReportEvidenceChecklistContent` | `(*core.EvidenceStrategyChecklist, string)` |
| 2 | `internal/core/agent_event_stream.go:190` | `decodeAgentEventLine` | `(AgentEvent, *AgentEventMalformedLine)` |
| 2 | `internal/core/evidence_strategy_checklist.go:225` | `parseEvidenceStrategyChecklistValue` | `(string, string)` |
| 2 | `internal/core/execution_report.go:240` | `executionReportChangedFiles` | `([]ExecutionReportChangedFile, ExecutionReportChangedFilesSummary)` |
| 2 | `internal/core/gate_evidence_classification.go:161` | `collectGateExecutionEvidence` | `(map[string]gateExecutionEvidence, []GateMissingAgentEvidence)` |
| 2 | `internal/core/harness_generation_run.go:167` | `harnessRunMeasuredUsage` | `(*HarnessRunUsage, string)` |
| 2 | `internal/core/loop_report.go:88` | `loopReportExecutions` | `([]LoopReportExecution, LoopReportExecutionSummary)` |
| 2 | `internal/core/quiz_curation.go:212` | `splitQuizTurnMetadata` | `(map[string]string, string)` |
| 2 | `internal/core/quiz_openai_model.go:753` | `quizModelRoute` | `(string, string)` |
| 2 | `internal/core/ticket_document.go:245` | `splitTicketDiscussion` | `(string, string)` |
| 2 | `internal/core/ticket_parser.go:126` | `parseTicketScalarList` | `([]string, int)` |
| 2 | `internal/core/ticket_parser.go:144` | `parseTicketReviewRefsBlock` | `([]TicketReviewRef, int)` |
| 2 | `internal/core/ticket_review_command.go:166` | `ticketReviewCommandStateAndStatus` | `(string, string)` |
| 2 | `internal/core/ticket_review_metadata.go:234` | `ticketFrontmatterBlockRange` | `(int, int)` |
| 2 | `internal/core/work_readiness_service.go:929` | `splitWorktreeStatus` | `([]string, []string)` |
| 2 | `internal/core/worker_context.go:494` | `parseWorkerEvidenceChecklist` | `(*EvidenceStrategyChecklist, string)` |
| 2 | `internal/core/workflow_continuation_steps.go:64` | `workflowContinuationAfter` | `(int, string)` |
| 2 | `internal/core/workflow_frontier_validate.go:153` | `workflowRetainedRoleCounts` | `(int, int)` |
| 2 | `internal/core/workflow_integration_budget_reserve.go:272` | `workflowIntegrationRoleUsage` | `(int64, int64)` |
| 2 | `internal/core/workflow_process_supervisor.go:44` | `workflowSupervisedCommand` | `(*exec.Cmd, *workflowCommandGuard)` |
| 2 | `internal/core/workflow_stop.go:80` | `WorkflowStore.watchWorkflowStop` | `(context.Context, context.CancelFunc)` |
| 2 | `internal/tui/agent_work_refresh.go:88` | `Model.handleAgentWorkRefreshTick` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/artifact_actions.go:133` | `Model.updateArtifactActionKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/case_metadata_actions.go:30` | `Model.updateCasesKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/command_palette.go:152` | `Model.updateCommandPaletteKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/command_palette.go:174` | `Model.updateCommandPaletteRunes` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/dump_actions.go:88` | `Model.updateDumpActionKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/dump_actions.go:99` | `Model.updateDumpEditKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/dumps.go:172` | `Model.updateDumpsKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/dumps.go:52` | `Model.renderDumpDetailPanel` | `(string, []string)` |
| 2 | `internal/tui/guided_input.go:150` | `Model.updateGuidedInputKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/guided_input.go:28` | `Model.updateGuidedKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/guided_input.go:54` | `Model.guidedActionKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/harness_selection.go:104` | `Model.updateHarnessKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/model.go:1575` | `Model.updateFilterKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/model.go:318` | `Model.Update` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/model.go:455` | `Model.updateKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/model.go:585` | `Model.updateMouse` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/model.go:613` | `Model.updateMousePress` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/model.go:669` | `Model.updateMouseMotion` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/model.go:691` | `Model.updateNavigationKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/model.go:724` | `Model.handleNavigationActionKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/model.go:887` | `Model.shellBodyGeometry` | `(int, int)` |
| 2 | `internal/tui/plan_actions.go:134` | `Model.updatePlanActionKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/plan_actions.go:83` | `Model.updatePlansKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/project_metadata_actions.go:37` | `Model.updateProjectKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/quiz_actions.go:106` | `Model.updateQuizAnswerKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/quiz_actions.go:170` | `Model.confirmQuizAction` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/quiz_actions.go:179` | `Model.saveQuizAnswer` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/quiz_actions.go:95` | `Model.updateQuizActionKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/quizzes.go:245` | `Model.updateQuizzesKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/shell.go:1474` | `Model.splitWidths` | `(int, int)` |
| 2 | `internal/tui/shell.go:1482` | `Model.splitHeights` | `(int, int)` |
| 2 | `internal/tui/spec_actions.go:108` | `Model.updateSpecActionKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/spec_actions.go:76` | `Model.updateSpecsKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/ticket_actions.go:127` | `Model.updateTicketActionKey` | `(tea.Model, tea.Cmd)` |
| 2 | `internal/tui/work_actions.go:303` | `Model.updateWorkActionKey` | `(tea.Model, tea.Cmd)` |
