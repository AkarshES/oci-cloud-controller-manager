**Script Purpose**

This script validates your CCM/CSI release images by:

1. Comparing your local `MANIFEST.csv` against the `image_list.json` extracted from the release archive.
2. Reporting any missing entries or count mismatches.
3. Updating the provided Stampede JSON spec with the new artifacts list and Git commit ID for the specified branch.

**How to Execute**

```bash
${REPO_ROOT}/stampede/update_stampede_spec.sh \
  -r <path-to-oci-ccm-repo> \
  -s <path-to-stampede.json> \
  -c <commit-id-of-ccm-internal-branch-head>
```

* `-r, --repo`    : Path to the local OCI CCM repository containing `MANIFEST.csv`.
* `-s, --stampede`: Path to the Stampede spec JSON file to update.
* `-c, --commit-id`  : Commit id of internal branch
