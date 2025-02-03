resource "odo_deployment" "deployment" {
  ad = var.apps[0].ad
  alias = var.apps[0].alias
  artifact {
    url = var.artifact_version.uri
    build_tag = var.artifact_version.version
    type = var.artifact_version.type
  }

  is_overlay = true
}