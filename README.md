# Terraform Provider for MAAS

This repository contains the source code for the Terraform MAAS provider.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
- [Go](https://golang.org/doc/install) >= 1.16

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

#### `maas`

```hcl
provider "maas" {
  api_version = "2.0"
  api_key = "YOUR MAAS API KEY"
  api_url = "http://<MAAS_SERVER>[:MAAS_PORT]/MAAS"
}
```

### Resources and Data Sources Configuration

The [docs](/docs) section contains details about each supported Terraform resource and data source.

### Release process

1. Checkout main and create a new branch as `release-vX.X.X`
   ```bash
   git checkout main  
   git checkout -b release-vX.X.X
   git push -u origin release-vX.X.X
   ```
2. Raise a PR on github, title of the PR should be in the following format
   `Release vX.X.X`
3. Update the `CHANGELOG.md` with your release version, date and change details.
4. Go to [Releases](https://github.com/maas/terraform-provider-maas/releases) over on github
5. Click [Draft a new release](https://github.com/maas/terraform-provider-maas/releases/new)
6. On `Target` choose the lastest commit you want to release for
7. Set the `release title` to the release version, for example `v1.0.1`
8. Copy and paste the relevant CHNAGELOG.md entries to the release description
9. Click `Publish release`
10. The new version should be available on the [Releases](https://github.com/maas/terraform-provider-maas/releases) page
   
