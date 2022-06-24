// saving the format for phase 2
/*variable "root_compartment_ocid" {}
variable "compartment_name" {}
variable "alarms_enabled" {}
variable "project" {}
variable "mapi_fleet_format" {}
variable "mapi_hostmetrics_fleet" {}
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
variable "leadership_jira_component" {}
variable "runbook_base" {}
variable "mapi_jira_item" {}
variable "mapi_api_alarm_label_format" {}
variable "mapi_infra_alarm_label_format" {}
variable "mapi_availability_severity" {}
variable "mapi_latency_severity" {}
variable "mapi_api_alarms_enabled" {}
variable "mapi_availability_threshold" {}
variable "mapi_latency_threshold" {}
variable "mapi_dashboard" {}
variable "realm" {}
variable "mapi_availability_runbook" {}
variable "mapi_latency_runbook" {}
variable "mapi_failure_runbook" {}
variable "kmon_jira_item" {}
variable "kmon_infra_alarm_label_format" {}
variable "kmon_alarm_label_format" {}
variable "kmon_fleet_format" {}
variable "kmon_alarms_enabled" {}
variable "kmon_dashboard" {}
variable "worker_jira_item" {}
variable "worker_alarm_label_format" {}
variable "worker_fleet_format" {}
variable "worker_hostmetrics_fleet" {}
variable "worker_alarms_enabled" {}
variable "worker_dashboard" {}
variable "mapi_lb" {}
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

  mapi_hostmetrics_fleet   = var.realm != "oc1" ? "${local.prod_tenancy_name}.${var.mapi_hostmetrics_fleet}" : var.mapi_hostmetrics_fleet
  worker_hostmetrics_fleet = var.realm != "oc1" ? "${local.prod_tenancy_name}.${var.worker_hostmetrics_fleet}" : var.worker_hostmetrics_fleet

  cell_name = format(var.cell_compartment_format, var.cell_index)
  mapi = {
    t2_fleet     = var.realm != "oc1" ? "${local.prod_tenancy_name}.${format(var.mapi_fleet_format, var.cell_index)}" : format(var.mapi_fleet_format, var.cell_index)
    lj_namespace = format("oke%s-api-%s", var.env_name != "prd" ? "-${var.env_name}" : "", local.cell_name)
  }
  kmon = {
    t2_fleet     = var.realm != "oc1" ? "${local.prod_tenancy_name}.${format(var.kmon_fleet_format, var.cell_index)}" : format(var.kmon_fleet_format, var.cell_index)
    lj_namespace = format("oke%s-monitor-%s", var.env_name != "prd" ? "-${var.env_name}" : "", local.cell_name)
  }
  worker = {
    t2_fleet     = var.realm != "oc1" ? "${local.prod_tenancy_name}.${format(var.worker_fleet_format, var.cell_index)}" : format(var.worker_fleet_format, var.cell_index)
    lj_namespace = format("oke%s-worker-%s", var.env_name != "prd" ? "-${var.env_name}" : "", local.cell_name)
  }

  wf_tenant_prefix    = var.env_name != "prd" ? "oke-${var.env_name}-mp-${local.cell_name}" : "oke-mp-${local.cell_name}"
  wf_cp_tenant_prefix = "oke-cp-${var.env_name == "prd" ? "prod" : var.env_name}"
  splat_name          = var.env_name != "prd" ? "oke-mapi-${local.cell_name}" : "oke-mapi-${local.cell_name}-${var.env_name}"

  mapi_lb_host = concat(regex("([^.]*)$", var.mapi_lb.id))
  grafana_uri = {
    oc5 = "https://grafana.us-tacoma-1.oci.oraclerealm5.com"
    oc6 = "https://grafana.us-gov-fortworth-1.oci.oraclerealm.ic.gov"
    oc7 = "https://grafana.us-gov-sterling-1.oci.oci.ic.gov"
  }

  envSetup                          = var.env_name == "dev" ? var.env_name : "${var.env_name == "prd" ? "prod" : var.env_name}-cell${var.cell_index}"
  grafana_base                      = lookup(local.grafana_uri, var.realm, "https://grafana.oci.oraclecorp.com")
  grafana_mapi_lb                   = "${local.grafana_base}/d/P8zcjpsMz/mapi-load-balancer-dashboard?orgId=1"
  grafana_mapi_operation_template   = "${local.grafana_base}/d/E1URIW5Mz/oke-management-plane-api-operations?orgId=1&var-realm=${var.realm}&var-region=${var.region}&var-granularity=1m&var-project=kubernetes&var-mapi_fleet=${local.mapi.t2_fleet}&var-hostmetrics_project=hostmetrics&var-hostmetrics_hosts=All&var-splat_proxy_project=splat-proxy&var-splat_proxy_fleet=splat-proxy-overlay&var-ad=All&var-apis=All&var-splat_mapi_project=${local.splat_name}&var-envSetup=${local.envSetup}&var-buckets=CLUSTER_STATE&var-bucket_name=${local.wf_tenant_prefix}&var-hostmetrics_fleet=${local.mapi_hostmetrics_fleet}&var-hostmetrics_cell=${local.cell_name}"
  grafana_kmon_operation_template   = "${local.grafana_base}/d/2iaa26KGz/oke-management-plane-monitor-operations?orgId=1&refresh=10s&var-realm=${var.realm}&var-region=${var.region}&var-granularity=1m&var-hostmetrics_project=hostmetrics&var-project=kubernetes&var-kmon_fleet=${local.worker.t2_fleet}&var-hm_fleet=${local.worker_hostmetrics_fleet}&var-hostmetrics_hosts=All&var-ad=All&var-envSetup=${local.envSetup}&var-hostmetrics_cell=${local.cell_name}&var-controllers=All&var-controller_metrics=All&var-leases=All"
  grafana_worker_operation_template = "${local.grafana_base}/d/oG0GFvpGk/oke-management-plane-worker-operations?orgId=1&refresh=5s&var-realm=${var.realm}&var-region=${var.region}&var-granularity=5m&var-project=kubernetes&var-hostmetrics_project=hostmetrics&var-hostmetrics_fleet=${local.worker_hostmetrics_fleet}&var-worker_fleet=${local.worker.t2_fleet}&var-hostmetrics_cell=${local.cell_name}&var-hostmetrics_hosts=All&var-ad=All&var-envSetup=${local.envSetup}&var-wf_tenant_prefix=${local.wf_tenant_prefix}&var-wf_cp_tenant_prefix=${local.wf_cp_tenant_prefix}"
  grafana_log_utilization_template  = "${local.grafana_base}/d/000009733/oke-lumberjack-metrics-draft-twilley?orgId=1&var-ad=All&var-region=T2-${var.region}&var-granularity=60&var-tenancy=${var.root_compartment_ocid}"
  is_cp_canary_successful           = data.capability.oci_containerengine_cluster.is_available
  alarms_enabled                    = var.alarms_enabled && local.is_cp_canary_successful
  mapi_api_alarms_enabled           = var.mapi_api_alarms_enabled && local.is_cp_canary_successful
  kmon_alarms_enabled               = var.kmon_alarms_enabled && local.is_cp_canary_successful
  worker_alarms_enabled             = var.worker_alarms_enabled && local.is_cp_canary_successful
}

# Duplicated in non-production
module "compartment_lookup" {
  source                = "../compartment/compartment_lookup"
  root_compartment_ocid = var.root_compartment_ocid
  name                  = var.compartment_name
}
*/