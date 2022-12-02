locals {
  tenancy_property_overrides = {
    "oc1" = {
      "ccm-image-version-mapping" = {
        overrides = [
          {
            regions      = ["iad"]
            env          = "prd"
            value        = "{\"default\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.16\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.17\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.18\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.19\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.20\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.21\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.22\": \"oke-multiarch-1.22-1762caa-120@sha256:b9f3cba2c9b3786885ec5601770ade5be94438d4fb89b29f87fba77601d7ba01\",\"v1.23\": \"oke-multiarch-1.23-24e21fc-44@sha256:3f6dd1e2a539c769acffea46b19bc875c5a2f267be6f1652915adc1c4c284e93\",\"v1.24\": \"oke-multiarch-1.24-bd118e8-72@sha256:ec635e9eb8fa9960e92cd2032bd6b82332f21aebdf3140a3e660e976a773b92c\"}"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaanrbjdkpfvz2w2ve67ecxy24pa6k3teofpvlbviyl2r5rsbwseqvq"
          },
          {
            regions      = ["phx"]
            env          = "prd"
            value        = "{\"default\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.16\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.17\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.18\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.19\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.20\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.21\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.22\": \"oke-multiarch-1.22-1762caa-120@sha256:b9f3cba2c9b3786885ec5601770ade5be94438d4fb89b29f87fba77601d7ba01\",\"v1.23\": \"oke-multiarch-1.23-24e21fc-44@sha256:3f6dd1e2a539c769acffea46b19bc875c5a2f267be6f1652915adc1c4c284e93\",\"v1.24\": \"oke-multiarch-1.24-bd118e8-72@sha256:ec635e9eb8fa9960e92cd2032bd6b82332f21aebdf3140a3e660e976a773b92c\"}"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaa7wzb7yxrfd3vyhwbkgt5yzgbp6nqmf2xegio5ssi6lvxfwps4w3a"
          },
          {
            regions      = ["iad"]
            env          = "prd"
            value        = "{\"default\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.16\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.17\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.18\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.19\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.20\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.21\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.22\": \"oke-multiarch-1.22-1762caa-120@sha256:b9f3cba2c9b3786885ec5601770ade5be94438d4fb89b29f87fba77601d7ba01\",\"v1.23\": \"oke-multiarch-1.23-24e21fc-44@sha256:3f6dd1e2a539c769acffea46b19bc875c5a2f267be6f1652915adc1c4c284e93\",\"v1.24\": \"oke-multiarch-1.24-bd118e8-72@sha256:ec635e9eb8fa9960e92cd2032bd6b82332f21aebdf3140a3e660e976a773b92c\"}"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaahp7giiq4smoqb5hmqnz5xmplfwuyfpovxivy22qczlrlgkzns7qq"
          },
          {
            regions      = ["iad"]
            env          = "prd"
            value        = "{\"default\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.16\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.17\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.18\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.19\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.20\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.21\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.22\": \"oke-multiarch-1.22-1762caa-120@sha256:b9f3cba2c9b3786885ec5601770ade5be94438d4fb89b29f87fba77601d7ba01\",\"v1.23\": \"oke-multiarch-1.23-24e21fc-44@sha256:3f6dd1e2a539c769acffea46b19bc875c5a2f267be6f1652915adc1c4c284e93\",\"v1.24\": \"oke-multiarch-1.24-bd118e8-72@sha256:ec635e9eb8fa9960e92cd2032bd6b82332f21aebdf3140a3e660e976a773b92c\"}"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaahm2iczdm7ekujxrwgeicw6k2fedhsl742gauuvpxxbz6uvhlcekq"
          },
          {
            regions      = ["iad"]
            env          = "prd"
            value        = "{\"default\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.16\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.17\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.18\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.19\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.20\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.21\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.22\": \"oke-multiarch-1.22-1762caa-120@sha256:b9f3cba2c9b3786885ec5601770ade5be94438d4fb89b29f87fba77601d7ba01\",\"v1.23\": \"oke-multiarch-1.23-24e21fc-44@sha256:3f6dd1e2a539c769acffea46b19bc875c5a2f267be6f1652915adc1c4c284e93\",\"v1.24\": \"oke-multiarch-1.24-bd118e8-72@sha256:ec635e9eb8fa9960e92cd2032bd6b82332f21aebdf3140a3e660e976a773b92c\"}"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaapagptfoo7j4ccd3j3vv3fbz3kwiz5jxcgmemxvt2fikluz6w3gqa"
          },
          {
            regions      = ["iad"]
            env          = "prd"
            value        = "{\"default\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.16\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.17\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.18\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.19\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.20\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.21\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.22\": \"oke-multiarch-1.22-1762caa-120@sha256:b9f3cba2c9b3786885ec5601770ade5be94438d4fb89b29f87fba77601d7ba01\",\"v1.23\": \"oke-multiarch-1.23-24e21fc-44@sha256:3f6dd1e2a539c769acffea46b19bc875c5a2f267be6f1652915adc1c4c284e93\",\"v1.24\": \"oke-multiarch-1.24-bd118e8-72@sha256:ec635e9eb8fa9960e92cd2032bd6b82332f21aebdf3140a3e660e976a773b92c\"}"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaa5uvuf7b3bmylrw725q4drkef5ftbeuoqfvqptzam655d44ncv5mq"
          },
          {
            regions      = ["phx"]
            env          = "prd"
            value        = "{\"default\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.16\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.17\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.18\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.19\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.20\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.21\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.22\": \"oke-multiarch-1.22-1762caa-120@sha256:b9f3cba2c9b3786885ec5601770ade5be94438d4fb89b29f87fba77601d7ba01\",\"v1.23\": \"oke-multiarch-1.23-24e21fc-44@sha256:3f6dd1e2a539c769acffea46b19bc875c5a2f267be6f1652915adc1c4c284e93\",\"v1.24\": \"oke-multiarch-1.24-bd118e8-72@sha256:ec635e9eb8fa9960e92cd2032bd6b82332f21aebdf3140a3e660e976a773b92c\"}"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaelop2bappx2pqqxzqv4myxtexgzuvwceexk3xi4zs44mk2tvvswa"
          },
          {
            regions      = ["iad"]
            env          = "dev"
            value        = "{\"default\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.16\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.17\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.18\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.19\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.20\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.21\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.22\": \"oke-multiarch-1.22-1762caa-120@sha256:b9f3cba2c9b3786885ec5601770ade5be94438d4fb89b29f87fba77601d7ba01\",\"v1.23\": \"oke-multiarch-1.23-24e21fc-44@sha256:3f6dd1e2a539c769acffea46b19bc875c5a2f267be6f1652915adc1c4c284e93\",\"v1.24\": \"oke-multiarch-1.24-bd118e8-72@sha256:ec635e9eb8fa9960e92cd2032bd6b82332f21aebdf3140a3e660e976a773b92c\"}"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaalbb2c4nf3ptr356lej2rpkjzy5cbcatbwqsphoseqokcjtueunha"
          },
          {
            regions      = ["iad"]
            env          = "integ"
            value        = "{\"default\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.16\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.17\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.18\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.19\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.20\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.21\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.22\": \"oke-multiarch-1.22-1762caa-120@sha256:b9f3cba2c9b3786885ec5601770ade5be94438d4fb89b29f87fba77601d7ba01\",\"v1.23\": \"oke-multiarch-1.23-24e21fc-44@sha256:3f6dd1e2a539c769acffea46b19bc875c5a2f267be6f1652915adc1c4c284e93\",\"v1.24\": \"oke-multiarch-1.24-bd118e8-72@sha256:ec635e9eb8fa9960e92cd2032bd6b82332f21aebdf3140a3e660e976a773b92c\"}"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaabg3uraonkvuh6kvczem6zuk6bwpvz4rca5jmhiqxzlcb4dhzxyqq"
          },
          {
            regions      = ["iad"]
            env          = "prd"
            value        = "{\"default\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.16\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.17\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.18\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.19\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.20\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.21\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.22\": \"oke-multiarch-1.22-1762caa-120@sha256:b9f3cba2c9b3786885ec5601770ade5be94438d4fb89b29f87fba77601d7ba01\",\"v1.23\": \"oke-multiarch-1.23-24e21fc-44@sha256:3f6dd1e2a539c769acffea46b19bc875c5a2f267be6f1652915adc1c4c284e93\",\"v1.24\": \"oke-multiarch-1.24-bd118e8-72@sha256:ec635e9eb8fa9960e92cd2032bd6b82332f21aebdf3140a3e660e976a773b92c\"}"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaafoai2hnvga3lzgy3ljsz4n6vlcl6wdgec25k3ogjqg43yloxwkoa"
          },
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