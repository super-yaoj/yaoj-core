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

// Get all processor's information (used by cook)
func All() map[string][][]string {
	res := map[string][][]string{}
	for name, input := range inLabel {
		res[name] = append(res[name], input)
		res[name] = append(res[name], ouLabel[name])
	}
	return res
}
