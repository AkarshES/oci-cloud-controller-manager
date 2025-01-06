variable "service_artifact_version" {}

resource "ocir_steward_image" "ocir-images" {
  # Any non ocir artifacts will have to be filtered out in future
  for_each = var.service_artifact_version

  artifact {
    url       = each.value.uri
    build_tag = each.value.version
  }
}