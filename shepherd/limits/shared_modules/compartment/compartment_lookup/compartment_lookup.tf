// A module to lookup a compartment given it's name
variable root_compartment_ocid {}
variable name {}

module "compartments_lookup" {
  source = "../compartments_lookup"
  root_compartment_ocid = var.root_compartment_ocid
  names = [var.name]
}

output "ocid" {
  value = [for compartment in module.compartments_lookup.compartments: compartment.id if lower(compartment.name) == var.name ][0]
}


