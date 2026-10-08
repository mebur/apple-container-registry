// Command ibmcloud-cr-apple-container is an IBM Cloud CLI plugin that adds
// support for authenticating Apple's `container` CLI
// (https://github.com/apple/container) against IBM Cloud Container Registry.
//
// It implements the behavior requested in
// https://github.com/IBM-Cloud/ibm-cloud-cli-release/issues/227 as a
// standalone plugin, since the official `ibmcloud cr` plugin is not an open
// source project that can be patched directly. Once installed, it exposes:
//
//	ibmcloud cr-container login
//
// which is the `container` CLI equivalent of the existing
//
//	ibmcloud cr login --client docker
//	ibmcloud cr login --client podman
//
// commands.
package main

import (
	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/plugin"

	crcontainer "github.com/patrikranudd/ibmcloud-cr-apple-container/internal/crcontainer"
)

func main() {
	plugin.Start(new(crcontainer.Plugin))
}
