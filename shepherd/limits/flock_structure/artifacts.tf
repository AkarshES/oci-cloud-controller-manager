# Configuration
variable "images" {
  type = list(object({
    name     = string
    location = string
  }))
}

resource "shepherd_artifacts" "artifacts" {
  dynamic "artifact" {
    for_each = [for image in var.images : {
      name     = lookup(image, "name")
      location = lookup(image, "location")
    }]
    content {
      name        = artifact.value.name
      type        = "ocir"
      location    = artifact.value.location
      description = "${artifact.value.name} OCIR image"
    }
  }
}