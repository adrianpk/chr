# Operational metadata scalars

This package owns the existing quoting, list and timestamp representation shared
by ticket, workflow and authoring documents. Callers import it directly; there
is no document discovery, model generation, approval or persistence policy here.

Quote uses the original scalar quoting. QuoteList and ParseList keep the same
brackets, comma splitting, empty-list handling and empty-item omission. Unquote
keeps whitespace trimming and permissive fallback for values that do not decode.
Timestamp keeps UTC RFC3339, including the original quoted zero-time spelling.
OptionalTimestamp retains the existing empty value and timestamp diagnostics.
The package uses only the standard library. Caller-owned input bounds remain.
