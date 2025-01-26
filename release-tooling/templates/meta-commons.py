# common variables shared across templates
# deployment ordering gathered from https://confluence.oci.oraclecorp.com/display/DS/CKS+CRB+Ticket+Hygiene+Guidelines#CKSCRBTicketHygieneGuidelines-DeploymentOrdering
regular_schedule_for_all = [
    {
        "bundle": "Day 1",
        "groups": [
            {
                "title": "OC5",
                "realms": ["OC5"],
                "labels": ["normal"]
            },
            {
                "title": "OC17 Deployments",
                "realms": ["OC17"],
                "labels": ["normal"]
            },
            {
                "title": "OC16 Deployments",
                "realms": ["OC16"],
                "labels": ["normal"]
            },
            {
                "title": "Group 1 - Flight 1",
                "airport_codes": ["qro"],
                "labels": ["normal"],
                "depends_on": {"groups": ["OC5", "OC16 Deployments", "OC17 Deployments"]}
            },
            {
                "title": "Group 1 - Flight 2",
                "airport_codes": ["auh", "cdg"],
                "labels": ["normal"],
                "depends_on": {"groups": ["Group 1 - Flight 1"]}
            },
            {
                "title": "Group 1 - Flight 3",
                "airport_codes": ["mtz", "mad"],
                "labels": ["normal"],
                "depends_on": {"groups": ["Group 1 - Flight 2"]}
            }
        ]
    },
    {
        "bundle": "Day 2",
        "groups": [
            {
                "title": "Group 2 - Flight 1",
                "airport_codes": ["syd", "vcp", "mrs"],
                "labels": ["normal"],
                "depends_on": {"groups": ["Group 1 - Flight 3"]}
            },
            {
                "title": "Group 2 - Flight 2",
                "airport_codes": ["hyd", "scl", "zrh", "yul"],
                "labels": ["normal"],
                "depends_on": {"groups": ["Group 2 - Flight 1"]}
            },
            {
                "title": "Group 2 - Flight 3",
                "airport_codes": ["nrt", "ams", "cwl", "jed"],
                "labels": ["normal"],
                "depends_on": {"groups": ["Group 2 - Flight 2"]}
            },
            {
                "title": "Group 2 - Flight 4",
                "airport_codes": ["sin", "icn", "arn", "dxb", "aga"],
                "labels": ["normal"],
                "depends_on": {"groups": ["Group 2 - Flight 3"]}
            }
        ]
    },
    {
        "bundle": "Day 3",
        "groups": [
            {
                "title": "OC8 NJA",
                "airport_codes": ["nja"],
                "labels": ["nri"],
                "depends_on": {"groups": ["Group 2 - Flight 4"]}
            },
            {
                "title": "Group 2 - Flight 5",
                "airport_codes": ["mel", "bom", "yyz", "sjc", "jnb"],
                "labels": ["normal"],
                "depends_on": {"groups": ["Group 2 - Flight 4"]}
            },
            {
                "title": "Group 2 - Flight 6",
                "airport_codes": ["kix", "yny", "gru", "lin", "mty"],
                "labels": ["normal"],
                "depends_on": {"groups": ["Group 2 - Flight 5"]}
            },
            {
                "title": "OC20 Deployments",
                "realms": ["OC20"],
                "labels": ["normal"],
                "depends_on": {"groups": ["Group 2 - Flight 6"]}
            },
            {
                "title": "Group 2 - Flight 7",
                "airport_codes": ["bog", "vap", "xsp", "ruh"],
                "labels": ["normal"]
            },
            {
                "title": "Group 3 - Flight 1",
                "airport_codes": ["lhr"],
                "labels": ["normal"],
                "depends_on": {"groups": ["Group 2 - Flight 6"]}
            },
            {
                "title": "Group 3 - Flight 2",
                "airport_codes": ["iad"],
                "labels": ["normal"],
                "depends_on": {"groups": ["Group 3 - Flight 1"]}
            },
            {
                "title": "Group 3 - Flight 3",
                "airport_codes": ["phx"],
                "labels": ["normal"],
                "depends_on": {"groups": ["Group 3 - Flight 2"]}
            }
        ]
    },
    {
        "bundle": "Day 4",
        "groups": [
            # phase 1 starts here
            {
                "title": "Group 3 - Flight 4",
                "airport_codes": ["fra"],
                "labels": ["normal"],
                "depends_on": {"groups": ["Group 3 - Flight 3"]}
            },
            {
                "title": "OC4 - uk-gov-cardiff-1",
                "airport_codes": ["brs"],
                "labels": ["ukgov"],
                "depends_on": {"groups": ["Group 3 - Flight 3"]}
            },
            {
                "title": "OC2 - us-langley-1",
                "airport_codes": ["lfi"],
                "labels": ["usgov"],
                "depends_on": {"groups": ["Group 3 - Flight 3"]}
            },
            {
                "title": "OC3 - us-gov-phoenix-1",
                "airport_codes": ["tus"],
                "labels": ["usgov"],
                "depends_on": {"groups": ["Group 3 - Flight 3"]}
            },
            {
                "title": "OC14 - eu-dcc-dublin-1",
                "airport_codes": ["ork"],
                "labels": ["normal"],
                "depends_on": {"groups": ["Group 3 - Flight 3"]}
            },
            # phase 2 starts here
            {
                "title": "Group 3 - Flight 5",
                "airport_codes": ["ord"],
                "labels": ["normal"],
                "depends_on": {"groups": ["Group 3 - Flight 4"]}
            },
            {
                "title": "OC4 - uk-gov-london-1",
                "airport_codes": ["ltn"],
                "labels": ["ukgov"],
                "depends_on": {"realms": ["OC1"]}
            },
            {
                "title": "OC2 - us-luke-1",
                "airport_codes": ["luf"],
                "labels": ["usgov"],
                "depends_on": {"groups": ["OC2 - us-langley-1"]}
            },
            {
                "title": "OC3 - us-gov-chicago-1",
                "airport_codes": ["pia"],
                "labels": ["usgov"],
                "depends_on": {"groups": ["OC3 - us-gov-phoenix-1"]}
            },
            {
                "title": "OC14 - eu-dcc-milan-1",
                "airport_codes": ["bgy"],
                "labels": ["normal"],
                "depends_on": {"groups": ["OC14 - eu-dcc-dublin-1"]}
            },
            {
                "title": "OC10 Deployments",
                "realms": ["OC10"],
                "labels": ["normal"],
                "depends_on": {"realms": ["OC1"]}
            },
            # phase 3 starts here
            {
                "title": "OC3 - us-gov-ashburn-1",
                "airport_codes": ["ric"],
                "labels": ["usgov"],
                "depends_on": {"groups": ["OC3 - us-gov-chicago-1"]}
            },
            {
                "title": "OC14 - eu-dcc-rating-1",
                "airport_codes": ["dus"],
                "labels": ["normal"],
                "depends_on": {"groups": ["OC14 - eu-dcc-milan-1"]}
            },
            {
                "title": "OC9 Deployments",
                "realms": ["OC9"],
                "labels": ["normal"],
                "depends_on": {"realms": ["OC1"]}
            },
            {
                "title": "OC15-ap-dcc-gazipur-1",
                "airport_codes": ["dac"],
                "labels": ["normal"],
                "depends_on": {"realms": ["OC1"]}
            },
            # phase 4 starts here
            {
                "title": "OC14 - remaining",
                "airport_codes": ["snn", "mxp", "dtm"],
                "labels": ["normal"],
                "depends_on": {"groups": ["OC14 - eu-dcc-rating-1"]}
            },
            {
                "title": "OC8 UKB",
                "airport_codes": ["ukb"],
                "labels": ["nri"],
                "depends_on": {"airport_codes": ["nja"]}
            },
            # phase 5 starts here
            {
                "title": "OC22 - nap",
                "airport_codes": ["nap"],
                "labels": ["normal"]
            },
            {
                "title": "OC24",
                "airport_codes": ["avz", "avf"],
                "labels": ["normal"],
                "depends_on": {"groups": ["OC22 - nap"]}
            },
            {
                "title": "OC6 - FTW",
                "airport_codes": ["ftw"],
                "labels": ["onsr"],
                "depends_on": {"realms": ["OC5"]}
            },
            {
                "title": "OC6 - DCA",
                "airport_codes": ["dca"],
                "labels": ["onsr"],
                "depends_on": {"realms": ["OC5"]}
            },
            {
                "title": "OC11 Deployments",
                "airport_codes": ["hef", "gyr", "dal"],
                "labels": ["onsr"],
                "depends_on": {"realms": ["OC1"]}
            },
            {
                "title": "OC19 Deployments",
                "realms": ["OC19"],
                "labels": ["normal"],
                "depends_on": {"realms": ["OC1"]}
            },
            {
                "title": "OC21 - me-dcc-doha-1",
                "airport_codes": ["doh"],
                "labels": ["normal"],
                "depends_on": {"realms": ["OC1"]}
            },
            {
                "title": "OC25",
                "airport_codes": ["tyo", "uky"],
                "labels": ["normal"],
                "depends_on": {"realms": ["OC1"]}
            },
            {
                "title": "OC27 Deployments",
                "realms": ["OC27"],
                "labels": ["normal"],
                "depends_on": {"realms": ["OC1"]}
            },
            {
                "title": "OC28 Deployments",
                "realms": ["OC28"],
                "labels": ["normal"],
                "depends_on": {"realms": ["OC1"]}
            },
            {
                "title": "OC26 Deployments",
                "airport_codes": ["ahu", "rba"],
                "labels": ["normal"],
                "depends_on": {"realms": ["OC1"]}
            },
            {
                "title": "OC31 Deployments",
                "airport_codes": ["izq", "jjt"],
                "labels": ["normal"],
                "depends_on": {"realms": ["OC1"]}
            },
            {
                "title": "OC29 Deployments",
                "airport_codes": ["rkt", "shj"],
                "labels": ["normal"],
                "depends_on": {"realms": ["OC1"]}
            },
            {
                "title": "OC23 Deployments",
                "airport_codes": ["ebb", "ebl"],
                "labels": ["normal"],
                "depends_on": {"realms": ["OC1"]}
            },
            {
                "title": "OC35 Deployments",
                "airport_codes": ["dln", "bno", "dtz"],
                "labels": ["normal"],
                "depends_on": {"realms": ["OC1"]}
            }
        ]
    }
]

excluded_locations = [
    "OC1-eu-kragujevac-1",
    "OC7-us-gov-sterling-1",
    "OC7-us-gov-fortworth-2",
    "OC9-me-duqm-1",
    "OC12-us-gov-manassas-1",
    "OC12-us-gov-saltlakecity-1",
]
