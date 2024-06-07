package analyze

import "testing"

func Test_GetPkgPath(t *testing.T) {
	expected := "github.com/jlewi/foyle/app/pkg/analyze"
	actual := getFullPackagePath()
	if actual != expected {
		t.Errorf("Got %s; want %s", actual, expected)
	}
}
