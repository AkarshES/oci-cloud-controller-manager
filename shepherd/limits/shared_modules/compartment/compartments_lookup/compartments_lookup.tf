// A module to lookup multiple compartments given the names
variable root_compartment_ocid {}
variable names { type = list(string) }

data "oci_identity_compartments" "all_compartments" {
  compartment_id = var.root_compartment_ocid
  access_level = "ANY"
  compartment_id_in_subtree = true

  filter {
    name   = "name"
    values = var.names
  }
}

output "compartments" {
    value = coalescelist(data.oci_identity_compartments.all_compartments.compartments, [{"id" : "id for ${var.names[0]} NOT FOUND", "name" : var.names[0]}])
}


