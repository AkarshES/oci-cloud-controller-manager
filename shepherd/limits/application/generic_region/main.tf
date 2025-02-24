variable "cpo-image-validation-enabled" {
  default = true
}

module "ad_map" {
  source                = "./ad_map"
  root_compartment_ocid = local.execution_target.tenancy_ocid
  realm = local.execution_target.region.realm
}

locals {
  physical_ad1                     = module.ad_map.physical_ad1
  image_validator_count            = var.cpo-image-validation-enabled ? 1 : 0

  enable_validation = var.cpo-image-validation-enabled && (length(data.odo_applications.image-release-validator-ccm-csi[0].applications) > 0)
}

module "oke-cpo-images" {
  source                   = "./shared_modules"
  service_artifact_version = local.artifact_versions
}

data "odo_applications" "image-release-validator-ccm-csi" {
  count = var.cpo-image-validation-enabled ? 1 : 0

  ad                     = module.ad_map.physical_ad1.name
  application_name_regex = "image-release-validator-ccm-csi-${local.execution_target.uniquifier}"
}

module "odo_deployment_ccm_csi" {
  source = "./odo_deployment"
  enable_validation = local.enable_validation

  artifact_version = local.artifact_versions["release-validator-ccm-csi"]
  apps             = [
    {
      ad = module.ad_map.physical_ad1.name
      alias = "image-release-validator-ccm-csi-${local.execution_target.uniquifier}"
    }
  ]
  depends_on            = [module.oke-cpo-images]
}

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