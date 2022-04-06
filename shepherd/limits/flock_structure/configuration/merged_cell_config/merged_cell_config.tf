variable "flock_config" {}
variable "overrides" {}
variable "cell_overrides" {}

locals {
  # Merged additional locals in the defined order of overrides
  additional_locals = {
    for cell_name in keys(var.cell_overrides):
        cell_name => merge(
          var.flock_config,
          lookup(var.overrides, split(".", cell_name)[0], {}),
          lookup(var.overrides, join(".", [split(".", cell_name)[0], split(".", cell_name)[1]]), {}),
          lookup(var.overrides, join(".", [split(".", cell_name)[0], split(".", cell_name)[1], split(".", cell_name)[2]]), {}),
          lookup(var.cell_overrides, cell_name, {}),
          length(regexall("region-build.+", lookup(lookup(var.cell_overrides, cell_name, {}), "phase", "")))> 0 ? {alarms_enabled = true, mapi_api_alarms_enabled = true, kmmon_alarms_enabled = true, worker_alarms_enabled = true} : {})
  }
  snowflake_config_locations = {
    for cell in keys(local.additional_locals):
      cell => lookup(lookup(local.additional_locals, cell, {}), "snowflake_config_location", "generic_region")
  }

  uniquifiers = {
    for cell in keys(local.additional_locals):
      cell => lookup(lookup(local.additional_locals, cell, {}), "shorten_uniquifier", "true") == "false" ? replace(cell, ".", "-") : replace(replace(replace(cell, "cell", ""), "-", ""), ".", "")
  }
}

output "additional_locals" {
  value = local.additional_locals
}

output "snowflake_config_locations" {
  value = local.snowflake_config_locations
}

output "uniquifiers" {
  value = local.uniquifiers
}
