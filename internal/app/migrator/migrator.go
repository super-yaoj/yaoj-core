package migrator

type Migrator interface {
	// migrate dumpfile to YaOJ's problem file
	Migrate(dest string) error
}
