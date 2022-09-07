<!-- "MAAS Terraform provider reference" -->
If you wish to use MAAS with [Terraform](https://www.terraform.io/), there is a [provider avaiable](https://github.com/maas/terraform-provider-maas), supplied by a third party.  This article provides reference information about the data sources and resources available through the provider.  It does not attempt to explain the mechanics or usage of Terraform or offer any tutorial information related to this MAAS Terraform provider.

<a href="#heading--what-is-this"><h1 id="heading--what-is-this">The MAAS Terraform provider</h1></a>

The MAAS provider is a Terraform provider that allows you to manage MAAS resources using the Terraform (CRUD) tool. This provider can be used to manage many aspects of a MAAS environment, including networking, users, machines, and VM hosts.

These aspects can be divided into three categories of Terraform-compliant HCL:

- [API linkages](#heading--terraform-api-linkage)
- [Data sources](#heading--data-sources)
- [Resources](#heading--resources)

We will deal with each of these categories in turn.  For each data source and resource, we will offer a brief definition and description of how that item is employed in MAAS.  If you are new to [Terraform](https://www.terraform.io/), or want to explore what terraforming may provide for your MAAS instance, you may wish to consult the [Terraform documentation](https://www.terraform.io/intro) or one of the many [tutorials available](https://learn.hashicorp.com/collections/terraform/aws-get-started?utm_source=WEBSITE&utm_medium=WEB_IO&utm_offer=ARTICLE_PAGE&utm_content=DOCS).

<a href="#heading--terraform-api-linkage"><h2 id="heading--terraform-api-linkage">API linkages</h2></a>

The schema that provides an API linkage to MAAS from Terraform consists of a standard HCL provider block and a provider API block.  As with all Terraform providers, the provider block contains at least two items:

- a source element, which in this case is "maas/maas".
- a version element, which can be sufficiently specified by "~>1.0".

The provider block would look something like this:

```nohighlight
terraform {
  required_providers {
    maas = {
      source  = "maas/maas"
      version = "~>1.0"
    }
  }
}
```

The provider API block contains the necessary credentials to allow Terraform to access your MAAS instance, which include three things:

- an API version.
- an API key.
- an API URL.

A typical provider API block might look like this:

```nohighlight
provider "maas" {
  api_version = "2.0"
  api_key = "<YOUR API KEY>"
  api_url = "http://127.0.0.1:5240/MAAS"
}
```

A completed definition would also include some data sources and resources, like this typical example:

```nohighlight
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
  api_url = "<YOUR API URL>"
}

resource "maas_space" "tf_space" {
  name = "tf-space"
}

resource "maas_fabric" "tf_fabric" {
  name = "tf-fabric"
}

resource "maas_vlan" "tf_vlan" {
  fabric = maas_fabric.tf_fabric.id
  vid = 14
  name = "tf-vlan14"
  space = maas_space.tf_space.name
}
resource "maas_subnet" "tf_subnet" {
  cidr = "10.88.88.0/24"
  fabric = maas_fabric.tf_fabric.id
  vlan = maas_vlan.tf_vlan.vid
  name = "tf_subnet"
  gateway_ip = "10.88.88.1"
  dns_servers = [
    "1.1.1.1",
  ]
  ip_ranges {
    type = "reserved"
    start_ip = "10.88.88.1"
    end_ip = "10.88.88.50"
  }
  ip_ranges {
    type = "dynamic"
    start_ip = "10.88.88.200"
    end_ip = "10.88.88.254"
  }
}
```

See the [Terraform HCL documentation](https://www.terraform.io/language) for more details about these blocks.

<a href="#heading--data-sources"><h2 id="heading--data-sources">Data sources</h2></a>

The MAAS Terraform provider offers three data sources, all representing network elements:

- a [fabric](https://discourse.maas.io/t/maas-concepts-and-terms-reference/5416#heading--fabrics), which is essentially a VLAN namespace -- that is, it connects two or more VLANs together.
- a [subnet](https://discourse.maas.io/t/maas-concepts-and-terms-reference/5416#heading--subnets), which is the traditional way of dividing up IP addresses into smaller networks, e.g., 192.168.15.0/24.
- a [VLAN](https://en.wikipedia.org/wiki/VLAN), a "virtual LAN", which is a collection of specific addresses or ports that are connected together to form a restricted network.

Each of these data sources has a specific HCL block with elements structured appropriately to manage that MAAS element.

<a href="#heading--fabric"><h3 id="heading--fabric">Fabric</h3></a>

The [fabric](https://github.com/maas/terraform-provider-maas/blob/master/docs/data_sources/maas_fabric.md) data source provides minimal details, namely, the fabric ID, of an existing MAAS fabric.  It takes one argument (the fabric name) and exports one attribute (the fabric ID):

```nohighlight
data "maas_fabric" "default" {
  name = "maas"
}
```

Fabrics within MAAS are not widely manipulated in and of themselves, but rather serve as containers for storing VLAN/subnet combinations.

<a href="#heading--subnet"><h3 id="heading--subnet">Subnet</h3></a>

The [subnet](https://github.com/maas/terraform-provider-maas/blob/master/docs/data_sources/maas_subnet.md) data source provides a number of details about an existing MAAS network subnet.  The element takes one argument, the subnet CIDR, and exports a number of attributes:

- id - The subnet ID.
- fabric - The subnet fabric.
- vid - The subnet VLAN traffic segregation ID.
- name - The subnet name.
- rdns_mode - How reverse DNS is handled for this subnet. It can have one of the following values:
-- 0 - Disabled, no reverse zone is created.
-- 1 - Enabled, generate reverse zone.
-- 2 - RFC2317, extends 1 to create the necessary parent zone with the appropriate CNAME resource records for the network, if the network is small enough to require the support described in RFC2317.
- allow_dns - Boolean value that indicates if the MAAS DNS resolution is enabled for this subnet.
- allow_proxy - Boolean value that indicates if maas-proxy allows requests from this subnet.
- gateway_ip - Gateway IP address for the subnet.
- dns_servers - List of IP addresses set as DNS servers for the subnet.

Declaring a subnet looks something like this example:

```nohighlight
data "maas_subnet" "vid10" {
  cidr = "10.10.0.0/16"
}
```

Subnets are the network backbone of MAAS, and thus provide a number of attributes that can be manipulated to alter the behavior of MAAS.

<a href="#heading--vlan"><h3 id="heading--vlan">VLAN</h3></a>

The [VLAN](https://github.com/maas/terraform-provider-maas/blob/master/docs/data_sources/maas_vlan.md) data source provides details about an existing MAAS VLAN.  A VLAN takes two arguments:

- fabric - (Required) The fabric identifier (ID or name) for the VLAN.
- vlan - (Required) The VLAN identifier (ID or traffic segregation ID).

A VLAN data source exports a few useful attributes:

- mtu - The MTU used on the VLAN.
- dhcp_on - Boolean value indicating if DHCP is enabled on the VLAN.
- name - The VLAN name.
- space - The VLAN space.

VLAN [spaces](https://discourse.maas.io/t/maas-concepts-and-terms-reference/5416#heading--spaces) are used mostly by Juju, but can be employed by other tools, if desired.

The typical definition of a MAAS VLAN in HCL might look like this:

```nohighlight
data "maas_vlan" "vid10" {
  fabric = data.maas_fabric.default.id
  vlan = 10
}
```

VLANs are available as data sources, but generally, subnets are the workhorses of most MAAS instances.

<a href="#heading--resources"><h2 id="heading--resources">Resources</h2></a>

The MAAS Terraform provider makes a large number of resources available, currently including the following items.  Because of the large number of items, details of arguments and attributes are not duplicated here, but instead provided from a single source at the given links:

- A [maas_instance](https://github.com/maas/terraform-provider-maas/blob/master/docs/resources/maas_instance.md) provides a resource to deploy and release machines already configured in MAAS, based on the specified parameters. If no parameters are given, a random machine will be allocated and deployed using the defaults.
- A [maas_vm_host](https://github.com/maas/terraform-provider-maas/blob/master/docs/resources/maas_vm_host.md) provides a resource to manage MAAS VM hosts.  Note that MAAS VM hosts are not machines, but the host(s) upon which virtual machines are created.
- A [maas_vm_host_machine](https://github.com/maas/terraform-provider-maas/blob/master/docs/resources/maas_vm_host_machine.md) provides a resource to manage MAAS VM host machines, which represent the individual machines that are spun up on a given VM host.
- A [maas_machine](https://github.com/maas/terraform-provider-maas/blob/master/docs/resources/maas_machine.md) provides a resource to manage MAAS machines; note that these are typically physical machines (rather than VMs), so they tend to respond differently at times.
- A [maas_network_interface_physical](https://github.com/maas/terraform-provider-maas/blob/master/docs/resources/maas_network_interface_physical.md) provides a resource to manage a physical network interface from an existing MAAS machine.  Network interfaces can be created and deleted at will via the MAAS CLI/UI, so there may be more than one of these associate with any given machine.
- A [maas_network_interface_link](https://github.com/maas/terraform-provider-maas/blob/master/docs/resources/maas_network_interface_link.md) provides a resource to manage network configuration on a network interface.  Note that this does not represent the interface itself, but the parameter set that configure that interface.
- A [maas_fabric](https://github.com/maas/terraform-provider-maas/blob/master/docs/resources/maas_fabric.md) provides a resource to manage MAAS network fabrics, which are [described above](#heading--fabric). 
- A [maas_vlan](https://github.com/maas/terraform-provider-maas/blob/master/docs/resources/maas_vlan.md) provides a resource to manage MAAS network VLANs, also [described above](#heading--vlan).
- A [maas_subnet](https://github.com/maas/terraform-provider-maas/blob/master/docs/resources/maas_subnet.md) provides a resource to manage MAAS network subnets, also [described above](#heading--subnet)
- A [maas_subnet_ip_range](https://github.com/maas/terraform-provider-maas/blob/master/docs/resources/maas_subnet_ip_range.md) provides a resource to manage MAAS network subnets IP ranges.  IP ranges carry particular importance when managing DHCP with multiple DHCP servers, for example.
- A [maas_dns_domain](https://github.com/maas/terraform-provider-maas/blob/master/docs/resources/maas_dns_domain.md) provides a resource to manage MAAS DNS domains.
- A [maas_dns_record](https://github.com/maas/terraform-provider-maas/blob/master/docs/resources/maas_dns_record.md) provides a resource to manage MAAS DNS domain records.
- A [maas_space](https://github.com/maas/terraform-provider-maas/blob/master/docs/resources/maas_space.md) provides a resource to manage MAAS network [spaces](https://juju.is/docs/olm/network-spaces).
- A [maas_block_device](https://github.com/maas/terraform-provider-maas/blob/master/docs/resources/maas_block_device.md) provides a resource to manage block devices on MAAS machines.
- A [maas_tag](https://github.com/maas/terraform-provider-maas/blob/master/docs/resources/maas_tag.md) provides a resource to manage a MAAS tag.  MAAS tags have multiple roles in controlling how machines are configured, booted, and monitored.
- A [maas_user](https://github.com/maas/terraform-provider-maas/blob/master/docs/resources/maas_user.md) provides a resource to manage MAAS users.  This resource does not provide any control over any Candid or RBAC restrictions that may be in place.

Please visit the links to get details on these resources, since the documentation at those links will always be the most current information available.
