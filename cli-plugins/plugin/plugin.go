package plugin

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/docker/cli/cli"
	cliplugins "github.com/docker/cli/cli-plugins"
	"github.com/docker/cli/cli/command"
	cliconfig "github.com/docker/cli/cli/config"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/cli/internal/containerizedengine"
	"github.com/docker/docker/pkg/term"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Command represents a top-level plugin command.
type Command struct {
	cobra.Command

	// RunPlugin may be used in preference to .Command.Run to get a Cli object.
	RunPlugin func(*cobra.Command, command.Cli, []string)
}

// Run is the top-level entry point to the CLI plugin framework. It should be called from your plugin's `main()` function.
func Run(plugin *Command, meta cliplugins.Metadata) {
	// Set terminal emulation based on platform as required.
	stdin, stdout, stderr := term.StdStreams()
	logrus.SetOutput(stderr)

	dockerCli := command.NewDockerCli(stdin, stdout, stderr, containerizedengine.NewClient)

	cmd := newPluginCommand(dockerCli, plugin, meta)

	if err := cmd.Execute(); err != nil {
		if sterr, ok := err.(cli.StatusError); ok {
			if sterr.Status != "" {
				fmt.Fprintln(stderr, sterr.Status)
			}
			// StatusError should only be used for errors, and all errors should
			// have a non-zero exit status, so never exit with 0
			if sterr.StatusCode == 0 {
				os.Exit(1)
			}
			os.Exit(sterr.StatusCode)
		}
		fmt.Fprintln(stderr, err)
		os.Exit(1)
	}
}

func newPluginCommand(dockerCli *command.DockerCli, plugin *Command, meta cliplugins.Metadata) *cobra.Command {
	opts := cliflags.NewClientOptions()
	var flags *pflag.FlagSet

	name := plugin.Use
	fullname := cliplugins.NamePrefix + name

	cmd := &cobra.Command{
		Use:              "docker" + " [OPTIONS] " + name + " [ARG...]",
		Short:            fullname + " is a Docker CLI plugin",
		SilenceUsage:     true,
		SilenceErrors:    true,
		TraverseChildren: false,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// flags must be the top-level command flags, not cmd.Flags()
			opts.Common.SetDefaultOptions(flags)
			return dockerCli.Initialize(opts)
		},
	}
	flags = cmd.Flags()
	flags.StringVar(&opts.ConfigDir, "config", cliconfig.Dir(), "Location of client config files")
	opts.Common.InstallFlags(flags)

	cmd.SetOutput(dockerCli.Out())

	// Setup plugin.Run if needed.
	if plugin.Command.Run == nil {
		plugin.Command.Run = func(cmd *cobra.Command, args []string) {
			plugin.RunPlugin(cmd, dockerCli, args)
		}
	}

	cmd.AddCommand(
		&plugin.Command,
		newMetadataSubcommand(plugin, meta),
	)

	return cmd
}

func newMetadataSubcommand(plugin *Command, meta cliplugins.Metadata) *cobra.Command {
	if meta.ShortDescription == "" {
		meta.ShortDescription = plugin.Short
	}
	cmd := &cobra.Command{
		Use:    cliplugins.MetadataSubcommandName,
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			enc := json.NewEncoder(os.Stdout)
			enc.SetEscapeHTML(false)
			enc.SetIndent("", "     ")
			return enc.Encode(meta)
		},
	}
	return cmd
}
