# Releasing

Builds and releases are automated with GitHub Actions and GoReleaser. There are several manual steps to complete:

1. Decide on the version number by reviewing commits since the last release. The provider versions follow [semantic versioning](https://semver.org/).

1. Start the release action:

   Assuming `upstream` is the remote name pointing to the `canonical/terraform-provider-maas` repository:
   1. First, checkout the master branch and pull the latest changes:

      ```shell
      git switch master
      git pull upstream master
      ```
   1. Verify the latest commit is the same as the latest on the remote master branch.

   1. Create a new tag and push it to the remote repository to trigger the release action:

      ```shell
      git tag vX.Y.Z
      git push upstream tag vX.Y.Z
      ```
   The release action can be viewed under [Actions](https://github.com/canonical/terraform-provider-maas/actions/workflows/release.yml).

1. Publish the release:

   The GitHub Action creates a "draft" release. Go to [Releases](https://github.com/canonical/terraform-provider-maas/releases) to open it, select `edit`, and select `Publish release` if you are happy.

1. Verify the release is published:
   1. Check the release is now the latest published under [Releases](https://github.com/canonical/terraform-provider-maas/releases).
   1. Check the [HashiCorp Registry provider page](https://registry.terraform.io/providers/canonical/maas/latest) is displaying the released version as latest. This could take approximately 30 minutes to update.
