# Gate

This package collects Git evidence, runs configured validation, classifies changed
paths and encodes dev-to-main gate reports. Callers supply the configuration path
and a report loader. Application composition allocates operational artifacts and
refreshes their metadata. Validation and execution evidence do not merge code.

The package depends on configuration literals and standard-library facilities.
It cannot import root core, workflow coordination or adapters. Git commands and
validation children use the operation context. Evidence diagnostics retain report
identities and original order; manual changes retain their high-risk patterns.
