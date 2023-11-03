locals {
  // From https://confluence.oci.oraclecorp.com/display/COM/Latest+Releases#tab-IO+Overlay+Image
  // oc1 (all pre-YNY regions), & oc4 (LTN) are on IO Overlay Image
  io_overlay_uek5_images = {
    "20201014" = {
      name : "ol79-x86_64-lvm-20201014-UEK5-75G-00F6-PV"
      url : "https://objectstorage.us-phoenix-1.oraclecloud.com/p/Tzut-zRU24r6wq9UGQs17Dn-cfYK06chggQrKl0HRviuXLydlT97Rv18p1a8hCJc/n/idlybogdd5kn/b/manual_tagged_images/o/ol79-x86_64-lvm-20201014-UEK5-75G-00F6-PV"
    }
    "20201116" = {
      name : "ol79-x86_64-lvm-20201116-UEK5-75G-51BC-PV"
      url : "https://objectstorage.us-phoenix-1.oraclecloud.com/p/ng8PSLSQw6rgU92c5pnHPbkSuxO9YstSqKdDUAH6ULh8z2dIUgcCErtpd7Lxxie_/n/idlybogdd5kn/b/manual_tagged_images/o/ol79-x86_64-lvm-20201116-UEK5-75G-51BC-PV"
    }
    "20210111" = {
      name : "ol79-x86_64-lvm-20210108-UEK5-75G-C66B-PV"
      url : "https://objectstorage.us-phoenix-1.oraclecloud.com/p/8CB3yIyEOCtF3q5klB6mY9KqMns_Wv_FJyMX9pMeKPPKpC3l28rZeDnSAPda7KRg/n/idlybogdd5kn/b/manual_tagged_images/o/ol79-x86_64-lvm-20210108-UEK5-75G-C66B-PV"
    }
    "20210409" = {
      name : "E446-20210409-3183-ol79-x86_64-lvm-UEK5-75G-PV"
      url : "https://objectstorage.us-phoenix-1.oraclecloud.com/p/Ii7Yl_9PYB4qWUUzVtJSBzVphZMyKFe_tWkJLFVTG3GC8uCBjIe5m-Je0-POMGiq/n/idlybogdd5kn/b/manual_tagged_images/o/E446-20210409-3183-ol79-x86_64-lvm-UEK5-75G-PV"
    }
  }
  // From https://confluence.oci.oraclecorp.com/pages/viewpage.action?spaceKey=COM&title=Latest+Releases#tab-SELinux-Enabled+Overlay+Image
  // oc1 (YNY and later regions), oc2, oc3, oc4 (BRS), oc5, oc6 and oc8 are on SELinux-Enabled Overlay Image
  selinux_overlay_uek5_images = {
    "20201116" = {
      name : "ol79-x86_64-lvm-20201116-UEK5-75G-C6A1-PV"
      url : "https://objectstorage.us-phoenix-1.oraclecloud.com/p/bsCrgAeebWfHuS434ley_EmkytDZbrQAo3T-_5tHrWDSY-jIz00bYFnoPSmwG18j/n/idlybogdd5kn/b/manual_tagged_images/o/ol79-x86_64-lvm-20201116-UEK5-75G-C6A1-PV"
    }
  }
}