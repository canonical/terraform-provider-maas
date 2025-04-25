# Terraform Provider for MAAS

This repository contains the source code for the Terraform MAAS provider.

## :warning: Repository ownership and provider name change

The Terraform Provider for MAAS repository now lives under the [Canonical GitHub organisation](https://github.com/canonical) with a new name `github.com/canonical/terraform-provider-maas`.

Ensure you are pointing at the new provider name inside your Terraform module(s), which is `canonical/maas`:

1. Manually update the list of required providers in your Terraform module(s):

    ```diff
    terraform {
      required_providers {
        maas = {
    -     source  = "maas/maas"
    +     source  = "canonical/maas"
          version = "~>2.0"
        }
      }
    }
    ```

1. Upgrade your provider dependencies to add the `canonical/maas` provider info:

    ```bash
    terraform init -upgrade
    ```

1. Replace the provider reference in your state:

    ```bash
    terraform state replace-provider maas/maas canonical/maas
    ```

1. Upgrade your provider dependencies to remove the `maas/maas` provider info:

    ```bash
    terraform init -upgrade
    ```

References:

- <https://developer.hashicorp.com/terraform/language/files/dependency-lock#dependency-on-a-new-provider>
- <https://developer.hashicorp.com/terraform/language/files/dependency-lock#providers-that-are-no-longer-required>
- <https://developer.hashicorp.com/terraform/cli/commands/state/replace-provider>

---

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.4.x
- [Go](https://golang.org/doc/install) >= 1.23

## Build The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider with:

    ```sh
    make build
    ```

1. (Optional): Install the freshly built provider with:

    ```sh
    make install
    ```

## Usage

### Provider Configuration

The provider accepts the following config options:

- **api_key**: [MAAS API key](https://maas.io/docs/snap/3.0/cli/maas-cli#heading--log-in-required).
- **api_url**: URL for the MAAS API server (eg: <http://127.0.0.1:5240/MAAS>).
- **api_version**: MAAS API version used. It is optional and it defaults to `2.0`.
- **installation_method**: MAAS Installation method used. Optional, defaults to `snap`. Valid options: `snap`, and `deb`.

#### `maas`

```hcl
provider "maas" {
  api_version = "2.0"
  api_key     = "YOUR MAAS API KEY"
  api_url     = "http://<MAAS_SERVER>[:MAAS_PORT]/MAAS"

  installation_method = "snap"
}
```

### Resources and Data Sources Configuration

The [docs](/docs) section contains details about each supported Terraform resource and data source.

## Release process

Builds and releases are automated with GitHub Actions and GoReleaser. There are a few manual steps to complete:

1. Start the release action: 
   
  ```shell
  git checkout master
  git tag vX.Y.Z
  git push upstream tag vX.Y.Z
  ```

  Where `upstream` is the remote name pointing to the canonical/terraform-provider-maas repository. 
  
  The provider versions follow [semantic versioning](https://semver.org/), and the release action can be viewed under [Actions](https://github.com/canonical/terraform-provider-maas/actions).

2. Publish the release: 
   
   The action creates a "draft" release. Go to [Releases](https://github.com/canonical/terraform-provider-maas/releases) to open it, select `edit` and click `Generate release notes`. Select `Publish release` when you are happy. 

3. Verify the published release looks good by checking in [Releases](https://github.com/canonical/terraform-provider-maas/releases).

## Additional Info

### Testing

Unit tests run with every pull request and merge to master. The end to end tests run on a nightly basis against a hosted MAAS deployment, results can be found [here](https://raw.githubusercontent.com/canonical/maas-terraform-e2e-tests/main/results.json?token=GHSAT0AAAAAAB3FX6R5C67Q4LH7ADOO5O3IY4ODCNA) and are checked on each PR, with a warning if failed.
