module "spectre" {
  source = "./shared_modules/spectre"
  root_compartment_ocid = local.execution_target.tenancy_ocid
  group_name = lookup(local.execution_target.additional_locals, "spectre_group_name")
  env = lookup(local.execution_target.additional_locals, "env")
  realm = local.execution_target.region.realm
  region = local.execution_target.region.name
}

resource "capability_require_capability" "oke_limits_management" {
  name = "oke_limits_management"
}