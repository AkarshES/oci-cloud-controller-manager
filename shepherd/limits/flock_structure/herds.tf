module "herds-oc1-config" {
  source               = "./configuration/merged_realm_config"
  flock_config         = local.flock_config
  overrides            = local.overrides
  qualified_realm_name = "herds.oc1"
}

locals {
  herds_scalar = contains(keys(local.prod_realm_by_name), "oc1") ? 1 : 0
  herds_region = ["eu-frankfurt-1"]
  herds_cell_overrides = local.herds_scalar == 1 ? {
    for key, value in local.cell_overrides : key => value if split(".", key)[0] == "herds" && split(".", key)[1] == "oc1"
  } : {}
  herds_env_setup_ets = local.herds_scalar == 1 ? {
    "herds.oc1" = {
      cell_count        = 1
      realm             = "oc1"
      env               = "rbaas"
      region            = "eu-frankfurt-1"
      additional_locals = module.herds-oc1-config.config
    }
  } : {}
  herds_spectre_setup_ets = local.herds_scalar == 1 ? {
    "herds.oc1" = {
      realm             = "oc1"
      env               = "rbaas"
      region            = "us-phoenix-1"
      additional_locals = module.herds-oc1-config.config
    }
  } : {}
  herds_spectre_regional_ets = local.herds_scalar == 1 ? toset([for key in local.spectre_regional_et : key if ! contains(local.build_regions_nocell, key) && split(".", key)[0] == "herds" && split(".", key)[1] == "oc1"]) : toset([])
  #labels_to_watch = [ for idx in range(each.value.cell_count) : format("oke-mp-release-cells%s",idx)]
}


resource "shepherd_release_phase" "herds_oc1" {
  count        = local.herds_scalar
  name         = "herds.oc1"
  realm        = "oc1"
  production   = false
  predecessors = []
}

resource "shepherd_release_phase" "herds_oc1_regional" {
  count        = local.herds_scalar * length(local.herds_region)
  name         = "herds.${local.herds_region[count.index]}"
  realm        = "oc1"
  production   = false
  predecessors = [shepherd_release_phase.herds_oc1[0].name]
}

resource "shepherd_execution_target" "herds_et" {
  for_each                  = local.herds_cell_overrides
  name                      = each.key
  region                    = split(".", each.key)[2]
  predecessors              = lookup(each.value, "predecessor", "") != "" ? [lookup(each.value, "predecessor", "")] : tonumber(split("cell", each.key)[1]) > 0 ? [format("%s.cell%s", split(".cell", each.key)[0], tonumber(split(".cell", each.key)[1]) - 1)] : []
  phase                     = lookup(merge(lookup(local.overrides, each.key, {})), "phase", join(".", ["herds", split(".", each.key)[2]]))
  uniquifier                = lookup(module.merged_cell_config.uniquifiers, each.key, "")
  tenancy_name              = lookup(lookup(local.overrides.tenancy_info, "rbaas", {}), split(".", each.key)[1], local.overrides.tenancy_info.default)
  snowflake_config_location = lookup(module.merged_cell_config.snowflake_config_locations, each.key, "")
  additional_locals         = merge({
    limits_region          = lower(lookup(local.region_by_name_all_regions, split(".", each.key)[2]).airport_code)
    manage_regional_values = "true"
    manage_definitions     = "false"
    pool_name_regex = "^oke-deploy-rbaas[0-9]*"
  }, lookup(module.merged_cell_config.additional_locals, each.key, {}))
  alarms_to_watch {
    compartment_name = "assets"
    labels           = [format(lookup(lookup(module.merged_cell_config.additional_locals, each.key, {}), "watch_mp_release_label_format"), split(lookup(lookup(module.merged_cell_config.additional_locals, each.key, {}), "cell_name_prefix"), each.key)[1])]
  }
  ignored_region_build_capabilities = ["grafana_dashboard"]
  labels = {
    herd = "993ddcbe-99c5-49ac-a791-889537ecb67a"
  }
  provider_override {
    name = "null"
    constraint = ">= 0.1"
  }
}

resource "shepherd_execution_target" "herds_env_setup_et" {
  for_each                          = local.herds_env_setup_ets
  name                              = format("env.setup.%s", each.key)
  region                            = each.value.region
  phase                             = lookup(each.value, "phase", each.key)
  predecessors                      = []
  uniquifier                        = format("env-setup-%s", replace(each.key, ".", "-"))
  tenancy_name                      = lookup(lookup(local.overrides.tenancy_info, each.value.env, {}), each.value.realm, local.overrides.tenancy_info.default)
  snowflake_config_location         = "generic_tenancy"
  additional_locals                 = merge(each.value.additional_locals, { cell_count : each.value.cell_count })
  ignored_region_build_capabilities = ["grafana_dashboard"]
  alarms_to_watch {
    compartment_name = "assets"
    labels = [ for idx in range(each.value.cell_count) : format("oke-mp-release-cells%s",idx)]
    #labels           = ["oke-mp-release-cell0", "oke-mp-release-cell1"]
  }
  labels = {
    herd = "993ddcbe-99c5-49ac-a791-889537ecb67a"
  }
}

resource "shepherd_execution_target" "herds_spectre_setup_et" {
  for_each                          = local.herds_spectre_setup_ets
  name                              = format("spectre.setup.%s", each.key)
  region                            = local.home_region_by_realm[split(".", each.key)[1]]
  phase                             = lookup(each.value, "phase", each.key)
  predecessors                      = ["env.setup.${each.key}"]
  uniquifier                        = format("spectre-setup-%s", replace(each.key, ".", "-"))
  tenancy_name                      = lookup(lookup(local.overrides.tenancy_info, each.value.env, {}), each.value.realm, local.overrides.tenancy_info.default)
  snowflake_config_location         = "spectre_region"
  additional_locals                 = each.value.additional_locals
  scope                             = format("%s%s", "eu-frankfurt-1", "~home-region")
  ignored_region_build_capabilities = ["grafana_dashboard"]
  alarms_to_watch {
    compartment_name = "assets"
    labels           = ["oke-mp-release-cell0", "oke-mp-release-cell1"]
  }
  labels = {
    herd = "993ddcbe-99c5-49ac-a791-889537ecb67a"
  }
}

resource "shepherd_execution_target" "herds_region_values" {
  for_each                  = local.herds_spectre_regional_ets
  name                      = format("spectre.values.%s", each.key)
  region                    = local.home_region_by_realm[split(".", each.key)[1]]
  predecessors              = []
  phase                     = lookup(merge(lookup(local.overrides, each.key, {})), "phase", join(".", ["herds", split(".", each.key)[2]]))
  uniquifier                = format("spectre-values-%s", lookup(module.merged_cell_config.uniquifiers, join(".", [each.key, "cell0"]), ""))
  tenancy_name              = lookup(lookup(local.overrides.tenancy_info, "rbaas", {}), split(".", each.key)[1], local.overrides.tenancy_info.default)
  snowflake_config_location = "generic_spectre_region"
  additional_locals = merge({
    limits_region          = lower(lookup(local.region_by_name_all_regions, split(".", each.key)[2]).airport_code)
    manage_regional_values = "true"
    manage_definitions     = "false"
    spectre_group_name     = lookup(lookup(module.merged_cell_config.additional_locals, join(".", [each.key, "cell0"])), "spectre_group_name")
    },
    lookup(module.merged_cell_config.additional_locals, join(".", [each.key, "cell0"]), {})
  )
  scope = format("%s%s", split(".", each.key)[2], "~home-region")
  labels = {
    herd = "993ddcbe-99c5-49ac-a791-889537ecb67a"
  }
  provider_override {
    name = "null"
    constraint = ">= 0.1"
  }
  alarms_to_watch {
    compartment_name = "assets"
    labels           = ["oke-mp-release-cell0", "oke-mp-release-cell1"]
  }
}

resource "shepherd_execution_target" "herds_region_capability" {
  for_each                  = local.herds_spectre_regional_ets
  name                      = format("capability.%s", each.key)
  region                    = split(".", each.key)[2]
  predecessors              = []
  phase                     = lookup(merge(lookup(local.overrides, each.key, {})), "phase", join(".", ["herds", split(".", each.key)[2]]))
  uniquifier                = format("capability-%s", lookup(module.merged_cell_config.uniquifiers, join(".", [each.key, "cell0"]), ""))
  tenancy_name              = lookup(lookup(local.overrides.tenancy_info, "rbaas", {}), split(".", each.key)[1], local.overrides.tenancy_info.default)
  snowflake_config_location = "capability_et"
  labels = {
    herd = "993ddcbe-99c5-49ac-a791-889537ecb67a"
  }
  provider_override {
    name       = "limit"
    constraint = ">= 1.0.962"
  }
  provider_override {
    name       = "property"
    constraint = ">= 1.0.962"
  }
  alarms_to_watch {
    compartment_name = "assets"
    labels           = ["oke-mp-release-cell0", "oke-mp-release-cell1"]
  }
}
