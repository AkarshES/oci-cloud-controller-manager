data "property_definitions" cluster_group_defns {
  group_name = local.execution_target.additional_locals.spectre_group_name
}

locals {
  ccm_image_version_mapping = [for property in data.property_definitions.cluster_group_defns.definitions : property.name if lower(property.name) == "ccm-image-version-mapping" ]
}

resource "capability" "oke_ccm_csi_internal_capability" {
  count = length(local.ccm_image_version_mapping)  >  0 ? 1 : 0
  name  = "oke_ccm_csi_internal_capability"
}