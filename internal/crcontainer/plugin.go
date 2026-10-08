// Package crcontainer implements an IBM Cloud CLI plugin that authenticates
// Apple's `container` CLI (https://github.com/apple/container) against
// IBM Cloud Container Registry, mirroring the existing
// `ibmcloud cr login --client docker|podman` flows.
package crcontainer

import (
	"os"

	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/bluemix/terminal"
	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/plugin"
)

const (
	pluginName  = "cr-container"
	aliasName   = "crc"
	namespace   = "cr-container"
	loginCmd    = "login"
	versionMaj  = 0
	versionMin  = 1
	versionBld  = 0
	minCLIMajor = 2
	minCLIMinor = 0
	minCLIBuild = 0
)

// Plugin implements the plugin.Plugin interface from the IBM Cloud CLI SDK.
type Plugin struct{}

// GetMetadata returns the metadata describing this plugin, its namespace and
// commands, as required by the IBM Cloud CLI plugin framework.
func (p *Plugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:          pluginName,
		Aliases:       []string{aliasName},
		Version:       plugin.VersionType{Major: versionMaj, Minor: versionMin, Build: versionBld},
		MinCliVersion: plugin.VersionType{Major: minCLIMajor, Minor: minCLIMinor, Build: minCLIBuild},
		Namespaces: []plugin.Namespace{
			{
				Name:        namespace,
				Aliases:     []string{aliasName},
				Description: "Authenticate Apple's container CLI against IBM Cloud Container Registry",
			},
		},
		Commands: []plugin.Command{
			{
				Namespace:   namespace,
				Name:        loginCmd,
				Description: "Log Apple's container CLI in to IBM Cloud Container Registry",
				Usage: "ibmcloud cr-container login [-r REGISTRY] [--iam-api-key API_KEY]\n" +
					"   ibmcloud crc login [-r REGISTRY] [--iam-api-key API_KEY]\n\n" +
					"   This command is the Apple `container` CLI equivalent of\n" +
					"   'ibmcloud cr login --client docker' / '--client podman'.",
				Flags: []plugin.Flag{
					{
						Name:        "r,registry",
						Description: "Registry domain to log in to (default: domain for the targeted region)",
						HasValue:    true,
					},
					{
						Name:        "iam-api-key",
						Description: "IAM API key to use instead of the current CLI session's IAM token",
						HasValue:    true,
					},
				},
			},
		},
	}
}

// Run dispatches the plugin command, as required by the plugin.Plugin
// interface.
func (p *Plugin) Run(c plugin.PluginContext, args []string) {
	ui := terminal.NewStdUI()

	if len(args) == 0 {
		ui.Failed("No command specified. Usage: %s or %s",
			terminal.CommandColor("ibmcloud cr-container login"),
			terminal.CommandColor("ibmcloud crc login"))
		os.Exit(1)
	}

	switch args[0] {
	case loginCmd:
		if err := runLogin(c, ui, args[1:]); err != nil {
			ui.Failed(err.Error())
			os.Exit(1)
		}
	default:
		ui.Failed("Unknown command %q", args[0])
		os.Exit(1)
	}
}
