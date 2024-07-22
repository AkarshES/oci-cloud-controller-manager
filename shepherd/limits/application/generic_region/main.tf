module "oke-cpo-images" {
  source = "./cpo-images"
  service_artifact_version = local.artifact_versions
  realm = local.execution_target.region.realm
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
