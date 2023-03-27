# oke-ccm-csi

## How to use this flock to push images to Steward Tenancy
### Step 1: Pull the latest changes in this project
### Step 2: Generate tfvars file for this flock
The artifact list for this flock is generated using the `MANIFEST.csv` file in this repo

```bash
$ cd ./sheperd/limits
$ ./scripts/gen_images_tfvars.sh > flock_structure/images.auto.tfvars.json
$ cd ../..
```
### Step 3: Add any artifacts that were removed from the `MANIFEST.csv` file to the `MANIFEST_ARCHIVE.csv` file.
### Step 4: Commit and push the above changes
### Step 5: Collect the commit-id for the above commit and create a flock-config using `sheperd-cli`.
Example command to be executed in sheperd-cli.

```bash
./compile_flock_config oke oke-ccm-csi ca5fab3660b2ea016a161606a87462392a7d453f
```

### Step 6: Go to the flock-config on the [devops portal](https://devops.oci.oraclecorp.com/shepherd/projects/oke/flocks/oke-ccm-csi/configs).
### Step 7: Create a release with the flock-config and approve the release with the required CHANGE ticket.