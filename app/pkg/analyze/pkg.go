package analyze

import (
	"runtime"
	"strings"
)

func getFullPackagePath() string {
	pc, _, _, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	name := f.Name()

	// The name includes the path to the package, like 'main.getFullPackagePath'
	// So we should split by '.'
	parts := strings.Split(name, "/")
	if len(parts) > 0 {
		name = parts[len(parts)-1]
	}

	// Now name is 'main.getFullPackagePath', we just want 'main'
	nameParts := strings.Split(name, ".")
	if len(nameParts) > 1 {
		name = strings.Join(nameParts[:len(nameParts)-1], ".")
	}
	parts[len(parts)-1] = name

	return strings.Join(parts, "/")
}
