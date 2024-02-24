variable "root_compartment_ocid" {}
variable "compartment_name" {}
variable "alarms_enabled" {}
variable "project" {}
variable "region" {}
variable "region_code" {}
variable "cell_index" {}
variable "cell_compartment_format" {}
variable "env_name" {}
variable "severity_2" {}
variable "severity_3" {}
variable "severity_4" {}
variable "jira_project" {}
variable "jira_component" {}
variable "okenp_jira_project" {}
variable "okenp_jira_component" {}
variable "leadership_jira_component" {}
variable "runbook_base" {}
variable "realm" {}
variable "ccm_alarms_enabled" {}
variable "csi_alarms_enabled" {}
variable "ccm_jira_item" {}
variable "csi_jira_item" {}
variable "ccm_alarm_label_format" {}
variable "ccm_alarms_fleet_format" {}
variable "csi_dataplane_alarms_fleet_format" {}
variable "watch_mp_release_label" {}
variable "skip_mapi_alarms" {}
variable "skip_kmon_alarms" {}
variable "skip_worker_alarms" {}

locals {
  prod_tenancy_realms = {
    oc1 = lookup({
      "dev" : "ociokedev",
      "integ" : "ociokeinteg",
      "polaris" : "okeplatformdev"
    }, var.env_name, "odx-oke")
    oc2 = "okeprodoc2"
    oc3 = "okeprodoc3"
    oc4 = "oke-prod-oc4"
  }
  prod_tenancy_name = lookup(local.prod_tenancy_realms, var.realm, "oke-prod")

  cell_name = format(var.cell_compartment_format, var.cell_index)

  ccm = {
    t2_fleet     = var.realm != "oc1" ? "${local.prod_tenancy_name}.${format(var.ccm_alarms_fleet_format, var.cell_index)}" : format(var.ccm_alarms_fleet_format, var.cell_index)
    lj_namespace = format("oke%s-kmi-%s", var.env_name != "prd" ? "-${var.env_name}" : "", local.cell_name)
  }

  csi = {
    t2_fleet     = var.realm != "oc1" ? "${local.prod_tenancy_name}.${format(var.csi_dataplane_alarms_fleet_format, var.region_code, "")}" : tonumber(var.cell_index) == 0 ? format(var.csi_dataplane_alarms_fleet_format, var.region_code, "") : format(var.csi_dataplane_alarms_fleet_format, var.region_code, format("-cell%s", var.cell_index))
    lj_namespace = format("oke%s-kmi-%s", var.env_name != "prd" ? "-${var.env_name}" : "", local.cell_name)
  }

  grafana_uri = {
    oc5 = "https://grafana.us-tacoma-1.oci.oraclerealm5.com"
    oc6 = "https://grafana.us-gov-fortworth-1.oci.oraclerealm.ic.gov"
  }

  envSetup                          = var.env_name == "dev" ? var.env_name : "${var.env_name == "prd" ? "prod" : var.env_name}-cell${var.cell_index}"
  grafana_base                      = lookup(local.grafana_uri, var.realm, "https://grafana.oci.oraclecorp.com")
  grafana_ccm_template              = "${local.grafana_base}/d/eaDdAjE7k/cpo-monitoring-calls-made-via-kmi?panelId=66&fullscreen&orgId=1&refresh=30s&from=now-6h&to=now-1m&var-realm=${var.realm}&var-granularity=1h&var-MQLRegion=${var.region}&var-fleet=${local.cell_name}"
  grafana_csi_template              = "${local.grafana_base}/d/1vt7X4gIz/storage-plugins-block-volume-observability?&orgId=1&refresh=1m&var-realm=${var.realm}&var-granularity=1h&var-MQLRegion=${var.region}&var-fleet=${local.csi.t2_fleet}"
  grafana_csi_oci_calls_template    = "${local.grafana_base}/d/dAA2EzvIz/storage-plugins-dependent-oci-service-observability?orgId=1&from=now-30m&to=now-1m&var-realm=${var.realm}&var-granularity=4m&var-MQLRegion=${var.region}&var-fleet=${local.ccm.t2_fleet}"
  is_cp_canary_successful           = data.capability.oci_containerengine_cluster.is_available
  alarms_enabled                    = var.alarms_enabled && local.is_cp_canary_successful
  ccm_alarms_enabled                = var.ccm_alarms_enabled && local.is_cp_canary_successful && var.realm == "oc1"
  csi_alarms_enabled                = var.csi_alarms_enabled && local.is_cp_canary_successful && var.realm == "oc1"
}

# Duplicated in non-production
module "compartment_lookup" {
  source                = "../compartment/compartment_lookup"
  root_compartment_ocid = var.root_compartment_ocid
  name                  = var.compartment_name
}

module "ad_map" {
  source                = "../ad_map"
  root_compartment_ocid = var.root_compartment_ocid
}

module "orchestration_compartment" {
  source                = "../compartment/compartment_lookup"
  root_compartment_ocid = var.root_compartment_ocid
  name                  = "${local.cell_name}.mp.orchestration"
}


module "ccm_alarms" {
  source      = "./ccm_alarms"
  skip_alarms = var.skip_kmon_alarms

  compartment_ocid            = module.compartment_lookup.ocid
  project                     = var.project
  fleet                       = local.ccm.t2_fleet
  region                      = var.region
  region_code                 = var.region_code
  realm                       = var.realm
  cell                        = local.cell_name
  env_name                    = var.env_name
  severity_2                  = var.severity_2
  severity_3                  = var.severity_3
  severity_4                  = var.severity_4
  jira_project                = var.jira_project
  jira_component              = var.jira_component
  leadership_jira_component   = var.leadership_jira_component
  ccm_jira_item               = var.ccm_jira_item
  runbook_base                = var.runbook_base
  enabled                     = local.alarms_enabled && local.ccm_alarms_enabled
  label                       = format(var.ccm_alarm_label_format, var.cell_index)
  dashboard                   = local.grafana_ccm_template
  num_of_ads                  = length(keys(module.ad_map.physical_to_logical_map))
  root_compartment_ocid       = var.root_compartment_ocid
  cell_name                   = local.cell_name
  prod_tenancy_name           = local.prod_tenancy_name
  lumberjack_cell_compartment = module.orchestration_compartment.ocid
  watch_mp_release_label      = var.watch_mp_release_label
  lumberjack_namespace        = local.ccm.lj_namespace
}

module "csi_alarms" {
  source      = "./csi_alarms"
  skip_alarms = var.skip_kmon_alarms

  compartment_ocid            = module.compartment_lookup.ocid
  project                     = var.project
  fleet                       = local.csi.t2_fleet
  kmi_fleet                   = local.ccm.t2_fleet
  region                      = var.region
  region_code                 = var.region_code
  realm                       = var.realm
  cell                        = local.cell_name
  env_name                    = var.env_name
  severity_2                  = var.severity_2
  severity_3                  = var.severity_3
  severity_4                  = var.severity_4
  jira_project                = var.okenp_jira_project
  jira_component              = var.okenp_jira_component
  leadership_jira_component   = var.leadership_jira_component
  jira_item                   = var.csi_jira_item
  runbook_base                = var.runbook_base
  enabled                     = local.alarms_enabled && local.csi_alarms_enabled
  label                       = format(var.ccm_alarm_label_format, var.cell_index)
  dashboard                   = local.grafana_csi_template
  dashboard_oci_calls         = local.grafana_csi_oci_calls_template
  num_of_ads                  = length(keys(module.ad_map.physical_to_logical_map))
  root_compartment_ocid       = var.root_compartment_ocid
  cell_name                   = local.cell_name
  prod_tenancy_name           = local.prod_tenancy_name
  lumberjack_cell_compartment = module.orchestration_compartment.ocid
  watch_mp_release_label      = var.watch_mp_release_label
  lumberjack_namespace        = local.csi.lj_namespace
}
