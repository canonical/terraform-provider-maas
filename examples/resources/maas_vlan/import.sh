# Existing MAAS VLANs can be imported using the fabric identifier (ID or name) and the VLAN identifier (ID or traffic segregation ID; if a value matches both, the traffic segregation ID wins). e.g.
$ terraform import maas_vlan.tf_vlan tf-fabric:14
