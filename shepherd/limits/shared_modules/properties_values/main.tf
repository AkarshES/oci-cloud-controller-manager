output "regional_properties_values_overrides" {
  value = local.regional_properties_values_overrides
}

module "all_values_validation" {
  source = "../validation"

  for_each = local.all_values
  image_mapping_values = lookup(each.value, "value", null)
}

module "overrides_validation" {
  source = "../validation"

  for_each = local.tenancy_overrides_final
  image_mapping_values = lookup(each.value, "value", null)
}

resource "property_value" "values" {
  for_each = local.all_values

  group  = each.value.group
  name   = each.value.name
  region = each.value.region
  ad     = each.value.ad
  min    = lookup(each.value, "min", null)
  max    = lookup(each.value, "max", null)
  value  = lookup(each.value, "value", null)
}

resource "property_override" "overrides" {
  for_each = local.tenancy_overrides_final

  group  = each.value.group
  name   = each.value.name
  region = each.value.region
  ad     = each.value.ad
  tag    = each.value.tag
  min    = lookup(each.value, "min", null)
  max    = lookup(each.value, "max", null)
  value  = lookup(each.value, "value", null)
}