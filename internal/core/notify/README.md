# Notifications

This package owns notification events, configuration, SQLite persistence, retained
Telegram admission, conditional dispatch claims, completion and document retention.
Each atomic operation remains in one transaction. Legacy, frozen, missing, failed,
uncertain and acknowledged outcomes retain their existing meanings and spellings.

The database schema, message payload and canonical project binding have independent
owners. HTTP sending and receiver/outbox persistence remain transport consumers.
This package imports no root core, application adapter or workflow coordinator.
