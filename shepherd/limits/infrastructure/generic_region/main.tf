module "cell_resources" {
  source                            = "./shared_modules/cell_infra_resources"
  root_compartment_ocid             = local.execution_target.tenancy_ocid
  cell_index                        = split(lookup(local.execution_target.additional_locals, "cell_name_prefix", "cell"), local.execution_target.name)[1]
  orchestration_compartment_format  = lookup(local.execution_target.additional_locals, "orchestration_compartment_format", "cell%d.mp.orchestration")
  kmi_compartment_format            = lookup(local.execution_target.additional_locals, "kmi_compartment_format", "cell%d.mp.kmi")
  vcn_format                        = lookup(local.execution_target.additional_locals, "vcn_format", "cell%d")
  cell_compartment_format           = lookup(local.execution_target.additional_locals, "cell_compartment_format", "cell%d")
  mapi_instance_subnet_name         = lookup(local.execution_target.additional_locals, "mapi_instance_subnet_name", "mapi-instance")
  mapi_lb_subnet_name               = lookup(local.execution_target.additional_locals, "mapi_lb_subnet_name", "mapi-lb")
  mp_worker_subnet_name             = lookup(local.execution_target.additional_locals, "mp_worker_subnet_name", "mp-worker")
  kmi_subnet_name                   = lookup(local.execution_target.additional_locals, "kmi_subnet_name", "kmi")
  env_name                          = lookup(local.execution_target.additional_locals, "env", "")
  tag_ns_name                       = lookup(local.execution_target.additional_locals, "tag_ns_name", "ManagementPlane")
  instance_type_tag_name            = lookup(local.execution_target.additional_locals, "instance_type_tag_name", "HostType")
  api_hostclass                     = lookup(local.execution_target.additional_locals, "api_hostclass", "")
  lb_shape                          = lookup(local.execution_target.additional_locals, "lb_shape", "100Mbps")
  enable_kaas_regional_instance     = lookup(local.execution_target.additional_locals, "enable_kaas_regional_instance", true)
  kaas_name_format                  = lookup(local.execution_target.additional_locals, "kaas_name_format", "oke-mp-cell%d")
  region_public_name                = local.execution_target.region.public_name
  region_code                       = lower(local.execution_target.region.airport_code)
  wfaas_name_format                 = lookup(local.execution_target.additional_locals, "wfaas_name_format", "oke-mp-cell%d")
  enable_wfaas                      = lookup(local.execution_target.additional_locals, "enable_wfaas", true)
  api_log_namespace_format          = lookup(local.execution_target.additional_locals, "api_log_namespace_format", "oke-api-cell%d")
  monitor_log_namespace_format      = lookup(local.execution_target.additional_locals, "monitor_log_namespace_format", "oke-monitor-cell%d")
  worker_log_namespace_format       = lookup(local.execution_target.additional_locals, "worker_log_namespace_format", "oke-worker-cell%d")
  phonebook_name                    = lookup(local.execution_target.additional_locals, "phonebook_name", "oracle-kubernetes-engine")
  sms_namespace_name_format         = lookup(local.execution_target.additional_locals, "sms_namespace_name_format", "oke-mapi-cell%d")
  certificate_name_format           = lookup(local.execution_target.additional_locals, "certificate_name_format", "mapi-pki-cert-cell%d")
  certificate_compartment           = lookup(local.execution_target.additional_locals, "certificate_compartment", "assets")
  alarms_compartment                = lookup(local.execution_target.additional_locals, "alarms_compartment", "")
  alarms_project                    = lookup(local.execution_target.additional_locals, "alarms_project", "kubernetes")
  severity_2                        = lookup(local.execution_target.additional_locals, "severity_2", 2)
  severity_3                        = lookup(local.execution_target.additional_locals, "severity_3", 3)
  severity_4                        = lookup(local.execution_target.additional_locals, "severity_4", 4)
  jira_project                      = lookup(local.execution_target.additional_locals, "jira_project", "OKE")
  jira_component                    = lookup(local.execution_target.additional_locals, "jira_component", "Management Plane")
  okenp_jira_project                = lookup(local.execution_target.additional_locals, "jira_project", "OKENP")
  okenp_jira_component              = lookup(local.execution_target.additional_locals, "jira_component", "Volumes")
  leadership_jira_component         = lookup(local.execution_target.additional_locals, "leadership_jira_component", "Leadership")
  runbook_base                      = lookup(local.execution_target.additional_locals, "runbook_base", "https://devops.oci.oraclecorp.com/runbooks/OKE/oke-how-tos")
  alarms_enabled                    = lookup(local.execution_target.additional_locals, "alarms_enabled", false)
  image_name                        = lookup(local.execution_target.additional_locals, "image_name", "")
  image_url                         = lookup(local.execution_target.additional_locals, "image_url", "")
  image_type                        = lookup(local.execution_target.additional_locals, "image_type", "BE24")
  splunk_enabled                    = lookup(local.execution_target.additional_locals, "splunk_enabled", false)
  mapi_lb_name                      = lookup(local.execution_target.additional_locals, "mapi_lb_name", "oke-mp-api-loadbalancer")
  realm                             = local.execution_target.region.realm
  splat_base_service_name           = lookup(local.execution_target.additional_locals, "splat_base_service_name", "")
  splat_service_name_format         = lookup(local.execution_target.additional_locals, "splat_service_name_format", "oke-mapi-cell%d")
  splat_service_fleet               = lookup(local.execution_target.additional_locals, "splat_service_fleet", "overlay-fleet")
  oke_secrets_namespace             = lookup(local.execution_target.additional_locals, "oke_secrets_namespace", "")
  cp_vcn_name_format                = lookup(local.execution_target.additional_locals, "cp_vcn_name_format", "%s-cpvcn")
  cp_vcn_compartment                = lookup(local.execution_target.additional_locals, "cp_vcn_compartment", "oke-cp-api")
  pl_vcn_name                       = lookup(local.execution_target.additional_locals, "pl_vcn_name", "admin")
  pl_vcn_compartment                = lookup(local.execution_target.additional_locals, "pl_vcn_compartment", "admin")
  prime_vcn_compartment             = lookup(local.execution_target.additional_locals, "prime_vcn_compartment", "admin")
  prime_vcn_name                    = lookup(local.execution_target.additional_locals, "prime_vcn_name", "prod")
  watch_mp_release_label_format     = lookup(local.execution_target.additional_locals, "watch_mp_release_label_format", "oke-mp-release-cell%d")
  skip_mapi_alarms                  = lookup(local.execution_target.additional_locals, "skip_mapi_alarms", false)
  skip_kmon_alarms                  = lookup(local.execution_target.additional_locals, "skip_kmon_alarms", false)
  skip_worker_alarms                = lookup(local.execution_target.additional_locals, "skip_worker_alarms", false)
  skip_dns                          = lookup(local.execution_target.additional_locals, "skip_dns", false)
  ccm_alarm_label_format            = lookup(local.execution_target.additional_locals, "ccm_alarm_label_format", "oke-kmi-cell%d")
  ccm_alarms_fleet_format           = lookup(local.execution_target.additional_locals, "ccm_alarms_fleet_format", "oke-kmi-cell%d")
  csi_dataplane_alarms_fleet_format = lookup(local.execution_target.additional_locals, "csi_dataplane_alarms_fleet_format", "prod-oke-clusters-%s-dataplane%s")
  ccm_alarms_enabled                = lookup(local.execution_target.additional_locals, "ccm_alarms_enabled", true)
  csi_alarms_enabled                = lookup(local.execution_target.additional_locals, "csi_alarms_enabled", true)
  ccm_jira_item                     = lookup(local.execution_target.additional_locals, "ccm_jira_item", "KMI")
  csi_jira_item                     = lookup(local.execution_target.additional_locals, "csi_jira_item", "CSI")
  has_mapi_grafana_dashboard        = ""
}

module "data" {
  source                = "./shared_modules/tenancies"
  canary_tenancies      = lookup(local.execution_target.additional_locals, "canary_tenancies", jsonencode([]))
  integration_tenancies = lookup(local.execution_target.additional_locals, "integration_tenancies", jsonencode([]))
}

resource "capability_require_capability" "oke_secrets_management" {
  name = "oke_secrets_management"
}

variable "cpo-image-validation-enabled" {
  default = true
}

locals {
  pop_version = "acd9f6513f3_17"

  regional_values = [for mapping in module.validation_module.regional_values: mapping.value if mapping.region == local.execution_target.additional_locals.limits_region]
  override_values = [for mapping in module.validation_module.override_values: mapping.value if mapping.region == local.execution_target.additional_locals.limits_region]

  raw_regional_image_list = [for v in local.regional_values : regexall("\"([^\"]+?)@sha256", v)]
  raw_override_image_list = [for v in local.override_values : regexall("\"([^\"]+?)@sha256", v)]

  regional_image_list = tolist(toset(flatten(local.raw_regional_image_list)))
  overrides_image_list = tolist(toset(flatten(local.raw_override_image_list)))

  regional_list_mid_index = floor(length(local.regional_image_list) / 2)
  overrides_list_mid_index = floor(length(local.overrides_image_list) / 2)

  regional_image_list_first_half = length(local.regional_image_list) > 0 ? slice(local.regional_image_list, 0, local.regional_list_mid_index) : []
  regional_image_list_second_half = length(local.regional_image_list) > 0 ? slice(local.regional_image_list, local.regional_list_mid_index, length(local.regional_image_list) - 1) : []

  overrides_image_list_first_half = length(local.overrides_image_list) > 0 ? slice(local.overrides_image_list, 0, local.overrides_list_mid_index) : []
  overrides_image_list_second_half = length(local.overrides_image_list) > 0 ? slice(local.overrides_image_list, local.overrides_list_mid_index, length(local.overrides_image_list) - 1) : []
}

data "odo_applications" "infra-release-validator-ccm-csi" {
  count = var.cpo-image-validation-enabled ? 1 : 0

  ad                     = module.ad_map.physical_ad1.name
  application_name_regex = "infra-release-validator-ccm-csi-${local.execution_target.additional_locals.stage}"
}

module "ad_map" {
  source                = "./ad_map"
  root_compartment_ocid = local.execution_target.tenancy_ocid
  realm = local.execution_target.region.realm
}

module "validation_module" {
  source = "./infra-validation"
  execution_target   = local.execution_target
  spectre_group_name = lookup(local.execution_target.additional_locals, "spectre_group_name")
  env                = lookup(local.execution_target.additional_locals, "env", "")
  realm              = local.execution_target.region.realm
}

module "odo_configuration_ccm_csi_image_push" {
  source = "./shared_modules/odo_configuration"
  execution_target = local.execution_target

  realm                   = lower(local.execution_target.region.realm)
  stage                   = local.execution_target.additional_locals.stage
  artifact_set_identifier = "release-validator-ccm-csi"
  compartment_id          = local.execution_target.tenancy_ocid
  pool_name_regex         = local.execution_target.additional_locals.pool_name_regex
  physical_ad1            = module.ad_map.physical_ad1.name
  application_alias = "image-release-validator-ccm-csi-${local.execution_target.additional_locals.stage}"
  env_vars = []
}

module "odo_configuration_ccm_csi_infra" {
  source = "./shared_modules/odo_configuration"
  execution_target = local.execution_target

  realm                   = lower(local.execution_target.region.realm)
  stage                   = local.execution_target.additional_locals.stage
  artifact_set_identifier = "release-validator-ccm-csi"
  compartment_id          = local.execution_target.tenancy_ocid
  pool_name_regex         = local.execution_target.additional_locals.pool_name_regex
  physical_ad1            = module.ad_map.physical_ad1.name
  application_alias = "infra-release-validator-ccm-csi-${local.execution_target.additional_locals.stage}"
  env_vars = [
    {
      name = "REGIONAL_IMAGE_LIST_1"
      value = join(",", local.regional_image_list_first_half)
    },
    {
      name = "REGIONAL_IMAGE_LIST_2"
      value = join(",", local.regional_image_list_second_half)
    },
    {
      name = "OVERRIDES_IMAGE_LIST_1"
      value = join(",", local.overrides_image_list_first_half)
    },
    {
      name = "OVERRIDES_IMAGE_LIST_2"
      value = join(",", local.overrides_image_list_second_half)
    }
  ]
}

#module "odo_deployment_ccm_csi_infra" {
#  source = "./odo_deployment"
#
#  artifact_version = {
#    uri = "release-validator-ccm-csi-${local.pop_version}.tar.gz"
#    type = "pop"
#    version = local.pop_version
#  }
#  apps             = [
#    for i in range(length(data.odo_applications.infra-release-validator-ccm-csi)) : {
#      ad    = module.ad_map.physical_ad1.name,
#      alias = lookup(data.odo_applications.infra-release-validator-ccm-csi[i].applications[0], "alias", null)
#    }
#  ]
#  depends_on = [module.odo_configuration_ccm_csi_infra]
#}

resource "capability_require_capability" "regional_infra" {
  name = "oke_deploy_odo"
}