// +build !windows

package manager

import (
	"github.com/docker/cli/cli/config"
)

var defaultPluginDirs = []string{
	config.Path("cli-plugins"),
	"/usr/local/lib/docker/cli-plugins", "/usr/local/libexec/docker/cli-plugins",
	"/usr/lib/docker/cli-plugins", "/usr/libexec/docker/cli-plugins",
}
