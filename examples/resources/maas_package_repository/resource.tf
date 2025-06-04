resource "maas_package_repository" "repo1" {
  name = "my public repo"
  key  = "supersecretkey"
  url  = "http://repo_url:8000/public"

  disable_sources = true
  enabled         = true

  arches = ["amd64", "arm64", "ppcel64"]
  components = [
    "main"
  ]
  distributions = [
    "jammy-prod",
    "jammy-updates-prod",
    "jammy-security-prod"
  ]
}

