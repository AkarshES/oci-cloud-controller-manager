# This file contains variable that are used by appliation as well as infrastructure
locals {
  flock_config = {
    shorten_uniquifier = "true"
    // The following configuration elements are are pulled in from realm-config
    //    clusters_tanden_group_ocid
    //    oke_tenancy_ocid
    //    odo_tenancy_ocid
    //    secinf_tenancy_ocid
    //    boat_tenancy_ocid
    //    os_namespace
    //    public_tld
    //    internal_tld
    //    steward_tenancy_namespace
    orchestration_compartment_format = "cell%d.mp.orchestration"
    kmi_compartment_format           = "cell%d.mp.kmi"
    vcn_format                       = "cell%d"
    cell_compartment_format          = "cell%d"
    kmi_subnet_name                  = "kmi"
    env                              = ""
    phonebook_name                   = "oracle-kubernetes-engine"
    cell_name_prefix                 = "cell"
    instance_type_tag_name           = "HostType"
    watch_mp_release_label_format    = "oke-mp-release-cell%d"

    // Configuration for PKI certificate
    certificate_name_format      = "mapi-pki-cert-cell%d"
    certificate_compartment      = "assets"
    mapi_subdomain               = "mapi" // Subdomain used in common name of certificate
    api_log_namespace_format     = ""
    monitor_log_namespace_format = ""
    worker_log_namespace_format  = ""
    oke_secrets_namespace        = ""
    canary_tenancies             = ""
    integration_tenancies        = ""
    splat_base_service_name      = ""
    // Configuration for image
    // IO Overlay Image from https://confluence.oci.oraclecorp.com/display/COM/Latest+Releases#tab-IO+Overlay+Image
    // SELinux-Enabled Overlay Image from https://confluence.oci.oraclecorp.com/pages/viewpage.action?spaceKey=COM&title=Latest+Releases#tab-SELinux-Enabled+Overlay+Image
    // Needs to be specified at environment, realm, region or cell level. Depends on
    // what image was available at the time cells were provisioned for that
    // environment, realm, region or cell
    // Please note:
    // oc1 (all pre-YNY regions), & oc4 (LTN) are on IO Overlay Image
    // oc1 (YNY and later regions), oc2, oc3, oc4 (BRS), oc5, oc6, oc7 and oc8 are on SELinux-Enabled Overlay Image
    image_name = ""
    image_url  = ""
    // Alarms Configuration
    alarms_compartment = "assets"
    alarms_project     = "kubernetes"
    //SPLAT
    splat_base_service_name = "" // By default environments won't have base service unless specified
    // Configuration for Spectre properties
    enable_spectre     = true
    spectre_group_name = "clusters"
    // SMS Configuration
    oke_secrets_prime_namespace_format = "oke-prime-%s-cell%d" #not used
    canary_tenancies                   = jsonencode([])
    integration_tenancies              = jsonencode([])
  }
}