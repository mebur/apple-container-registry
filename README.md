# ibmcloud-apple-container

An IBM Cloud CLI plug-in that authenticates [Apple's `container` CLI](https://github.com/apple/container)
against IBM Cloud Container Registry.

It implements the behavior requested in
[IBM-Cloud/ibm-cloud-cli-release#227](https://github.com/IBM-Cloud/ibm-cloud-cli-release/issues/227)
as a standalone, third-party plug-in: the official `ibmcloud cr` plug-in is
not open source, so `--client container` can't be added to it directly from
here. This plug-in provides the equivalent workflow instead:

```sh
ibmcloud cr login --client docker      # existing, built into the cr plug-in
ibmcloud cr login --client podman      # existing, built into the cr plug-in
ibmcloud acr login                     # this plug-in (alias: apple-container)
```

## What it does

`ibmcloud acr login` (or `ibmcloud apple-container login`):

1. Confirms you are logged in to IBM Cloud (`ibmcloud login`) or that you
   supplied an API key via `--iam-api-key`.
2. Confirms the `container` CLI is installed and on your `PATH`.
3. Resolves the IBM Cloud Container Registry domain for your targeted region
   (or uses the domain passed via `--registry`).
4. Runs `container registry login DOMAIN --username USER --password-stdin`,
   piping in your IAM bearer token (or API key) so it's never exposed in the
   process list or shell history.
5. Configures the macOS Keychain Access List to authorize `container` daemon
   binaries so push/pull operations run without prompt/error.

## Usage

```sh
ibmcloud acr login [-r REGISTRY] [--iam-api-key API_KEY]
# or
ibmcloud apple-container login [-r REGISTRY] [--iam-api-key API_KEY]
```

| Flag | Description |
| --- | --- |
| `-r, --registry` | Registry domain to log in to (default: the domain for your targeted region, e.g. `us.icr.io`) |
| `--iam-api-key` | Use an IAM API key instead of the current CLI session's IAM token |

## Build

Requires Go 1.22+.

```sh
go build -o ibmcloud-apple-container .
```

## Install locally

```sh
ibmcloud plugin install -f ./ibmcloud-apple-container
```

## Test

```sh
go test ./...
```

## Uninstall

```sh
ibmcloud plugin uninstall apple-container
```
