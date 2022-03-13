### Want to make a new database migration?

1. Create a new file with the path being something along the lines of `internal/migration/4_changed_column_type.go`
2. Copy structure of `internal/migration/1_init.go`
3. Make sure you've changed the migrationVersion, and the contents of both functions passed along to Register
