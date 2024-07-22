variable "service_artifact_version" {}
variable "realm" {}

module "tenancy_info" {
  source           = "../shared_modules/tenancies"
  canary_tenancies      = jsonencode([])
  integration_tenancies = jsonencode([])
}

resource "ocir_steward_image" "ocir-images" {
  # Any non ocir artifacts will have to be filtered out in future
  for_each = var.service_artifact_version

  artifact {
    url       = each.value.uri
    build_tag = each.value.version
  }
}

# Push to OCIR in a service tenancy
resource "ocir_federated_artifact" docker_image_federated {
  for_each = var.realm == "oc1" ? var.service_artifact_version : {}

  artifact {
    url       = each.value.uri
  }

  tenancy_id = module.tenancy_info.tenancy_ocid_map.okecust_tenancy_ocid
  ocir_namespace = module.tenancy_info.tenancy_os_namespace_map.okecust_tenancy_os_namespace
}
