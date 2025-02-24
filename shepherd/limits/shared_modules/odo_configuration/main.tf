data "odo_pools" "odo-pool" {
  ad                      = var.physical_ad1
  pool_name_regex         = var.pool_name_regex
}

resource "odo_application" "applications" {
  ad                      = var.physical_ad1
  alias                   = var.application_alias
  compartment_ocid        = var.compartment_id
  type                    = var.stage == "prod" ? "PRODUCTION" : "NON_PRODUCTION"
  default_artifact_source = "DEFAULT"
  artifact_set_identifier = var.artifact_set_identifier
  agent                   = "HOSTAGENT_V2"
  pools                   = data.odo_pools.odo-pool.pools[*].resource_id

  config {
    deployments {
      min_nodes_in_service_percent = 0
      parallelism_type = "HOSTS"
      parallelism = 1
      deploy_sequentially = true
      fault_domain_deploy_sequentially = true
      ttl_seconds_pull_image           = 240
      ttl_seconds_start_instance       = 240
      ttl_seconds_stop_instance        = 300
      ttl_seconds_validation           = 240
    }

    environment_variables {
      name = "STEWARD_TENANCY_OCID"
      value = local.steward_tenancy_ocid
    }

    environment_variables {
      name = "REGION"
      value = local.execution_target.region.name
    }

    dynamic environment_variables {
      for_each = {for variable in var.env_vars: variable.name => variable.value}
      content {
        name  = environment_variables.key
        value = environment_variables.value
      }
    }
  }

  lifecycle {
    ignore_changes = [ default_artifact_source ]
  }
}