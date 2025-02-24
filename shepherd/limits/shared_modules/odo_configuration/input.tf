variable "execution_target" {}

variable "realm" {
  type = string
}
variable "compartment_id" {
  type = string
}
variable "stage" {
  type = string
}
variable "env_vars" {
  type = list(object({
    name             = string
    value            = string
  }))
  description = "Environment variables"
}
variable "physical_ad1" {
  type = string
}
variable "application_alias" {
  type = string
}
variable "pool_name_regex" {
  type        = string
  description = "Pool to deploy from"
}
variable "artifact_set_identifier" {
  type        = string
  description = "The artifact name"
}