module "build-realm-config" {
  for_each             = local.build_realm_by_name
  source               = "./configuration/merged_realm_config"
  flock_config         = local.flock_config
  overrides            = local.overrides
  qualified_realm_name = format("prd.%s", each.key)
}

locals {
  build_phases       = keys(local.build_realm_by_name)
  build_realm_config = local.prod_scalar == 1 ? local.build_realm_by_name : {}
  // The keys here identify all the env setup ETs that will be created, these
  // are one per realm of each environments. The home region of the realm must
  // be used as a target since DG and policy changes are made with these ETs
  build_env_setup_ets = {
    for r in keys(local.build_realm_config) : format("prd.%s", r) => merge({
      cell_count        = 1
      realm             = r
      env               = "prd"
      region            = local.home_region_by_realm[r]
      additional_locals = module.build-realm-config[r].config
      },
    r == "oc1" ? { region = "us-ashburn-1", phase = length(shepherd_release_phase.prd-oc1-bake-1) == 1 ? shepherd_release_phase.prd-oc1-bake-1[0].name : null, cell_count = 2 } : {})
  }
  // The keys here identify all the spectre setup ETs that will be created, these
  // are one per realm of each environments only when needed. Spectre property
  // editing is allowed only in one of the regions, called the Master region.
  // That is where this ET will be targeted. Sample key: dev.oci1
  // The master regions were identified using this source:
  // https://confluence.oci.oraclecorp.com/pages/viewpage.action?pageId=91099575
  build_spectre_setup_ets = {
    for r in keys(local.build_realm_config) : format("prd.%s", r) => merge({
      realm             = r
      env               = "prd"
      region            = local.home_region_by_realm[r]
      additional_locals = module.build-realm-config[r].config
      },
    r == "oc1" ? { phase = length(shepherd_release_phase.prd-oc1-bake-1) == 1 ? shepherd_release_phase.prd-oc1-bake-1[0].name : null } : {})
  }
}

resource "shepherd_release_phase" "build_realm" {
  count        = length(local.build_phases) * local.prod_scalar
  name         = "realm-build.${local.build_phases[count.index]}"
  realm        = local.build_phases[count.index]
  production   = ! contains(local.realms_under_build, local.build_phases[count.index])
  predecessors = count.index == 0 ? ["prd.${local.onsr_realms[length(local.onsr_phases) - 1].name}"] : ["realm-build.${local.build_phases[count.index - 1]}"]
  dynamic "on_success" {
    for_each = local.build_phases[count.index] == "oc3" ? toset([local.build_phases[count.index]]) : toset([])
    content {
      # the default version set will get merged into the golden after it has been verified
      publish_region_build_version_sets = ["default", "*/break_glass"]
    }
  }
}

resource "shepherd_execution_target" "build_env_setup_et" {
  for_each                          = local.build_env_setup_ets
  name                              = format("env.setup.%s", each.key)
  region                            = each.value.region
  phase                             = lookup(each.value, "phase", join(".", ["realm-build", split(".", each.key)[1]]))
  predecessors                      = []
  uniquifier                        = format("env-setup-%s", replace(each.key, ".", "-"))
  tenancy_name                      = lookup(lookup(local.overrides.tenancy_info, each.value.env, {}), each.value.realm, local.overrides.tenancy_info.default)
  snowflake_config_location         = "generic_tenancy"
  additional_locals                 = merge(each.value.additional_locals, { cell_count : each.value.cell_count })
  ignored_region_build_capabilities = ["grafana_dashboard"]
}

resource "shepherd_execution_target" "build_spectre_setup_et" {
  for_each                          = local.build_spectre_setup_ets
  name                              = format("spectre.setup.%s", each.key)
  region                            = each.value.region
  phase                             = lookup(each.value, "phase", join(".", ["realm-build", split(".", each.key)[1]]))
  predecessors                      = ["env.setup.${each.value.env}.${each.value.realm}"]
  uniquifier                        = format("spectre-setup-%s", replace(each.key, ".", "-"))
  tenancy_name                      = lookup(lookup(local.overrides.tenancy_info, each.value.env, {}), each.value.realm, local.overrides.tenancy_info.default)
  snowflake_config_location         = "spectre_region"
  additional_locals                 = each.value.additional_locals
  ignored_region_build_capabilities = ["grafana_dashboard"]
}
