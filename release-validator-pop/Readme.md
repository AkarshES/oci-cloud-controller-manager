# How to use release-validator-ccm-csi

### New Resources

* `release-validator-ccm-csi` - name of the POP artifact that is built in build service as a subproject of oci-cloud-controller-manager. This artifact is a tar file created from `run-command/00-start.sh` and `image_versions.json`.

* `image-release-validator-ccm-csi` - name of the ODO Application responsible for validating image-push (app) releases.

* `infra-release-validator-ccm-csi` - name of the ODO Application responsible for validating images before infra release (mapping-update).

### Release Process

1. Create a release branch from the internal branch, e.g `release/v1.8.0`.
2. Add the images needed to be pushed via application release to `MANIFEST.csv` and run the following command from the root directory of the project
```bash
 sh ./shepherd/limits/scripts/gen_images_tfvars.sh MANIFEST.csv >> shepherd/limits/flock_structure/images.auto.tfvars.json
```
3. The above command will update these 2 files with information about the images to be pushed - `images.auto.tfvars.json` and `image_versions.json`.
4. Create a PR from the release branch to `internal` branch. Once the PR is created, a new build for `release-validator-ccm-csi` will be triggered. Create App releases from this commit as was done previously and provide the artifact tag for `release-validator-ccm-csi` corresponding to the new build triggered for the PR.
5. Now Copy and paste the new artifact tag to pop_version in `shared_modules/properties_values/default_values.tf`. Commit and push this change. Create infra releases to both `<env>.<realm>.<region>.cell0` and `spectre.values.<env>.<realm>.<region>` targets.

### Ways to disable image validation

If image validation needs to be disabled for some reason, please follow 1 of the following approaches:

* If it is an app release, provide the artifact version for `release-validator-ccm-csi` as `skip`. This will ensure that no ODO deployments are made for image validation.

* For both app and infra releases, follow these steps to disable validation:
  1. Go to the flock config you wish to create releases from.
  2. Add description, select change type and execution targets, select artifact versions (if applicable).
  3. Under `Input Variables (optional)`, click on `Add Variable +`.
  4. Enter key = `cpo-image-validation-enabled` and value = `false`.
  5. Create the release. This release should have the module odo_deployment as empty.