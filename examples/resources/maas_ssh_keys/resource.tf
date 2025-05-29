# Manage a single key in MAAS
resource "maas_ssh_keys" "single_key" {
  keys = [
    "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIP55TGiiUJ8ShRvhvg1tq2Rrhn4fjbzy7hYAopT6QVYE"
  ]
}

# Import all keys from a Launchpad user and manage them together
resource "maas_ssh_keys" "from_launchpad" {
  keysource = "lp:mylaunchpadid"
}

# Import all keys from a GitHub user and manage them together
resource "maas_ssh_keys" "from_github" {
  keysource = "gh:mygithubusername"
}

# Manage multiple keys in MAAS together
resource "maas_ssh_keys" "multiple_keys" {
  keys = [
    "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIP55TGiiUJ8ShRvhvg1tq2Rrhn4fjbzy7hYAopT6QVYE",
    "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAINGSy3uNIEXvrmSc96uqqbLt1iNHK2HOC8YtFmPADZye",
  ]
}
