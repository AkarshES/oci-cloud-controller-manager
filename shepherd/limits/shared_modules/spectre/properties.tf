variable "root_compartment_ocid" {}
variable "group_name" {}
variable "env" {}
variable "realm" {}
variable "region" {}

locals {
  region         = var.region
  ## Current realms is a list of realms that already have property values defined at 
  ## global/realm level before limits started blocking global updates. It is to main backward
  ## compatibility so that TF/Shepherd does not try to delete them. Any new realm will not have these
  ## properties created at realm level
  current_realms = ["oc1", "oc2", "oc3", "oc4", "oc6", "oc8", "oc9", "oc10"]

  property_defs = {
    "csi-image-version-mapping" = {
      type        = "STRING"
      description = "Json map of kubernetes versions to CSI tag+sha256"
      default_value = chomp(<<-EOF
      {}
      EOF
      )
    }
    "ccm-image-version-mapping" = {
      type        = "STRING"
      description = "Json map of kubernetes versions to Cloud Provider OCI image tag+sha256"
      default_value = chomp(<<-EOF
      {}
      EOF
      )
    }
    // check if this property is needed to be imported
    "csi-fss-node-driver-registrar-image-version-mapping" = {
      type        = "STRING"
      description = "Json map of kubernetes versions to CSI node driver registrar tag+sha256"
      default_value = chomp(<<-EOF
      {}
      EOF
      )
    }
    // check if this property is needed to be imported
    "csi-fss-node-driver-image-version-mapping" = {
      type        = "STRING"
      description = "Json map of kubernetes versions to CSI node driver tag+sha256"
      default_value = chomp(<<-EOF
      {}
      EOF
      )
    }
    // check if this property is needed to be imported
    "fss-csi-driver-enabled" = {
      type        = "ENUM"
      description = "Enable fss encryption"
      options = [
        "true",
      "false"]
      default_value = "false"
    }
    // check if this property is needed to be imported
    "csi-bv-expansion-enabled" = {
      type        = "ENUM"
      description = "Enable block volume expansion"
      options = [
        "true",
      "false"]
      default_value = "false"
    }
    "oci-service-controller-enabled" = {
      type = "ENUM"
      description = "Enable custom OCI service controller for SKE support"
      options = [
        "true",
        "false"]
      default_value = "false"
    }
  }

  

  // The "default" mapping uses a SHA that references a manifest list. At runtime, the manifest list will resolve to
  // the actual image based on the node's architecture.
  csi_image_version_mapping = {
    "default" = {
      "all" : "{}"
    }
    // To override property values for an environment declare a key
    // by environment name. To override values for a qualified realm, declare
    // a key with qualified realm name
  }

  ccm_image_version_mapping = {
    "default" = {
      "all" : "{}"
    }
    // To override property values for an environment declare a key
    // by environment name. To override values for a qualified realm, declare
    // a key with qualified realm name
  }

  // The "default" mapping uses a SHA that references a manifest list. At runtime, the manifest list will resolve to
  // the actual image based on the node's architecture.
  csi_fss_node_driver_registrar_image_version_mapping = {
    "default" = {
      "all" : "{}"
    }


    // To override property values for an environment declare a key
    // by environment name. To override values for a qualified realm, declare
    // a key with qualified realm name
  }

  // The "default" mapping uses a SHA that references a manifest list. At runtime, the manifest list will resolve to
  // the actual image based on the node's architecture.
  csi_fss_node_driver_image_version_mapping = {
    "default" = {
      "all" : "{}"
    }

    // To override property values for an environment declare a key
    // by environment name. To override values for a qualified realm, declare
    // a key with qualified realm name
  }
}

resource "property_definition" property {
  for_each      = local.property_defs
  group         = var.group_name
  name          = each.key
  type          = each.value.type
  description   = each.value.description
  default_value = each.value.default_value
  options       = lookup(each.value, "options", null)
}

resource "property_value" csi_image_version_mapping {
  for_each = contains(local.current_realms, var.realm) && var.env != "rbaas" ? merge(
    lookup(local.csi_image_version_mapping, "default", {}),
    lookup(local.csi_image_version_mapping, var.env, {}),
    lookup(local.csi_image_version_mapping, "${var.env}.${var.realm}", {})
  ) : {}
  group      = var.group_name
  name       = "csi-image-version-mapping"
  region     = each.key
  ad         = "all"
  value      = each.value
  depends_on = [property_definition.property["csi-image-version-mapping"]]
}

resource "property_value" ccm_image_version_mapping {
  for_each = contains(local.current_realms, var.realm) && var.env != "rbaas" ? merge(
    lookup(local.ccm_image_version_mapping, "default", {}),
    lookup(local.ccm_image_version_mapping, var.env, {}),
    lookup(local.ccm_image_version_mapping, "${var.env}.${var.realm}", {})
  ) : {}
  group      = var.group_name
  name       = "ccm-image-version-mapping"
  region     = each.key
  ad         = "all"
  value      = each.value
  depends_on = [property_definition.property["ccm-image-version-mapping"]]
}

resource "property_value" csi_fss_node_driver_registrar_image_version_mapping {
  for_each = contains(local.current_realms, var.realm) && var.env != "rbaas" ? merge(
    lookup(local.csi_fss_node_driver_registrar_image_version_mapping, "default", {}),
    lookup(local.csi_fss_node_driver_registrar_image_version_mapping, var.env, {}),
    lookup(local.csi_fss_node_driver_registrar_image_version_mapping, "${var.env}.${var.realm}", {})
  ) : {}
  group      = var.group_name
  name       = "csi-fss-node-driver-registrar-image-version-mapping"
  region     = each.key
  ad         = "all"
  value      = each.value
  depends_on = [property_definition.property["csi-fss-node-driver-registrar-image-version-mapping"]]
}

resource "property_value" csi_fss_node_driver_image_version_mapping {
  for_each = contains(local.current_realms, var.realm) && var.env != "rbaas" ? merge(
    lookup(local.csi_fss_node_driver_image_version_mapping, "default", {}),
    lookup(local.csi_fss_node_driver_image_version_mapping, var.env, {}),
    lookup(local.csi_fss_node_driver_image_version_mapping, "${var.env}.${var.realm}", {})
  ) : {}
  group      = var.group_name
  name       = "csi-fss-node-driver-image-version-mapping"
  region     = each.key
  ad         = "all"
  value      = each.value
  depends_on = [property_definition.property["csi-fss-node-driver-image-version-mapping"]]
}


resource "property_override" overrides {
  for_each = lookup(local.tenancy_property_overrides, "${var.env}.${var.realm}", {})
  group    = var.group_name
  name     = each.key
  region   = lookup(each.value, "region", null)
  ad       = "all"
  tag      = lookup(each.value, "tenancy_ocid", null)
  value    = lookup(each.value, "value", null)
  min      = lookup(each.value, "min", null)
  max      = lookup(each.value, "max", null)
}
