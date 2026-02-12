resource "maas_package_repository" "repo1" {
  name = "my public repo"
  key  = "supersecretkey"
  url  = "http://foo.bar.com/foobar"

  disable_sources = true
  enabled         = true

  arches = ["amd64", "arm64", "ppc64el"]
  components = [
    "main"
  ]
  distributions = [
    "jammy-prod",
    "jammy-updates-prod",
    "jammy-security-prod"
  ]
}
