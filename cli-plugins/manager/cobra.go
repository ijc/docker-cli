package manager

import (
	"github.com/spf13/cobra"
)

const (
	// CommandAnnotPlugin is added to every stub command added by
	// AddPluginCommandStubs with the value "true" and so can be
	// used to distinguish plugin stubs from regular commands.
	CommandAnnotPlugin = "com.docker.cli.plugin"

	// CommandAnnotPluginVendor is added to every stub command
	// added by AddPluginCommandStubs and contains the vendor of
	// that plugin.
	CommandAnnotPluginVendor = "com.docker.cli.plugin.vendor"

	// CommandAnnotPluginInvalid is added to any stub command
	// added by AddPluginCommandStubs for an invalid command (that
	// is, one which failed it's candidate test) and contains the
	// reason for the failure.
	CommandAnnotPluginInvalid = "com.docker.cli.plugin-invalid"
)

// AddPluginCommandStubs adds a stub cobra.Commands for each plugin
// (optionally including invalid ones). The command stubs will have
// several annotations added, see `CommandAnnotPlugin*`.
func AddPluginCommandStubs(cmd *cobra.Command, includeInvalid bool) error {
	//fmt.Fprintf(os.Stderr, "Fall thru to HelpFunc\n")
	plugins, err := ListPlugins(cmd)
	if err != nil {
		return err
	}
	for _, p := range plugins {
		if !includeInvalid && p.Err != nil {
			continue
		}
		vendor := p.Vendor
		if vendor == "" {
			vendor = "unknown"
		}
		annots := map[string]string{
			CommandAnnotPlugin:       "true",
			CommandAnnotPluginVendor: vendor,
		}
		if p.Err != nil {
			annots[CommandAnnotPluginInvalid] = p.Err.Error()
		}
		cmd.AddCommand(&cobra.Command{
			Use:         p.Name,
			Short:       p.ShortDescription,
			Run:         func(_ *cobra.Command, _ []string) {},
			Annotations: annots,
		})
	}
	return nil
}
