locals {
  tenancy_property_overrides = {
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
    }*/
    "oc1" = {
       "ccm-image-version-mapping" = {
          overrides = [
            {
              regions = ["lhr", "iad", "fra", "phx", "jed", "sjc", "syd", "nrt", "kix", "yyz"]
              env = "prd"
              value = "{\"default\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.16\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.17\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.18\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.19\": \"oke-multiarch-1.19-a7eaef8-235@sha256:8a76473d22732c9c8a8215d66cc0409e3f185d2e010695cf4f8160bceb4d4a53\",\"v1.20\": \"oke-multiarch-1.19-a7eaef8-235@sha256:8a76473d22732c9c8a8215d66cc0409e3f185d2e010695cf4f8160bceb4d4a53\",\"v1.21\": \"oke-multiarch-1.19-a7eaef8-235@sha256:8a76473d22732c9c8a8215d66cc0409e3f185d2e010695cf4f8160bceb4d4a53\",\"v1.22\": \"oke-multiarch-1.22-571aa0f-97@sha256:7f142f1ae32a2e942ed538d3a36f2f5734092ca7cf807a1c88d4dab5d0802c72\",\"v1.23\": \"oke-multiarch-1.23-94dc7a6-27@sha256:e5a5ebb0740a3d642c63b7590cb0d6becb4d3ce274555c3752a75beb2bda4710\",\"v1.24\": \"oke-multiarch-1.24-7d1247f-23@sha256:bc10657a08cd0eaf27288c6e8432d7764dff5483dd2eda3ac42a5907d51abc28\"}"
              tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaa7y7jlxtgzeufk25jj22kmcrc37wzdbtdi6w3gkesuyffgyasyh7q"
            }
          ]
      },
      "csi-image-version-mapping" = {
        overrides = [
          {
            regions = ["lhr", "iad", "fra", "phx", "jed", "sjc", "syd", "nrt", "kix", "yyz"]
            env = "prd"
            value = "{\"default\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\", \"v1.16\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\", \"v1.17\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\", \"v1.18\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\", \"v1.19\": \"oke-multiarch-1.19-a7eaef8-235@sha256:8a76473d22732c9c8a8215d66cc0409e3f185d2e010695cf4f8160bceb4d4a53\", \"v1.20\": \"oke-multiarch-1.19-a7eaef8-235@sha256:8a76473d22732c9c8a8215d66cc0409e3f185d2e010695cf4f8160bceb4d4a53\", \"v1.21\": \"oke-multiarch-1.19-a7eaef8-235@sha256:8a76473d22732c9c8a8215d66cc0409e3f185d2e010695cf4f8160bceb4d4a53\", \"v1.22\": \"oke-multiarch-1.22-ba49ae1-87@sha256:66cbb0e0c4dd3820fbef9d45d12644df280cc3889b0084cefa1a37649efbf7a6\", \"v1.23\": \"oke-multiarch-1.23-9edde43-21@sha256:ed871b6147b24ac3f6ae203996f16e659566f1908561b74647cc90efed9a7ff4\",\"v1.24\": \"oke-multiarch-1.24-7d1247f-23@sha256:bc10657a08cd0eaf27288c6e8432d7764dff5483dd2eda3ac42a5907d51abc28\"}"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaa7y7jlxtgzeufk25jj22kmcrc37wzdbtdi6w3gkesuyffgyasyh7q"
          }
        ]
      }
    }
  }
}
