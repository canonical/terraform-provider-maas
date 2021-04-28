# Terraform Provider for MAAS

This repository contains the source code for the Terraform MAAS provider.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
- [Go](https://golang.org/doc/install) >= 1.16

## Build The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider with:

    ```sh
    make build
    ```

1. (Optional): Install the freshly built provider with:

    ```sh
    make install
    ```

## Usage

### Provider Configuration

The provider accepts the following config options:

- **api_key**: [MAAS API key](https://maas.io/docs/snap/3.0/cli/maas-cli#heading--log-in-required).
- **api_url**: URL for the MAAS API server (eg: <http://127.0.0.1:5240/MAAS>).
- **api_version**: MAAS API version used. It is optional and it defaults to `2.0`.

#### `maas`

```hcl
provider "maas" {
  api_version = "2.0"
  api_key = "YOUR MAAS API KEY"
  api_url = "http://<MAAS_SERVER>[:MAAS_PORT]/MAAS"
}
```

### Resource Configuration

#### `maas_instance`

It is used to deploy and release machines already registered and configured in MAAS based on the specified parameters. If no parameters are given, a random machine will be allocated and deployed using the defaults.

Example:

```hcl
resource "maas_instance" "two_random_nodes_2G" {
  count = 2
  allocate_params {
    min_cpu_count = 1
    min_memory = 2048
  }
}
```

Parameters:

| Name | Type | Required | Description
| ---- | ---- | -------- | -----------
| `allocate_params.min_cpu_count` | `string` | `false` | The minimum CPU cores count used for MAAS machine allocation.
| `allocate_params.min_memory` | `int` | `false` | The minimum RAM memory (in MB) used for MAAS machine allocation.
| `allocate_params.hostname` | `string` | `false` | Hostname used for MAAS machine allocation.
| `allocate_params.zone` | `string` | `false` | Zone name used for MAAS machine allocation.
| `allocate_params.pool` | `string` | `false` | Pool name used for MAAS machine allocation.
| `allocate_params.tags` | `[]string` | `false` | List of tag names used for MAAS machine allocation.
| `deploy_params.distro_series` | `string` | `false` | Distro series used to deploy the MAAS machine. It defaults to `focal`.
| `deploy_params.hwe_kernel` | `string` | `false` | Hardware enablement kernel to use with the image. Only used when deploying Ubuntu.
| `deploy_params.user_data` | `string` | `false` | Cloud-init user data script that gets run on the machine once it has deployed.
| `deploy_params.install_kvm` | `string` | `false` | Install KVM on machine.

#### `maas_pod`

Creates a new MAAS pod.

Example:

```hcl
resource "maas_pod" "kvm" {
  type = "virsh"
  power_address = "qemu+ssh://ubuntu@10.113.1.10/system"
  name = "kvm-host-01"
}
```

Parameters:

| Name | Type | Required | Description
| ---- | ---- | -------- | -----------
| `type` | `string` | `true` | The type of pod to create: `rsd` or `virsh`.
| `power_address` | `string` | `true` | Address that gives MAAS access to the pod's power control. For example: `qemu+ssh://172.16.99.2/system`.
| `power_user` | `string` | `false` | Username to use for power control of the pod. Required for `rsd` pods or `virsh` pods that do not have SSH set up for public-key authentication.
| `power_pass` | `string` | `false` | Password to use for power control of the pod. Required for `rsd` pods or `virsh` pods that do not have SSH set up for public-key authentication.
| `name` | `string` | `false` | The new pod's name.
| `zone` | `string` | `false` | The new pod's zone.
| `pool` | `string` | `false` | The name of the resource pool the new pod will belong to. Machines composed from this pod will be assigned to this resource pool by default.
| `tags` | `[]string` | `false` | A list of tags to assign to the new pod.
| `cpu_over_commit_ratio` | `float` | `false` | CPU overcommit ratio.
| `memory_over_commit_ratio` | `float` | `false` | RAM memory overcommit ratio.
| `default_macvlan_mode` | `string` | `false` |  Default macvlan mode for pods that use it: `bridge`, `passthru`, `private`, `vepa`.

### `maas_pod_machine`

Composes a new MAAS machine from an existing MAAS pod.

Example:

```hcl
resource "maas_pod_machine" "kvm" {
  pod = maas_pod.kvm.id
  cores = 1
  memory = 2048
  storage = "disk1:32,disk2:20"
}
```

Parameters:

| Name | Type | Required | Description
| ---- | ---- | -------- | -----------
| `pod` | `string` | `true` | The `id` or `name` of an existing MAAS pod.
| `cores` | `int` | `false` | The number of CPU cores (defaults to `1`).
| `pinned_cores` | `int` | `false` | List of host CPU cores to pin the VM to. If this is passed, the `cores` parameter is ignored.
| `memory` | `int` | `false` | The amount of memory, specified in MiB.
| `storage` | `string` | `false` | A list of storage constraint identifiers in the form `label:size(tag,tag,...),label:size(tag,tag,...)`. For more information, see [this](https://maas.io/docs/composable-hardware#heading--storage).
| `interfaces` | `string` | `false` | A labeled constraint map associating constraint labels with desired interface properties. MAAS will assign interfaces that match the given interface properties. For more information, see [this](https://maas.io/docs/composable-hardware#heading--interfaces).
| `hostname` | `string` | `false` | The hostname of the newly composed machine.
| `domain` | `string` | `false` | The name of the domain in which to put the newly composed machine.
| `zone` | `string` | `false` | The name of the zone in which to put the newly composed machine.
| `pool` | `string` | `false` | The name of the pool in which to put the newly composed machine.
