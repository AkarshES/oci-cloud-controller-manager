module "prd-onsr-config" {
  for_each             = local.onsr_realm_by_name
  source               = "./configuration/merged_realm_config"
  flock_config         = local.flock_config
  overrides            = local.overrides
  qualified_realm_name = format("prd.%s", each.key)
}

locals {
  onsr_scalar = length(local.onsr_phases) > 0 ? 1 : 0
  onsr_realms = values(local.onsr_realm_by_name)
  onsr_cell_overrides = local.onsr_scalar == 1 ? {
    for key, value in local.cell_overrides : key => value if split(".", key)[0] == "prd" && contains(local.onsr_phases, split(".", key)[1])
  } : {}
  // The keys here identify all the env setup ETs that will be created, these
  // are one per realm of each environments. The home region of the realm must
  // be used as a target since DG and policy changes are made with these ETs
  onsr_env_setup_ets = {
    for r in keys(local.onsr_realm_by_name) : format("prd.%s", r) => merge({
      cell_count        = 1
      realm             = r
      env               = "prd"
      region            = local.home_region_by_realm[r]
      additional_locals = module.prd-onsr-config[r].config
      },
    r == "oc1" ? { region = "us-ashburn-1", phase = length(shepherd_release_phase.prd-oc1-bake-1) == 1 ? shepherd_release_phase.prd-oc1-bake-1[0].name : null, cell_count = 2 } : {})
  }
  // The keys here identify all the spectre setup ETs that will be created, these
  // are one per realm of each environments only when needed. Spectre property
  // editing is allowed only in one of the regions, called the Master region.
  // That is where this ET will be targeted. Sample key: dev.oci1
  // The master regions were identified using this source:
  // https://confluence.oci.oraclecorp.com/pages/viewpage.action?pageId=91099575
  onsr_spectre_setup_ets = {
    for r in keys(local.onsr_realm_by_name) : format("prd.%s", r) => merge({
      realm             = r
      env               = "prd"
      region            = local.home_region_by_realm[r]
      additional_locals = module.prd-onsr-config[r].config
      },
    r == "oc1" ? { phase = length(shepherd_release_phase.prd-oc1-bake-1) == 1 ? shepherd_release_phase.prd-oc1-bake-1[0].name : null } : {})
  }
  onsr_spectre_regional_ets = local.onsr_scalar == 1 ? toset([for key in local.spectre_regional_et : key if ! contains(local.build_regions_nocell, key) && split(".", key)[0] == "prd" && contains(local.onsr_phases, split(".", key)[1])]) : toset([])
}

resource "shepherd_release_phase" "onsr_prod" {
  count        = length(local.onsr_realms) * local.onsr_scalar
  name         = "prd.${local.onsr_realms[count.index].name}"
  realm        = local.onsr_realms[count.index].name
  production   = ! contains(local.realms_under_build, local.onsr_realms[count.index].name)
  predecessors = count.index == 0 ? local.onsr_polaris_scalar == 0 ? [] : local.onsr_realms[count.index].name != "oc5" ? ["prd.${local.prod_realms[length(local.prod_realms) - 1].name}"] : ["polaris.oc5"] : local.onsr_polaris_scalar == 1 && local.onsr_realms[count.index].name == "oc5" ? ["polaris.oc5"] : ["prd.${local.onsr_realms[count.index - 1].name}"]
  dynamic "on_success" {
    for_each = local.onsr_phases[count.index] == "oc3" ? toset([local.onsr_phases[count.index]]) : toset([])
    content {
      # the default version set will get merged into the golden after it has been verified
      publish_region_build_version_sets = ["default", "*/break_glass"]
    }
  }
}

resource "shepherd_execution_target" "onsr_et" {
  for_each                  = local.onsr_cell_overrides
  name                      = each.key
  region                    = split(".", each.key)[2]
  predecessors              = lookup(each.value, "predecessor", "") != "" ? [lookup(each.value, "predecessor", "")] : tonumber(split("cell", each.key)[1]) > 0 ? [format("%s.cell%s", split(".cell", each.key)[0], tonumber(split(".cell", each.key)[1]) - 1)] : []
  phase                     = lookup(merge(each.value, lookup(local.overrides, split(".cell", each.key)[0], {})), "phase", join(".", [split(".", each.key)[0], split(".", each.key)[1]]))
  uniquifier                = lookup(module.merged_cell_config.uniquifiers, each.key, "")
  tenancy_name              = lookup(lookup(local.overrides.tenancy_info, split(".", each.key)[0], {}), split(".", each.key)[1], local.overrides.tenancy_info.default)
  snowflake_config_location = lookup(module.merged_cell_config.snowflake_config_locations, each.key, "")
  additional_locals         = lookup(module.merged_cell_config.additional_locals, each.key, {})
  alarms_to_watch {
    compartment_name = "assets"
    #compartment_name = format("cell%d:cell%d.mp:cell%d.mp.orchestration", split(lookup(lookup(module.merged_cell_config.additional_locals, each.key, {}), "cell_name_prefix"), each.key)[1], split(lookup(lookup(module.merged_cell_config.additional_locals, each.key, {}), "cell_name_prefix"), each.key)[1], split(lookup(lookup(module.merged_cell_config.additional_locals, each.key, {}), "cell_name_prefix"), each.key)[1]) # This is a compartment in the tenancy above
    labels           = ["oke-mp-release-cell0", "oke-mp-release-cell1"]
  }
  ignored_region_build_capabilities = ["grafana_dashboard"]
}

resource "shepherd_execution_target" "onsr_env_setup_et" {
  for_each                          = local.onsr_env_setup_ets
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

resource "shepherd_execution_target" "onsr_spectre_setup_et" {
  for_each                          = local.onsr_spectre_setup_ets
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
    //labels           = [format(lookup(lookup(module.merged_cell_config.additional_locals, each.key, {}), "watch_mp_release_label_format"), split(lookup(lookup(module.merged_cell_config.additional_locals, each.key, {}), "cell_name_prefix"), each.key)[1])]
  }
}

resource "shepherd_execution_target" "onsr_region_values" {
  for_each                  = local.onsr_spectre_regional_ets
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
    stage = "prod"
    pool_name_regex = "^oke-deploy-prod[0-9]*"
    spectre_group_name     = lookup(lookup(module.merged_cell_config.additional_locals, join(".", [each.key, "cell0"])), "spectre_group_name")
  }, lookup(module.merged_cell_config.additional_locals, join(".", [each.key, "cell0"]), {}))
  provider_override {
    name = "null"
    constraint = ">= 0.1"
  }
  alarms_to_watch {
    compartment_name = "assets"
    labels           = ["oke-mp-release-cell0", "oke-mp-release-cell1"]
   //labels           = [format(lookup(lookup(module.merged_cell_config.additional_locals, each.key, {}), "watch_mp_release_label_format"), split(lookup(lookup(module.merged_cell_config.additional_locals, each.key, {}), "cell_name_prefix"), each.key)[1])]
  }
}
