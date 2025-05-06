locals {
  // Update the pop version corresponding to the pop build for app release
  pop_version = "462627232d7_72"

  // Update the ccm image sha value here for updating CCM versions for respective k8s versions across all realms
  ccm_default_mapping = {
    "default" = {
      "all" : {
        "default" : "oke-multiarch-1.23-526d1e6-171@sha256:85235e1fa24c41e5fb158346e3339fc680dcdce791735bfca25c7755a479e4c8",
        "v1.23" : "oke-multiarch-1.23-526d1e6-171@sha256:85235e1fa24c41e5fb158346e3339fc680dcdce791735bfca25c7755a479e4c8",
        "v1.24" : "v1.24-32be19ef595-4@sha256:3eda1610412ce5a3f6009b1d1a9219b3fdcc59009a8e3077a83f2b82142a586e",
        "v1.25" : "v1.25-d0b59914251-38@sha256:ffa44cb1e6bc5793859a1ebf7762bcd1731ec0dea426034c44bc77ffb12ece48",
        "v1.26" : "v1.26-8744a6c9ccd-42@sha256:bc20c825c5e3b5f40b56467e3b931597a6edef41cd0ab0cb20524b4cd8e603a0",
        "v1.27" : "v1.27-7e9bf7a9189-52@sha256:fd510337e52ef609ffb86953ef080b3f2fa1a19a0647f65dfbd2c1dfbda0df7a",
        "v1.28" : "v1.28-cb1635cc6c7-80@sha256:d9e51c8c78b3ef040e5739d6ac079d6587bb4b03d8b65579a877b4ec84c4f219",
        "v1.29" : "v1.29-e259bc56402-106@sha256:8f0e0e655672311396a631d19688fbaf4cf473ce991b72791586dce8ced607b8",
        "v1.30" : "v1.30-c1f46fdfd72-96@sha256:96c29df6e94d0dbcb2956d99f08de612f866eff2cca6263a33ab90470f633afa",
        "v1.31" : "v1.31-5712723a30a-54@sha256:78bec501d421ae30f133b754956b60da57493f1f2adcdd79ea18d9da15d84b3b",
        "v1.32" : "v1.32-78e6d79de3c-22@sha256:f832a34c0c8c887e564d405243d69284ec777f94144a0b26313b4e8460bf3e35",
        "v1.33" : "v1.32-78e6d79de3c-22@sha256:f832a34c0c8c887e564d405243d69284ec777f94144a0b26313b4e8460bf3e35"
      }
    }
  }

  csi_default_mapping = {
    "default" = {
      "all" : {
        "default": "oke-multiarch-1.23-4c65264-146@sha256:defa64f16f0a16b84f3009cd6d902e4ce547dff68d3a5b85e510749d982de164",
        "v1.23" : "oke-multiarch-1.23-4c65264-146@sha256:defa64f16f0a16b84f3009cd6d902e4ce547dff68d3a5b85e510749d982de164",
        "v1.24" : "v1.24-32be19ef595-4@sha256:3eda1610412ce5a3f6009b1d1a9219b3fdcc59009a8e3077a83f2b82142a586e",
        "v1.25" : "v1.25-e402eabbe0a-25@sha256:af01ae775ba15f965f797ccea9aa6dece9e98136f42fb23c56775d7047307394",
        "v1.26" : "v1.26-d5c95b4f813-25@sha256:d401f00fc5f6d2710f916b44450c5f4264673f1403a4aabeba98393e318fcec1",
        "v1.27" : "v1.27-bff7091ad1d-51@sha256:10555f13db26cbcb35dcd684be88db9897e01ad25bdcd7d65e9cb4bd1c5f386f",
        "v1.28" : "v1.28-cb1635cc6c7-80@sha256:d9e51c8c78b3ef040e5739d6ac079d6587bb4b03d8b65579a877b4ec84c4f219",
        "v1.29" : "v1.29-0911461af79-92@sha256:ed06eeb4d53c3af81e3caf662c7057922dc1a39400b3a38ef281db0b3707113f",
        "v1.30" : "v1.30-a67f7b269a7-85@sha256:f0224f684ef8e32ea6435e51cb72dba3a6bd140f6ef8d7190cc1f6d7c669c1b0",
        "v1.31" : "v1.31-c7e5bd92e29-43@sha256:eff28f23a6c82e1d1539e5f3c4b65efe8fdf8ea615c6f9158592c630fc9f09d1",
        "v1.32" : "v1.32-c01d1d4113e-11@sha256:a797ab6fd015af36ce0a5570ccd741860672bccdd879b167fa16b1212ab77bc1",
        "v1.33" : "v1.32-c01d1d4113e-11@sha256:a797ab6fd015af36ce0a5570ccd741860672bccdd879b167fa16b1212ab77bc1"

      }
    }
  }

  // Test any upstream version
  ccm_upstream_k8s_new_version_test = {
    "default" = {
      "all" : {
      }
    }
  }

  csi_upstream_k8s_new_version_test = {
    "default" = {
      "all" : {
      }
    }
  }

  global_default_values = {
    // The "default" mapping uses a SHA that references a manifest list. At runtime, the manifest list will resolve to
    // the actual image based on the node's architecture.


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

    oci-service-controller-enabled = {
      "default" = {
        "all": "true"
      }
    }

    cpo-enable-resource-attribution = {
      "default" = {
        "all": "false"
      }
      "dev.oc1" = {
        "all": "true"
      }
      "integ.oc1" = {
        "all": "true"
      }
      "prd.oc1" = {
        "all": "true"
      }
      "prd.oc16" = {
        "all": "true"
      }
    }

    lustre-csi-driver-enabled = {
      "default" = {
        "all": "true"
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
          },
          local.ccm_upstream_k8s_new_version_test.default.all
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
            "v1.29" : "v1.29-0911461af79-92@sha256:ed06eeb4d53c3af81e3caf662c7057922dc1a39400b3a38ef281db0b3707113f",
            "v1.30" : "v1.30-a67f7b269a7-85@sha256:f0224f684ef8e32ea6435e51cb72dba3a6bd140f6ef8d7190cc1f6d7c669c1b0",
            "v1.31" : "v1.31-c7e5bd92e29-43@sha256:eff28f23a6c82e1d1539e5f3c4b65efe8fdf8ea615c6f9158592c630fc9f09d1",
            "v1.32" : "v1.32-c01d1d4113e-11@sha256:a797ab6fd015af36ce0a5570ccd741860672bccdd879b167fa16b1212ab77bc1",
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
            "v1.29" : "v1.29-0911461af79-92@sha256:ed06eeb4d53c3af81e3caf662c7057922dc1a39400b3a38ef281db0b3707113f",
            "v1.30" : "v1.30-a67f7b269a7-85@sha256:f0224f684ef8e32ea6435e51cb72dba3a6bd140f6ef8d7190cc1f6d7c669c1b0",
            "v1.31" : "v1.31-c7e5bd92e29-43@sha256:eff28f23a6c82e1d1539e5f3c4b65efe8fdf8ea615c6f9158592c630fc9f09d1",
            "v1.32" : "v1.32-c01d1d4113e-11@sha256:a797ab6fd015af36ce0a5570ccd741860672bccdd879b167fa16b1212ab77bc1",
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
            "v1.25" : "v1.25-756ad2ecbcd-31@sha256:9653868283b1b285daa6773383e0a77e5068dd67a3c32cd36b0a8f10d92ffbda",
            "v1.26" : "v1.26-046c0c74dc9-35@sha256:672d5c011956e4621b563e658c589dc2ae368d6a04dc04cc804765d8195bf029",
            "v1.29" : "v1.29-0911461af79-92@sha256:ed06eeb4d53c3af81e3caf662c7057922dc1a39400b3a38ef281db0b3707113f",
            "v1.30" : "v1.30-a67f7b269a7-85@sha256:f0224f684ef8e32ea6435e51cb72dba3a6bd140f6ef8d7190cc1f6d7c669c1b0",
            "v1.31" : "v1.31-c7e5bd92e29-43@sha256:eff28f23a6c82e1d1539e5f3c4b65efe8fdf8ea615c6f9158592c630fc9f09d1",
            "v1.32" : "v1.32-c01d1d4113e-11@sha256:a797ab6fd015af36ce0a5570ccd741860672bccdd879b167fa16b1212ab77bc1",
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
            "v1.25" : "v1.25-756ad2ecbcd-31@sha256:9653868283b1b285daa6773383e0a77e5068dd67a3c32cd36b0a8f10d92ffbda",
            "v1.26" : "v1.26-046c0c74dc9-35@sha256:672d5c011956e4621b563e658c589dc2ae368d6a04dc04cc804765d8195bf029",
            "v1.29" : "v1.29-0911461af79-92@sha256:ed06eeb4d53c3af81e3caf662c7057922dc1a39400b3a38ef281db0b3707113f",
            "v1.30" : "v1.30-a67f7b269a7-85@sha256:f0224f684ef8e32ea6435e51cb72dba3a6bd140f6ef8d7190cc1f6d7c669c1b0",
            "v1.31" : "v1.31-c7e5bd92e29-43@sha256:eff28f23a6c82e1d1539e5f3c4b65efe8fdf8ea615c6f9158592c630fc9f09d1",
            "v1.32" : "v1.32-c01d1d4113e-11@sha256:a797ab6fd015af36ce0a5570ccd741860672bccdd879b167fa16b1212ab77bc1",
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
            "v1.25" : "v1.25-756ad2ecbcd-31@sha256:9653868283b1b285daa6773383e0a77e5068dd67a3c32cd36b0a8f10d92ffbda",
            "v1.26" : "v1.26-046c0c74dc9-35@sha256:672d5c011956e4621b563e658c589dc2ae368d6a04dc04cc804765d8195bf029",
            "v1.29" : "v1.29-13ef016091f-80@sha256:2c7bf660d548975c3c7cf079153eec6e716a9399e6eb01cde5f56102125ef40b",
            "v1.30" : "v1.30-424c53e6898-73@sha256:0fde8115356b9654282bf361c7ae55aa9b95c510c34775bec501264bca999648",
            "v1.31" : "v1.31-46a28c37e3d-32@sha256:64472a97837f58f3f33489077923b0babf8802453f16b195b260b559fbd52b6d",
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

      "prd.oc23" = {
        "all" : jsonencode(merge(local.ccm_default_mapping.default.all,
          {
            "v1.29" : "v1.29-0911461af79-92@sha256:ed06eeb4d53c3af81e3caf662c7057922dc1a39400b3a38ef281db0b3707113f",
            "v1.30" : "v1.30-a67f7b269a7-85@sha256:f0224f684ef8e32ea6435e51cb72dba3a6bd140f6ef8d7190cc1f6d7c669c1b0",
            "v1.31" : "v1.31-c7e5bd92e29-43@sha256:eff28f23a6c82e1d1539e5f3c4b65efe8fdf8ea615c6f9158592c630fc9f09d1",
            "v1.32" : "v1.32-c01d1d4113e-11@sha256:a797ab6fd015af36ce0a5570ccd741860672bccdd879b167fa16b1212ab77bc1",
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

    // CSI Image Mappings
    csi_image_version_mapping = {
      "default" = {
        "all" : jsonencode(local.csi_default_mapping.default.all)
      }
      // To override property values for an environment declare a key
      // by environment name. To override values for a qualified realm, declare
      // a key with qualified realm name
      "dev.oc1" = {
        "all" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.20": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.21": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.22": "oke-multiarch-1.22-d0bafe8-232@sha256:4697113594971e55df52d9e72dda7c381b431ec11d8d4622d82c7fafeb6c2689"
          }
        ))
        "iad" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.20": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.21": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.22": "oke-multiarch-1.22-d0bafe8-232@sha256:4697113594971e55df52d9e72dda7c381b431ec11d8d4622d82c7fafeb6c2689"
          }
        ))
        "phx" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.20": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.21": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.22": "oke-multiarch-1.22-d0bafe8-232@sha256:4697113594971e55df52d9e72dda7c381b431ec11d8d4622d82c7fafeb6c2689"
          }
        ))
        "fra" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.20": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.21": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.22": "oke-multiarch-1.22-d0bafe8-232@sha256:4697113594971e55df52d9e72dda7c381b431ec11d8d4622d82c7fafeb6c2689"
          }
        ))
      }

      "integ.oc1" = {
        "all" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.20": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.21": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.22": "oke-multiarch-1.22-d0bafe8-232@sha256:4697113594971e55df52d9e72dda7c381b431ec11d8d4622d82c7fafeb6c2689"

          }
        ))
        "iad" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.20": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.21": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.22": "oke-multiarch-1.22-d0bafe8-232@sha256:4697113594971e55df52d9e72dda7c381b431ec11d8d4622d82c7fafeb6c2689"
          }
        ))
        "phx" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.20": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.21": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.22": "oke-multiarch-1.22-d0bafe8-232@sha256:4697113594971e55df52d9e72dda7c381b431ec11d8d4622d82c7fafeb6c2689"
          },
          local.csi_upstream_k8s_new_version_test.default.all
        ))
        "fra" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.20": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.21": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.22": "oke-multiarch-1.22-d0bafe8-232@sha256:4697113594971e55df52d9e72dda7c381b431ec11d8d4622d82c7fafeb6c2689"
          }
        ))
      }

      "prd.oc1" = {
        "all" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.20": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.21": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.22": "oke-multiarch-1.22-d0bafe8-232@sha256:4697113594971e55df52d9e72dda7c381b431ec11d8d4622d82c7fafeb6c2689"
          }
        ))
      }

      "prd.oc8" = {
        "all" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.20": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.21": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.22": "oke-multiarch-1.22-aa9948e-192@sha256:5bf9fef2c99f2c77fdb8027d8690900c99adef873bde9bcf6cd30db39ce467eb",
            "v1.23": "oke-multiarch-1.23-16022dc-95@sha256:2782b14bc755aa839ca68a86591007cf4dfb2c8419be14a5eae94b24a77f1031"
          }
        ))
      }

      "prd.oc9" = {
        "all" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.20": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.21": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.22": "oke-multiarch-1.22-aa9948e-192@sha256:5bf9fef2c99f2c77fdb8027d8690900c99adef873bde9bcf6cd30db39ce467eb",
            "v1.23": "oke-multiarch-1.23-16022dc-95@sha256:2782b14bc755aa839ca68a86591007cf4dfb2c8419be14a5eae94b24a77f1031"
          }
        ))
      }

      "prd.oc10" = {
        "all" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.20": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.21": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.22": "oke-multiarch-1.22-aa9948e-192@sha256:5bf9fef2c99f2c77fdb8027d8690900c99adef873bde9bcf6cd30db39ce467eb",
            "v1.23": "oke-multiarch-1.23-16022dc-95@sha256:2782b14bc755aa839ca68a86591007cf4dfb2c8419be14a5eae94b24a77f1031"
          }
        ))
      }

      // US-GOV Realms
      "prd.oc2" = {
        "all" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19": "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.20": "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.21": "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22": "oke-multiarch-1.22-aa9948e-192@sha256:5bf9fef2c99f2c77fdb8027d8690900c99adef873bde9bcf6cd30db39ce467eb",
            "v1.23": "oke-multiarch-1.23-16022dc-95@sha256:2782b14bc755aa839ca68a86591007cf4dfb2c8419be14a5eae94b24a77f1031",
          }
        ))
      }

      "prd.oc3" = {
        "all" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19": "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.20": "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.21": "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22": "oke-multiarch-1.22-aa9948e-192@sha256:5bf9fef2c99f2c77fdb8027d8690900c99adef873bde9bcf6cd30db39ce467eb",
            "v1.23": "oke-multiarch-1.23-16022dc-95@sha256:2782b14bc755aa839ca68a86591007cf4dfb2c8419be14a5eae94b24a77f1031",
          }
        ))
      }

      // UK - GOV Realms
      "prd.oc4" = {
        "all" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-1.16-520cc1d-11@sha256:ba64f2c4ebf862d8e00d5c762e510bf41a7c9b735b5bb3a80a5ee33e344da7bf",
            "v1.16": "oke-1.16-520cc1d-11@sha256:ba64f2c4ebf862d8e00d5c762e510bf41a7c9b735b5bb3a80a5ee33e344da7bf",
            "v1.17": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.20": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.21": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.22": "oke-multiarch-1.22-aa9948e-192@sha256:5bf9fef2c99f2c77fdb8027d8690900c99adef873bde9bcf6cd30db39ce467eb",
            "v1.23": "oke-multiarch-1.23-16022dc-95@sha256:2782b14bc755aa839ca68a86591007cf4dfb2c8419be14a5eae94b24a77f1031",
          }
        ))
      }

      // ONSR Realms
      "prd.oc5" = {
        "all" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-1.16-520cc1d-11@sha256:7900589b191fb6a77b77172c3800428e4c435b54605cc033c8ebea4c60ae5df1",
            "v1.16": "oke-1.16-520cc1d-11@sha256:7900589b191fb6a77b77172c3800428e4c435b54605cc033c8ebea4c60ae5df1",
            "v1.17": "oke-1.17-40e9a7a-13@sha256:424470ceb13e4b76f0a07e645bf2b37838f9a30709fdfdbce51b79748a6ff364",
            "v1.18": "oke-1.17-40e9a7a-13@sha256:424470ceb13e4b76f0a07e645bf2b37838f9a30709fdfdbce51b79748a6ff364",
            "v1.19": "oke-1.19-64ab664-255@sha256:fb4144e0480f120c54ea4688f70ef834b9a1fb0c03ffe0f03eca5afe61ae6765",
            "v1.20": "oke-1.19-64ab664-255@sha256:fb4144e0480f120c54ea4688f70ef834b9a1fb0c03ffe0f03eca5afe61ae6765",
            "v1.21": "oke-1.19-64ab664-255@sha256:fb4144e0480f120c54ea4688f70ef834b9a1fb0c03ffe0f03eca5afe61ae6765",
            "v1.22": "oke-1.22-aa9948e-192@sha256:6407276d0df64c6b29e9beafe84bfe55191cb3ef153424a41d969bc1c0b0de52",
            "v1.23": "oke-1.23-16022dc-95@sha256:51de94592812ef652c880cd7ebfd0ddf1187d79e5fecf263253dfe72a72e8f38",
          }
        ))
      }

      "prd.oc6" = {
        "all" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-1.16-520cc1d-11@sha256:7900589b191fb6a77b77172c3800428e4c435b54605cc033c8ebea4c60ae5df1",
            "v1.16": "oke-1.16-520cc1d-11@sha256:7900589b191fb6a77b77172c3800428e4c435b54605cc033c8ebea4c60ae5df1",
            "v1.17": "oke-1.17-40e9a7a-13@sha256:424470ceb13e4b76f0a07e645bf2b37838f9a30709fdfdbce51b79748a6ff364",
            "v1.18": "oke-1.17-40e9a7a-13@sha256:424470ceb13e4b76f0a07e645bf2b37838f9a30709fdfdbce51b79748a6ff364",
            "v1.19": "oke-1.19-64ab664-255@sha256:fb4144e0480f120c54ea4688f70ef834b9a1fb0c03ffe0f03eca5afe61ae6765",
            "v1.20": "oke-1.19-64ab664-255@sha256:fb4144e0480f120c54ea4688f70ef834b9a1fb0c03ffe0f03eca5afe61ae6765",
            "v1.21": "oke-1.19-64ab664-255@sha256:fb4144e0480f120c54ea4688f70ef834b9a1fb0c03ffe0f03eca5afe61ae6765",
            "v1.22": "oke-1.22-9f30fe0-220@sha256:fdbb7926235ecaa790b4d775ca7befe57f0fd8cd559afd6ead9a6ff765de8312",
            "v1.23": "oke-1.23-b8a8dee-140@sha256:922eb0404d72f9abc63218ac4275b5676519af7698d770362a8eb9e4f3ba65a8",
          }
        ))
      }

      "prd.oc11" = {
        "all" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-1.17-40e9a7a-13@sha256:424470ceb13e4b76f0a07e645bf2b37838f9a30709fdfdbce51b79748a6ff364",
            "v1.18": "oke-1.17-40e9a7a-13@sha256:424470ceb13e4b76f0a07e645bf2b37838f9a30709fdfdbce51b79748a6ff364",
            "v1.19": "oke-1.19-64ab664-255@sha256:fb4144e0480f120c54ea4688f70ef834b9a1fb0c03ffe0f03eca5afe61ae6765",
            "v1.20": "oke-1.19-64ab664-255@sha256:fb4144e0480f120c54ea4688f70ef834b9a1fb0c03ffe0f03eca5afe61ae6765",
            "v1.21": "oke-1.19-64ab664-255@sha256:fb4144e0480f120c54ea4688f70ef834b9a1fb0c03ffe0f03eca5afe61ae6765",
            "v1.22": "oke-1.22-aa9948e-192@sha256:6407276d0df64c6b29e9beafe84bfe55191cb3ef153424a41d969bc1c0b0de52",
            "v1.23": "oke-1.23-16022dc-95@sha256:51de94592812ef652c880cd7ebfd0ddf1187d79e5fecf263253dfe72a72e8f38",
          }
        ))
      }

      "prd.oc14" = {
        "all" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22": "oke-multiarch-1.22-aa9948e-192@sha256:5bf9fef2c99f2c77fdb8027d8690900c99adef873bde9bcf6cd30db39ce467eb",
            "v1.23": "oke-multiarch-1.23-16022dc-95@sha256:2782b14bc755aa839ca68a86591007cf4dfb2c8419be14a5eae94b24a77f1031",
          }
        ))
      }

      "prd.oc16" = {
        "all" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22": "oke-multiarch-1.22-aa9948e-192@sha256:5bf9fef2c99f2c77fdb8027d8690900c99adef873bde9bcf6cd30db39ce467eb",
            "v1.23": "oke-multiarch-1.23-16022dc-95@sha256:2782b14bc755aa839ca68a86591007cf4dfb2c8419be14a5eae94b24a77f1031",
          }
        ))
      }

      "prd.oc17" = {
        "all" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22": "oke-multiarch-1.22-aa9948e-192@sha256:5bf9fef2c99f2c77fdb8027d8690900c99adef873bde9bcf6cd30db39ce467eb",
            "v1.23": "oke-multiarch-1.23-16022dc-95@sha256:2782b14bc755aa839ca68a86591007cf4dfb2c8419be14a5eae94b24a77f1031",
          }
        ))
      }

      "prd.oc19" = {
        "all" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22": "oke-multiarch-1.22-aa9948e-192@sha256:5bf9fef2c99f2c77fdb8027d8690900c99adef873bde9bcf6cd30db39ce467eb",
            "v1.23": "oke-multiarch-1.23-16022dc-95@sha256:2782b14bc755aa839ca68a86591007cf4dfb2c8419be14a5eae94b24a77f1031",
          }
        ))
      }

      "prd.oc20" = {
        "all" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
            "v1.22": "oke-multiarch-1.22-aa9948e-192@sha256:5bf9fef2c99f2c77fdb8027d8690900c99adef873bde9bcf6cd30db39ce467eb",
            "v1.23": "oke-multiarch-1.23-16022dc-95@sha256:2782b14bc755aa839ca68a86591007cf4dfb2c8419be14a5eae94b24a77f1031",
          }
        ))
      }

      "prd.oc22" = {
        "all" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.20": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.21": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.22": "oke-multiarch-1.22-d0bafe8-232@sha256:4697113594971e55df52d9e72dda7c381b431ec11d8d4622d82c7fafeb6c2689",
          }
        ))
      }

      "prd.oc24" = {
        "all" : jsonencode(merge(local.csi_default_mapping.default.all,
          {
            "default": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.16": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
            "v1.17": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.18": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
            "v1.19": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.20": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.21": "oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa",
            "v1.22": "oke-multiarch-1.22-d0bafe8-232@sha256:4697113594971e55df52d9e72dda7c381b431ec11d8d4622d82c7fafeb6c2689",
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

output "pop_version" {
  value = local.pop_version
}
