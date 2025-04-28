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
   
   The action creates a "draft" release. Go to [Releases](https://github.com/canonical/terraform-provider-maas/releases) to open it, select `edit` and click `Generate release notes`. Select `Publish release` when you are happy. 

3. Verify the published release looks good by checking in [Releases](https://github.com/canonical/terraform-provider-maas/releases).