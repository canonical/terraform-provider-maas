# An existing SSH key can be imported using its id. e.g.
$ terraform import maas_ssh_keys.single_key 1

# Multiple SSH keys can be imported and managed together as a single resource by using "/". e.g.
$ terraform import maas_ssh_keys.multiple_keys 1/2/3

