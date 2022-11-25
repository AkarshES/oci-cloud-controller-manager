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
  grafana_template = "${var.dashboard}&fullscreen&panelId=%d"
  grafana_template_with_no_panel_id = "${var.dashboard}&fullscreen"
  lumberjack_uri = "https://devops.oci.oraclecorp.com/logs?region=${var.region_code}&ad=_all&namespace=&from=-30m&to=now&query=&tenant=${var.prod_tenancy_name}&apiVersion=v2&keywords%5B0%5D%5Bfield%5D%5Bvalue%5D%5Bname%5D=fileId&keywords%5B0%5D%5Bfield%5D%5Bvalue%5D%5BfieldType%5D=SORTED&keywords%5B0%5D%5Bfield%5D%5Bvalue%5D%5Btype%5D=STRING&keywords%5B0%5D%5Bfield%5D%5Blabel%5D=fileId&keywords%5B0%5D%5Bop%5D%5Bvalue%5D=STRING_MATCH&keywords%5B0%5D%5Bop%5D%5Blabel%5D=%3D&keywords%5B0%5D%5Bop%5D%5Btypes%5D%5B0%5D=NONE&keywords%5B0%5D%5Bop%5D%5Btypes%5D%5B1%5D=SORTED&keywords%5B0%5D%5Bvalue%5D=cpo.logs&keywords%5B0%5D%5Bid%5D=fileId%3Dcpo.logs&group=none&sort=oldest&timezone=utc&isReverse=false&pivotAggregation=true&timeGroup%5Bvalue%5D=FIVE_MINUTES&timeGroup%5Blabel%5D=5%20Minute&resultLimit%5Bvalue%5D=none&resultLimit%5Blabel%5D=No%20Limit&dimension1%5Bvalue%5D%5Bname%5D=NONE&dimension1%5Blabel%5D=None&aggregation1%5Bvalue%5D%5Bname%5D=NONE&aggregation1%5Blabel%5D=None&operation%5Bvalue%5D=COUNT&operation%5Blabel%5D=Count&columns%5B0%5D=ts&columns%5B1%5D=msg&tabId=search&namespaces%5B0%5D%5Bname%5D=_${var.lumberjack_namespace}&namespaces%5B0%5D%5Bcompartment%5D=${var.lumberjack_cell_compartment}"
  ocelot_uri_template = "https://devops.oci.oraclecorp.com/kubernetes-engine/?regionId=${var.region}&tabId=clusters"
  cross_region_dedupe = true
}

resource "telemetry_alarm" "cloud_provider_oci_container_restarts" {
  compartment_id = var.compartment_ocid
  project = var.project
  fleet = var.fleet
  display_name = "${var.fleet}-cloud-provider-oci-restarts in ${var.region}"
  query = "OKE.KMI.HostAgent.Pod.Container.Restarts[10m]{name=\"cloud-provider-oci\"}.groupBy(clusterId,resourceId).filter(x=>x>0).max().groupBy(clusterId).filter(x=>x>10).count().filter(x => x==1)"
  severity = var.severity_3
  dedupe_key = "CpoRestartsForSingleKmi"
  is_enabled = var.enabled
  is_dedupe_key_cross_region = local.cross_region_dedupe
  pending_duration = "PT5M"
  body = <<EOT
OKE.KMI.HostAgent.Pod.Container.Restarts - cloud-provider-oci container is continuously restarting on 1 out of 3 KMI
See [OCI Grafana Dashboard.|${format(local.grafana_template, 630)}]
For Runbook instructions, please see [this runbook here|${var.runbook_base}/oke-cloud-provider-oci-restarts].
For service logs, see [lumberjack link|${local.lumberjack_uri}]
For OCELOT (Oracle Container Engine Live Operations Tool), see [OCELOT link|${local.ocelot_uri_template}]
EOT
  destinations {
    jira {
      project = var.jira_project
      component = var.jira_component
      item = var.ccm_jira_item
    }
  }
  labels = [ var.label, var.watch_mp_release_label, "KMI-CPO-restart" ]
}
