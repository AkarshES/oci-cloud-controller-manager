variable "flock_config" {}
variable "overrides" {}
variable "qualified_realm_name" {}

locals {
  config = merge(
        var.flock_config,
        lookup(var.overrides, split(".", var.qualified_realm_name)[0], {}),
        lookup(var.overrides, var.qualified_realm_name, {}),
  )
}

output "config" {
  value = local.config
}
