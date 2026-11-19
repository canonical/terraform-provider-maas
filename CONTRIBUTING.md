# Contributing to the Terraform MAAS Provider

Thank you for your interest in contributing to the Terraform MAAS Provider! We appreciate your help in making this project better. This document provides information on how to set up your development environment and best practices for contributing to the project.


## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.4.x
- [Go](https://golang.org/doc/install) >= 1.23
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


## Create a development environment
There are two options for developing the provider:
- **Locally**, with Go and Terraform installed on your machine, as described in this section and in [Testing](#testing).
- **In a workshop**, an isolated development environment with the toolchain pre-installed. See [Develop in a workshop](#develop-in-a-workshop).

### Setup
1. Run `make build` to build the provider binary locally, located in `./bin`.
1. Run `make create_dev_overrides` and follow any output instructions. More info [here](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers).
1. Run `make create_dev_env` to create a development directory and follow any output instructions.
1. In your `.devenv` directory, use these commands to get started:
   1. Run `terraform fmt` to format the `main.tf` file.
   1. Run `terraform init` to initialize the provider.
   1. Run `terraform plan` to see the changes that will be applied.
   1. Run `terraform apply` to apply the changes. These should be reflected in the MAAS environment.
   1. Run `terraform destroy` to destroy the resources.

### Workflow
Assuming you have already setup dev-overrides:
1. Make a change to the provider.
1. Rebuild the provider binary locally with `make build`.
1. In your dev-env directory, you can immediately run `terraform apply` with your new changes.

## Testing
Tests are written as advised in the [Terraform docs](https://developer.hashicorp.com/terraform/plugin/sdkv2/testing). They are split into unit tests and acceptance tests, with the latter creating real resources in the MAAS environment. Therefore, you will need to ensure MAAS is running locally for these tests to pass.

To run the tests:
1. Ensure `MAAS_API_KEY`, `MAAS_API_URL`, and `MAAS_VERSION` environment variables are set in your shell, corresponding to your running MAAS installation.

2. Run the tests:
    - Run the unit tests:
        ```bash
        make test
        ```
    - Run both the unit tests and all Terraform acceptance tests:
        ```bash
        make testacc
        ```
        > [!NOTE]
        > You may need to set specific environment variables for some tests to pass, for example machine ids. Add these to your `env.sh` file before sourcing it again, if required:
        > ```bash
        > export TF_ACC_NETWORK_INTERFACE_MACHINE=<system_id>   # b68rn4
        > export TF_ACC_TAG_MACHINES=<system_id>                # b68rn4
        > export TF_ACC_VM_HOST_ID=<system_id>                  # maas-host
        > export TF_ACC_BLOCK_DEVICE_MACHINE=<system_id>        # b68rn4
        > export TF_ACC_RACK_CONTROLLER_HOSTNAME=<name>         # maas-dev
        > ```
    - Run a specific acceptance test:
        ```bash
        make testacc TESTARGS="-run=TestAcc<resource_name>_basic"
        ```
    - Run a specific acceptance test without using the cached result:
        ```bash
        make testacc TESTARGS="-run=TestAcc<resource_name>_basic -count=1"
        ```
    - Run a group of acceptance tests with names matching the regex:
        ```bash
        make testacc TESTARGS="-run=TestAcc<resource_name>"
        ```

### Cleaning Up Test Resources

If acceptance tests fail or are interrupted, they may leave resources in your MAAS environment. Use sweepers to clean up these leftover test resources:

```bash
make sweep
```

**Important:** Sweepers will delete infrastructure. Only run them in test environments.

Sweepers require the same environment variables as acceptance tests (`MAAS_API_KEY`, `MAAS_API_URL`) plus any resource-specific variables (e.g., `TF_ACC_NETWORK_INTERFACE_MACHINE`). Test resources are identified by naming patterns (e.g., `tf-nic-*` for network interfaces) and will be automatically removed.

To run specific sweepers or pass additional options:

```bash
make sweep SWEEPARGS="-sweep-run=maas_network_interface"
```

## Develop in a workshop

As an alternative to a local setup, you can develop in a [workshop](https://ubuntu.com/workshop/docs/): an LXD container defined in `.workshop/dev.yaml` that comes with Go, Terraform, `make`, and `golangci-lint` installed. Your checkout is mounted into the workshop at `/project`, so you keep editing and using git on your host as usual.

### Setup
1. Install Workshop by following the [Workshop documentation](https://ubuntu.com/workshop/docs/).
1. From the repository root, launch the workshop:
    ```bash
    workshop launch dev
    ```
1. Check it is ready:
    ```bash
    workshop info dev
    ```

> [!NOTE]
> If the launch fails to resolve or download packages, your host firewall may be blocking traffic on the `workshopbr0` bridge. Run `workshop warnings` for details. With UFW, allow it with:
> ```bash
> sudo ufw allow in on workshopbr0
> sudo ufw route allow in on workshopbr0
> sudo ufw route allow out on workshopbr0
> ```

### Workflow
The common tasks are available as workshop actions, which wrap the `make` targets:

| Action | Runs |
|--------|------|
| `build` | `make build` |
| `test` | `make test` |
| `testacc` | `make testacc` |
| `lint` / `lint-fix` | `make lint` / `make lint-go-fix` |
| `fmt` / `fmt-check` | `terraform fmt -recursive` (with `-check`) |
| `generate` | `go generate ./...` and `make generate_docs` |
| `docs-validate` | `tfplugindocs validate` |
| `dev-overrides` | `make create_dev_overrides` |

Run an action with:
```bash
workshop run dev build
```

To work interactively, for example to use Terraform with your development build, open a shell in the workshop:
```bash
workshop run dev dev-overrides
workshop shell
# inside the workshop:
export TF_CLI_CONFIG_FILE=/project/bin/developer_overrides.tfrc
make create_dev_env
cd /project/.devenv && terraform init && terraform apply
```

### Testing
Unit tests need no extra setup:
```bash
workshop run dev test
```

Acceptance tests need the same environment variables as a local run. Set them in your host shell and pass them into the workshop by name with `--env`:
```bash
export MAAS_API_URL=http://<maas-address>:5240/MAAS
export MAAS_API_KEY=<api-key>
export MAAS_VERSION=<maas-version>

workshop run --env MAAS_API_URL --env MAAS_API_KEY --env MAAS_VERSION \
  dev testacc TESTARGS="-run=TestAcc<resource_name>_basic -count=1"
```

Pass any `TF_ACC_*` variables that a test needs the same way, with an extra `--env` flag for each.

> [!NOTE]
> `localhost` inside the workshop refers to the workshop itself, not your host. Set `MAAS_API_URL` to an address the workshop can reach, such as the address of your MAAS on its bridge network or LAN.

To run sweepers, use `workshop shell`, export the variables inside the workshop, and run `make sweep`.

## Getting Help

Check for existing issues [here](https://github.com/canonical/terraform-provider-maas/issues), or open a new one for bugs and feature requests.

## Release Process

Releases are handled by the maintainers, see [RELEASING.md](RELEASING.md).

## Additional Resources

- [Terraform Provider Development](https://developer.hashicorp.com/terraform/plugin)
- [Go Documentation](https://golang.org/doc/)
- [MAAS API Documentation](https://maas.io/docs/api)
