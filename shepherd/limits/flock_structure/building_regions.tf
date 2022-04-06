locals {
  build_region_cell_overrides = local.prod_scalar == 1 ? {
    for key, value in local.cell_overrides : key => value if split(".", key)[0] == "prd" && contains(keys(local.build_region_to_realm), split(".", key)[2])
  } : {}
}

resource "shepherd_release_phase" "region_build" {
  count        = length(local.regions_under_build) * local.prod_scalar
  name         = "region-build_${local.regions_under_build[count.index].public_name}"
  realm        = local.regions_under_build[count.index].realm
  production   = false
  predecessors = count.index == 0 ? ["realm-build.${local.build_phases[length(local.build_phases) - 1]}"] : ["region-build_${local.regions_under_build[count.index - 1].public_name}"]
}

resource "shepherd_execution_target" "prod_build_spectre_region_et" {
  for_each                          = local.build_region_cell_overrides
  name                              = format("spectre.values.setup.%s", split(".cell", each.key)[0])
  region                            = local.home_region_by_realm[split(".", each.key)[1]]
  predecessors                      = []
  phase                             = format("region-build_%s", split(".", each.key)[2])
  uniquifier                        = format("spectre-values-setup-%s", replace(each.key, ".", "-"))
  tenancy_name                      = lookup(lookup(local.overrides.tenancy_info, split(".", each.key)[0], {}), split(".", each.key)[1], local.overrides.tenancy_info.default)
  snowflake_config_location         = "generic_spectre_region"
  ignored_region_build_capabilities = ["grafana_dashboard"]
  scope                             = format("%s%s", split(".", each.key)[2], "~home-region") ## Adding scope to overcome this https://confluence.oci.oraclecorp.com/pages/viewpage.action?spaceKey=SHEP&title=FAQ#FAQ-WhyamIgettingMFOticketsformultipleregionsinaphasewhenIonlyhaveonebuildingregion+ahomeregioninaphase?
  additional_locals = merge({
    limits_region          = lower(lookup(local.region_by_name_all_regions, split(".", each.key)[2]).airport_code)
    manage_regional_values = "true"
    manage_definitions     = "false"
    spectre_group_name     = lookup(lookup(module.merged_cell_config.additional_locals, each.key), "spectre_group_name")
  }, lookup(module.merged_cell_config.additional_locals, each.key, {}))
}

resource "shepherd_execution_target" "prod_build_region_et" {
  for_each                  = local.build_region_cell_overrides
  name                      = each.key
  region                    = split(".", each.key)[2]
  predecessors              = tonumber(split(".cell", each.key)[1]) == 0 ? [format("spectre.values.setup.%s", split(".cell", each.key)[0])] : [format("%s.cell%s", split(".cell", each.key)[0], tonumber(split(".cell", each.key)[1]) - 1)]
  phase                     = lookup(merge(each.value, lookup(local.overrides, split(".cell", each.key)[0], {})), "phase", join(".", [split(".", each.key)[0], split(".", each.key)[1]]))
  uniquifier                = lookup(module.merged_cell_config.uniquifiers, each.key, "")
  tenancy_name              = lookup(lookup(local.overrides.tenancy_info, split(".", each.key)[0], {}), split(".", each.key)[1], local.overrides.tenancy_info.default)
  snowflake_config_location = lookup(module.merged_cell_config.snowflake_config_locations, each.key, "")
  additional_locals         = lookup(module.merged_cell_config.additional_locals, each.key, {})
  alarms_to_watch {
    compartment_name = format("cell%d:cell%d.mp:cell%d.mp.orchestration", split(lookup(lookup(module.merged_cell_config.additional_locals, each.key, {}), "cell_name_prefix"), each.key)[1], split(lookup(lookup(module.merged_cell_config.additional_locals, each.key, {}), "cell_name_prefix"), each.key)[1], split(lookup(lookup(module.merged_cell_config.additional_locals, each.key, {}), "cell_name_prefix"), each.key)[1]) # This is a compartment in the tenancy above
    labels           = [format(lookup(lookup(module.merged_cell_config.additional_locals, each.key, {}), "watch_mp_release_label_format"), split(lookup(lookup(module.merged_cell_config.additional_locals, each.key, {}), "cell_name_prefix"), each.key)[1])]
  }
  ignored_region_build_capabilities = ["grafana_dashboard"]
}