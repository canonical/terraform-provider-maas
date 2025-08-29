resource "maas_user" "maas_admin" {
  name     = "maas_admin"
  password = "P8ssw0rd12"
  email    = "admin@maas.com"
  is_admin = true
}

resource "maas_user" "cloudbase" {
  name             = "cloudbase"
  password         = "Passw0rd123"
  email            = "admin@cloudbase.local"
  is_admin         = true
  transfer_to_user = maas_admin.name
}
