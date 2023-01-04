# This module is responsible for creating any compartments required by MP
variable "root_compartment_ocid" {}

resource "oci_identity_compartment" "etcd_backups_compartment" {
  compartment_id = var.root_compartment_ocid
  description = var.etcd_backups_compartment_description
  name = var.etcd_backups_compartment_name
}