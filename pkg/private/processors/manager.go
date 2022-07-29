package processors

var processors map[string]Processor = make(map[string]Processor)

func Get(name string) Processor {
	return processors[name]
}

func GetAll() map[string]Processor {
	return processors
}

// register a processor to system
func Register(name string, proc Processor) {
	processors[name] = proc
}

func init() {
	Register("checker:hcmp", CheckerHcmp{})
	Register("checker:testlib", CheckerTestlib{})
	Register("compiler", Compiler{})
	Register("compiler:auto", CompilerAuto{})
	Register("compiler:testlib", CompilerTestlib{})
	Register("generator:testlib", GeneratorTestlib{})
	Register("inputmaker", Inputmaker{})
	Register("runner", Runner{})
	Register("runner:fileio", RunnerFileio{})
	Register("runner:stdio", RunnerStdio{})
}
