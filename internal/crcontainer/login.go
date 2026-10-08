package crcontainer

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/bluemix/terminal"
	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/plugin"
	"github.com/spf13/pflag"
)

// containerBinary is the name of Apple's container CLI executable.
// See https://github.com/apple/container
const containerBinary = "container"

// registryDomains maps IBM Cloud region names to their IBM Cloud Container
// Registry domain, as documented at
// https://cloud.ibm.com/docs/Registry?topic=Registry-registry_access#registry_access_apikey_auth
var registryDomains = map[string]string{
	"global":     "icr.io",
	"au-syd":     "au.icr.io",
	"ap-south":   "au.icr.io",
	"br-sao":     "br.icr.io",
	"ca-mon":     "ca2.icr.io",
	"ca-tor":     "ca.icr.io",
	"eu-de":      "de.icr.io",
	"eu-central": "de.icr.io",
	"eu-es":      "es.icr.io",
	"eu-gb":      "uk.icr.io",
	"uk-south":   "uk.icr.io",
	"in-che":     "in.icr.io",
	"in-mum":     "in2.icr.io",
	"jp-osa":     "jp2.icr.io",
	"jp-tok":     "jp.icr.io",
	"ap-north":   "jp.icr.io",
	"us-south":   "us.icr.io",
}

// defaultRegistryDomain is used when the targeted region has no known
// dedicated registry domain.
const defaultRegistryDomain = "icr.io"

// runLogin implements `ibmcloud cr-container login`. It authenticates
// Apple's `container` CLI against IBM Cloud Container Registry using the
// IAM bearer token of the current CLI session, mirroring the behavior of
// `ibmcloud cr login --client docker` / `--client podman`.
func runLogin(c plugin.PluginContext, ui terminal.UI, args []string) error {
	flags := pflag.NewFlagSet(loginCmd, pflag.ContinueOnError)
	registry := flags.StringP("registry", "r", "", "Registry domain to log in to")
	apiKey := flags.String("iam-api-key", "", "IAM API key to use instead of the current CLI session's IAM token")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("incorrect usage: %w", err)
	}

	if !c.IsLoggedIn() && *apiKey == "" {
		return fmt.Errorf("not logged in. Log in by running %s, or pass %s",
			terminal.CommandColor("ibmcloud login"), terminal.CommandColor("--iam-api-key"))
	}

	containerPath, err := exec.LookPath(containerBinary)
	if err != nil {
		return fmt.Errorf(
			"the Apple container CLI was not found on the PATH.\n"+
				"Install it from %s and try again.",
			terminal.CommandColor("https://github.com/apple/container"))
	}

	domain := *registry
	if domain == "" {
		domain = registryDomainForRegion(c.CurrentRegion().Name)
	}

	username := "iambearer"
	password := c.IAMToken()
	if *apiKey != "" {
		username = "iamapikey"
		password = *apiKey
	} else {
		if password == "" {
			return fmt.Errorf("no IAM token available. Log in by running %s", terminal.CommandColor("ibmcloud login"))
		}
		password = strings.TrimPrefix(password, "Bearer ")
	}

	ui.Say("Logging in to %s as %s...", terminal.EntityNameColor(domain), terminal.EntityNameColor(username))

	if err := containerRegistryLogin(domain, username, password); err != nil {
		return err
	}

	// Ask the user if they wish to authorize the Apple container CLI and daemon in macOS Keychain
	// to avoid non-interactive Keychain authorization errors (-25293) on image push/pull.
	var authorizeKeychain bool
	promptErr := ui.Prompt(
		"Authorize Apple container binaries (/usr/local/bin/container*) to access this credential in Keychain without prompts? [Y/n]",
		nil,
	).Resolve(&authorizeKeychain)

	// If prompt succeeded and user confirmed (or pressed enter/default), configure access control
	if promptErr == nil && authorizeKeychain {
		if err := authorizeKeychainAccess(domain, username, password, containerPath); err != nil {
			ui.Warn("Warning: Could not configure Keychain access list: %s", err.Error())
		} else {
			ui.Say("Keychain access configured for container CLI.")
		}
	}

	ui.Ok()
	ui.Say("Login succeeded.")
	return nil
}

// registryDomainForRegion returns the IBM Cloud Container Registry domain
// for the given region. It falls back to the global domain if the region is
// unknown.
func registryDomainForRegion(region string) string {
	if domain, ok := registryDomains[region]; ok {
		return domain
	}
	return defaultRegistryDomain
}

// containerRegistryLogin shells out to `container registry login` and feeds
// the password through stdin so it never appears in the process list or
// shell history.
func containerRegistryLogin(domain, username, password string) error {
	return containerRegistryLoginWithBinary(containerBinary, domain, username, password)
}

// containerRegistryLoginWithBinary is like containerRegistryLogin but allows
// overriding the binary name, which is useful for testing.
func containerRegistryLoginWithBinary(binary, domain, username, password string) error {
	cmd := exec.Command(binary, "registry", "login", domain, "--username", username, "--password-stdin")
	cmd.Stdin = strings.NewReader(password)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("%s %s failed: %s", containerBinary, "registry login", msg)
	}

	return nil
}

// authorizeKeychainAccess updates the Keychain internet-password entry with
// trusted application paths (e.g. /usr/local/bin/container and /usr/local/bin/container-apiserver)
// so the daemon and CLI can access the registry credentials without encountering
// non-interactive authorization errors (-25293).
func authorizeKeychainAccess(domain, username, password, containerPath string) error {
	securityBin, err := exec.LookPath("security")
	if err != nil {
		return err
	}

	args := []string{
		"add-internet-password",
		"-a", username,
		"-s", domain,
		"-w", password,
		"-d", "com.apple.container.registry",
		"-U",
	}

	if containerPath != "" {
		args = append(args, "-T", containerPath)
	}

	// Also add container-apiserver if present in same directory
	apiServerPath, err := exec.LookPath("container-apiserver")
	if err == nil && apiServerPath != "" {
		args = append(args, "-T", apiServerPath)
	}

	cmd := exec.Command(securityBin, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("security command failed: %s", msg)
	}

	return nil
}
