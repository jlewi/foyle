package server

import (
	"fmt"
	"github.com/pkg/errors"
	"regexp"
)

// vcodeCors is an allow function for CORS requests from the vscode-test-web server
// This is needed because the vscode-test-web server generates a random prefix so the server name will be something like
// http://v--19cf5ppcsqee9rrkfifq1ajc8b7nv2t96593a6n6bn95st339ul8.localhost:3000.
// So to allow the backend server to accept requests from the test server, we need to allow requests from the test server
type vscodeCors struct {
	match *regexp.Regexp
}

// NewVscodeCors creates a new vscodeCors instance for the given pot
func NewVscodeCors(port int) (*vscodeCors, error) {
	m, err := regexp.Compile(`^http://v--\w+\.localhost:` + fmt.Sprintf("%d", port) + `$`)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to compile regex")
	}

	return &vscodeCors{
		match: m,
	}, nil
}

func (f *vscodeCors) allowOrigin(origin string) bool {
	return f.match.MatchString(origin)
}
