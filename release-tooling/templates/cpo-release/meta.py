# constants
integ_phx_target = "integ.oc1.us-phoenix-1.cell0"
prod_phx_target = "prd.oc1.us-phoenix-1.cell0"

app     = meta_variables.get('app')
infra   = meta_variables.get('infra')

# load common variables

with open("templates/meta-commons.py") as commons:
    exec(commons.read())

config_id="6a6abdc6-c48e-4524-a039-b841231caba9"

app_release_template = {
    "alias": "image-push",
    "flock_name": "oke-ccm-csi",
    "type": "Application"
}

infra_release_template = {
    "alias": "mapping-update",
    "flock_name": "oke-ccm-csi",
    "type": "Infrastructure"
}

meta_template = {
    "template_name": "cpo-release",
    "release_template_location": "../../releases/cpo-release/",
    "disclaimer": "",
    "components": [],
    "flocks": [],
    "release_templates": [],
    "schedule": oke_common_release_schedule,
    "exclude_locations": excluded_locations
}

if app:
    meta_template["release_templates"].append(app_release_template)
    meta_template["flocks"].append({
        "shepherd_project_name": "oke",
        "flock_name": "oke-ccm-csi",
        "new_deployment_scope": {
            "config_id_or_commit_id": config_id,
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
    })
    meta_template["components"].append({
        "flock": "oke-ccm-csi",
        "artifact_version_resolvers": [
            {
                "resolver_type": "static",
                "artifacts": {
                    "oke-public-cloud-provider-oci__v1_DOT_29-0911461af79-92",
                    "oke-public-cloud-provider-oci__v1_DOT_30-a67f7b269a7-85",
                    "oke-public-cloud-provider-oci__v1_DOT_31-c7e5bd92e29-43",
                    "oke-public-cloud-provider-oci__v1_DOT_32-c01d1d4113e-11",
                    "release-validator-ccm-csi__f12d27156e9_10"
                },
                "resolver_params": {
                    "static_versions": {
                        "oke-public-cloud-provider-oci__v1_DOT_29-0911461af79-92": {
                            "version": "v1.29-0911461af79-92",
                            "summary": "CPO image to be pushed for release"
                        },
                        "oke-public-cloud-provider-oci__v1_DOT_30-a67f7b269a7-85": {
                            "version": "v1.30-a67f7b269a7-85",
                            "summary": "CPO image to be pushed for release"
                        },
                        "oke-public-cloud-provider-oci__v1_DOT_31-c7e5bd92e29-43": {
                            "version": "v1.31-c7e5bd92e29-43",
                            "summary": "CPO image to be pushed for release"
                        },
                        "oke-public-cloud-provider-oci__v1_DOT_32-c01d1d4113e-11": {
                            "version": "v1.32-c01d1d4113e-11",
                            "summary": "CPO image to be pushed for release"
                        },
                        "release-validator-ccm-csi__4fe27dbfb8e_20": {
                            "version": "4fe27dbfb8e_20",
                            "summary": "Release validator POP image to be pushed for release"
                        },
                    }
                }
            }
        ]
    })

if infra:
    meta_template["release_templates"].append(infra_release_template)
    meta_template["flocks"].append({
        "shepherd_project_name": "oke",
        "flock_name": "oke-ccm-csi",
        "new_deployment_scope": {
            "config_id_or_commit_id": config_id,
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
                "execution_target": lambda scope: f"prd.{scope.realm.lower()}.{scope.region_name.lower()}.cell0"
            },
            {
                "scope": "region",
                "name": "main",
                "shepherd_phase": lambda scope: f"prd.{scope.realm.lower()}",
                "execution_target": lambda scope: f"spectre.values.prd.{scope.realm.lower()}.{scope.region_name.lower()}"
            }
        ]
    })
