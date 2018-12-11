package manager

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	cliplugins "github.com/docker/cli/cli-plugins"
	"github.com/spf13/cobra"
)

// ErrPluginNotFound is the error returned when a plugin could not be found.
type ErrPluginNotFound string

func (e ErrPluginNotFound) Error() string {
	return "Error: No such CLI plugin: " + string(e)
}

var (
	pluginDirs     []string
	pluginDirsOnce sync.Once
)

func getPluginDirs() []string {
	pluginDirsOnce.Do(func() {
		// Mostly for test.
		if ds := os.Getenv("DOCKER_CLI_PLUGIN_EXTRA_DIRS"); ds != "" {
			pluginDirs = append(pluginDirs, strings.Split(ds, ":")...)
		}

		pluginDirs = append(pluginDirs,
			filepath.Join(os.Getenv("HOME"), ".docker/cli-plugins"),
			"/usr/local/lib/docker/cli-plugins", "/usr/local/libexec/docker/cli-plugins",
			"/usr/lib/docker/cli-plugins", "/usr/libexec/docker/cli-plugins",
		)
	})
	return pluginDirs
}

// FindPlugin finds a valid plugin, if the first candidate is invalid then returns an error
func FindPlugin(name string, rootcmd *cobra.Command, includeShadowed bool) (Plugin, error) {
	if !pluginNameRe.MatchString(name) {
		// We treat this as "not found" so that callers will
		// fallback to their "invalid" command path.
		return Plugin{}, ErrPluginNotFound(name)
	}
	exename := cliplugins.NamePrefix + name
	if runtime.GOOS == "windows" {
		exename = exename + ".exe"
	}
	var plugin Plugin
	for _, d := range getPluginDirs() {
		path := filepath.Join(d, exename)

		// We stat here rather than letting the exec tell us
		// ENOENT because the latter does not distinguish a
		// file not existing from its dynamic loader or one of
		// its libraries not existing.
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		if plugin.Path == "" {
			c := &candidate{path: path}
			var err error
			if plugin, err = NewPlugin(c, rootcmd); err != nil {
				return Plugin{}, err
			}
			if !includeShadowed {
				return plugin, nil
			}
		} else {
			plugin.ShadowedPaths = append(plugin.ShadowedPaths, path)
		}
	}
	if plugin.Path == "" {
		return Plugin{}, ErrPluginNotFound(name)
	}
	return plugin, nil
}

func runPluginCommand(name string, rootcmd *cobra.Command, args []string) (*exec.Cmd, error) {
	plugin, err := FindPlugin(name, rootcmd, false)
	if err != nil {
		return nil, err
	}
	if err := plugin.IsValid(); err != nil {
		return nil, ErrPluginNotFound(name)
	}
	return exec.Command(plugin.Path, args...), nil
}

// PluginRunCommand returns an "os/exec".Cmd which when .Run() will execute the named plugin.
// The rootcmd argument is referenced to determine the set of builtin commands in order to detect conficts.
// The error returned is an ErrPluginNotFound if no plugin was found or if the first candidate plugin was invalid somehow.
func PluginRunCommand(name string, rootcmd *cobra.Command) (*exec.Cmd, error) {
	// This uses the full original args, not the args which may
	// have been provided by cobra to our caller. This is because
	// they lack e.g. global options which we must propagate here.
	return runPluginCommand(name, rootcmd, os.Args[1:])
}
