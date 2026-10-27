# Telegram messages

This package owns bounded visible blocks, UTF-16 entities, complete frozen document
values and retained-presentation validation. Receiver and notification storage use
the same byte, hash, version, text, entity and attachment-binding rules. It imports
no transport client, database, root core or workflow coordinator.

Rendering keeps the existing text/document revisions, optional fields, JSON order,
Unicode limits and original document bytes. Storage callers own admission, quotas,
attempt allowance, acknowledgement and legacy-delivery state. SplitResponse retains
the existing bounded compatibility segmentation used by retained help responses.
