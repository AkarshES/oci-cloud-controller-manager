
module "merged_cell_config" {
  source         = "./configuration/merged_cell_config"
  flock_config   = local.flock_config
  overrides      = local.overrides
  cell_overrides = local.cell_overrides
}

resource "shepherd_resource_lock_rules" oke_mp_custom_rules {
  rule {
    name                    = "auto_unlock_resources_in_dev"
    resource_type_filter    = "/.*/"
    execution_target_filter = "/*dev*/"
    default_locked          = false
  }
  rule {
    name                    = "auto_unlock_resources_in_dev"
    resource_type_filter    = "/.*/"
    execution_target_filter = "/*integ*/"
    default_locked          = false
  }
  rule {
    name                    = "auto_unlock_resources_in_dev"
    resource_type_filter    = "/.*/"
    execution_target_filter = "/*legacy*/"
    default_locked          = false
  }
  rule {
    name                    = "lock_odo_pool_resources"
    resource_type_filter    = "odo_pool"
    execution_target_filter = "/*prd*/"
    default_locked          = true
  }
}