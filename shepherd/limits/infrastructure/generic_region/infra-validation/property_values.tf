variable "execution_target" {}
variable "spectre_group_name" {}
variable "env" {}
variable "realm" {}

locals {
  default_properties_values = [
    for filename in fileset(path.module, "${var.realm}/*.json") : jsondecode(file("${path.module}/${filename}"))
  ]

  default_values = flatten([
    for property in local.default_properties_values : [
      for realm in property.realms : [
        for value in property.values : {
          group  = var.spectre_group_name
          name   = property.name
          region = var.execution_target.additional_locals.limits_region
          ad     = value.ad
          min    = lookup(value, "min", null)
          max    = lookup(value, "max", null)
          value  = lookup(value, "value", null)
        }
      ] if realm == var.execution_target.region.realm
    ]
  ])

  /*
    Create Map so that we can concat with the regional values map
  */
  default_values_map = {
    for entry in local.default_values : "${entry.group}/${entry.name}/${entry.region}/${entry.ad}" => entry
  }

  tenancy_overrides = flatten([
    for p, v in lookup(local.tenancy_property_overrides, "${var.realm}", {}) : [
      for val in v.overrides : [
        for region in val.regions : {
          group  = var.spectre_group_name
          name   = p
          region = region
          ad     = lookup(val, "ad", "all")
          min    = lookup(val, "min", null)
          max    = lookup(val, "max", null)
          value  = lookup(val, "value")
          tag    = lookup(val, "tenancy_ocid")
        } if var.env == val.env && region == var.execution_target.additional_locals.limits_region && length(val.regions) != 0
      ]
    ]
  ])

  tenancy_overrides_all_regions = try(flatten([
    for p, v in lookup(local.tenancy_property_overrides, "${var.realm}", {}) : [
      for val in v.overrides : {
        group  = var.spectre_group_name
        name   = p
        region = var.execution_target.additional_locals.limits_region
        ad     = lookup(val, "ad", "all")
        min    = lookup(val, "min", null)
        max    = lookup(val, "max", null)
        value  = lookup(val, "value")
        tag    = lookup(val, "tenancy_ocid")
      } if length(val.regions) == 0 && var.env == val.env
    ]
  ]), [{}])

  tenancy_overrides__all_regions_map = {
    for entry in local.tenancy_overrides_all_regions : "${entry.group}/${entry.name}/${entry.region}/${entry.ad}/${entry.tag}" => entry
  }

  tenancy_overrides_map = {
    for entry in local.tenancy_overrides : "${entry.group}/${entry.name}/${entry.region}/${entry.ad}/${entry.tag}" => entry }

  tenancy_overrides_final = merge(local.tenancy_overrides_map, local.tenancy_overrides__all_regions_map)
  /*
   We read the property values from the folder regional_values. These are template files where we can pass the variable and
   can write code in template file to generate output file. We pass lower case region airport code coming from the execution target
 */

  regional_properties_values_overrides = try([
    for filename in fileset(path.module, "regional_values/${var.realm}/${var.env}/${var.execution_target.additional_locals.limits_region}/*.tpl") : jsondecode(templatefile("${path.module}/${filename}", { region = var.execution_target.additional_locals.limits_region }))
    //for filename in fileset(path.module, "regional_values/${var.realm}/${var.env}/${var.execution_target.additional_locals.limits_region}/*.tpl") : filename
  ], [{}])

  /*
    Traverse through the json created in previous statement and create object required for property_values resources
    We have special case here, we can add the snow-flake value for region in case want to have different value for property in
    region. Ignore the snowflake property value in below statement
    check shepherd-provider-testing_test_region1.tpl.
  */

  regional_values_overrides = try(flatten([
    for property in local.regional_properties_values_overrides : [
      for realm in property.realms : [
        for value in property.values : {
          group  = var.spectre_group_name
          name   = property.name
          region = property.region
          ad     = value.ad
          min    = lookup(value, "min", null)
          max    = lookup(value, "max", null)
          value  = lookup(value, "value", null)
        } if can(value.region) == false
      ] if realm == var.execution_target.region.realm && var.execution_target.additional_locals.manage_regional_values
    ]
  ]), {})

  /*
    create map with key as
    group/property_name/region/ad
  */
  regional_values_map_overrides = {
    for entry in local.regional_values_overrides : "${entry.group}/${entry.name}/${entry.region}/${entry.ad}" => entry
  }

  //regional_overrides = length(local.regional_snowflake_values_map) > 0 ? merge(local.regional_values_map, local.regional_snowflake_values_map) : local.regional_values_map

  all_values = merge(local.global_default_values_map, local.default_values_map, local.regional_values_map_overrides)
}

output "regional_values" {
  value = local.all_values
}

output "override_values" {
  value = local.tenancy_overrides_final
}