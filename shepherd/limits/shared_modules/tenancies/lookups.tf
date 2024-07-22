variable "canary_tenancies" {}
variable "integration_tenancies" {}


locals {
  canary_tenancies = jsondecode(replace(var.canary_tenancies, "&quot;", "\""))
  integration_tenancies = jsondecode(replace(var.integration_tenancies, "&quot;", "\""))
}

// constants
locals {
  okecustprod = "okecustprod"
}

data tenancylookup_ocid canary {
  for_each = toset(local.canary_tenancies)
  name = each.value
}

data tenancylookup_ocid integration {
  for_each = toset(local.integration_tenancies)
  name = each.value
}

data "tenancylookup_ocid" "okecust_tenancy" {
  name = local.okecustprod
}

data "oci_objectstorage_namespace" "custprod_namespace" {
  compartment_id = data.tenancylookup_ocid.okecust_tenancy.id
}

output "canary" {
  value = [for i in data.tenancylookup_ocid.canary: i.id]
}

output "integration" {
  value = [for i in data.tenancylookup_ocid.integration: i.id]
}
output tenancy_ocid_map {
  value = {
    okecust_tenancy_ocid = data.tenancylookup_ocid.okecust_tenancy.id
  }
}

output tenancy_os_namespace_map {
  value = {
    okecust_tenancy_os_namespace = data.oci_objectstorage_namespace.custprod_namespace.namespace
  }
}
