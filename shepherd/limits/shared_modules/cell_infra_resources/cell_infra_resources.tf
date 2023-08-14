variable "root_compartment_ocid" {}
variable "region_code" {}
variable "cell_index" {}
variable "cell_compartment_format" {}
variable "orchestration_compartment_format" {}
variable "kmi_compartment_format" {}
variable "vcn_format" {}
variable "mapi_instance_subnet_name" {}
variable "mapi_lb_subnet_name" {}
variable "mp_worker_subnet_name" {}
variable "kmi_subnet_name" {}
variable "env_name" {}
variable "tag_ns_name" {}
variable "instance_type_tag_name" {}
variable "api_hostclass" {}
variable "lb_shape" {}
variable "enable_kaas_regional_instance" {}
variable "kaas_name_format" {}
variable "region_public_name" {}
variable "enable_wfaas" {}
variable "wfaas_name_format" {}
variable "api_log_namespace_format" {}
variable "monitor_log_namespace_format" {}
variable "worker_log_namespace_format" {}
variable "phonebook_name" {}
variable "sms_namespace_name_format" {}
variable "certificate_name_format" {}
variable "certificate_compartment" {}
variable "alarms_enabled" {}
variable "alarms_compartment" {}
variable "alarms_project" {}
variable "jira_project" {}
variable "jira_component" {}
variable "leadership_jira_component" {}
variable "runbook_base" {}
variable "severity_2" {}
variable "severity_3" {}
variable "severity_4" {}
variable "has_mapi_grafana_dashboard" {}
variable "image_name" {}
variable "image_url" {}
variable "splunk_enabled" {}
variable "mapi_lb_name" {}
variable "realm" {}
variable "splat_base_service_name" {}
variable "splat_service_name_format" {}
variable "splat_service_fleet" {}
variable "oke_secrets_namespace" {}
variable "cp_vcn_name_format" {}
variable "cp_vcn_compartment" {}
variable "pl_vcn_name" {}
variable "pl_vcn_compartment" {}
variable "prime_vcn_compartment" {}
variable "prime_vcn_name" {}
variable "watch_mp_release_label_format" {}
variable "skip_mapi_alarms" {}
variable "skip_kmon_alarms" {}
variable "skip_worker_alarms" {}
variable "skip_dns" {}
variable "image_type" {}
variable "ccm_jira_item" {}
variable "ccm_alarms_enabled" {}
variable "csi_alarms_enabled" {}
variable "ccm_alarm_label_format" {}
variable "ccm_alarms_fleet_format" {}
variable "csi_dataplane_alarms_fleet_format" {}

locals {
  cp_vcn_name = var.env_name == "prd" ? format(var.cp_vcn_name_format,"prod") : format(var.cp_vcn_name_format,var.env_name)
  watch_mp_release_label = format(var.watch_mp_release_label_format, var.cell_index)
  certificate_compartment = tonumber(var.cell_index) == 0 && var.env_name == "prd" ? "assets" : format(var.orchestration_compartment_format, var.cell_index)
  alarms_compartment      = coalesce(var.alarms_compartment, format(var.orchestration_compartment_format, var.cell_index))
}

module "alarms" {
  source = "../alarms"
  root_compartment_ocid = var.root_compartment_ocid
  region_code = var.region_code
  compartment_name = local.alarms_compartment
  cell_compartment_format = var.cell_compartment_format
  cell_index = var.cell_index
  env_name = var.env_name
  alarms_enabled = var.alarms_enabled
  project = var.alarms_project
  jira_project = var.jira_project
  jira_component = var.jira_component
  leadership_jira_component = var.leadership_jira_component
  runbook_base = var.runbook_base
  severity_2 = var.severity_2
  severity_3 = var.severity_3
  severity_4 = var.severity_4
  realm = var.realm
  region = var.region_public_name
  watch_mp_release_label = local.watch_mp_release_label
  skip_kmon_alarms = var.skip_kmon_alarms
  skip_mapi_alarms = var.skip_mapi_alarms
  skip_worker_alarms = var.skip_worker_alarms
  ccm_alarm_label_format = var.ccm_alarm_label_format
  ccm_alarms_enabled = var.ccm_alarms_enabled
  csi_alarms_enabled = var.csi_alarms_enabled
  ccm_jira_item = var.ccm_jira_item
  ccm_alarms_fleet_format = var.ccm_alarms_fleet_format
  csi_dataplane_alarms_fleet_format = var.csi_dataplane_alarms_fleet_format
}


/*
module "grafana_dashboard" {
  source = "../grafana_dashboard"
  has_mapi_grafana_dashboard = var.has_mapi_grafana_dashboard
}
*/
