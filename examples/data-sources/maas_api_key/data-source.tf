data "maas_api_key" "existing" {
  name = "my-automation-token"
}

output "existing_api_key" {
  value     = data.maas_api_key.existing.api_key
  sensitive = true
}
