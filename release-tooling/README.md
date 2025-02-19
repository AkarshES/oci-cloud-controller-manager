# Install release tooling

1. Create an empty virtual environment (e.g. python3 -m virtualenv --python=python3.8 venv3)
2. Access the virtual environment with source venv3/bin/activate

```
pip install release-tooling \
--index-url=https://artifactory.oci.oraclecorp.com/api/pypi/global-release-pypi/simple \
--extra-index-url=https://artifactory.oci.oraclecorp.com/api/pypi/seeks-dev-pypi-local/simple \
--extra-index-url=https://artifactory.oci.oraclecorp.com/api/pypi/global-dev-pypi/simple \
--extra-index-url=https://artifactory.oci.oraclecorp.com/api/pypi/nre-tools-release-pypi-local/simple```

pip install --upgrade jira 
```

# Creating a release branch

1. Please cut a release branch from the internal branch which has the merged app/infra changes
2. Name the release branch against the release number (for ex: release/v1.5.0)
3. Add a new release commit to the newly created branch which removes all the older artifacts and updates the ocibuild.conf for release specific changes

# Updating CM relevant information

1. Update all the relevant info in release-tooling/templates/cpo-release/config.j2
2. Update commit ID and artifact information (against the "artifacts" and "artifact_version_resolvers" keys) in release-tooling/templates/cpo-release/meta.py
3. Update the Jira SD personal access token (generate one in profile section in Jira SD if you don't have already) in jira_pat.txt

**Notes:**
1. The change locations field is from config.j2, this file need to be manually edited every time there is a new region added (please check the (region build page)[https://devops.oci.oraclecorp.com/region-build/regions?regionsFilter=state%20%3D%20%22Production%22&sortInfo%5BsortBy%5D=Generation&sortInfo%5BsortDirection%5D=Desc]) for recently GA regions (Note: Even though this is a good starting point, this page does not include regions where OKE is GA but the whole region is not GA yet so please check our slack channels for latest information)
2. Currently, spectre.values.setup execution targets are not supported (upstream sheepy limitation), we will need to manually add in case these targets need to be included (usually found for RB regions)

#TODO: Use region build [capabilities](https://devops.oci.oraclecorp.com/region-build/capabilities?owner=oracle-kubernetes-engine) to automate the above

## Update gitmodules (template/shared_modules)
This gitmodule is reading release schedule from https://bitbucket.oci.oraclecorp.com/projects/OKE/repos/oke-common-release-tooling/browse to create Shepherd links.
Use the following make commands to add/ update gitmodule:

```
make add-gitmodule
```

```
make update-gitmodule
```

# Creating the CM

```
python3 create_cm.py
```

## Steps being run by the script (Only for reference)

```
sheepy init -t templates/cpo-release/meta.py -m "app=true" --output-file app.json 
sheepy deploy -d releases/cpo-release/app.json create --all
sheepy cm -d releases/cpo-release/app.json create
sheepy init -t templates/cpo-release/meta.py -m "infra=true" --output-file infra.json
sheepy deploy -d releases/cpo-release/infra.json create --all 
```
