resource "maas_user" "cloudbase" {
  name     = "cloudbase"
  password = "Passw0rd123"
  email    = "admin@cloudbase.local"
  is_admin = true
}
