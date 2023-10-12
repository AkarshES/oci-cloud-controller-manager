locals {
  tenancy_property_overrides = {
    "oc1" = {
      "ccm-image-version-mapping" = {
        overrides = [
          // pm tenancy
          {
            regions      = ["phx"]
            env          = "prd"
            value        = "{\"default\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",  \"v1.16\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",  \"v1.17\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",  \"v1.18\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\", \"v1.19\": \"oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b\", \"v1.20\": \"oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b\", \"v1.21\": \"oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b\", \"v1.22\": \"oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db\", \"v1.23\": \"oke-multiarch-1.23-526d1e6-171@sha256:85235e1fa24c41e5fb158346e3339fc680dcdce791735bfca25c7755a479e4c8\", \"v1.24\": \"v1.24-32be19ef595-4@sha256:3eda1610412ce5a3f6009b1d1a9219b3fdcc59009a8e3077a83f2b82142a586e\", \"v1.25\": \"v1.25-f1ae2f635ad-2170@sha256:65c616a592162928f0f85e19b3f3564056bf717577b1ee859e96358af4e0ee69\", \"v1.26\": \"v1.26-3e5bbc3314d-2149@sha256:a47df8d8ca2faa8c5b00b93a04b822c6fc6b4978024910ed95b4a330b58aab12\", \"v1.27\": \"v1.27-75a708e7486-2171@sha256:dde76c1bf8cf70335303e5b4b8b618995a7698f79ea21b4522ec2afa3597cca6\"}"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaa5gtuu5ao3rl7tdj4fqe445ou3uvyn64c3muvnkigi2a72drw3rya"
          },
          // telesis tenancy
          {
            regions      = ["ams", "lhr"]
            env          = "prd"
            value        = "{\"default\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",  \"v1.16\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",  \"v1.17\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",  \"v1.18\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\", \"v1.19\": \"oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b\", \"v1.20\": \"oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b\", \"v1.21\": \"oke-multiarch-1.19-64ab664-255@sha256:c0b0b665735d3288d0f8991c792c51aa00f9aaa031e2ffdd5ecca0238c03f28b\", \"v1.22\": \"oke-multiarch-1.22-9893434-269@sha256:ceba7b8788c84d494113c862cd03dce2cc2c7b52c451ebeaa6eee88a97a4d8db\", \"v1.23\": \"oke-multiarch-1.23-526d1e6-171@sha256:85235e1fa24c41e5fb158346e3339fc680dcdce791735bfca25c7755a479e4c8\", \"v1.24\": \"v1.24-32be19ef595-4@sha256:3eda1610412ce5a3f6009b1d1a9219b3fdcc59009a8e3077a83f2b82142a586e\", \"v1.25\": \"v1.25-f1ae2f635ad-2170@sha256:65c616a592162928f0f85e19b3f3564056bf717577b1ee859e96358af4e0ee69\", \"v1.26\": \"v1.26-3e5bbc3314d-2149@sha256:a47df8d8ca2faa8c5b00b93a04b822c6fc6b4978024910ed95b4a330b58aab12\", \"v1.27\": \"v1.27-75a708e7486-2171@sha256:dde76c1bf8cf70335303e5b4b8b618995a7698f79ea21b4522ec2afa3597cca6\"}"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaa5kcse4xrdxteesgdei3dz42intf46jpwloof6upjqzqb6cxd7krq"
          }
        ]
      },
      "csi-image-version-mapping" = {
        overrides = [
          // pm tenancy
          {
            regions      = ["phx"]
            env          = "prd"
            value        = "{\"default\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.16\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.17\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.18\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.19\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.20\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.21\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.22\": \"oke-multiarch-1.22-d0bafe8-232@sha256:4697113594971e55df52d9e72dda7c381b431ec11d8d4622d82c7fafeb6c2689\",\"v1.23\": \"oke-multiarch-1.23-4c65264-146@sha256:defa64f16f0a16b84f3009cd6d902e4ce547dff68d3a5b85e510749d982de164\",\"v1.24\": \"v1.24-32be19ef595-4@sha256:3eda1610412ce5a3f6009b1d1a9219b3fdcc59009a8e3077a83f2b82142a586e\", \"v1.25\": \"v1.25-f1ae2f635ad-2170@sha256:65c616a592162928f0f85e19b3f3564056bf717577b1ee859e96358af4e0ee69\", \"v1.26\": \"v1.26-3e5bbc3314d-2149@sha256:a47df8d8ca2faa8c5b00b93a04b822c6fc6b4978024910ed95b4a330b58aab12\", \"v1.27\": \"v1.27-75a708e7486-2171@sha256:dde76c1bf8cf70335303e5b4b8b618995a7698f79ea21b4522ec2afa3597cca6\"}"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaa5gtuu5ao3rl7tdj4fqe445ou3uvyn64c3muvnkigi2a72drw3rya"
          },
          // telesis tenancy
          {
            regions      = ["ams", "lhr"]
            env          = "prd"
            value        = "{\"default\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.16\": \"oke-multiarch-1.16-520cc1d-11@sha256:5a38b559cbb0a027b06f9381973974854b7bc5c5085ddd9e225ddf02820cdc78\",\"v1.17\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.18\": \"oke-multiarch-1.17-40e9a7a-13@sha256:60b1e805918f93e14bf618df8e224d8ac6de004496cf484c1ffd6bc74d1e38d9\",\"v1.19\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.20\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.21\": \"oke-multiarch-1.19-73d694a-238@sha256:7215c4dcaeae4f199e82939a8a3dc73e519b00d0ac3714350b22002f8ae4f7aa\",\"v1.22\": \"oke-multiarch-1.22-d0bafe8-232@sha256:4697113594971e55df52d9e72dda7c381b431ec11d8d4622d82c7fafeb6c2689\",\"v1.23\": \"oke-multiarch-1.23-4c65264-146@sha256:defa64f16f0a16b84f3009cd6d902e4ce547dff68d3a5b85e510749d982de164\",\"v1.24\": \"v1.24-32be19ef595-4@sha256:3eda1610412ce5a3f6009b1d1a9219b3fdcc59009a8e3077a83f2b82142a586e\", \"v1.25\": \"v1.25-f1ae2f635ad-2170@sha256:65c616a592162928f0f85e19b3f3564056bf717577b1ee859e96358af4e0ee69\", \"v1.26\": \"v1.26-3e5bbc3314d-2149@sha256:a47df8d8ca2faa8c5b00b93a04b822c6fc6b4978024910ed95b4a330b58aab12\", \"v1.27\": \"v1.27-75a708e7486-2171@sha256:dde76c1bf8cf70335303e5b4b8b618995a7698f79ea21b4522ec2afa3597cca6\"}"
            tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaa5kcse4xrdxteesgdei3dz42intf46jpwloof6upjqzqb6cxd7krq"
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
