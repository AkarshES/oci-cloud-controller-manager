locals {
  tally_csi_overrides = {
    "v1.31" : "v1.31-d02f08265e8-67@sha256:4d0b75d5875307be05859d157e9d647b0703f878100e49a6a0df38b53d2dcaa3",
    "v1.32" : "v1.32-bf20a443646-33@sha256:6ca6fd685627f329769e110f34c4d48d78621859539ce64aa2de4fa868ed10a0"
  }

  // !!! Do not play with fire, only used by ci-cd pipeline
  oci_cnp_dev_override = {
  }

  // !!! Do not play with fire, only used by ci-cd pipeline
  oci_cnp_dev_ccm_override = {
    "default": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
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
    "v1.27": "v1.27-7e9bf7a9189-52@sha256:fd510337e52ef609ffb86953ef080b3f2fa1a19a0647f65dfbd2c1dfbda0df7a",
    "v1.28": "v1.28-79d4f40b682-84@sha256:d8e0957a384781955a1e6b1dfb5498fa8ffb585face9a407486692ca685343fa",
    "v1.29": "v1.29-393f7c992a6-112@sha256:cfc0512abe6e31a1e6020e45ebd7e79f89dc8f7bf5edc2655757762d5c662886",
    "v1.30": "v1.30-c435522d651-125@sha256:3035c0f0f0c3d31fa6ca1e0db6e7bf9a1122160cca9848001381927c941e3fd4",
    "v1.31": "v1.31-cb3b26234f1-102@sha256:8f9b9816ad153b127e27d898785c1d79928dbe4b6564db2dd935241aaa296f18",
    "v1.32": "v1.32-796d1c8d4a1-71@sha256:9e92facfdb41b844ccff3a492afcad68535870bf554812582310b54101bf9f73",
    "v1.33": "v1.33-03761e1e1b3-41@sha256:dd236948ebe6f2f672840fce4af1a75f3533c32d62a0841e0fb4b839ea01031b",
    "v1.34": "v1.34-56769a515a4-14@sha256:d0b10ed378ea025615a6a5835db4a61bc31df1ef20ff2fce8a207839cb99bd85",
    "v1.35": "v1.34-56769a515a4-14@sha256:d0b10ed378ea025615a6a5835db4a61bc31df1ef20ff2fce8a207839cb99bd85"
  }
  // !!! Do not play with fire, only used by ci-cd pipeline
  oci_cnp_dev_csi_override = {
    "default": "oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78",
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
    "v1.27": "v1.27-bff7091ad1d-51@sha256:10555f13db26cbcb35dcd684be88db9897e01ad25bdcd7d65e9cb4bd1c5f386f",
    "v1.28": "v1.28-cb1635cc6c7-80@sha256:d9e51c8c78b3ef040e5739d6ac079d6587bb4b03d8b65579a877b4ec84c4f219",
    "v1.29": "v1.29-0f63a5020b8-110@sha256:466c7d32860ef68c4c98feba232409bf1580c732ee23fd3294ce59e2653bd125",
    "v1.30": "v1.30-9696a00641f-5952@sha256:ef50fb8445e15b6e816bc68d6261f07e10e66b758da2923df023dbe4bd82da47",
    "v1.31": "v1.31-4e30e28b828-80-csi@sha256:bf55f642531ebcb3e8ec09c3adeb9507552733df58e9fc6b7692bc241d5df2ad",
    "v1.32": "v1.32-2c5fcd2e853-46-csi@sha256:fb9e892af78589a74bf8a85fa47af4de66cb97a5fbe33846a5e4380f97c024ec",
    "v1.33": "v1.33-87690329d0a-20-csi@sha256:5c7e230d58e1b6faed400bbe3744a3608fca42f33c2c4b5e281abc7df5489a0f",
    "v1.34": "v1.34-8f0fbf7e71e-9-csi@sha256:2eca76b52bc3198f86b839c199f67476100bc614e401dec25ac0a2310f609c28",
    "v1.35": "v1.34-8f0fbf7e71e-9-csi@sha256:2eca76b52bc3198f86b839c199f67476100bc614e401dec25ac0a2310f609c28"
  }

  tenancy_property_overrides = {
    "oc1" = {
      "ccm-image-version-mapping" = {
        overrides = [
          /*
            oci-cnp-dev tenancy override. Not to be removed as it is used by pipelines
          */
          {
            regions      = ["phx"]
            env          = "integ"
            value        = jsonencode(merge(local.ccm_default_mapping.default.all, local.oci_cnp_dev_ccm_override))
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaajol5woa4is3merb234fy4b46bps2nsjr3lcz7rvgj25dr5dxfmnq"
          }
        ]
      },
      "csi-image-version-mapping" = {
        overrides = [
          /*
            oci-cnp-dev tenancy override. Not to be removed as it is used by pipelines
          */
          {
            regions      = ["phx"]
            env          = "integ"
            value        = jsonencode(merge(local.csi_default_mapping.default.all, local.oci_cnp_dev_csi_override))
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaajol5woa4is3merb234fy4b46bps2nsjr3lcz7rvgj25dr5dxfmnq"
          },
          /*
            Tally Prod CSI Pin Override
          */
          {
            regions      = ["bom"]
            env          = "prd"
            value        = jsonencode(merge(local.csi_default_mapping.default.all, local.tally_csi_overrides))
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaosp7wej2fqrvacexolafmu3kw3euc2k5lwuirycvwzhspoleqrsq"
          },
          /*
            Tally QA CSI Pin Override
          */
          {
            regions      = ["bom"]
            env          = "prd"
            value        = jsonencode(merge(local.csi_default_mapping.default.all, local.tally_csi_overrides))
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaxtpwu4jog5czcjmawmwptkuqjfimnc57iij6asca4xwjx3wkf3pq"
          },
          /*
            Tally DEV CSI Pin Override
          */
          {
            regions      = ["bom"]
            env          = "prd"
            value        = jsonencode(merge(local.csi_default_mapping.default.all, local.tally_csi_overrides))
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaacihe4hnurwedhwfbgvexk2uemvwpyoq37vywgl6yh5n3a7wpopca"
          },
        ]
      },
      "lustre-csi-driver-enabled" = {
        /* hilbert tenancy - Temple*/
        overrides = [
          {
            regions      = ["aga"]
            env          = "prd"
            value        = "false"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaad6ghxkqwame6d4ioz2ejmrhksjxb7pif3fpp3r44fweejt2jvptq"
          }
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
          // Hurray, no snowflakes
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
