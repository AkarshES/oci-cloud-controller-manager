module "ad_map" {
  source                = "./ad_map"
  root_compartment_ocid = local.execution_target.tenancy_ocid
  realm = local.execution_target.region.realm
}

locals {
  physical_ad1                     = module.ad_map.physical_ad1
  image_validator_count            = local.artifact_versions["release-validator-ccm-csi"].version == "skip" ? 0 : 1
}

module "oke-cpo-images" {
  source                   = "./ocir_images"
  service_artifact_version = local.artifact_versions
}

// Uncomment the following code to enable validation of images once the odo application is configured in all regions.

#data "odo_applications" "image_release_validator_pop_ccm_csi" {
#  count = local.image_validator_count
#  ad                     = local.physical_ad1.name
#  application_name_regex = "image-release-validator-ccm-csi-${local.execution_target.additional_locals.stage}"
#}
#
#module "odo_deployment_ccm_csi" {
#  source = "./shared_modules/odo_deployment"
#  image_validator_count = local.image_validator_count
#
#  artifact_version = local.artifact_versions["release-validator-ccm-csi"]
#  apps             = [
#    {
#      ad = local.physical_ad1.name
#      alias = "image-release-validator-ccm-csi-${local.execution_target.additional_locals.stage}"
#    }
#  ]
#  depends_on            = [module.oke-cpo-images]
#}

resource "capability_require_capability" "oke_regional_infrastructure" {
  name = "oke_regional_infrastructure"
}

resource "capability_require_capability" "ocir_steward_tenancy" {
  name = "ocir_steward_image"
}

resource "capability" "oke_ccm_csi" {
  name       = "oke_ccm_csi"
  depends_on = [module.oke-cpo-images]
}