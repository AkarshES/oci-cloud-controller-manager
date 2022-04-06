variable "canary_tenancies" {}
variable "integration_tenancies" {}


locals {
  canary_tenancies = jsondecode(replace(var.canary_tenancies, "&quot;", "\""))
  integration_tenancies = jsondecode(replace(var.integration_tenancies, "&quot;", "\""))
}

data tenancylookup_ocid canary {
  for_each = toset(local.canary_tenancies)
  name = each.value
}

data tenancylookup_ocid integration {
  for_each = toset(local.integration_tenancies)
  name = each.value
}

output "canary" {
  value = [for i in data.tenancylookup_ocid.canary: i.id]
}

output "integration" {
  value = [for i in data.tenancylookup_ocid.integration: i.id]
}