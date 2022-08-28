package processor

var inLabel, ouLabel map[string][]string = map[string][]string{}, map[string][]string{}

// Get input label of builtin processors.
func InputLabel(name string) []string {
	return inLabel[name]
}

// Get output label of builtin processors.
func OutputLabel(name string) []string {
	return ouLabel[name]
}
