//go:build integration

package integration

import (
	"fmt"
	"sync/atomic"

	"github.com/rogpeppe/go-internal/testscript"
)

var tagCounter uint64

// cmdGenTag generates unique tag for test isolation.
// Usage: gentag VARNAME
func cmdGenTag(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("gentag does not support negation")
	}
	if len(args) != 1 {
		ts.Fatalf("usage: gentag VARNAME")
	}
	tag := fmt.Sprintf("t%d", atomic.AddUint64(&tagCounter, 1))
	ts.Setenv(args[0], tag)
}
