package migration

var (
	migrations []Migration
)

func Register(version int64, up MigrationFunction, down MigrationFunction) {
	migrations = append(migrations, Migration{
		Version: version,
		Up:      up,
		Down:    down,
	})
}
