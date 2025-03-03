module "properties_values" {
  source             = "./shared_modules/properties_values/"
  execution_target   = local.execution_target
  spectre_group_name = lookup(local.execution_target.additional_locals, "spectre_group_name")
  env                = lookup(local.execution_target.additional_locals, "env", "")
  realm              = local.execution_target.region.realm
}

resource "capability_require_capability" "oke_ccm_csi_internal_capability" {
  name  = "oke_ccm_csi_internal_capability"
}