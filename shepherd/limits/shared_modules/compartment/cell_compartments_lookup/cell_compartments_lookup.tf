// Lookup v2 compartments for a given cell (identified by cell_index)

variable "root_compartment_ocid" {}
variable "cell_index" {}
variable "cell_compartment_format" {}
variable "orchestration_compartment_format" {}
variable "kmi_compartment_format" {}
locals {
  cell_compartment_name = format(var.cell_compartment_format, var.cell_index)
  orchestration_compartment_name = format(var.orchestration_compartment_format, var.cell_index)
  kmi_compartment_name = format(var.kmi_compartment_format, var.cell_index)
  orchestration_compartment_fqcn = format("%s:%s.mp:%s", local.cell_compartment_name, local.cell_compartment_name, local.orchestration_compartment_name)
  etcd_backups_compartment_name = "etcd-backups"
}

data "oci_identity_compartments" "cell_compartment" {
  compartment_id = var.root_compartment_ocid
  access_level = "ANY"
  compartment_id_in_subtree = true

  filter {
    name   = "name"
    values = [ local.cell_compartment_name ]
  }
}

data "oci_identity_compartments" "orchestration_compartment" {
  compartment_id = var.root_compartment_ocid
  access_level = "ANY"
  compartment_id_in_subtree = true

  filter {
    name   = "name"
    values = [ local.orchestration_compartment_name ]
  }
}

data "oci_identity_compartments" "kmi_compartment" {
  compartment_id = var.root_compartment_ocid
  access_level = "ANY"
  compartment_id_in_subtree = true

  filter {
    name   = "name"
    values = [ local.kmi_compartment_name ]
  }
}

data "oci_identity_compartments" "etcd_backups_compartment" {
  compartment_id = var.root_compartment_ocid
  access_level = "ANY"
  compartment_id_in_subtree = true

  filter {
    name   = "name"
    values = [ local.etcd_backups_compartment_name ]
  }
}

output "cell_compartment" {
  value = data.oci_identity_compartments.cell_compartment.compartments[0]
}

output "orchestration_compartment" {
  value = data.oci_identity_compartments.orchestration_compartment.compartments[0]
}

output "kmi_compartment" {
  value = data.oci_identity_compartments.kmi_compartment.compartments[0]
}

output "orchestration_compartment_fqcn" {
  value = local.orchestration_compartment_fqcn
}

output "etcd_backups_compartment_id" {
  value = length(data.oci_identity_compartments.etcd_backups_compartment.compartments) == 0 ? "not-exists" : data.oci_identity_compartments.etcd_backups_compartment.compartments[0].id
}