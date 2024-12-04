
module "polaris-oc5-config" {
  source               = "./configuration/merged_realm_config"
  flock_config         = local.flock_config
  overrides            = local.overrides
  qualified_realm_name = "polaris.oc5"
}

locals {
  onsr_polaris_scalar     = contains(local.onsr_phases, "oc5") ? 1 : 0
  index_prior_to_oc5      = try(index(keys(local.onsr_realm_by_name), "oc5"), 0)
  index_0_and_commercial  = local.prod_scalar == 1 ? ["prd.${local.prod_realms[length(local.prod_realms) - 1].name}"] : []
  polaris_oc5_predecessor = local.index_prior_to_oc5 == 0 ? local.index_0_and_commercial : ["prd.${local.onsr_realms[local.index_prior_to_oc5 - 1].name}"]
  onsr_polaris_cell_overrides = local.onsr_polaris_scalar == 1 ? {
    for key, value in local.cell_overrides : key => value if split(".", key)[0] == "polaris" && split(".", key)[1] == "oc5"
  } : {}
  onsr_polaris_env_setup_ets = local.onsr_polaris_scalar == 1 ? {
    "polaris.oc5" = {
      cell_count        = 1
      realm             = "oc5"
      env               = "polaris"
      region            = "us-tacoma-1"
      additional_locals = module.polaris-oc5-config.config
    }
  } : {}
  onsr_polaris_spectre_setup_ets = local.onsr_polaris_scalar == 1 ? {
    "polaris.oc5" = {
      realm             = "oc5"
      env               = "polaris"
      region            = "us-tacoma-1"
      additional_locals = module.polaris-oc5-config.config
    }
  } : {}
  onsr_polaris_spectre_regional_ets = local.onsr_polaris_scalar == 1 ? toset([for key in local.spectre_regional_et : key if ! contains(local.build_regions_nocell, key) && split(".", key)[0] == "polaris" && split(".", key)[1] == "oc5"]) : toset([])
}

resource "shepherd_release_phase" "polaris_oc5" {
  count        = local.onsr_polaris_scalar
  name         = "polaris.oc5"
  realm        = "oc5"
  production   = false
  predecessors = local.polaris_oc5_predecessor
}

resource "shepherd_execution_target" "polaris_onsr_et" {
  for_each                  = local.onsr_polaris_cell_overrides
  name                      = each.key
  region                    = split(".", each.key)[2]
  predecessors              = lookup(each.value, "predecessor", "") != "" ? [lookup(each.value, "predecessor", "")] : tonumber(split("cell", each.key)[1]) > 0 ? [format("%s.cell%s", split(".cell", each.key)[0], tonumber(split(".cell", each.key)[1]) - 1)] : []
  phase                     = lookup(merge(each.value, lookup(local.overrides, split(".cell", each.key)[0], {})), "phase", join(".", [split(".", each.key)[0], split(".", each.key)[1]]))
  uniquifier                = lookup(module.merged_cell_config.uniquifiers, each.key, "")
  tenancy_name              = lookup(lookup(local.overrides.tenancy_info, split(".", each.key)[0], {}), split(".", each.key)[1], local.overrides.tenancy_info.default)
  snowflake_config_location = lookup(module.merged_cell_config.snowflake_config_locations, each.key, "")
  additional_locals         = merge({
    stage = "polaris"
    pool_name_regex = "^oke-deploy-prod[0-9]*"
  }, lookup(module.merged_cell_config.additional_locals, each.key, {}))
  alarms_to_watch {
    compartment_name = "assets"
    labels           = [format(lookup(lookup(module.merged_cell_config.additional_locals, each.key, {}), "watch_mp_release_label_format"), split(lookup(lookup(module.merged_cell_config.additional_locals, each.key, {}), "cell_name_prefix"), each.key)[1])]
  }
  ignored_region_build_capabilities = ["grafana_dashboard"]
}

resource "shepherd_execution_target" "polaris_onsr_env_setup_et" {
  for_each                          = local.onsr_polaris_env_setup_ets
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
  }
}

resource "shepherd_execution_target" "polaris_onsr_spectre_setup_et" {
  for_each                          = local.onsr_polaris_spectre_setup_ets
  name                              = format("spectre.setup.%s", each.key)
  region                            = each.value.region
  phase                             = lookup(each.value, "phase", each.key)
  predecessors                      = ["env.setup.${each.value.env}.${each.value.realm}"]
  uniquifier                        = format("spectre-setup-%s", replace(each.key, ".", "-"))
  tenancy_name                      = lookup(lookup(local.overrides.tenancy_info, each.value.env, {}), each.value.realm, local.overrides.tenancy_info.default)
  snowflake_config_location         = "spectre_region"
  additional_locals                 = each.value.additional_locals
  ignored_region_build_capabilities = ["grafana_dashboard"]
  alarms_to_watch {
    compartment_name = "assets"
    labels           = ["oke-mp-release-cell0", "oke-mp-release-cell1"]
  }
}

resource "shepherd_execution_target" "polaris_onsr_region_values" {
  for_each                  = local.onsr_polaris_spectre_regional_ets
  name                      = format("spectre.values.%s", each.key)
  region                    = local.home_region_by_realm[split(".", each.key)[1]]
  predecessors              = [join(".", [each.key, "cell0"])]
  phase                     = lookup(merge(lookup(local.overrides, each.key, {})), "phase", join(".", [split(".", each.key)[0], split(".", each.key)[1]]))
  uniquifier                = format("spectre-values-%s", lookup(module.merged_cell_config.uniquifiers, join(".", [each.key, "cell0"]), ""))
  tenancy_name              = lookup(lookup(local.overrides.tenancy_info, split(".", each.key)[0], {}), split(".", each.key)[1], local.overrides.tenancy_info.default)
  snowflake_config_location = "generic_spectre_region"
  additional_locals = merge({
    limits_region          = lower(lookup(local.region_by_name_all_regions, split(".", each.key)[2]).airport_code)
    manage_regional_values = "true"
    manage_definitions     = "false"
    stage = "polaris"
    pool_name_regex = "^oke-deploy-prod[0-9]*"
    spectre_group_name     = lookup(lookup(module.merged_cell_config.additional_locals, join(".", [each.key, "cell0"])), "spectre_group_name")
    },
    lookup(module.merged_cell_config.additional_locals, join(".", [each.key, "cell0"]), {})
  )
  provider_override {
    name = "null"
    constraint = ">= 0.1"
  }
  alarms_to_watch {
    compartment_name = "assets"
    labels           = ["oke-mp-release-cell0", "oke-mp-release-cell1"]
  }
}
