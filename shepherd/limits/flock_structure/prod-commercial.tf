module "prd-config" {
  for_each             = local.prod_realm_by_name
  source               = "./configuration/merged_realm_config"
  flock_config         = local.flock_config
  overrides            = local.overrides
  qualified_realm_name = format("prd.%s", each.key)
}

locals {
  prod_scalar = contains(keys(local.prod_realm_by_name), "oc1") ? 1 : 0
  prod_realms = values(local.prod_realm_by_name)
  prod_cell_overrides = local.prod_scalar == 1 ? {
    for key, value in local.cell_overrides : key => value if split(".", key)[0] == "prd" && contains(local.prod_phases, split(".", key)[1]) && !contains(keys(local.build_region_to_realm), split(".", key)[2])
  } : {}
  // The keys here identify all the env setup ETs that will be created, these
  // are one per realm of each environments. The home region of the realm must
  // be used as a target since DG and policy changes are made with these ETs
  prod_env_setup_ets = {
    for r in keys(local.prod_realm_by_name) : format("prd.%s", r) => merge({
      cell_count        = 1
      realm             = r
      env               = "prd"
      region            = local.home_region_by_realm[r]
      additional_locals = module.prd-config[r].config
      },
    r == "oc1" ? { region = "us-ashburn-1", phase = length(shepherd_release_phase.prd-oc1-bake-1) == 1 ? shepherd_release_phase.prd-oc1-bake-1[0].name : null, cell_count = 2 } : {})
  }
  // The keys here identify all the spectre setup ETs that will be created, these
  // are one per realm of each environments only when needed. Spectre property
  // editing is allowed only in one of the regions, called the Master region.
  // That is where this ET will be targeted. Sample key: dev.oci1
  // The master regions were identified using this source:
  // https://confluence.oci.oraclecorp.com/pages/viewpage.action?pageId=91099575
  prod_spectre_setup_ets = {
    for r in keys(local.prod_realm_by_name) : format("prd.%s", r) => merge({
      realm             = r
      env               = "prd"
      region            = local.home_region_by_realm[r]
      additional_locals = module.prd-config[r].config
      },
    r == "oc1" ? { phase = length(shepherd_release_phase.prd-oc1-bake-1) == 1 ? shepherd_release_phase.prd-oc1-bake-1[0].name : null } : {})
  }
  prod_spectre_regional_ets = local.prod_scalar == 1 ? toset([for key in local.spectre_regional_et : key if !contains(local.build_regions_nocell, key) && split(".", key)[0] == "prd" && contains(local.prod_phases, split(".", key)[1])]) : toset([])
}

resource "shepherd_release_phase" "prd-oc1-bake-1" {
  count        = local.prod_scalar
  name         = "prd.oc1.bake1"
  realm        = "oc1"
  production   = true
  predecessors = contains(local.preprod_phases, "integ") ? [shepherd_release_phase.preprod[index(local.preprod_phases, "integ")].name] : ["integ.oc1"]
}

resource "shepherd_release_phase" "prd-oc1-bake-2" {
  count        = local.prod_scalar
  name         = "prd.oc1.bake2"
  realm        = "oc1"
  production   = true
  predecessors = [shepherd_release_phase.prd-oc1-bake-1[0].name]
}

resource "shepherd_release_phase" "prd-oc1-bake-3" {
  count        = local.prod_scalar
  name         = "prd.oc1.bake3"
  realm        = "oc1"
  production   = true
  predecessors = [shepherd_release_phase.prd-oc1-bake-2[0].name]
}

resource "shepherd_release_phase" "prd-oc1-phx" {
  count        = local.prod_scalar
  name         = "prd.oc1.phx"
  realm        = "oc1"
  production   = true
  predecessors = [shepherd_release_phase.prd-oc1-bake-3[0].name]
}

resource "shepherd_release_phase" "prd-oc1-fra" {
  count        = local.prod_scalar
  name         = "prd.oc1.fra"
  realm        = "oc1"
  production   = true
  predecessors = [shepherd_release_phase.prd-oc1-phx[0].name]
}

resource "shepherd_release_phase" "prd-oc1-single-ad-part-1" {
  count        = local.prod_scalar
  name         = "prd.oc1.single.ad.part1"
  realm        = "oc1"
  production   = true
  predecessors = [shepherd_release_phase.prd-oc1-fra[0].name]
}

resource "shepherd_release_phase" "prd-oc1-single-ad-part-2" {
  count        = local.prod_scalar
  name         = "prd.oc1.single.ad.part2"
  realm        = "oc1"
  production   = true
  predecessors = [shepherd_release_phase.prd-oc1-single-ad-part-1[0].name]
}

resource "shepherd_release_phase" "prod" {
  count        = length(local.prod_realms) * local.prod_scalar
  name         = "prd.${local.prod_realms[count.index].name}"
  realm        = local.prod_realms[count.index].name
  production   = !contains(local.realms_under_build, local.prod_realms[count.index].name)
  predecessors = count.index == 0 ? [shepherd_release_phase.prd-oc1-single-ad-part-2[0].name] : ["prd.${local.prod_realms[count.index - 1].name}"]
  dynamic "on_success" {
    for_each = local.prod_realms[count.index].name == "oc3" ? toset([local.prod_realms[count.index].name]) : toset([])
    content {
      # the default version set will get merged into the golden after it has been verified
      publish_region_build_version_sets = ["default", "*/break_glass"]
    }
  }
}

resource "shepherd_execution_target" "prod_et" {
  for_each                  = local.prod_cell_overrides
  name                      = each.key
  region                    = split(".", each.key)[2]
  predecessors              = lookup(each.value, "predecessor", "") != "" ? [lookup(each.value, "predecessor", "")] : tonumber(split("cell", each.key)[1]) > 0 ? [format("%s.cell%s", split(".cell", each.key)[0], tonumber(split(".cell", each.key)[1]) - 1)] : []
  phase                     = lookup(merge(each.value, lookup(local.overrides, split(".cell", each.key)[0], {})), "phase", join(".", [split(".", each.key)[0], split(".", each.key)[1]]))
  uniquifier                = lookup(module.merged_cell_config.uniquifiers, each.key, "")
  tenancy_name              = lookup(lookup(local.overrides.tenancy_info, split(".", each.key)[0], {}), split(".", each.key)[1], local.overrides.tenancy_info.default)
  snowflake_config_location = lookup(module.merged_cell_config.snowflake_config_locations, each.key, "")
  additional_locals         = merge({
    stage = "prod"
    pool_name_regex = "^oke-deploy-prod[0-9]*"
  }, lookup(module.merged_cell_config.additional_locals, each.key, {}))
  alarms_to_watch {
    compartment_name = "assets"
    #compartment_name = format("cell%d:cell%d.mp:cell%d.mp.orchestration", split(lookup(lookup(module.merged_cell_config.additional_locals, each.key, {}), "cell_name_prefix"), each.key)[1], split(lookup(lookup(module.merged_cell_config.additional_locals, each.key, {}), "cell_name_prefix"), each.key)[1], split(lookup(lookup(module.merged_cell_config.additional_locals, each.key, {}), "cell_name_prefix"), each.key)[1]) # This is a compartment in the tenancy above
    labels = [format(lookup(lookup(module.merged_cell_config.additional_locals, each.key, {}), "watch_mp_release_label_format"), split(lookup(lookup(module.merged_cell_config.additional_locals, each.key, {}), "cell_name_prefix"), each.key)[1])]
  }
  ignored_region_build_capabilities = ["grafana_dashboard"]
}

resource "shepherd_execution_target" "env_setup_et" {
  for_each                          = local.prod_env_setup_ets 
  name                              = format("env.setup.%s", each.key) # herds.oc1
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

resource "shepherd_execution_target" "spectre_setup_et" {
  for_each                          = local.prod_spectre_setup_ets
  name                              = format("spectre.setup.%s", each.key) # dev.oc1
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
    labels           = [ for idx in range(lookup(local.prod_env_setup_ets, each.key, {cell_count = 1}).cell_count) : format("oke-mp-release-cells%s",idx)]
  }
}

resource "shepherd_execution_target" "region_values" {
  for_each                  = local.prod_spectre_regional_ets
  name                      = format("spectre.values.%s", each.key) # dev.oc1.eu-frankfurt-1
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
    name       = "null"
    constraint = ">= 0.1"
  }
  alarms_to_watch {
    compartment_name = "assets"
    labels           = [ for idx in range(lookup(local.prod_env_setup_ets, join(".", [split(".", each.key)[0], split(".", each.key)[1]]), {cell_count = 1}).cell_count) : format("oke-mp-release-cells%s",idx)]
  }
}

