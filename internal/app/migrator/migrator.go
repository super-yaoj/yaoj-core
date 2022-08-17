package migrator

import "github.com/super-yaoj/yaoj-core/pkg/problem"

type Migrator interface {
	// migrate dumpfile to YaOJ's problem in specific dir
	Migrate(src string, dir string) (Problem, error)
}

type Problem = problem.Problem
