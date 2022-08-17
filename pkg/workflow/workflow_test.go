package workflow_test

//go:generate go build -buildmode=plugin -o ./testdata ./testdata/custom_analyzer.go
// func TestLoadAnalyzer(t *testing.T) {
// 	a, err := workflow.LoadAnalyzer("testdata/custom_analyzer.so")
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	t.Log(pp.Sprint(a))
// }
