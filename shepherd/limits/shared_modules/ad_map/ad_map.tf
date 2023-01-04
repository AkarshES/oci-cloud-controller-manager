// A helper module to return physical -> logical and physical -> number name map
// of availability domains

variable "root_compartment_ocid" {}

data "ad_availability_domains" ads {
  tenancy_id = var.root_compartment_ocid
}

locals {
  # Physical ad to logical ad map
  physical_to_logical_map = {
  for ad in data.ad_availability_domains.ads.ads :
  ad.name => ad.logical_ad
  }
  # Physical ad to ad number name map
  physical_to_number_name_map = {
  for ad in data.ad_availability_domains.ads.ads :
  ad.name => ad.ad_number_name
  }
}

output "physical_to_logical_map" {
  value = local.physical_to_logical_map
}

output "physical_to_number_name_map" {
  value = local.physical_to_number_name_map
}
