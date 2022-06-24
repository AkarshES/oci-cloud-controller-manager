variable "has_mapi_grafana_dashboard" {}

// saving the format for phase 2
/*resource "grafana_dashboard" "oke_mapi_dashboard" {
  for_each = toset(compact(var.has_mapi_grafana_dashboard == "true" ? ["run"] : []))
  config_json = templatefile("${path.module}/mapi/oke-mapi-operations.json", {})
}
*/