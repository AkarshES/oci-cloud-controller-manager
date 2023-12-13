locals {
  global_default_values = {
    // The "default" mapping uses a SHA that references a manifest list. At runtime, the manifest list will resolve to
    // the actual image based on the node's architecture.
    csi_image_version_mapping = {
      "default" = {
        "all" : chomp(<<-EOF
      {
        "default": "oke-multiarch-1.23-4c65264-146@sha256:defa64f16f0a16b84f3009cd6d902e4ce547dff68d3a5b85e510749d982de164",
        "v1.16": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
        "v1.17": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
        "v1.18": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
        "v1.19": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
        "v1.20": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
        "v1.21": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
        "v1.22": "oke-multiarch-1.22-d0bafe8-232@sha256:4697113594971e55df52d9e72dda7c381b431ec11d8d4622d82c7fafeb6c2689",
        "v1.23": "oke-multiarch-1.23-4c65264-146@sha256:defa64f16f0a16b84f3009cd6d902e4ce547dff68d3a5b85e510749d982de164",
        "v1.24": "v1.24-32be19ef595-4@sha256:3eda1610412ce5a3f6009b1d1a9219b3fdcc59009a8e3077a83f2b82142a586e",
        "v1.25": "v1.25-ed706a20a85-7@sha256:89418f404b65b858ef9f06fcb384b0166623066d589a68661367a3aec2e82ff9",
        "v1.26": "v1.26-46a75cc4065-7@sha256:bcacc2d28c1af3e09c336533089fd3551146ce3e2c1a24db81c8564dff3d28fe",
        "v1.27": "v1.27-621480e7d56-7@sha256:52357cfd18ed38843ae54438369095520a95fe6abcd18aac53b74000b1e222fb",
        "v1.28": "v1.28-94e5f680ebe-3@sha256:dae5ada4a06ecc101d87897712322a1d452f146a5373f32ba2a760b45146759b"
      }
      EOF
        )
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
        "all" : chomp(<<-EOF
      {
        "default": "oke-multiarch-1.23-4c65264-146@sha256:defa64f16f0a16b84f3009cd6d902e4ce547dff68d3a5b85e510749d982de164",
        "v1.16": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
        "v1.17": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
        "v1.18": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
        "v1.19": "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
        "v1.20": "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
        "v1.21": "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
        "v1.22": "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
        "v1.23": "oke-multiarch-1.23-526d1e6-171@sha256:85235e1fa24c41e5fb158346e3339fc680dcdce791735bfca25c7755a479e4c8",
        "v1.24": "v1.24-32be19ef595-4@sha256:3eda1610412ce5a3f6009b1d1a9219b3fdcc59009a8e3077a83f2b82142a586e",
        "v1.25": "v1.25-f6e6131af43-12@sha256:33df70cbac597f198002bff06a52bd614cd60dc60125cddbb554da3585a18ef3",
        "v1.26": "v1.26-73ac4e8121d-12@sha256:a1b8c9433c2b7be37f32f98b4027e1d48afc0a4eb8db659d04d5d69c1c2476ab",
        "v1.27": "v1.27-84a82e6e492-12@sha256:f262c59eea6063bc918443d0f35a4168772301f46e95ff1273d701272613024d",
        "v1.28": "v1.28-94e5f680ebe-3@sha256:dae5ada4a06ecc101d87897712322a1d452f146a5373f32ba2a760b45146759b"
      }
      EOF
        )
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

    csi-bv-expansion-enabled = {
      "default" = {
        "all" : "true"
      }
    }

    fss-csi-driver-enabled = {
      "default" = {
        "all" : "true"
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
