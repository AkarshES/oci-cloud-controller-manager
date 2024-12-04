locals {
  regional_values = [for mapping in module.properties_values.regional_property_values: mapping.value if mapping.region == local.execution_target.additional_locals.limits_region]
  override_values = [for mapping in module.properties_values.regional_property_overrides: mapping.value if mapping.region == local.execution_target.additional_locals.limits_region]

  raw_regional_image_list = [for v in local.regional_values : regexall("\"([^\"]+?)@sha256", v)]
  raw_override_image_list = [for v in local.override_values : regexall("\"([^\"]+?)@sha256", v)]

  regional_image_list = flatten(local.raw_regional_image_list)
  overrides_image_list = flatten(local.raw_override_image_list)

  image_name = "oke-public-cloud-provider-oci"
}

module "ad_map" {
  source                = "./ad_map"
  root_compartment_ocid = local.execution_target.tenancy_ocid
  realm = local.execution_target.region.realm
}

module "odo_configuration_ccm_csi_image_push" {
  source = "./shared_modules/odo_configuration"
  execution_target = local.execution_target

  realm                   = lower(local.execution_target.region.realm)
  stage                   = local.execution_target.additional_locals.stage
  artifact_set_identifier = "image-release-validator-ccm-csi"
  compartment_id          = local.execution_target.tenancy_ocid
  pool_name_regex         = local.execution_target.additional_locals.pool_name_regex
  physical_ad1            = module.ad_map.physical_ad1.name
  application_alias = "image-release-validator-ccm-csi-${local.execution_target.additional_locals.stage}"
  env_vars = []
}

module "odo_configuration_ccm_csi_infra" {
  source = "./shared_modules/odo_configuration"
  execution_target = local.execution_target

  realm                   = lower(local.execution_target.region.realm)
  stage                   = local.execution_target.additional_locals.stage
  artifact_set_identifier = "infra-release-validator-ccm-csi"
  compartment_id          = local.execution_target.tenancy_ocid
  pool_name_regex         = local.execution_target.additional_locals.pool_name_regex
  physical_ad1            = module.ad_map.physical_ad1.name
  application_alias = "infra-release-validator-ccm-csi-${local.execution_target.additional_locals.stage}"
  env_vars = [
    {
      name = regional_image_list
      value = local.regional_image_list
    },
    {
      name = override_image_list
      value = local.overrides_image_list
    }
  ]
}

module "odo_deployment_ccm_csi_infra" {
  source = "./shared_modules/odo_deployment"

  artifact_version = local.artifact_versions["infra-release-validator-ccm-csi"]
  apps             = [
    {
      ad = module.ad_map.physical_ad1.name
      alias = "infra-release-validator-ccm-csi-${local.execution_target.additional_locals.stage}"
    }
  ]
  depends_on = [module.odo_configuration_ccm_csi_infra]
}

resource "capability_require_capability" "regional_infra" {
  name = "oke_deploy_odo"
}

module "properties_values" {
  source             = "./shared_modules/properties_values/"
  execution_target   = local.execution_target
  spectre_group_name = lookup(local.execution_target.additional_locals, "spectre_group_name")
  env                = lookup(local.execution_target.additional_locals, "env", "")
  realm              = local.execution_target.region.realm

  depends_on = [module.odo_deployment_ccm_csi_infra]
}

resource "capability_require_capability" "oke_ccm_csi_internal_capability" {
  name  = "oke_ccm_csi_internal_capability"
}
