# User database

This package owns the existing SQLite migration runner, schema inspection,
connection configuration, database path, timestamp encoding and transaction error
handling. Repositories keep their own SQL, row mappings and transactions. The
schema assets, migration ordering, one-connection limit and busy timeout remain
unchanged. No workflow coordinator or adapter is imported.
