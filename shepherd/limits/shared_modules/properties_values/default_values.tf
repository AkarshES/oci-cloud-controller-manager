locals {

  // Update the ccm image sha value here for updating CCM versions for respective k8s versions across all realms
  ccm_default_mapping = {
    "default" = {
      "all" : {
        "default" : "oke-multiarch-1.23-526d1e6-171@sha256:85235e1fa24c41e5fb158346e3339fc680dcdce791735bfca25c7755a479e4c8",
        "v1.23" : "oke-multiarch-1.23-526d1e6-171@sha256:85235e1fa24c41e5fb158346e3339fc680dcdce791735bfca25c7755a479e4c8",
        "v1.24" : "v1.24-32be19ef595-4@sha256:3eda1610412ce5a3f6009b1d1a9219b3fdcc59009a8e3077a83f2b82142a586e",
        "v1.25" : "v1.25-756ad2ecbcd-31@sha256:9653868283b1b285daa6773383e0a77e5068dd67a3c32cd36b0a8f10d92ffbda",
        "v1.26" : "v1.26-046c0c74dc9-35@sha256:672d5c011956e4621b563e658c589dc2ae368d6a04dc04cc804765d8195bf029",
        "v1.27" : "v1.27-88f1aaae3b3-33@sha256:7888a14ff06d0c8b4ebcdf1d27e33ff1da507e75e0692869b2f8bb918700ec12",
        "v1.28" : "v1.28-82f08e51cb0-27@sha256:0f1de79957e0cbfa1cee9aaeb902db10ffaa45e91805e3c279b6cf2ef176c488",
        "v1.29" : "v1.29-b4676b971bb-13@sha256:26c1e2c2d6b0f908bf11448bc13c5cb5c6625ea995ccaee59012cfa2872d9dc6"
      }
    }
  }

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
        "v1.25": "v1.25-e402eabbe0a-25@sha256:af01ae775ba15f965f797ccea9aa6dece9e98136f42fb23c56775d7047307394",
        "v1.26": "v1.26-d5c95b4f813-25@sha256:d401f00fc5f6d2710f916b44450c5f4264673f1403a4aabeba98393e318fcec1",
        "v1.27": "v1.27-ff66a5f1ae0-24@sha256:9c5992e61469ccf706979741bb32390b2986591464cc08c02958adad0572bd66",
        "v1.28": "v1.28-5f2e34f4a04-16@sha256:705d41ca2f58182e445471a4188d6bd48ddbe0fe55b1246e32af3ca168030c08",
        "v1.29": "v1.29-eaf7462813a-5@sha256:5608f1ea0b4dda66169d9d1076017dc04c437dec0d33c0fed2e1f25d96affb70"
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




    // CCM related mappings
    ccm_image_version_mapping = {
      "default" = {
        "all" : jsonencode(local.ccm_default_mapping.default.all)
      }

      "dev.oc1" = {
        "all" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.20" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.21" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
          }
        ))
        "iad" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.20" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.21" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
          }
        ))
        "phx" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.20" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.21" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
          }
        ))
        "fra" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.20" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.21" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
          }
        ))
      }

      "integ.oc1" = {
        "all" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.20" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.21" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
          }
        ))
        "iad" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.20" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.21" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
          }
        ))
        "phx" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.20" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.21" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
          }
        ))
        "fra" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.20" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.21" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
          }
        ))
      }
      // To override property values for an environment declare a key
      // by environment name. To override values for a qualified realm, declare
      // a key with qualified realm name

      // https://developer.hashicorp.com/terraform/language/functions/merge
      "prd.oc1" = {
        "all" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.20" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.21" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
          }
        ))
      }

      "prd.oc8" = {
        "all" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.20" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.21" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
          }
        ))
      }

      "prd.oc9" = {
        "all" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.20" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.21" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
          }
        ))
      }

      "prd.oc10" = {
        "all" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.20" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.21" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
          }
        ))
      }

      // US - GOV realms
      "prd.oc2" = {
        "all" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.20" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.21" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
          }
        ))
      }

      "prd.oc3" = {
        "all" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.20" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.21" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
          }
        ))
      }

      // UK - GOV realms
      "prd.oc4" = {
        "all" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-1.16-520cc1d-11@sha256:ba64f2c4ebf862d8e00d5c762e510bf41a7c9b735b5bb3a80a5ee33e344da7bf",
            "v1.16" : "oke-1.16-520cc1d-11@sha256:ba64f2c4ebf862d8e00d5c762e510bf41a7c9b735b5bb3a80a5ee33e344da7bf",
            "v1.17" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.20" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.21" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
          }
        ))
      }


      // ONSR realms
      "prd.oc5" = {
        "all" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-1.16-520cc1d-11@sha256:7900589b191fb6a77b77172c3800428e4c435b54605cc033c8ebea4c60ae5df1",
            "v1.16" : "oke-1.16-520cc1d-11@sha256:7900589b191fb6a77b77172c3800428e4c435b54605cc033c8ebea4c60ae5df1",
            "v1.17" : "oke-1.17-40e9a7a-13@sha256:424470ceb13e4b76f0a07e645bf2b37838f9a30709fdfdbce51b79748a6ff364",
            "v1.18" : "oke-1.17-40e9a7a-13@sha256:424470ceb13e4b76f0a07e645bf2b37838f9a30709fdfdbce51b79748a6ff364",
            "v1.19" : "oke-1.19-64ab664-255@sha256:fb4144e0480f120c54ea4688f70ef834b9a1fb0c03ffe0f03eca5afe61ae6765",
            "v1.20" : "oke-1.19-64ab664-255@sha256:fb4144e0480f120c54ea4688f70ef834b9a1fb0c03ffe0f03eca5afe61ae6765",
            "v1.21" : "oke-1.19-64ab664-255@sha256:fb4144e0480f120c54ea4688f70ef834b9a1fb0c03ffe0f03eca5afe61ae6765",
            "v1.22" : "oke-1.22-9893434-269@sha256:4d9775fe579b47b3e53c6a361c27114d55b97775d6040206f1b48d7f002684e6",
            "v1.23" : "oke-1.23-526d1e6-171@sha256:b738bf71baccc009e8eda5d818942a7d4c1be6d6dd63abe79c731aa2cafb036f",
          }
        ))
      }

      "prd.oc6" = {
        "all" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-1.16-520cc1d-11@sha256:7900589b191fb6a77b77172c3800428e4c435b54605cc033c8ebea4c60ae5df1",
            "v1.16" : "oke-1.16-520cc1d-11@sha256:7900589b191fb6a77b77172c3800428e4c435b54605cc033c8ebea4c60ae5df1",
            "v1.17" : "oke-1.17-40e9a7a-13@sha256:424470ceb13e4b76f0a07e645bf2b37838f9a30709fdfdbce51b79748a6ff364",
            "v1.18" : "oke-1.17-40e9a7a-13@sha256:424470ceb13e4b76f0a07e645bf2b37838f9a30709fdfdbce51b79748a6ff364",
            "v1.19" : "oke-1.19-64ab664-255@sha256:fb4144e0480f120c54ea4688f70ef834b9a1fb0c03ffe0f03eca5afe61ae6765",
            "v1.20" : "oke-1.19-64ab664-255@sha256:fb4144e0480f120c54ea4688f70ef834b9a1fb0c03ffe0f03eca5afe61ae6765",
            "v1.21" : "oke-1.19-64ab664-255@sha256:fb4144e0480f120c54ea4688f70ef834b9a1fb0c03ffe0f03eca5afe61ae6765",
            "v1.22" : "oke-1.22-9f30fe0-220@sha256:fdbb7926235ecaa790b4d775ca7befe57f0fd8cd559afd6ead9a6ff765de8312",
            "v1.23" : "oke-1.23-b8a8dee-140@sha256:922eb0404d72f9abc63218ac4275b5676519af7698d770362a8eb9e4f3ba65a8",
          }
        ))
      }

      "prd.oc11" = {
        "all" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-1.17-40e9a7a-13@sha256:424470ceb13e4b76f0a07e645bf2b37838f9a30709fdfdbce51b79748a6ff364",
            "v1.18" : "oke-1.17-40e9a7a-13@sha256:424470ceb13e4b76f0a07e645bf2b37838f9a30709fdfdbce51b79748a6ff364",
            "v1.19" : "oke-1.19-64ab664-255@sha256:fb4144e0480f120c54ea4688f70ef834b9a1fb0c03ffe0f03eca5afe61ae6765",
            "v1.20" : "oke-1.19-64ab664-255@sha256:fb4144e0480f120c54ea4688f70ef834b9a1fb0c03ffe0f03eca5afe61ae6765",
            "v1.21" : "oke-1.19-64ab664-255@sha256:fb4144e0480f120c54ea4688f70ef834b9a1fb0c03ffe0f03eca5afe61ae6765",
            "v1.22" : "oke-1.22-9893434-269@sha256:4d9775fe579b47b3e53c6a361c27114d55b97775d6040206f1b48d7f002684e6",
            "v1.23" : "oke-1.23-526d1e6-171@sha256:b738bf71baccc009e8eda5d818942a7d4c1be6d6dd63abe79c731aa2cafb036f",
          }
        ))
      }

      "prd.oc14" = {
        "all" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
          }
        ))
      }

      "prd.oc16" = {
        "all" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
          }
        ))
      }

      "prd.oc17" = {
        "all" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
          }
        ))
      }

      "prd.oc19" = {
        "all" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
          }
        ))
      }

      "prd.oc20" = {
        "all" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
          }
        ))
      }

      "prd.oc22" = {
        "all" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
          }
        ))
      }

      "prd.oc24" = {
        "all" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "default" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
          }
        ))
      }

    }
  }
  global_default_values_by_property = {
  for property_name, property_value in local.global_default_values : property_name => merge(
    lookup(property_value, "default", {}),
    lookup(property_value, var.env, {}),
    lookup(property_value, "${var.env}.${var.realm}", {})
  )
  }

  global_default_values_list = flatten([
  for property_name, property_value in local.global_default_values_by_property : [
    {
      ad     = lookup(property_value, "ad", "all")
      group  = var.spectre_group_name
      name   = replace(property_name, "_", "-")
      region = var.execution_target.additional_locals.limits_region
      value  = lookup(property_value, var.execution_target.additional_locals.limits_region, lookup(property_value, "all", ""))
      min    = lookup(property_value, "min", null)
      max    = lookup(property_value, "max", null)
    }
  ] if length(property_value) > 0
  ])

  global_default_values_map = {
  for property in local.global_default_values_list : "${property.group}/${property.name}/${property.region}/${property.ad}" => property
  }
}
