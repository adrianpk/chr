# Telegram

This package owns bounded receiver HTTP, configuration and OS exclusion, peer
routing, retained operations, confirmations and delivery. Existing operations
keep their one SQL transaction across conversations, requests, notices, actions
and document snapshots. A separate outbox package would split that transaction
without supplying an independent consumer, so it is not introduced.

Application operations perform domain execution and full domain validation.
The receiver binds validated routing values and frozen outcomes; it does not
import the root coordinator or an adapter. The separate message child owns
pure render/document values and cannot import this store or notification IO.

Imports are bounded by the syntax quality check. Native domain integration tests
remain with root core. Receiver SQL/HTTP/restart/retention tests live here;
integration fixtures own their independent SQL inspection connections.

HTTP calls take operation-specific send requests. Context and credentials remain
explicit. Document requests retain label-list cardinality so omission and an
explicit empty label list preserve their original validation behavior.

Transport health, operation, delivery, action, conversation and inbox vocabularies
have distinct string types with unchanged stored constants. SQL cursor cleanup
errors are returned; HTTP body cleanup preserves the already observed send
acknowledgement. No cleanup outcome grants another delivery attempt.
