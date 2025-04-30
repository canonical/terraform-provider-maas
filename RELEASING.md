# Releasing

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
   
   The action creates a "draft" release. Go to [Releases](https://github.com/canonical/terraform-provider-maas/releases) to open it, select `edit`, and select `Publish release` if you are happy. 

3. Verify the release is published by: 
   1. Checking the release is now the latest published under [Releases](https://github.com/canonical/terraform-provider-maas/releases). 
   2. Checking the [HashiCorp Registry provider page](https://registry.terraform.io/providers/canonical/maas/latest) is displaying the released version as latest. This could take approximately 30 minutes to update.
