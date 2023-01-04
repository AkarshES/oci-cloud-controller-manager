// Lookup v2 compartments for all cells, given cell_count

variable "root_compartment_ocid" {}
variable "cell_count" {}
variable "cell_compartment_format" {}
variable "orchestration_compartment_format" {}
variable "kmi_compartment_format" {}
locals {
  orchestration_compartment_fqcns = [
    for i in range(var.cell_count):
      format("${var.cell_compartment_format}:${var.cell_compartment_format}.mp:${var.orchestration_compartment_format}", i, i, i)
  ]
}

data "oci_identity_compartments" "cell_compartment" {
  count = var.cell_count
  compartment_id = var.root_compartment_ocid
  access_level = "ANY"
  compartment_id_in_subtree = true

  filter {
    name   = "name"
    values = [ format(var.cell_compartment_format, count.index) ]
  }
}

data "oci_identity_compartments" "orchestration_compartment" {
  count = var.cell_count
  compartment_id = var.root_compartment_ocid
  access_level = "ANY"
  compartment_id_in_subtree = true

  filter {
    name   = "name"
    values = [ format(var.orchestration_compartment_format, count.index) ]
  }
}

data "oci_identity_compartments" "kmi_compartment" {
  count = var.cell_count
  compartment_id = var.root_compartment_ocid
  access_level = "ANY"
  compartment_id_in_subtree = true

  filter {
    name   = "name"
    values = [ format(var.kmi_compartment_format, count.index) ]
  }
}

output "orchestration_compartments" {
  value = flatten([data.oci_identity_compartments.orchestration_compartment.*.compartments[*]])
}

output "orchestration_compartment_fqcns" {
  value = local.orchestration_compartment_fqcns
}