resource "maas_api_key" "my_token" {
  name = "my-automation-token"
}

output "api_key" {
  value     = maas_api_key.my_token.api_key
  sensitive = true
}
