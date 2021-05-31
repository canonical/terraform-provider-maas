# `maas_instance`

It is used to deploy and release machines already registered and configured in MAAS based on the specified parameters. If no parameters are given, a random machine will be allocated and deployed using the defaults.

Example:

```hcl
resource "maas_instance" "two_random_nodes_2G" {
  count = 2
  allocate_min_cpu_count = 1
  allocate_min_memory = 2048
  deploy_distro_series = "focal"
}
```

Parameters:

| Name | Type | Required | Description
| ---- | ---- | -------- | -----------
| `allocate_min_cpu_count` | `string` | `false` | The minimum CPU cores count used for MAAS machine allocation.
| `allocate_min_memory` | `int` | `false` | The minimum RAM memory (in MB) used for MAAS machine allocation.
| `allocate_hostname` | `string` | `false` | Hostname used for MAAS machine allocation.
| `allocate_zone` | `string` | `false` | Zone name used for MAAS machine allocation.
| `allocate_pool` | `string` | `false` | Pool name used for MAAS machine allocation.
| `allocate_tags` | `[]string` | `false` | List of tag names used for MAAS machine allocation.
| `deploy_distro_series` | `string` | `false` | Distro series used to deploy the MAAS machine. It defaults to `focal`.
| `deploy_hwe_kernel` | `string` | `false` | Hardware enablement kernel to use with the image. Only used when deploying Ubuntu.
| `deploy_user_data` | `string` | `false` | Cloud-init user data script that gets run on the machine once it has deployed.
| `deploy_install_kvm` | `string` | `false` | Install KVM on machine.
