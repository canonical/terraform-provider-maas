#!/bin/bash

ENV_DIR=$PWD/.devenv

if [ -d "$ENV_DIR" ]; then
  echo ""
  echo "A development folder already exists at $ENV_DIR."
  echo "Would you like to overwrite it? (y/n)"
  read -r OVERWRITE_CHOICE
  echo ""

  if [[ ! "$OVERWRITE_CHOICE" =~ ^[Yy]$ ]]; then
    echo "Exiting."
    echo ""
    exit 0
  fi
fi

if [ -d "$ENV_DIR" ]; then
  rm -rf "$ENV_DIR"
fi
mkdir -p "$ENV_DIR"
cat << EOF > $ENV_DIR/main.tf
terraform {
  required_providers {
    maas = {
      source  = "registry.terraform.io/canonical/maas"
      # A version is not required during development of the provider.
    }
  }
}

provider "maas" {
  api_version         = "2.0"
  api_key             = ... # Fill me in from MAAS
  api_url             = ... # Fill me in from MAAS
  installation_method = "snap"
}

# Try me out!
# resource "maas_fabric" "myfabric" {
#   name = "myfabric"
# }

EOF

echo ""
echo "A development directory has been created at $ENV_DIR."
echo "This is a space to get started with during development. Fill in the relevant fields in the provider block from MAAS, and you should be able to start terraforming in MAAS."
echo ""
echo "Happy Terraforming! 🚀"
echo ""
