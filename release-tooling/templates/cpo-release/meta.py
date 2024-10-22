# constants
integ_phx_target = "integ.oc1.us-phoenix-1.cell0"
prod_phx_target = "prd.oc1.us-phoenix-1.cell0"


# load common variables

with open("templates/meta-commons.py") as commons:
    exec(commons.read())

# template definition

release_templates = [
    {
        "alias": "image-push",
        "flock_name": "oke-ccm-csi",
        "type": "Application"
    },
    {
        "alias": "mapping-update",
        "flock_name": "oke-ccm-csi",
        "type": "Infrastructure"
    }
]

meta_template = {
    "template_name": "cpo-release",
    "release_template_location": "../../releases/cpo-release/",
    "disclaimer": "",
    "components": [
        {
            "flock": "oke-ccm-csi",
            "artifact_version_resolvers": [
                {
                    "resolver_type": "static",
                    "artifacts": {
                        "oke-public-cloud-provider-oci-linux_arm64_v8__v1_DOT_30-2617eb746ba-5",
                        "oke-public-cloud-provider-oci-linux_x86_64__v1_DOT_30-2617eb746ba-5",
                        "oke-public-cloud-provider-oci__v1_DOT_30-2617eb746ba-5"
                    },
                    "resolver_params": {
                        "static_versions": {
                            "oke-public-cloud-provider-oci-linux_arm64_v8__v1_DOT_30-2617eb746ba-5": {
                                "version": "v1.30-2617eb746ba-5",
                                "summary": "CPO image to be pushed for release"
                            },
                            "oke-public-cloud-provider-oci-linux_x86_64__v1_DOT_30-2617eb746ba-5": {
                                "version": "v1.30-2617eb746ba-5",
                                "summary": "CPO image to be pushed for release"
                            },
                            "oke-public-cloud-provider-oci__v1_DOT_30-2617eb746ba-5": {
                                "version": "v1.30-2617eb746ba-5",
                                "summary": "CPO image to be pushed for release"
                            }
                        }
                    }
                }
            ]
        }
    ],
    "flocks": [
        {
            "shepherd_project_name": "oke",
            "flock_name": "oke-ccm-csi",
            "new_deployment_scope": {
                "config_id_or_commit_id": "96824fb5-b15c-4965-8480-91dd06d28c74",
            },
            "deployment_scope_type": "regional",
            "current_prod_deployment_scope": {
                "execution_target": prod_phx_target,
                "change_type": "Application"
            },
            "logical_phases": [
                {
                    "scope": "region",
                    "name": "main",
                    "shepherd_phase": lambda scope: f"prd.{scope.realm.lower()}",
                    "execution_target": lambda scope: f"prd.{scope.realm.lower()}.{scope.region_name.lower()}.cell0"
                }
            ]
        },
        {
            "shepherd_project_name": "oke",
            "flock_name": "oke-ccm-csi",
            "new_deployment_scope": {
                "config_id_or_commit_id": "96824fb5-b15c-4965-8480-91dd06d28c74",
            },
            "deployment_scope_type": "regional",
            "current_prod_deployment_scope": {
                "execution_target": prod_phx_target,
                "change_type": "Infrastructure"
            },
            "logical_phases": [
                {
                    "scope": "region",
                    "name": "main",
                    "shepherd_phase": lambda scope: f"prd.{scope.realm.lower()}",
                    "execution_target": lambda scope: f"spectre.values.prd.{scope.realm.lower()}.{scope.region_name.lower()}"
                }
            ]
        }
    ],
    "release_templates": release_templates,
    "interleave_phases": False,
    "schedule": regular_schedule_for_all,
    "exclude_locations": excluded_locations
}
