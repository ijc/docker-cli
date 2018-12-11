package cliplugins

import (
	"testing"

	"gotest.tools/assert"
	is "gotest.tools/assert/cmp"
	"gotest.tools/golden"
	"gotest.tools/icmd"
)

// TestRunNonexisting ensures correct behaviour when running a nonexistent plugin.
func TestRunNonexisting(t *testing.T) {
	res := icmd.RunCmd(icmd.Command("docker", "nonexistent"))
	res.Assert(t, icmd.Expected{
		ExitCode: 1,
	})
	assert.Assert(t, is.Equal(res.Stdout(), ""))
	golden.Assert(t, res.Stderr(), "docker-nonexistent-err.golden")
}

// TestRunBad ensures correct behaviour when running an existent but invalid plugin
func TestRunBad(t *testing.T) {
	res := icmd.RunCmd(icmd.Command("docker", "badmeta"))
	res.Assert(t, icmd.Expected{
		ExitCode: 1,
	})
	assert.Assert(t, is.Equal(res.Stdout(), ""))
	golden.Assert(t, res.Stderr(), "docker-badmeta-err.golden")
}

// TestRunGood ensures correct behaviour when running a valid plugin
func TestRunGood(t *testing.T) {
	res := icmd.RunCmd(icmd.Command("docker", "helloworld"))
	res.Assert(t, icmd.Expected{
		ExitCode: 0,
		Out:      "Hello World!",
	})
}
