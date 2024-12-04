data "ad_availability_domains" ads {
  tenancy_id = var.root_compartment_ocid
}

locals {
  ad = var.realm == "region1" ? "ad2" : "ad1"
  physical_ad1 = [
    for ad in data.ad_availability_domains.ads.ads:
      ad if ad.ad_number_name == local.ad
  ][0]
}

output "physical_ad1" {
  value = local.physical_ad1
}