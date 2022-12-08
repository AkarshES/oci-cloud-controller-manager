locals {
  tenancy_property_overrides = {
    "oc1" = {
      "ccm-image-version-mapping" = {
        overrides = [
          {
            regions      = ["lhr", "mel"]
            env          = "prd"
            value        = "{\"default\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.16\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.17\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.18\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.19\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.20\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.21\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.22\": \"oke-multiarch-1.22-aa9948e-192@sha256:5bf9fef2c99f2c77fdb8027d8690900c99adef873bde9bcf6cd30db39ce467eb\",\"v1.23\": \"oke-multiarch-1.23-16022dc-95@sha256:2782b14bc755aa839ca68a86591007cf4dfb2c8419be14a5eae94b24a77f1031\",\"v1.24\": \"oke-multiarch-1.24-9448c95-98@sha256:c09599b01eb0127fba552e4c9ade8e5e81ba33f115aac293e42de45929e2fce9\"}"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaatxzd3axgkv7gybg2dtii3ecaetdg42wwx3x723bi6j55dgi3a7uq"
          },
          {
            regions      = ["lhr", "mel"]
            env          = "prd"
            value        = "{\"default\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.16\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.17\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.18\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.19\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.20\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.21\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.22\": \"oke-multiarch-1.22-aa9948e-192@sha256:5bf9fef2c99f2c77fdb8027d8690900c99adef873bde9bcf6cd30db39ce467eb\",\"v1.23\": \"oke-multiarch-1.23-16022dc-95@sha256:2782b14bc755aa839ca68a86591007cf4dfb2c8419be14a5eae94b24a77f1031\",\"v1.24\": \"oke-multiarch-1.24-9448c95-98@sha256:c09599b01eb0127fba552e4c9ade8e5e81ba33f115aac293e42de45929e2fce9\"}"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaznz6dvb7goamdqoenmhlpngfqjalldgdox6kvs5w3pnkptpiehyq"
          },
          {
            regions      = ["lhr", "mel"]
            env          = "prd"
            value        = "{\"default\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.16\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.17\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.18\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.19\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.20\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.21\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.22\": \"oke-multiarch-1.22-aa9948e-192@sha256:5bf9fef2c99f2c77fdb8027d8690900c99adef873bde9bcf6cd30db39ce467eb\",\"v1.23\": \"oke-multiarch-1.23-16022dc-95@sha256:2782b14bc755aa839ca68a86591007cf4dfb2c8419be14a5eae94b24a77f1031\",\"v1.24\": \"oke-multiarch-1.24-9448c95-98@sha256:c09599b01eb0127fba552e4c9ade8e5e81ba33f115aac293e42de45929e2fce9\"}"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaa5gtuu5ao3rl7tdj4fqe445ou3uvyn64c3muvnkigi2a72drw3rya"
          }
        ]
      },
      "csi-image-version-mapping" = {
        overrides = [
          {
            regions      = ["lhr", "mel"]
            env          = "prd"
            value        = "{\"default\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.16\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.17\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.18\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.19\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.20\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.21\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.22\": \"oke-multiarch-1.22-aa9948e-192@sha256:5bf9fef2c99f2c77fdb8027d8690900c99adef873bde9bcf6cd30db39ce467eb\",\"v1.23\": \"oke-multiarch-1.23-16022dc-95@sha256:2782b14bc755aa839ca68a86591007cf4dfb2c8419be14a5eae94b24a77f1031\",\"v1.24\": \"oke-multiarch-1.24-9448c95-98@sha256:c09599b01eb0127fba552e4c9ade8e5e81ba33f115aac293e42de45929e2fce9\"}"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaatxzd3axgkv7gybg2dtii3ecaetdg42wwx3x723bi6j55dgi3a7uq"
          },
          {
            regions      = ["lhr", "mel"]
            env          = "prd"
            value        = "{\"default\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.16\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.17\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.18\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.19\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.20\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.21\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.22\": \"oke-multiarch-1.22-aa9948e-192@sha256:5bf9fef2c99f2c77fdb8027d8690900c99adef873bde9bcf6cd30db39ce467eb\",\"v1.23\": \"oke-multiarch-1.23-16022dc-95@sha256:2782b14bc755aa839ca68a86591007cf4dfb2c8419be14a5eae94b24a77f1031\",\"v1.24\": \"oke-multiarch-1.24-9448c95-98@sha256:c09599b01eb0127fba552e4c9ade8e5e81ba33f115aac293e42de45929e2fce9\"}"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaznz6dvb7goamdqoenmhlpngfqjalldgdox6kvs5w3pnkptpiehyq"
          },
          {
            regions      = ["lhr", "mel"]
            env          = "prd"
            value        = "{\"default\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.16\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.17\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.18\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.19\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.20\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.21\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.22\": \"oke-multiarch-1.22-aa9948e-192@sha256:5bf9fef2c99f2c77fdb8027d8690900c99adef873bde9bcf6cd30db39ce467eb\",\"v1.23\": \"oke-multiarch-1.23-16022dc-95@sha256:2782b14bc755aa839ca68a86591007cf4dfb2c8419be14a5eae94b24a77f1031\",\"v1.24\": \"oke-multiarch-1.24-9448c95-98@sha256:c09599b01eb0127fba552e4c9ade8e5e81ba33f115aac293e42de45929e2fce9\"}"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaa5gtuu5ao3rl7tdj4fqe445ou3uvyn64c3muvnkigi2a72drw3rya"
          }
        ]
      }
    }
  }
}
#Usage:
/* "oc1" = {
     "property-name" = {
        overrides = [
           {
             regions = ["region"]
             env = "environment"
             value = "value"
             tenancy_ocid = "ocid1.tenancy.ocx..aaaaaaaaaaaaaaaaaaaaaaa"
           }
        ]
     }
  }
*/
