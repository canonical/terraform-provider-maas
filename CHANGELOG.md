## 1.0.1 (Unreleased)

IMPROVEMENTS:

* Updates documentation structure to use terraform-plugin-docs tool

## 1.0.0 (Jul 21, 2021)

NEW:

* Resource `maas_block_device`
* Resource `maas_fabric`
* Resource `maas_vlan`
* Resource `maas_subnet`
* Resource `maas_space`
* Resource `maas_subnet_ip_range`
* Resource `maas_dns_domain`
* Resource `maas_dns_record`
* Resource `maas_user`
* Resource `Implement importers for the existing managed resources:`
* Resource `maas_machine`
* Resource `maas_instance`
* Resource `maas_tag`
* Resource `maas_network_interface_physical`
* Resource `maas_vm_host`
* Resource `maas_vm_host_machine`

IMPROVEMENTS:

* Allow env vars `MAAS_API_KEY` and `MAAS_API_URL` to be used for the provider configuration.
* Use VM host naming instead of Pod.
* Add validation for maas_machine resource power_type argument.
* Update VM host machine network and storage params.
* Properly implement the network and storage parameters for the `maas_vm_host_machine` resource.
* Remove managed argument from `maas_subnet` resource and data source. This is considered `true` by default on MAAS 2.0 and newer.
* Update docs and examples.