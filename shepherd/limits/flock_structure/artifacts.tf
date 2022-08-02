# Configuration
locals {
  artifact_config = {
    management-plane-api = {
      type        = "docker"
      description = "API Image"
    }
    management-plane-monitor = {
      type        = "docker"
      description = "Monitor Image"
    }
    management-plane-worker = {
      type        = "docker"
      description = "Worker Image"
    }
  }
}
# Artifacts
/*Commenting out as the application release is not currently enabled in this flock.
This block might be needed at a later point.
*/
/*
resource "shepherd_artifacts" "artifact" {
  for_each = local.artifact_config
  artifact {
    name        = each.key
    type        = each.value.type
    location    = each.key
    description = each.value.description
  }
}
*/