locals {
  shepherd_current_regions   = [for region in local.shepherd_all_regions : region if region.realm != "region1" && region.state == "Production" && ! contains(local.blacklist_realms, region.realm)]
  regions_under_build        = [for region in local.shepherd_all_regions : region if(region.state == "Building") && region.airport_code != "DCA" && region.airport_code != "QDF" && ! contains(local.blacklist_realms, region.realm)]
  build_region_to_realm      = { for region in local.regions_under_build : region.public_name => region.realm }
  region_to_realm            = { for region in local.shepherd_all_regions : region.public_name => region.realm }
  region_by_name             = { for region in local.shepherd_current_regions : region.public_name => region }
  region_by_name_all_regions = merge({ for region in local.shepherd_all_regions : region.public_name => region }, { for region in local.regions : region.public_name => region })
  realm_by_name              = { for realm in local.shepherd_all_realms : realm.name => realm if ! contains(local.blacklist_realms, realm.name) }
  prod_phases                = [for realm in local.shepherd_all_realms : realm.name if ! contains(local.blacklist_realms, realm.name) && ! (realm.attributes.is_disconnected || realm.is_sovereign_realm)]
  home_region_by_realm       = { for realm in local.realm_by_name : realm.name => realm.attributes.first_region }
  realms_under_build         = [for region in local.regions_under_build : region.realm if local.home_region_by_realm[region.realm] == region.public_name]
  blacklist_realms           = ["region1", "integ-next", "integ-stable", "dev", "oc7", "oc12"]
  onsr_realm_by_name         = { for realm in local.shepherd_all_realms : realm.name => realm if ! contains(local.blacklist_realms, realm.name) && (realm.attributes.is_disconnected || realm.is_sovereign_realm) && ! contains(keys(local.build_region_by_name), realm.attributes.first_region) }
  prod_realm_by_name         = { for realm in local.shepherd_all_realms : realm.name => realm if ! contains(local.blacklist_realms, realm.name) && ! (realm.attributes.is_disconnected || realm.is_sovereign_realm) && ! contains(keys(local.build_region_by_name), realm.attributes.first_region) }
  build_region_by_name       = { for region in local.shepherd_all_regions : region.name => region if ! contains(local.blacklist_realms, region.realm) && (region.state == "Building") && region.airport_code != "DCA" && region.airport_code != "QDF" && ! contains(local.blacklist_realms, region.realm) }
  build_realm_by_name        = { for realm in local.shepherd_all_realms : realm.name => merge(realm, { airport_code : lower(lookup(local.build_region_by_name, realm.attributes.first_region).airport_code) }) if ! contains(local.blacklist_realms, realm.name) && lookup(local.prod_realm_by_name, realm.name, {}) == {} && lookup(local.onsr_realm_by_name, realm.name, {}) == {} && lookup(local.build_region_by_name, realm.attributes.first_region, {}) != {} }
  onsr_phases                = [for realm in local.shepherd_all_realms : realm.name if ! contains(local.blacklist_realms, realm.name) && (realm.attributes.is_disconnected || realm.is_sovereign_realm) && ! contains(keys(local.build_region_by_name), realm.attributes.first_region)]

  // Overrides for environment, realm or region level
  overrides = {
    // Sample key: dev
    // Sample key: dev.oc1
    // Sample key: dev.oc1.us-ashburn-1
    "herds" = {
      env              = "rbaas"
      oke_tenancy_ocid = "TODO"

      // Configuration for PKI certificate
      mapi_subdomain          = "mapi-rbaas"
      certificate_compartment = "" // Defaults to orchestration compartment

      // Spectre Configuration
      spectre_group_name = "clusters_test"

      // KaaS Configuration
      kaas_name_format = "oke-rbaas-mp-cell%d"

      // WFaaS Configuration
      wfaas_name_format = "oke-rbaas-mp-cell%d"

      // API instance configuration
      api_hostclass = "oke-mp-api-rbaas"

      // Worker instance configuration
      worker_hostclass = "oke-mp-worker-rbaas"

      // Logging configuration
      api_log_namespace_format     = "oke-rbaas-api-cell%d"
      monitor_log_namespace_format = "oke-rbaas-monitor-cell%d"
      worker_log_namespace_format  = "oke-rbaas-worker-cell%d"

      // Generic OKE configuration
      service_name = "okepldev"

      // SMS Configuration
      sms_namespace_name_format = "oke-rbaas-mapi-cell%d"
      oke_secrets_namespace     = "oke-rbaas0"

      // Configuration about other OKE components
      etcdop_s3compat_bucket = "tkc-etcd-backup-rbaas-0"
      os_namespace           = "ax8rjegmraam"

      // Alarms configuration
      alarms_compartment              = "" // Defaults to orchestration compartment
      alarms_enabled                  = false
      mapi_api_alarms_enabled         = false
      mapi_alarms_fleet_format        = "okerbaas.oke-rbaas-mapi-cell%d"
      mapi_alarms_hostmetrics_fleet   = "okerbaas.oke-mp-api-rbaas"
      kmon_alarms_enabled             = false
      kmon_alarms_fleet_format        = "okerbaas.oke-rbaas-kmon-cell%d"
      worker_alarms_enabled           = false
      worker_alarms_fleet_format      = "okerbaas.oke-rbaas-wfworker-cell%d"
      worker_alarms_hostmetrics_fleet = "okerbaas.oke-mp-worker-rbaas"

      // API ODO Configuration
      api_pool_alias_format = "oke-rbaas-mapi-cell%d"
      api_app_alias_format  = "oke-rbaas-mapi-cell%d"

      // Monitor ODO Configuration
      monitor_app_alias_format = "oke-rbaas-kmon-cell%d"

      // Worker ODO Configuration
      worker_pool_alias_format = "oke-rbaas-mpworker-cell%d"
      worker_app_alias_format  = "oke-rbaas-wfworker-cell%d"

      // SPLAT - base service already exists because this is still part of oc1
      splat_service_name_format           = "oke-mapi-cell%d-rbaas"
      splat_operational_spec_fleet_format = "oke-rbaas-mapi-cell%d"
      splat_host_header_format            = "oke-rbaas-mapi-cell%d.%s.oci.%s"
      // The following hardcoded compartment name in api.yaml is replaced with
      // orchestration compartment OCID for non-production environments
      splat_compartment_token_to_replace  = "ocid1.compartment.oc1..aaaaaaaamhk5oakipanjsuf3g6xrejzsa5e2tgsgugngtr7bo67wygrffkoq"
      splat_mapi_subdomain_format         = "oke-rbaas-mapi-cell%d"
      splat_allowed_service_principal     = "okepldev"

      // Configuration for image
      image_name = local.io_overlay_uek5_images["20210409"].name
      image_url  = local.io_overlay_uek5_images["20210409"].url

      // T2
      t2_namespace      = "oci_oke_rbaas"
      t2_fleet_template = "okerbaas.rbaas-oke-clusters-%s"

      cp_vcn_compartment = "oke-cp-api"
      pl_vcn_name        = "oke-admin"
      pl_vcn_compartment = "admin"

      prime_vcn_compartment = "rbaas-0"
      prime_vcn_name        = "rbaas"

      skip_dns = true
    }
    "polaris" = {
      env              = "polaris"
      oke_tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaxc346rx7oshe74elt6upirg54l62b4iaxalrkc44hqpv4nxq333a"

      // Configuration for PKI certificate
      mapi_subdomain          = "mapi-polaris"
      certificate_compartment = "" // Defaults to orchestration compartment

      // Spectre Configuration
      spectre_group_name = "clusters-polaris"

      // KaaS Configuration
      kaas_name_format = "oke-polaris-mp-cell%d"

      // WFaaS Configuration
      wfaas_name_format = "oke-polaris-mp-cell%d"

      // API instance configuration
      api_hostclass = "oke-mp-api-polaris"

      // Worker instance configuration
      worker_hostclass = "oke-mp-worker-polaris"

      // Logging configuration
      api_log_namespace_format     = "oke-polaris-api-cell%d"
      monitor_log_namespace_format = "oke-polaris-monitor-cell%d"
      worker_log_namespace_format  = "oke-polaris-worker-cell%d"

      // Generic OKE configuration
      service_name = "okepldev"

      // SMS Configuration
      sms_namespace_name_format = "oke-polaris-mapi-cell%d"
      oke_secrets_namespace     = "oke-polaris0"

      // Configuration about other OKE components
      etcdop_s3compat_bucket = "tkc-etcd-backup-polaris-0"
      os_namespace           = "ax8rjegmraam"

      // Alarms configuration
      alarms_compartment              = "" // Defaults to orchestration compartment
      alarms_enabled                  = false
      mapi_api_alarms_enabled         = false
      mapi_alarms_fleet_format        = "okeplatformdev.oke-polaris-mapi-cell%d"
      mapi_alarms_hostmetrics_fleet   = "okeplatformdev.oke-mp-api-polaris"
      kmon_alarms_enabled             = false
      kmon_alarms_fleet_format        = "okeplatformdev.oke-polaris-kmon-cell%d"
      worker_alarms_enabled           = false
      worker_alarms_fleet_format      = "okeplatformdev.oke-polaris-wfworker-cell%d"
      worker_alarms_hostmetrics_fleet = "okeplatformdev.oke-mp-worker-polaris"

      // API ODO Configuration
      api_pool_alias_format = "oke-polaris-mapi-cell%d"
      api_app_alias_format  = "oke-polaris-mapi-cell%d"

      // Monitor ODO Configuration
      monitor_app_alias_format = "oke-polaris-kmon-cell%d"

      // Worker ODO Configuration
      worker_pool_alias_format = "oke-polaris-mpworker-cell%d"
      worker_app_alias_format  = "oke-polaris-wfworker-cell%d"

      // SPLAT - base service already exists because this is still part of oc1
      splat_service_name_format           = "oke-mapi-cell%d-polaris"
      splat_operational_spec_fleet_format = "oke-polaris-mapi-cell%d"
      splat_host_header_format            = "oke-polaris-mapi-cell%d.%s.oci.%s"
      // The following hardcoded compartment name in api.yaml is replaced with
      // orchestration compartment OCID for non-production environments
      splat_compartment_token_to_replace  = "ocid1.compartment.oc1..aaaaaaaamhk5oakipanjsuf3g6xrejzsa5e2tgsgugngtr7bo67wygrffkoq"
      splat_mapi_subdomain_format         = "oke-polaris-mapi-cell%d"
      splat_allowed_service_principal     = "okepldev"

      // Configuration for image
      image_name = local.io_overlay_uek5_images["20210409"].name
      image_url  = local.io_overlay_uek5_images["20210409"].url

      // T2
      t2_namespace      = "oci_oke_polaris"
      t2_fleet_template = "okeplatformdev.polaris-oke-clusters-%s"

      cp_vcn_compartment = "oke-cp-api"
      pl_vcn_name        = "oke-admin"
      pl_vcn_compartment = "admin"

      prime_vcn_compartment = "polaris-0"
      prime_vcn_name        = "polaris"

      skip_dns = true
    }

    "dev" = {
      env              = "dev"
      oke_tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaal2r2bftstipxui5mx2sxr4tc5snnr5ajojgwxoj3xojtyghqrx4a"

      shorten_uniquifier = "false"

      // Configuration for PKI certificate
      mapi_subdomain          = "mapi-dev"
      certificate_compartment = "" // Defaults to orchestration compartment

      // Spectre Configuration
      spectre_group_name = "clusters-dev"

      // KaaS Configuration
      kaas_name_format = "oke-dev-mp-cell%d"

      // WFaaS Configuration
      wfaas_name_format = "oke-dev-mp-cell%d"

      // API instance configuration
      api_hostclass = "oke-mp-api-dev"

      // Worker instance configuration
      worker_hostclass = "oke-mp-worker-dev"

      // Logging configuration
      api_log_namespace_format     = "oke-dev-api-cell%d"
      monitor_log_namespace_format = "oke-dev-monitor-cell%d"
      worker_log_namespace_format  = "oke-dev-worker-cell%d"

      // Generic OKE configuration
      service_name = "okedev"

      // SMS Configuration
      sms_namespace_name_format = "oke-dev-mapi-cell%d"
      oke_secrets_namespace     = "oke-dev0"

      // Configuration about other OKE components
      etcdop_s3compat_bucket = "tkc-etcd-backup-dev-0"
      os_namespace           = "ociokedev"

      // Alarms configuration
      alarms_compartment                = "" // Defaults to orchestration compartment
      alarms_enabled                    = false
      mapi_api_alarms_enabled           = false
      mapi_alarms_fleet_format          = "ociokedev.oke-dev-mapi-cell%d"
      mapi_alarms_hostmetrics_fleet     = "ociokedev.oke-mp-api-dev"
      kmon_alarms_enabled               = false
      kmon_alarms_fleet_format          = "ociokedev.oke-dev-kmon-cell%d"
      worker_alarms_enabled             = false
      worker_alarms_fleet_format        = "ociokedev.oke-dev-wfworker-cell%d"
      worker_alarms_hostmetrics_fleet   = "ociokedev.oke-mp-worker-dev"
      ccm_alarms_fleet_format           = "ociokedev.oke-kmi-cell%d"
      csi_dataplane_alarms_fleet_format = "dev-oke-clusters-%s-dataplane%s"

      // API ODO Configuration
      api_pool_alias_format = "oke-dev-mapi-cell%d"
      api_app_alias_format  = "oke-dev-mapi-cell%d"

      // Monitor ODO Configuration
      monitor_app_alias_format = "oke-dev-kmon-cell%d"

      // Worker ODO Configuration
      worker_pool_alias_format = "oke-dev-mpworker-cell%d"
      worker_app_alias_format  = "oke-dev-wfworker-cell%d"

      // SPLAT
      splat_service_name_format           = "oke-mapi-cell%d-dev"
      splat_service_fleet                 = "overlay-dev-fleet"
      splat_operational_spec_fleet_format = "oke-dev-mapi-cell%d"
      splat_host_header_format            = "oke-dev-mapi-cell%d.%s.oci.%s"
      splat_public_domain_override        = "oc-test.com"
      // The following hardcoded compartment name in api.yaml is replaced with
      // orchestration compartment OCID for non-production environments
      splat_compartment_token_to_replace  = "ocid1.compartment.oc1..aaaaaaaa25nt3xxdztunf4mtccovffzqe4mjtyjoerraxrlf7wbjrtu4crxq"
      splat_mapi_subdomain_format         = "oke-dev-mapi-cell%d"
      splat_allowed_service_principal     = "okedev"

      // Configuration for image
      image_name = local.io_overlay_uek5_images["20201014"].name
      image_url  = local.io_overlay_uek5_images["20201014"].url

      // T2
      t2_namespace      = "oci_oke_dev"
      t2_fleet_template = "ociokedev.dev-oke-clusters-%s"

      cp_vcn_compartment = "oke-cp-api"
      pl_vcn_name        = "admin"
      pl_vcn_compartment = "admin"

      prime_vcn_compartment = "dev-0"
      prime_vcn_name        = "dev"
      skip_dns              = true


      canary_tenancies      = jsonencode(["oke-canary", "okecanaryla"]),
      integration_tenancies = jsonencode(["odx-mockcustomer"])


    }

    "integ" = {
      env              = "integ"
      oke_tenancy_ocid = "ocid1.tenancy.oc1..aaaaaaaaiconelbquhzjhrwgkkbubef3u3b3rezwhtcjmf6o3gwr6c3vewfq"

      service_name       = "okeinteg"
      spectre_group_name = "clusters-integ"

      // Configuration for PKI certificate
      certificate_compartment = "" // Defaults to orchestration compartment
      mapi_subdomain          = "mapi-integ"

      // Spectre properties' configuration
      enable_spectre = false // Reuse the group from dev

      // SMS Configuration
      sms_namespace_name_format = "oke-integ-mapi-cell%d"
      oke_secrets_namespace     = "oke-integ0"

      // KaaS Configuration
      kaas_name_format = "oke-integ-mp-cell%d"

      // WFaaS Configuration
      wfaas_name_format = "oke-integ-mp-cell%d"

      // MAPI and workers instances' configuration
      api_hostclass    = "oke-mp-api-integ"
      worker_hostclass = "oke-mp-worker-integ"

      // Logging Configuration
      api_log_namespace_format     = "oke-integ-api-cell%d"
      monitor_log_namespace_format = "oke-integ-monitor-cell%d"
      worker_log_namespace_format  = "oke-integ-worker-cell%d"

      // Integration with other OKE components
      etcdop_s3compat_bucket = "tkc-etcd-backup-int-0"
      os_namespace           = "idqknvtm9e3k"

      // Alarms Configuration
      severity_2                        = 4
      severity_3                        = 4
      mapi_availability_severity        = 4
      mapi_latency_severity             = 4
      alarms_compartment                = "" // Defaults to orchestration compartment
      alarms_enabled                    = true
      mapi_api_alarms_enabled           = true
      mapi_alarms_fleet_format          = "ociokeinteg.oke-integ-mapi-cell%d"
      mapi_alarms_hostmetrics_fleet     = "ociokeinteg.oke-mp-api-integ"
      kmon_alarms_enabled               = true
      kmon_alarms_fleet_format          = "ociokeinteg.oke-integ-kmon-cell%d"
      worker_alarms_enabled             = true
      worker_alarms_fleet_format        = "ociokeinteg.oke-integ-wfworker-cell%d"
      worker_alarms_hostmetrics_fleet   = "ociokeinteg.oke-mp-worker-integ"
      ccm_alarms_fleet_format           = "ociokeinteg.oke-kmi-cell%d"
      csi_dataplane_alarms_fleet_format = "integ-oke-clusters-%s-dataplane%s"

      // ODO Configuration
      api_pool_alias_format    = "oke-integ-mapi-cell%d"
      api_app_alias_format     = "oke-integ-mapi-cell%d"
      worker_pool_alias_format = "oke-integ-mpworker-cell%d"
      monitor_app_alias_format = "oke-integ-kmon-cell%d"
      worker_app_alias_format  = "oke-integ-wfworker-cell%d"

      // SPLAT
      splat_service_name_format           = "oke-mapi-cell%d-integ"
      splat_operational_spec_fleet_format = "oke-integ-mapi-cell%d"
      splat_host_header_format            = "oke-integ-mapi-cell%d.%s.oci.%s"
      // The following hardcoded compartment name in api.yaml is replaced with
      // orchestration compartment OCID for non-production environments
      splat_compartment_token_to_replace  = "ocid1.compartment.oc1..aaaaaaaa25nt3xxdztunf4mtccovffzqe4mjtyjoerraxrlf7wbjrtu4crxq"
      splat_mapi_subdomain_format         = "oke-integ-mapi-cell%d"
      splat_allowed_service_principal     = "okeinteg"
      splat_allowed_tenancies             = "boat,ocid1.tenancy.oc1..aaaaaaaat37ab62ltpvzgyoydasbfig3gcmccxwzvbi6yoh6ewqiiswps6sq"
      canary_vcn                          = "ocid1.vcn.oc1.iad.amaaaaaa2qd24aiaiy4a4kyv7x4ilzoamlqovopehhomli23oauzsox4sbsa"
      canary_cidr                         = "10.0.0.0/26"

      // Configuration for image
      image_name = local.io_overlay_uek5_images["20201116"].name
      image_url  = local.io_overlay_uek5_images["20201116"].url

      cp_vcn_compartment = "oke-cp-api"
      pl_vcn_name        = "admin"
      pl_vcn_compartment = "admin"

      prime_vcn_compartment = "int-0"
      prime_vcn_name        = "integ"

      t2_fleet_template = "integ-oke-clusters-%s"
      skip_dns          = true

      canary_tenancies      = jsonencode(["oke-canary", "okecanaryla"]),
      integration_tenancies = jsonencode(["odx-mockcustomer"])
    }
    // For tenancy OCID
    tenancy_info = {
      rbaas = {
        oc1 = "okerbaas"
      }
      dev = {
        oc1 = "ociokedev"
      }
      integ = {
        oc1 = "ociokeinteg"
      }
      polaris = {
        oc1 = "okeplatformdev"
        oc5 = "oke-prod-v2"
      }
      prd = {
        oc1 = "odx-oke"
        oc2 = "okeprodoc2"
        oc3 = "okeprodoc3"
        oc4 = "oke-prod-oc4"
      }
      default = "oke-prod"
    }

    "prd" = {
      env          = "prd"
      service_name = "oke"

      // Generic ODO Configuration
      odo_app_type = "PRODUCTION"

      // Configuration for MAPI and worker instances
      api_hostclass    = "oke-mp-api-prod"
      worker_hostclass = "oke-mp-worker-prod"

      // Logging configuration
      api_log_namespace_format     = "oke-api-cell%d"
      monitor_log_namespace_format = "oke-monitor-cell%d"
      worker_log_namespace_format  = "oke-worker-cell%d"
      // SMS Configuration
      oke_secrets_namespace        = "oke-prod0"

      // Alarms configuration
      alarms_enabled                    = true
      mapi_api_alarms_enabled           = true
      mapi_alarms_fleet_format          = "oke-mapi-cell%d"
      mapi_alarms_hostmetrics_fleet     = "oke-mp-api-prod"
      kmon_alarms_enabled               = true
      kmon_alarms_fleet_format          = "oke-kmon-cell%d"
      worker_alarms_enabled             = true
      worker_alarms_fleet_format        = "oke-wfworker-cell%d"
      ccm_alarms_enabled                = true
      csi_alarms_enabled                = true
      ccm_alarms_fleet_format           = "oke-kmi-cell%d"
      csi_dataplane_alarms_fleet_format = "prod-oke-clusters-%s-dataplane%s"
      worker_alarms_hostmetrics_fleet   = "oke-mp-worker-prod"

      // Integration with other OKE components
      etcdop_s3compat_bucket = "tkc-etcd-backup-prd-0"

      // SPLAT
      splat_base_service_name             = "management-plane-api"
      splat_service_name_format           = "oke-mapi-cell%d"
      splat_operational_spec_fleet_format = "oke-mapi-cell%d"
      splat_host_header_format            = "oke-mapi-cell%d.%s.oci.%s"

      cp_vcn_compartment = "oke-cp-api"
      pl_vcn_name        = "admin"
      pl_vcn_compartment = "admin"

      prime_vcn_compartment = "prd-0"
      prime_vcn_name        = "prod"

      t2_fleet_template = "prod-oke-clusters-%s"
    }

    "polaris.oc5" = {
      oke_tenancy_ocid        = "ocid1.tenancy.oc5..aaaaaaaaf7y2p6foiobqpwhjmg6niejayxb56uj7edhox6indmqrhidvx44q"
      splat_base_service_name = "management-plane-api"

      // Configuration about other OKE components
      os_namespace = "ax7fn3khatef"

      // Alarms
      alarms_enabled                = false
      skip_mapi_alarms              = true
      mapi_alarms_fleet_format      = "oke-prod-v2.oke-polaris-mapi-cell%d"
      mapi_alarms_hostmetrics_fleet = "oke-prod-v2.oke-mp-api-polaris"

      skip_kmon_alarms                = true
      skip_worker_alarms              = true
      kmon_alarms_fleet_format        = "oke-prod-v2.oke-polaris-kmon-cell%d"
      worker_alarms_fleet_format      = "oke-prod-v2.oke-polaris-wfworker-cell%d"
      worker_alarms_hostmetrics_fleet = "oke-prod-v2.oke-mp-worker-polaris"

      splat_compartment_token_to_replace = "unused-remove"

      // T2
      t2_fleet_template     = "oke-prod-v2.polaris-oke-clusters-%s"
      mapi_instance_shape   = "VM.Standard2.2"
      worker_instance_shape = "VM.Standard2.4"
    }

    "herds.oc1" = {
      has_mapi_grafana_dashboard = false
      image_type                 = "E446"
      worker_instance_count      = 3
    }

    "dev.oc1" = {
      has_mapi_grafana_dashboard = false
      image_type                 = "E446"
      worker_instance_count      = 3
    }

    "prd.oc1" = {
      // Configuration for image
      image_name = local.io_overlay_uek5_images["20201116"].name
      image_url  = local.io_overlay_uek5_images["20201116"].url
    }

    "integ.oc1.us-ashburn-1" = {
      // Configuration for worker instances in any cell
      worker_instance_count = 9
      image_type            = "E446"
      cell_count            = 2
    }
    "integ.oc1.us-phoenix-1" = {
      worker_instance_count = 9
      image_type            = "E446"
    }

    "prd.oc1.ap-melbourne-1" = {
      phase      = length(shepherd_release_phase.prd-oc1-bake-1) == 1 ? shepherd_release_phase.prd-oc1-bake-1[0].name : null
      image_type = "E446"
    }

    "prd.oc1.ca-montreal-1" = {
      phase      = length(shepherd_release_phase.prd-oc1-single-ad-part-1) == 1 ? shepherd_release_phase.prd-oc1-single-ad-part-1[0].name : null
      image_type = "E446"
    }
    "prd.oc1.ap-mumbai-1" = {
      phase      = length(shepherd_release_phase.prd-oc1-single-ad-part-1) == 1 ? shepherd_release_phase.prd-oc1-single-ad-part-1[0].name : null
      image_type = "E446"
    }
    "prd.oc1.ap-chuncheon-1" = {
      phase      = length(shepherd_release_phase.prd-oc1-single-ad-part-1) == 1 ? shepherd_release_phase.prd-oc1-single-ad-part-1[0].name : null
      image_type = "E446"
    }
    "prd.oc1.ap-hyderabad-1" = {
      phase      = length(shepherd_release_phase.prd-oc1-single-ad-part-2) == 1 ? shepherd_release_phase.prd-oc1-single-ad-part-2[0].name : null
      image_type = "E446"
    }
    "prd.oc1.eu-zurich-1" = {
      phase      = length(shepherd_release_phase.prd-oc1-single-ad-part-1) == 1 ? shepherd_release_phase.prd-oc1-single-ad-part-1[0].name : null
      image_type = "E446"
    }
    "prd.oc1.ap-osaka-1" = {
      phase      = length(shepherd_release_phase.prd-oc1-single-ad-part-2) == 1 ? shepherd_release_phase.prd-oc1-single-ad-part-2[0].name : null
      image_type = "E446"
    }
    "prd.oc1.eu-amsterdam-1" = {
      phase      = length(shepherd_release_phase.prd-oc1-single-ad-part-1) == 1 ? shepherd_release_phase.prd-oc1-single-ad-part-1[0].name : null
      image_type = "E446"
    }
    "prd.oc1.ap-tokyo-1" = {
      phase      = length(shepherd_release_phase.prd-oc1-single-ad-part-1) == 1 ? shepherd_release_phase.prd-oc1-single-ad-part-1[0].name : null
      image_type = "E446"
    }
    "prd.oc1.ap-seoul-1" = {
      phase      = length(shepherd_release_phase.prd-oc1-single-ad-part-2) == 1 ? shepherd_release_phase.prd-oc1-single-ad-part-2[0].name : null
      image_type = "E446"
    }
    "prd.oc1.ap-sydney-1" = {
      phase      = length(shepherd_release_phase.prd-oc1-single-ad-part-1) == 1 ? shepherd_release_phase.prd-oc1-single-ad-part-1[0].name : null
      image_type = "E446"
    }
    "prd.oc1.ca-toronto-1" = {
      phase      = length(shepherd_release_phase.prd-oc1-single-ad-part-2) == 1 ? shepherd_release_phase.prd-oc1-single-ad-part-2[0].name : null
      image_type = "E446"
    }
    "prd.oc1.sa-saopaulo-1" = {
      phase      = length(shepherd_release_phase.prd-oc1-single-ad-part-2) == 1 ? shepherd_release_phase.prd-oc1-single-ad-part-2[0].name : null
      image_type = "E446"
    }
    "prd.oc1.me-jeddah-1" = {
      phase      = length(shepherd_release_phase.prd-oc1-single-ad-part-2) == 1 ? shepherd_release_phase.prd-oc1-single-ad-part-2[0].name : null
      image_type = "E446"
    }

    "prd.oc1.us-ashburn-1" = {
      // Configuration for worker instances in any cell
      // we need to add cell_count here when we have more than one cell in a region
      // example: cell_count = 10
      cell_count            = 2
      skip_dns              = true
      worker_instance_count = 9
      phase                 = length(shepherd_release_phase.prd-oc1-bake-3) == 1 ? shepherd_release_phase.prd-oc1-bake-3[0].name : null
    }
    "prd.oc1.us-phoenix-1" = {
      // Configuration for worker instances in any cell
      worker_instance_count = 9
      phase                 = length(shepherd_release_phase.prd-oc1-phx) == 1 ? shepherd_release_phase.prd-oc1-phx[0].name : null
    }
    "prd.oc1.uk-london-1" = {
      // Configuration for worker instances in any cell
      worker_instance_count = 9
      phase                 = length(shepherd_release_phase.prd-oc1-bake-2) == 1 ? shepherd_release_phase.prd-oc1-bake-2[0].name : null
      image_type            = "E446"
    }
    "prd.oc1.eu-frankfurt-1" = {
      // Configuration for worker instances in any cell
      worker_instance_count = 9
      phase                 = length(shepherd_release_phase.prd-oc1-fra) == 1 ? shepherd_release_phase.prd-oc1-fra[0].name : null
      image_type            = "E446"
    }

    "prd.oc1.ap-chuncheon-1" = {
      // Configuration for image
      image_name = local.selinux_overlay_uek5_images["20201116"].name
      image_url  = local.selinux_overlay_uek5_images["20201116"].url
      phase      = length(shepherd_release_phase.prd-oc1-single-ad-part-1) == 1 ? shepherd_release_phase.prd-oc1-single-ad-part-1[0].name : null
    }

    "prd.oc1.us-sanjose-1" = {
      // Configuration for image
      image_name = local.selinux_overlay_uek5_images["20201116"].name
      image_url  = local.selinux_overlay_uek5_images["20201116"].url
      phase      = length(shepherd_release_phase.prd-oc1-single-ad-part-1) == 1 ? shepherd_release_phase.prd-oc1-single-ad-part-1[0].name : null
    }

    "prd.oc1.me-dubai-1" = {
      // Configuration for image
      image_name = local.selinux_overlay_uek5_images["20201116"].name
      image_url  = local.selinux_overlay_uek5_images["20201116"].url
    }

    "prd.oc1.uk-cardiff-1" = {
      // Configuration for image
      image_name = local.selinux_overlay_uek5_images["20201116"].name
      image_url  = local.selinux_overlay_uek5_images["20201116"].url
      phase      = length(shepherd_release_phase.prd-oc1-single-ad-part-2) == 1 ? shepherd_release_phase.prd-oc1-single-ad-part-2[0].name : null
    }

    "prd.oc1.sa-santiago-1" = {
      // Configuration for image
      image_name = local.selinux_overlay_uek5_images["20201116"].name
      image_url  = local.selinux_overlay_uek5_images["20201116"].url
      phase      = length(shepherd_release_phase.prd-oc1-single-ad-part-2) == 1 ? shepherd_release_phase.prd-oc1-single-ad-part-2[0].name : null
    }

    "prd.oc1.sa-vinhedo-1" = {
      image_name = local.io_overlay_uek5_images["20210409"].name
      image_url  = local.io_overlay_uek5_images["20210409"].url
    }

    "polaris.oc1.us-sanjose-1" = {
      cell_count            = 2
      mapi_instance_count   = 2
      worker_instance_count = 2
    }

    "prd.oc2" = {
      splunk_enabled              = true
      // Configuration for image
      image_name                  = local.selinux_overlay_uek5_images["20201116"].name
      image_url                   = local.selinux_overlay_uek5_images["20201116"].url
      // Configuration for policies and dynamic groups
      cluster_secrets_compartment = ""
      // This override to no compartment can be removed once the compartment is provisioned in this realm
    }

    "prd.oc3" = {
      splunk_enabled              = true
      // Configuration for image
      image_name                  = local.selinux_overlay_uek5_images["20201116"].name
      image_url                   = local.selinux_overlay_uek5_images["20201116"].url
      // Configuration for policies and dynamic groups
      cluster_secrets_compartment = ""
      // This override to no compartment can be removed once the compartment is provisioned in this realm
    }

    "prd.oc4" = {
      // Alarms configuration
      // disable alarm while we build v2 in the realm
      alarms_enabled          = false
      mapi_api_alarms_enabled = false
      kmon_alarms_enabled     = false
      worker_alarms_enabled   = false
      mapi_instance_shape     = "VM.Standard2.2"
      worker_instance_shape   = "VM.Standard2.4"
    }

    "prd.oc4.uk-gov-london-1" = {
      // Configuration for image
      // FIXME: Aseem Bajaj: At the time of provisioning of this cell, fill in the following values from local.io_overlay_uek5_images
      image_name = ""
      image_url  = ""
      image_type = "E446"
    }

    "prd.oc4.uk-gov-cardiff-1" = {
      // Configuration for image
      // FIXME: Aseem Bajaj: At the time of provisioning of this cell, fill in the following values from local.selinux_overlay_uek5_images
      image_name = ""
      image_url  = ""
    }

    "prd.oc5" = {
      splunk_enabled = true
      // Configuration for image
      // FIXME: Aseem Bajaj: At the time of provisioning of this cell, fill in the following values from local.selinux_overlay_uek5_images
      image_name     = ""
      image_url      = ""
    }

    "prd.oc6" = {
      splunk_enabled = true
      // Configuration for image
      // FIXME: Aseem Bajaj: At the time of provisioning of this cell, fill in the following values from local.selinux_overlay_uek5_images
      image_name     = ""
      image_url      = ""
    }

    "prd.oc8" = {
      // Configuration for image
      image_name                  = local.selinux_overlay_uek5_images["20201116"].name
      image_url                   = local.selinux_overlay_uek5_images["20201116"].url
      // Configuration for policies and dynamic groups
      cluster_secrets_compartment = ""
      // This override to no compartment can be removed once the compartment is provisioned in this realm
    }
  }

  defined_cell_overrides = {
    "herds.oc1.eu-frankfurt-1.cell0" = {}
    "polaris.oc1.us-sanjose-1.cell0" = {}
    "dev.oc1.us-ashburn-1.cell0"     = {}
    "dev.oc1.eu-frankfurt-1.cell0"   = {}
    "dev.oc1.us-phoenix-1.cell0"     = {
      predecessor = "dev.oc1.us-ashburn-1.cell0"
      image_name  = local.io_overlay_uek5_images["20210111"].name
      image_url   = local.io_overlay_uek5_images["20210111"].url
    }
    "integ.oc1.us-ashburn-1.cell0"   = {}
    "integ.oc1.us-ashburn-1.cell1"   = {
      predecessor                   = "integ.oc1.us-ashburn-1.cell0"
      enable_kaas_regional_instance = false
      oke_secrets_namespace         = "oke-prime-integ-cell1"
    }
    "integ.oc1.us-phoenix-1.cell0" = {
      predecessor = "integ.oc1.us-ashburn-1.cell0"
      image_name  = local.io_overlay_uek5_images["20210111"].name
      image_url   = local.io_overlay_uek5_images["20210111"].url
    }
    "integ.oc1.eu-frankfurt-1.cell0" = {}
    "polaris.oc1.us-sanjose-1.cell1" = {
      oke_secrets_namespace = "oke-prime-polaris-cell1"
    },
    "polaris.oc5.us-tacoma-1.cell0" = {}
    "prd.oc1.ap-melbourne-1.cell0"  = {
      predecessor = "env.setup.prd.oc1"
      skip_dns    = true
    }
    "prd.oc1.uk-london-1.cell0" = {
      skip_dns = true
    }
    "prd.oc1.us-ashburn-1.cell0" = {
      worker_instance_ocpus   = 8
      worker_instance_mem_gbs = 120
    }
    "prd.oc1.us-phoenix-1.cell0" = {
      skip_dns = true
    }
    "prd.oc1.eu-frankfurt-1.cell0" = {
      skip_dns = true
    }
    "prd.oc1.us-sanjose-1.cell0" = {
      skip_dns = true
    }
    "prd.oc1.ca-montreal-1.cell0" = {
      skip_dns = true
    }
    "prd.oc1.ap-mumbai-1.cell0" = {
      skip_dns = true
    }
    "prd.oc1.ap-chuncheon-1.cell0" = {
      skip_dns = true
    }
    "prd.oc1.ap-hyderabad-1.cell0" = {
      skip_dns = true
    }
    "prd.oc1.eu-zurich-1.cell0" = {
      skip_dns = true
    }
    "prd.oc1.ap-osaka-1.cell0" = {
      skip_dns = true
    }
    "prd.oc1.eu-amsterdam-1.cell0" = {
      skip_dns = true
    }
    "prd.oc1.ap-tokyo-1.cell0" = {
      skip_dns = true
    }
    "prd.oc1.ap-seoul-1.cell0" = {
      skip_dns = true
    }
    "prd.oc1.ap-sydney-1.cell0" = {
      skip_dns = true
    }
    "prd.oc1.ca-toronto-1.cell0" = {
      skip_dns = true
    }
    "prd.oc1.sa-saopaulo-1.cell0" = {
      skip_dns = true
    }
    "prd.oc1.me-jeddah-1.cell0" = {
      skip_dns = true
    }
    "prd.oc1.me-dubai-1.cell0" = {
      skip_dns = true
    }
    "prd.oc1.uk-cardiff-1.cell0" = {
      skip_dns = true
    }
    "prd.oc1.sa-santiago-1.cell0" = {
      skip_dns = true
    }
    "prd.oc1.sa-vinhedo-1.cell0" = {
      skip_dns = true
    }
    "prd.oc2.us-langley-1.cell0" = {
      predecessor = "env.setup.prd.oc2"
      skip_dns    = true
    }
    "prd.oc2.us-luke-1.cell0" = {
      predecessor = "prd.oc2.us-langley-1.cell0"
      skip_dns    = true
    }
    "prd.oc3.us-gov-ashburn-1.cell0" = {
      predecessor = "env.setup.prd.oc3"
      skip_dns    = true
    }
    "prd.oc3.us-gov-phoenix-1.cell0" = {
      predecessor = "prd.oc3.us-gov-ashburn-1.cell0"
      skip_dns    = true
    }
    "prd.oc3.us-gov-chicago-1.cell0" = {
      predecessor = "prd.oc3.us-gov-phoenix-1.cell0"
      skip_dns    = true
    }
    "prd.oc4.uk-gov-london-1.cell0" = {
      predecessor = "env.setup.prd.oc4"
      skip_dns    = true
    }
    "prd.oc4.uk-gov-cardiff-1.cell0" = {
      predecessor = "prd.oc4.uk-gov-london-1.cell0"
      skip_dns    = true
    }
    "prd.oc5.us-tacoma-1.cell0" = {
      predecessor = "env.setup.prd.oc5"
    }
    "prd.oc6.us-gov-fortworth-1.cell0" = {
      predecessor                = "env.setup.prd.oc6"
      has_mapi_grafana_dashboard = true
    }
    "prd.oc6.us-gov-sterling-2.cell0" = {
      predecessor = "prd.oc6.us-gov-fortworth-1.cell0"
    }
    "prd.oc8.ap-chiyoda-1.cell0" = {
      predecessor = "env.setup.prd.oc8"
      skip_dns    = true
    }
    "prd.oc1.us-ashburn-1.cell1" = {
      predecessor           = "prd.oc1.us-ashburn-1.cell0"
      oke_secrets_namespace = "oke-prime-prod-cell1"
    }
    "prd.oc1.us-phoenix-1.cell1" = {
      predecessor           = "prd.oc1.us-phoenix-1.cell0"
      oke_secrets_namespace = "oke-prime-prod-cell1"
    }
    "prd.oc1.eu-frankfurt-1.cell1" = {
      predecessor           = "prd.oc1.eu-frankfurt-1.cell0"
      oke_secrets_namespace = "oke-prime-prod-cell1"
    }
    "prd.oc8.ap-ibaraki-1.cell0" = {
      alarms_enabled          = true
      mapi_api_alarms_enabled = true
      kmon_alarms_enabled     = true
      worker_alarms_enabled   = true
    }
  }

  build_regions = flatten([
  for region in local.regions_under_build :
  format("prd.%s.%s.cell0", lookup(local.build_region_to_realm, region.public_name, "oc1"), region.public_name)
  ])

  production_regions = flatten([
  for cell, region in local.shepherd_current_regions :
  [
  for cell in range(merge({
    cell_count = 1
  }, lookup(local.overrides, format("prd.%s.%s", lookup(local.region_to_realm, region.public_name, "oc1"), region.public_name), {})).cell_count) :
  format("prd.%s.%s.cell%s", lookup(local.region_to_realm, region.public_name, "oc1"), region.public_name, cell)
  ]
  ])

  polaris_regions = flatten([
  for region in ["us-sanjose-1", "us-phoenix-1"] :
  [
  for cell in range(merge({
    cell_count = 1
  }, lookup(local.overrides, format("polaris.%s.%s", lookup(local.region_to_realm, region, "oc1"), "us-sanjose-1"), {})).cell_count) :
  format("polaris.%s.%s.cell%s", lookup(local.region_to_realm, region, "oc1"), region, cell)
  ]
  ])

  cell_overrides = merge(
    //     Sample key: dev.oc1.us-ashburn-1.cell0
    //     Unlike overrides above, the declaration of cell name is required below
    //     This allows creation of relevant ETs
    local.defined_cell_overrides,
    {
    for region in local.production_regions :
    region => {} if lookup(local.defined_cell_overrides, region, {}) == {}
    },
    {
    for region in local.build_regions : region => merge({
      phase = "region-build_${split(".", region)[2]}" }, lookup(local.defined_cell_overrides, region, {}))
    },
    {
    for region in local.polaris_regions :
    region => {} if lookup(local.defined_cell_overrides, region, {}) == {}
    })
  cell_override_keys   = keys(local.cell_overrides)
  build_regions_nocell = [for region in local.build_regions : split(".cell", region)[0]]
  spectre_regional_et  = compact(distinct([for key in local.cell_override_keys : split(".cell", key)[0]]))
}
