# constants
integ_phx_target = "integ.oc1.us-phoenix-1.cell0"
prod_phx_target = "prd.oc1.us-phoenix-1.cell0"

app     = meta_variables.get('app')
infra   = meta_variables.get('infra')

# load common variables

with open("templates/meta-commons.py") as commons:
    exec(commons.read())

config_id="5c133ea6-0dd7-4f63-8797-ed178d2db098"

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
                    "oke-public-cloud-provider-oci__v1_DOT_31-4e30e28b828-80-csi",
                    "oke-public-cloud-provider-oci__v1_DOT_32-2c5fcd2e853-46-csi",
                    "oke-public-cloud-provider-oci__v1_DOT_33-87690329d0a-20-csi",
                    "release-validator-ccm-csi"
                },
                "resolver_params": {
                    "static_versions": {
                        "oke-public-cloud-provider-oci__v1_DOT_31-4e30e28b828-80-csi": {
                            "version": "v1.31-4e30e28b828-80-csi",
                            "summary": "CPO image to be pushed for release"
                        },
                        "oke-public-cloud-provider-oci__v1_DOT_32-2c5fcd2e853-46-csi": {
                            "version": "v1.32-2c5fcd2e853-46-csi",
                            "summary": "CPO image to be pushed for release"
                        },
                        "oke-public-cloud-provider-oci__v1_DOT_33-87690329d0a-20-csi": {
                            "version": "v1.33-87690329d0a-20-csi",
                            "summary": "CPO image to be pushed for release"
                        },
                        "release-validator-ccm-csi": {
                            "version": "0833a431093_111",
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
