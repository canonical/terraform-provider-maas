# Contributing to the Terraform MAAS Provider

Thank you for your interest in contributing to the Terraform MAAS Provider! We appreciate your help in making this project better. This document provides information on how to set up your development environment and best practices for contributing to the project. For information on what this provider is and does, see [docs/index.md](docs/index.md).


## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.4.x
- [Go](https://golang.org/doc/install) >= 1.22
- A MAAS installation running. See the [maas-dev-setup](https://github.com/canonical/maas-dev-setup) repository for more information on a development setup.
- [CLA](https://ubuntu.com/legal/contributors) signed with the email used with git and GitHub.

## Branching strategy

This project follows a fork-based development model with a single long-running master branch. All contributions should be made via pull requests (PRs) from forked repositories.

1. Fork the Repository:
    1. Go to the repository on GitHub and click "Fork" in the top-right corner.
    1. Clone your fork locally:
       ```bash
       git clone <https-or-ssh-url-to-your-fork>
       cd terraform-provider-maas
       ```
    1. Add the upstream repository (the original repo) as a second remote:
       ```bash
       git remote add upstream <https-or-ssh-url-to-original>
       ```
1. Create a Feature Branch:
    ```bash
    git checkout -b feat/feature-name
    ```
1. Keep Your Branch Up to Date:
    1. Before working, sync your branch with the latest changes from master:
       ```bash
       git fetch upstream
       git checkout master
       git merge upstream/master
       ```
    1. Then, rebase or merge your feature branch if necessary:
        ```bash
        git checkout feat/feature-name
        git rebase master
        ```   
1. Commit and Push Changes:
    1. Follow commit message guidelines (e.g., fix: correct typo in readme).
    1. Push your branch to your forked repository:
        ```bash
        git push origin feat/feature-name
        ```
1. Submit a Pull Request:
    1. Go to the your forked repository on GitHub.
    1. Click "New Pull Request". Select your feature branch to merge from your forked repo, into the master branch of the original repo.
    1. Ensure your PR includes:
       - A clear description of changes.
       - Links to relevant issues (e.g., Fixes #113).
       - Passing tests, if applicable.
1. Address Review Feedback. Once approved, a maintainer will merge your PR. 🎉

## Commit messages

We follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification. Conventional Commits defines the following structure for the Git commit message:

```bash
<type>[scope][!]: <description>

[body]

[footer(s)]
```

Where 
- `type` is the kind of the change (e.g. feature, bug fix, documentation change, refactor).
- `scope` may be used to provide additional contextual information (e.g. which system component is affected). If scope is provided, it’s enclosed in parentheses.
- `!` MUST be added if commit introduces a breaking change.
- `description` is a brief summary of a change (try to keep it short, so overall title no more than 72 characters).
- `footer` is detailed information about the change (e.g. breaking change, related bugs, etc.).


## Running the local provider

1. Build a local version of the provider. At the root of the repository run:
   1. Run `make build` to build the provider.
   1. Run `make install` to install the provider locally. This installs the provider binary in the `~/.terraform.d/plugins` directory.
1. Create a terraform 
   1. Create a new directory with a `main.tf` file:
       ```bash
       mkdir -p ./terraform-provider-maas-dev
       cd ./terraform-provider-maas-dev
       touch main.tf
       ```
   1. Add the Terraform configuration below to `main.tf`. For more information, see [docs/index.md](docs/index.md):
       ```hcl
       terraform {
           required_providers {
               maas = {
                   source  = "registry.terraform.io/canonical/maas"
                   version = "=1.0.1"
               }
           }
       }

       provider "maas" {
           api_version = "2.0"
       }

       resource "maas_space" "tf_space" {
           name = "tf-space"
       }

       resource "maas_fabric" "tf_fabric" {
           name = "tf-fabric"
       }

       resource "maas_vlan" "tf_vlan" {
           fabric = maas_fabric.tf_fabric.id
           vid    = 14
           name   = "tf-vlan14"
           space  = maas_space.tf_space.name
       }

       resource "maas_subnet" "tf_subnet" {
           cidr       = "10.88.88.0/24"
           fabric     = maas_fabric.tf_fabric.id
           vlan       = maas_vlan.tf_vlan.vid
           name       = "tf_subnet"
           gateway_ip = "10.88.88.1"
           dns_servers = [
               "1.1.1.1",
           ]
           ip_ranges {
               type     = "reserved"
               start_ip = "10.88.88.1"
               end_ip   = "10.88.88.50"
           }
           ip_ranges {
               type     = "dynamic"
               start_ip = "10.88.88.200"
               end_ip   = "10.88.88.254"
           }
       }
       ```
1. Create and source your environment variables:
   1. In your environment running MAAS, obtain the MAAS API key:
       ```bash
       sudo maas apikey --username=maas
       ```
   1. Create an `env.sh` file and put in it the following variables: 
       ```shell
       export MAAS_API_KEY=<api-key>
       export MAAS_API_URL=<maas-api-url> # e.g. http://10.10.0.18:5240/MAAS/
       ```
   1. Run `source env.sh` to load the environment variables.
1. Use this configuration to start terraforming:
   1. Run `terraform fmt` to format the `main.tf` file.
   1. Run `terraform init` to initialize the provider.
   1. Run `terraform plan` to see the changes that will be applied.
   1. Run `terraform apply` to apply the changes. These should be reflected in the MAAS environment.
   1. Run `terraform destroy` to destroy the resources.

## Testing

- Ensure MAAS_API_KEY and MAAS_API_URL are set in your environment (see [Running the local provider](#running-the-local-provider)).
- Run the unit tests with:
    ```bash
    make test
    ```
- Run the unit tests and terraform acceptance tests with:
    ```bash
    make testacc
    ```
    Note that you may need to specify a specific machine or fabric to test against as other environment variables. Add these to your `env.sh` file before sourcing it again, if required:
    ```bash
    export TF_ACC=1
    export TF_ACC_FABRIC=<fabric_id> # e.g. 8
    export TF_ACC_NETWORK_INTERFACE_MACHINE=<machine_id> # e.g. b68rn4
    export TF_ACC_TAG_MACHINES=<machine_id> # e.g. b68rn4
    ```

## Getting Help

Check for existing issues [here](https://github.com/canonical/terraform-provider-maas/issues), or open a new one for bugs and feature requests.

## Release Process

Releases are handled by the maintainers, see [README.md](../README.md).

## Additional Resources

- [Terraform Provider Development](https://developer.hashicorp.com/terraform/plugin)
- [Go Documentation](https://golang.org/doc/)
- [MAAS API Documentation](https://maas.io/docs/api)
