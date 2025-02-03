variable artifact_version {
  type = object({
    uri = string,
    version = string,
    type = string
  })
  description = "Artifact to deploy"
}

variable apps {
  type = list(object({
    ad = string,
    alias = string
  }))
  description = "ODO apps to use for the deployments"
}