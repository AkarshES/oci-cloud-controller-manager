variable "compartment_ocid" {}
variable "root_compartment_ocid" {}
variable "lumberjack_cell_compartment" {}
variable "prod_tenancy_name" {}
variable "cell_name" {}
variable "region_code" {}
variable "project" {}
variable "fleet" {}
variable "region" {}
variable "realm" {}
variable "cell" {}
variable "env_name" {}
variable "severity_2" {}
variable "severity_3" {}
variable "jira_project" {}
variable "jira_component" {}
variable "leadership_jira_component" {}
variable "ccm_jira_item" {}
variable "runbook_base" {}
variable "enabled" {}
variable "label" {}
variable "dashboard" {}
variable "num_of_ads" {}
variable "skip_alarms" {}
variable "lumberjack_namespace" {}
variable "watch_mp_release_label" {}

locals {
  skip = var.skip_alarms ? 0 : 1
  grafana_template = "${var.dashboard}&fullscreen&panelId=%d&fullscreen"
  grafana_template_with_no_panel_id = "${var.dashboard}&fullscreen"
  attach_lumberjack_uri = "https://devops.oci.oraclecorp.com/logs/v3?ad=All%20ADs&aggregation=count&displayedFields%5B0%5D=ts&displayedFields%5B1%5D=logGroup&displayedFields%5B2%5D=msg&from=2023-08-09T08%3A14%3A12.227Z&granularity=1h&groupBy=%23volumeID&namespaces%5B0%5D%5Bcompartment%5D=${var.lumberjack_cell_compartment}&namespaces%5B0%5D%5BlogType%5D=standard&namespaces%5B0%5D%5Bnamespace%5D=${var.lumberjack_namespace}&region=${var.region}&sortOrder=ASC&sumBy=NONE&phonebook=oracle-kubernetes-engine&to=2023-08-11T08%3A14%3A12.227Z&tenant=${var.prod_tenancy_name}&viewFormat=table&searchBy=classic&fieldFilters%5B0%5D%5Bvalue%5D=BV&fieldFilters%5B0%5D%5BfieldName%5D=%23logger&fieldFilters%5B0%5D%5Boperator%5D=%3D&fieldFilters%5B0%5D%5BfieldType%5D=SORTED&fieldFilters%5B0%5D%5Bhidden%5D=false&fieldFilters%5B1%5D%5Bvalue%5D=timed%20out%20waiting%20for%20condition&fieldFilters%5B1%5D%5BfieldName%5D=%23error&fieldFilters%5B1%5D%5Boperator%5D=%3D&fieldFilters%5B1%5D%5BfieldType%5D=SORTED&fieldFilters%5B1%5D%5Bhidden%5D=false&lineWrap=true&serviceLog=false&showAggregation=true&clickType=2"
  detach_lumberjack_uri = "https://devops.oci.oraclecorp.com/logs/v3?ad=All%20ADs&aggregation=count&displayedFields%5B0%5D=ts&displayedFields%5B1%5D=logGroup&displayedFields%5B2%5D=msg&from=2023-08-09T08%3A14%3A12.227Z&granularity=1h&groupBy=%23volumeID&namespaces%5B0%5D%5Bcompartment%5D=${var.lumberjack_cell_compartment}&namespaces%5B0%5D%5BlogType%5D=standard&namespaces%5B0%5D%5Bnamespace%5D=${var.lumberjack_namespace}&region=${var.region}&sortOrder=ASC&sumBy=NONE&phonebook=oracle-kubernetes-engine&to=2023-08-11T08%3A14%3A12.227Z&tenant=${var.prod_tenancy_name}&viewFormat=table&searchBy=classic&fieldFilters%5B0%5D%5Bvalue%5D=BV&fieldFilters%5B0%5D%5BfieldName%5D=%23logger&fieldFilters%5B0%5D%5Boperator%5D=%3D&fieldFilters%5B0%5D%5BfieldType%5D=SORTED&fieldFilters%5B0%5D%5Bhidden%5D=false&fieldFilters%5B1%5D%5Bvalue%5D=timed%20out%20waiting%20for%20volume%20to%20be%20detached&fieldFilters%5B1%5D%5BfieldName%5D=%23error&fieldFilters%5B1%5D%5Boperator%5D=%3D&fieldFilters%5B1%5D%5BfieldType%5D=SORTED&fieldFilters%5B1%5D%5Bhidden%5D=false&lineWrap=true&serviceLog=false&showAggregation=true&clickType=2"
}

resource "telemetry_alarm" "csi_block_volume_attaching_timeout" {
  compartment_id = var.compartment_ocid
  project = var.project
  fleet = var.fleet
  display_name = "${var.fleet}-csi-block-volume-attaching-timeout in ${var.region}"
  query = "(OKE.CPO.PV_ATTACH[60m]{component=\"CSI_CTX_TIMEOUT\"}.groupBy(resourceOCID).count()).grouping().count().filter(x=>x>10)"
  severity = var.severity_3
  is_enabled = var.enabled
  pending_duration = "PT30M"
  body = <<EOT
OKE.CPO.PV_ATTACH - More than 10 persistent volumes are timing out on attach
See [OCI Grafana Dashboard.|${format(local.grafana_template, 90)}]
For Runbook instructions, please see [this runbook here|${var.runbook_base}/oke-csi-block-volumes-stuck-detaching].
For service logs, see [lumberjack link|${local.attach_lumberjack_uri}]
EOT
  destinations {
    jira {
      project = var.jira_project
      component = var.jira_component
      item = var.ccm_jira_item
    }
  }
  labels = [ var.label, var.watch_mp_release_label ]
}

resource "telemetry_alarm" "csi_block_volume_detaching_timeout" {
  compartment_id = var.compartment_ocid
  project = var.project
  fleet = var.fleet
  display_name = "${var.fleet}-csi-block-volume-detaching-timeout in ${var.region}"
  query = "(OKE.CPO.PV_DETACH[60m]{component=\"CSI_CTX_TIMEOUT\"}.groupBy(resourceOCID).count()).grouping().count().filter(x=>x>10)"
  severity = var.severity_3
  is_enabled = var.enabled
  pending_duration = "PT5M"
  body = <<EOT
OKE.CPO.PV_ATTACH - More than 10 persistent volumes are timing out on detach
See [OCI Grafana Dashboard.|${format(local.grafana_template, 83)}]
For Runbook instructions, please see [this runbook here|${var.runbook_base}/oke-csi-block-volumes-stuck-detaching].
For service logs, see [lumberjack link|${local.detach_lumberjack_uri}]
EOT
  destinations {
    jira {
      project = var.jira_project
      component = var.jira_component
      item = var.ccm_jira_item
    }
  }
  labels = [ var.label, var.watch_mp_release_label ]
}
