locals {
  execution_target = var.execution_target
  steward_tenancy_info = {
    region1 = {
      ocid = "ocid1.tenancy.region1..aaaaaaaad74wznrnar2var53t6kghhpa3xkial2cfmize7tujsq6fxz73sza"
    }
    oc1 = {
      ocid = "ocid1.tenancy.oc1..aaaaaaaafslvxmi2x2p2tgzxv3fgyzogwfzjjpypfvp2zvck2wk6t3bocnqa"
    }
    oc2 = {
      ocid = "ocid1.tenancy.oc2..aaaaaaaa7ztiqj65updqb3ky2tbkxgtbbxwukdgxtenjm772lqyuovo54wsa"
    }
    oc3 = {
      ocid = "ocid1.tenancy.oc3..aaaaaaaaahwpcsjg6dlq3n32sfeuijlx6f3g66j5mtf2ejg5i7evaxgwyscq"
    }
    oc4 = {
      ocid = "ocid1.tenancy.oc4..aaaaaaaaytm6y4lhmpbj6gjwqh27fzgqn7uzipjiqnrgg4efyym726g5afzq"
    }
    default = {
      ocid = "ocid1.tenancy.%s..aaaaaaaaguapt5ory5p7yl54pj4pcua3tdtagkqcoorsx43mp6c6es6kqkva"
    }
  }

  default_ocid = format(local.steward_tenancy_info.default.ocid, local.execution_target.region.realm)
  steward_tenancy_ocid = lookup(local.steward_tenancy_info, local.execution_target.region.realm, { ocid = local.default_ocid }).ocid
}