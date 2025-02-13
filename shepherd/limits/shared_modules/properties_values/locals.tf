locals {

  roma_ccm_mapping_override = {
    "default" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
    "v1.16" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
    "v1.17" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
    "v1.18" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
    "v1.19" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
    "v1.20" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
    "v1.21" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
    "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
    "v1.29" : "v1.29-b42bf1ab166-4766@sha256:9124dc96512a7998910cd4bb5cc07dc2e0f9e8ae243ff56f97898d53b357c977",
    "v1.30" : "v1.30-dd9a6803a33-4765@sha256:efb767e5f1d53e223794b91a287f0fe099a4eba9ac7054852dace9646ec24b85",
    "v1.31" : "v1.31-405cad3eb26-4768@sha256:e6ec4e4c031a086188660d52c822abfd4828742b2a6269e66c0f6b034a71f2f7"
  }

  yubi_ccm_mapping_override = {
    "default" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
    "v1.16" : "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
    "v1.17" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
    "v1.18" : "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
    "v1.19" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
    "v1.20" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
    "v1.21" : "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
    "v1.22" : "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
    "v1.30" : "v1.30-9996b0758fd-4095@sha256:9b9e7a0e3fd8f124065bc41c998da5187bc35ef3e14de7aafee7bce18d79573d",
  }

  spectra_ccm_mapping_override_hotfix_override = {
    "default": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
    "v1.16": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
    "v1.17": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
    "v1.18": "oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9",
    "v1.19": "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
    "v1.20": "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
    "v1.21": "oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b",
    "v1.22": "oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db",
    "v1.23": "oke-multiarch-1.23-526d1e6-171@sha256:85235e1fa24c41e5fb158346e3339fc680dcdce791735bfca25c7755a479e4c8",
    "v1.24": "v1.24-32be19ef595-4@sha256:3eda1610412ce5a3f6009b1d1a9219b3fdcc59009a8e3077a83f2b82142a586e",
    "v1.25": "v1.25-d0b59914251-38@sha256:ffa44cb1e6bc5793859a1ebf7762bcd1731ec0dea426034c44bc77ffb12ece48",
    "v1.26": "v1.26-8744a6c9ccd-42@sha256:bc20c825c5e3b5f40b56467e3b931597a6edef41cd0ab0cb20524b4cd8e603a0",
    "v1.27": "v1.27-bff7091ad1d-51@sha256:10555f13db26cbcb35dcd684be88db9897e01ad25bdcd7d65e9cb4bd1c5f386f",
    "v1.28": "v1.28-96b9f4622cc-66@sha256:e025418727bf3a7cbf4e33627fa434c0dec25ade6bd6393422c74c6b2ba6e6df",
    "v1.29": "v1.29-ad69a5aaaf2-4426@sha256:003224bebc9188458645ba42cfac02f48525da5cc350120c0ec2c9554e0c02b2",
    "v1.30": "v1.30-33eb1206a53-4413@sha256:174436d74662a9cd8cf444acfd1cfcc9991a5e6eb0da2e36bc70b47e6c4a43da",
    "v1.31": "v1.31-2bebf836dd5-4427@sha256:e485910f101c53006334d13fdead41d6b5a9bea5852b4b7616a358bae5681e1d"
  }

  tenancy_property_overrides = {
    "oc1" = {
      "ccm-image-version-mapping" = {
        overrides = [
          // YUBI GRPC override https://jira.oci.oraclecorp.com/browse/OKE-33286
          {
            regions = ["bom", "hyd"]
            env     = "prd"
            value   = jsonencode(merge(local.ccm_default_mapping.default.all, local.yubi_ccm_mapping_override))
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaxcauqzilnjm4aaabx35cjcfjvzvef5yuh3e77xmja2ehnoxtdc7a"
          },
          /*
          Pin Spectra tenancies with CPO 1.4.1 mappings to avoid any future reconcilation impacting NLB
          https://jira.oci.oraclecorp.com/browse/OKE-34092
          */
          {
            regions = ["phx"]
            env = "prd"
            value = jsonencode(local.spectra_ccm_mapping_override_hotfix_override)
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaa2d33riulr5do2zd6s3qpp44ljaji3t47433p43k3rcw7tsm3n47a"
          },
          {
            regions = ["phx"]
            env = "prd"
            value = jsonencode(local.spectra_ccm_mapping_override_hotfix_override)
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaziiwdjytm4z2gbij4xchpuywbwxozkfod3z4g3qorsz5zcyufz7a"
          },
          /*
          Pinned OMK tenancies with CPO 1.6.1 ccm mappings for Roma LA
          https://jira.oci.oraclecorp.com/browse/OKE-35076
          */
          {
            regions = ["iad", "phx"]
            env = "integ"
            value = jsonencode(merge(local.ccm_default_mapping.default.all, local.roma_ccm_mapping_override))
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaafyd4yolgubij6qz7kgxxl5ps2qhqkfoqrvxtjvdccy66ifxmglia"
          },
          {
            regions = ["iad", "phx"]
            env = "integ"
            value = jsonencode(merge(local.ccm_default_mapping.default.all, local.roma_ccm_mapping_override))
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaj4ymkytn4cwasyktbc5rbqiald2enkoarepdv7pawyuw2tkazisa"
          },
          {
            regions = ["iad", "phx"]
            env = "prd"
            value = jsonencode(merge(local.ccm_default_mapping.default.all, local.roma_ccm_mapping_override))
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaqeq267dfoqjqjvcfgazlo6tqv7nfnv2rv767dm5ecf2wqlbsh5sa"
          },
          {
            regions = ["iad", "phx"]
            env = "prd"
            value = jsonencode(merge(local.ccm_default_mapping.default.all, local.roma_ccm_mapping_override))
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaepbccznc477ibxxopf2o3ki4bvfn47ji2zqvdfluqtutcduggn4q"
          },
          {
            regions = ["iad", "phx"]
            env = "prd"
            value = jsonencode(merge(local.ccm_default_mapping.default.all, local.roma_ccm_mapping_override))
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaa3jtqmhbxuyvz42zvw7ltv5elx3dgypxjmtqd7yvhveqer7r7vwqa"
          },
          {
            regions = ["iad", "phx", "sjc", "syd", "fra"]
            env = "prd"
            value = jsonencode(merge(local.ccm_default_mapping.default.all, local.roma_ccm_mapping_override))
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaa3gudsvkwk52fnlmglncjuir56hhbuxoffdvc5muc2aqsuelmfuma"
          },
          {
            regions = ["iad", "phx", "sjc", "syd", "fra", "cwl"]
            env = "prd"
            value = jsonencode(merge(local.ccm_default_mapping.default.all, local.roma_ccm_mapping_override))
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaar4ugvbhybeczkonzvnyz3h4bqlqezfhpibviyqtyenzozj5xh4va"
          },
          {
            regions = ["iad", "phx", "sjc", "syd", "fra"]
            env = "prd"
            value = jsonencode(merge(local.ccm_default_mapping.default.all, local.roma_ccm_mapping_override))
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaarf636mkhnvtlwm77bhcukivkpql4qz2lhcximxyhsgnliud5ylwa"
          },
          {
            regions = ["iad", "phx", "sjc", "syd", "fra", "cwl"]
            env = "prd"
            value = jsonencode(merge(local.ccm_default_mapping.default.all, local.roma_ccm_mapping_override))
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaauw2eedsk3udqtclhqpphcsy4xhxfrtr4uiqwa3efx5hzzaaas45a"
          }
        ]
      },
      "csi-image-version-mapping" = {
        overrides = [
          // Hurray, no snowflakes
        ]
      },
      "oci-service-controller-enabled" = {
        overrides = [
          {
            regions      = ["iad"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaanrbjdkpfvz2w2ve67ecxy24pa6k3teofpvlbviyl2r5rsbwseqvq"
          },
          {
            regions      = ["phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaa7wzb7yxrfd3vyhwbkgt5yzgbp6nqmf2xegio5ssi6lvxfwps4w3a"
          },
          {
            regions      = ["iad"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaahp7giiq4smoqb5hmqnz5xmplfwuyfpovxivy22qczlrlgkzns7qq"
          },
          {
            regions      = ["iad"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaahm2iczdm7ekujxrwgeicw6k2fedhsl742gauuvpxxbz6uvhlcekq"
          },
          {
            regions      = ["iad"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaapagptfoo7j4ccd3j3vv3fbz3kwiz5jxcgmemxvt2fikluz6w3gqa"
          },
          {
            regions      = ["iad"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaa5uvuf7b3bmylrw725q4drkef5ftbeuoqfvqptzam655d44ncv5mq"
          },
          {
            regions      = ["phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaelop2bappx2pqqxzqv4myxtexgzuvwceexk3xi4zs44mk2tvvswa"
          },
          {
            regions      = ["iad"]
            env          = "dev"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaalbb2c4nf3ptr356lej2rpkjzy5cbcatbwqsphoseqokcjtueunha"
          },
          {
            regions      = ["iad"]
            env          = "integ"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaabg3uraonkvuh6kvczem6zuk6bwpvz4rca5jmhiqxzlcb4dhzxyqq"
          },
          {
            regions      = ["iad"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaafoai2hnvga3lzgy3ljsz4n6vlcl6wdgec25k3ogjqg43yloxwkoa"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaauklway4qjfxmzffkzloi65uxf2r5ak4rp7cccxlq6krk4seww7pa"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaakjncznwynlcsjpxrqub3jskbzmz3qlkgoffiv7yjmyrfqqgy7gaq"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaapbxhbccxcjrb4wnlyaix44kkcl2tgtj2lyrlcqcctqmmi26yehxa"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaogcmvxwqogkitjiepz3kqdv2irjuyhcrcmp7yinh3ytca3v56aza"
          },
          {
            regions      = ["iad"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaa5husazs2ytt4vmvztlaifts4zraa26vz6p7i76vgugoic7a5ddga"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaooqw2ki26evdbangxtahcgcbqr7ycygxkxr5y5d6vd5xwl7eqj2a"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaapofprzxfp4munko4mvqc3u4nqegtkru6uzbaut774xrgmvsco2qa"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaao5xkfdkgtgakm2gngfktpoxddhgak3sqlbuuc6ipsj4uqijhavsa"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaalgye6or5foisxcezh5hrrloiqedji3gseyz4jlkjy4ixa7en4yra"
          },
          {
            regions      = ["iad", "phx", "fra", "yyz"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaa62vpswgwthtfeiqsapy5dydvyk2bzzq6v3txo6gvh5g6o65ljuha"
          },
          {
            regions      = ["fra"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaafipe4lmow7rfrn5f3egpg3xgur6v2q2wgvb3id4ehwujnpu5mb5q"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaahzy3x4boh7ipxyft2rowu2xeglvanlfewudbnueugsieyuojkldq"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaazo2fdm3n4e5ldbcuw4nqbszce4wzynnff3vswas3frbdznidf73a"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaafyd4yolgubij6qz7kgxxl5ps2qhqkfoqrvxtjvdccy66ifxmglia"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaj4ymkytn4cwasyktbc5rbqiald2enkoarepdv7pawyuw2tkazisa"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaa3iy7zjbsehobgcdr3hr54ohhfzz6gb3mlrx3nil3rc5mefibhdqq"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaageglfk6dntbw6xlw3a2zwyxmojqdltjrrk6qrhergy2mmnuvgzga"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaa6myvkd2ri4u2p5rdgydenosfxgjrorlgxzede2dlmq37cbdagffq"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaiirvljo5b5ps4etbubykf3sorn46ih2cq47zbjwqpj3x3jmcxara"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaal4apn7aq4e3h4zv6cqz56omhwgmnrd6sepnjc3x2vifhdaes4oq"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaqeltzy3b3ioqnyqusa2po52d3ygn24yb72b44ceau4jdwoy2pnzq"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaa64d4rod2cgrqleyskgao4wkbeosylqellmeqbvxxkybhonlrbgk"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaa4fdy3gbpphhp2v7lrworctpyjli5yvwhoa7db6uutr4xqwquq5eq"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaasyc4pviq2l4mjafhhk2hejvsprv52jblh53jtujqoxi4knifzgya"
          },
          {
            regions      = ["iad", "phx", "fra", "yyz"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaatf553tfabldbvae4sfe3gqizjetiitaa6y7zyx2pazejqrufxlya"
          },
          {
            regions      = ["iad"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaagwykkho6dzrdst3nur52t3z5p6qk77z7dibjrnkbsfqm4peycqra"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaavyaz2rpyk5f54eqs4izohq2go3t54l437njmig2aonxqwboyqo3a"
          },
          {
            regions      = ["iad", "phx", "fra"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaa4wptnxymnypvjjltnejidchjhz6uimlhru7rdi5qb6qlnmrtgu3a"
          },
          {
            regions      = ["iad", "phx", "fra", "yyz"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaypz3jeouf67rycnobkfaj7zepotuvpoqtqipfai3qs4qw7ok6yna"
          },
          {
            regions      = ["iad", "phx", "fra", "yyz"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaymp6s54butpavvjb6fcwq3pzrrqbb33v7sh6qvzdqkimnbuqasla"
          },
          {
            regions      = ["phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaajznex5attydtrmrgudwayqu7kn4krasw2ct4h4pwz7nwbfxoyd4q"
          },
          {
            regions      = ["phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaax7tm7jtfarexna447cmubjxwou6lug42jss2ddyis63wqo3lrpda"
          },
          {
            regions      = ["iad", "phx"]
            env          = "dev"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaizwb7xbe7nt3wfb2jdrnsehrbip53s64qtwi2hx4y3ydkdmaeywq"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaat37ab62ltpvzgyoydasbfig3gcmccxwzvbi6yoh6ewqiiswps6sq"
          },
          {
            regions      = ["iad", "phx"]
            env          = "prd"
            value        = "true"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaaft2mucrx352mvjgie7ez7lobrnvn56hv5lkxm7kbv6m6wrwo22q"
          }
        ]
      }
    },
    "oc16" = {
      "ccm-image-version-mapping" = {
        overrides = [
          /*
          Pinned OMK tenancies with CPO 1.6.1 ccm mappings for Roma LA
          https://jira.oci.oraclecorp.com/browse/OKE-35076
          */
          {
            regions = ["sgu"]
            env = "prd"
            value = jsonencode(merge(local.ccm_default_mapping.default.all, local.roma_ccm_mapping_override))
            tenancy_ocid = "ocid1.tenancy.oc16..aaaaaaaai5vbmtglv6yngv6gldw6u4iqksco3zt5ynaxgy5sutaiqsord45q"
          },
          {
            regions = ["sgu"]
            env = "prd"
            value = jsonencode(merge(local.ccm_default_mapping.default.all, local.roma_ccm_mapping_override))
            tenancy_ocid = "ocid1.tenancy.oc16..aaaaaaaauv7ahg73tey6jj2kx3vhl56qs2anwehvtqluwrk5gitggsqlvt3q"
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
