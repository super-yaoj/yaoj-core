package analyzers

var processors map[string]Analyzer = make(map[string]Analyzer)

func Get(name string) Analyzer {
	return processors[name]
}

func GetAll() map[string]Analyzer {
	return processors
}

// register an analyzer to system
func Register(name string, proc Analyzer) {
	processors[name] = proc
}

func init() {
	Register("traditional", Traditional{})
}
