package manager

import (
	"github.com/docker/cli/cli/config"
)

var defaultPluginDirs = []string{
	config.Path("cli-plugins"),
	`C:\ProgramData\Docker\cli-plugins`,
}
