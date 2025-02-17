# Contributing to the Terraform MAAS Provider

Thank you for your interest in contributing to the Terraform MAAS Provider! We appreciate your help in making this project better.

## Getting Started

### Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.4.x
- [Go](https://golang.org/doc/install) >= 1.22
- A MAAS installation running. See the [maas-dev-setup](https://github.com/canonical/maas-dev-setup) repository for a quick setup.

### Setting Up Your Development Environment

1. Fork the repository on GitHub.
1. Clone your forked repository to your local machine:
   ```bash
   git clone <https-or-ssh-url>
   ```
1. Add the original repository as upstream:
   ```bash
   git remote add upstream <https-or-ssh-url-to-original-repo>
   ```
1. Verify you have two remotes:
   ```bash
   $ git remote -v
   origin    git@github.com:username/terraform-provider-maas.git (fetch)
   origin    git@github.com:username/terraform-provider-maas.git (push)
   upstream  git@github.com:maas/terraform-provider-maas.git (fetch)
   upstream  git@github.com:maas/terraform-provider-maas.git (push)
   ```

## Development Workflow

1. Create a new branch for your work:
   ```bash
   git checkout -b feat/your-feature-name
   ```
1. Make your changes, following our coding standards.
1. Write or update tests as needed.
1. Commit and push your changes to your fork.
1. Raise a PR (see below).

## Pull Request Process

1. Update your fork with the latest upstream changes. We recommend merging the upstream master branch into your fork master branch. On your feature branch:
   ```bash
   git fetch upstream
   git merge upstream/master
   ```
1. Push your changes to your fork:
   ```bash
   git push origin feat/your-feature-name
   ```
1. Open a Pull Request through the GitHub interface, with `canonical/terraform-provider-maas` as the base repository, and `master` as the target branch.
1. Ensure your PR has a clear title and description. 

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

## Running the provider
See [docs/index.md](../docs/index.md) for more information.

## Getting Help

Open an issue for bugs or feature requests.

## Release Process

Releases are handled by the maintainers, see [README.md](../README.md).

## Additional Resources

- [Terraform Provider Development](https://www.terraform.io/docs/extend/writing-custom-providers.html)
- [Go Documentation](https://golang.org/doc/)
- [MAAS API Documentation](https://maas.io/docs/api)
