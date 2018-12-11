package manager

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"
)

// errPluginNotFound is the error returned when a plugin could not be found.
type errPluginNotFound string

func (e errPluginNotFound) NotFound() {}

func (e errPluginNotFound) Error() string {
	return "Error: No such CLI plugin: " + string(e)
}

type notFound interface{ NotFound() }

// IsNotFound is true if the given error is due to a plugin not being found.
func IsNotFound(err error) bool {
	_, ok := err.(notFound)
	return ok
}

func getPluginDirs(dockerCli command.Cli) []string {
	var pluginDirs []string

	if cfg := dockerCli.ConfigFile(); cfg != nil {
		pluginDirs = append(pluginDirs, cfg.CLIPluginsExtraDirs...)
	}

	pluginDirs = append(pluginDirs, defaultPluginDirs...)
	return pluginDirs
}

// findPlugin finds a valid plugin, if the first candidate is invalid then returns an error
func findPlugin(dockerCli command.Cli, name string, rootcmd *cobra.Command, includeShadowed bool) (Plugin, error) {
	if !pluginNameRe.MatchString(name) {
		// We treat this as "not found" so that callers will
		// fallback to their "invalid" command path.
		return Plugin{}, errPluginNotFound(name)
	}
	exename := NamePrefix + name
	if runtime.GOOS == "windows" {
		exename = exename + ".exe"
	}
	var plugin Plugin
	for _, d := range getPluginDirs(dockerCli) {
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
			if plugin, err = newPlugin(c, rootcmd); err != nil {
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
		return Plugin{}, errPluginNotFound(name)
	}
	return plugin, nil
}

func runPluginCommand(dockerCli command.Cli, name string, rootcmd *cobra.Command, args []string) (*exec.Cmd, error) {
	plugin, err := findPlugin(dockerCli, name, rootcmd, false)
	if err != nil {
		return nil, err
	}
	if plugin.Err != nil {
		return nil, errPluginNotFound(name)
	}
	return exec.Command(plugin.Path, args...), nil
}

// PluginRunCommand returns an "os/exec".Cmd which when .Run() will execute the named plugin.
// The rootcmd argument is referenced to determine the set of builtin commands in order to detect conficts.
// The error returned satisfies the IsNotFound() predicate if no plugin was found or if the first candidate plugin was invalid somehow.
func PluginRunCommand(dockerCli command.Cli, name string, rootcmd *cobra.Command) (*exec.Cmd, error) {
	// This uses the full original args, not the args which may
	// have been provided by cobra to our caller. This is because
	// they lack e.g. global options which we must propagate here.
	return runPluginCommand(dockerCli, name, rootcmd, os.Args[1:])
}
