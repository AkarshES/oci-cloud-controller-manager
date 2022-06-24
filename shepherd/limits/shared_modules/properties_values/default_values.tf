locals {
  global_default_values = {
    // The "default" mapping uses a SHA that references a manifest list. At runtime, the manifest list will resolve to
    // the actual image based on the node's architecture.
    csi_image_version_mapping = {
      "default" = {
        "all" : "{}"
      }
      // To override property values for an environment declare a key
      // by environment name. To override values for a qualified realm, declare
      // a key with qualified realm name
      "prd.oc1" = {
        "all" : chomp(<<-EOF
      {
        "default": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
        "v1.16": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
        "v1.17": "oke-multiarch-1.17-6c62578-11@sha256:60f395b0e5688bfd39ea4c6272cf1dfb24016c372015de3226f0fdea55df0a26",
        "v1.18": "oke-multiarch-1.17-6c62578-11@sha256:60f395b0e5688bfd39ea4c6272cf1dfb24016c372015de3226f0fdea55df0a26",
        "v1.19": "oke-multiarch-1.19-b8c736a-86@sha256:beac2bf9a7efccc76d66f57a51a5b4c6efa1cb3a0166c4a6051063677ee456e5",
        "v1.20": "oke-multiarch-1.19-b8c736a-86@sha256:beac2bf9a7efccc76d66f57a51a5b4c6efa1cb3a0166c4a6051063677ee456e5",
        "v1.21": "oke-multiarch-1.19-b8c736a-86@sha256:beac2bf9a7efccc76d66f57a51a5b4c6efa1cb3a0166c4a6051063677ee456e5",
        "v1.22": "oke-multiarch-1.22-dbf1420-19@sha256:9ec69f55ead3b0be33bcfeac639ece2267b851ab0f184f2de2f5f343a09c9aef"
      }
      EOF
        )
      }
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


    ccm_image_version_mapping = {
      "default" = {
        "all" : "{}"
      }
      // To override property values for an environment declare a key
      // by environment name. To override values for a qualified realm, declare
      // a key with qualified realm name
      "prd.oc1" = {
        "all" : chomp(<<-EOF
      {
        "default": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
        "v1.16": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
        "v1.17": "oke-multiarch-1.17-6c62578-11@sha256:60f395b0e5688bfd39ea4c6272cf1dfb24016c372015de3226f0fdea55df0a26",
        "v1.18": "oke-multiarch-1.17-6c62578-11@sha256:60f395b0e5688bfd39ea4c6272cf1dfb24016c372015de3226f0fdea55df0a26",
        "v1.19": "oke-multiarch-1.19-b8c736a-86@sha256:beac2bf9a7efccc76d66f57a51a5b4c6efa1cb3a0166c4a6051063677ee456e5",
        "v1.20": "oke-multiarch-1.19-b8c736a-86@sha256:beac2bf9a7efccc76d66f57a51a5b4c6efa1cb3a0166c4a6051063677ee456e5",
        "v1.21": "oke-multiarch-1.19-b8c736a-86@sha256:beac2bf9a7efccc76d66f57a51a5b4c6efa1cb3a0166c4a6051063677ee456e5",
        "v1.22": "oke-multiarch-1.22-dbf1420-19@sha256:9ec69f55ead3b0be33bcfeac639ece2267b851ab0f184f2de2f5f343a09c9aef"
      }
      EOF
        )
      }
    }
  }
  global_default_values_by_property = { for property_name, property_value in local.global_default_values : property_name => merge(
    lookup(property_value, "default", {}),
    lookup(property_value, var.env, {}),
    lookup(property_value, "${var.env}.${var.realm}", {})
    )
  }

  global_default_values_list = flatten([for property_name, property_value in local.global_default_values_by_property : [
    {
      ad     = lookup(property_value, "ad", "all")
      group  = var.spectre_group_name
      name   = replace(property_name, "_", "-")
      region = var.execution_target.additional_locals.limits_region
      value  = lookup(property_value, var.execution_target.additional_locals.limits_region, lookup(property_value, "all", ""))
      min    = lookup(property_value, "min", null)
      max    = lookup(property_value, "max", null)
    }] if length(property_value) > 0
  ])

  global_default_values_map = {
    for property in local.global_default_values_list : "${property.group}/${property.name}/${property.region}/${property.ad}" => property
  }
}
