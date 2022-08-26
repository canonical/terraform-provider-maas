---
page_title: "MAAS Provider"
description: |-
  /* TODO: Add description of provider */
---

# MAAS Provider
/* TODO: Add detailed description of provider */<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `api_key` (String) The MAAS API key
- `api_url` (String) The MAAS API URL (eg: http://127.0.0.1:5240/MAAS)
- `api_version` (String) The MAAS API version (default 2.0)



## Example Usage

```terraform
terraform {
  required_providers {
    maas = {
      source  = "maas/maas"
      version = "~>1.0"
    }
  }
}

provider "maas" {
  api_version = "2.0"
  api_key = "<YOUR API KEY>"
  api_url = "http://127.0.0.1:5240/MAAS"
}
```