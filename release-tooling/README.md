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

**Notes:**
1. Due to current upstream sheepy limitation, we can only create one type of template at one time, and we will need to manually combine the releases tables after creating two CMs.
2. The regions part of the CM are hardcoded in release-tooling/meta-commons.py and config.j2, these files need to be manually edited every time there is a new region added (please check the (region build page)[https://devops.oci.oraclecorp.com/region-build/regions?regionsFilter=state%20%3D%20%22Production%22&sortInfo%5BsortBy%5D=Generation&sortInfo%5BsortDirection%5D=Desc]) for recently GA regions (Note: Even though this is a good starting point, this page does not include regions where OKE is GA but the whole region is not GA yet so please check our slack channels for latest information)

#TODO: Use region build [capabilities](https://devops.oci.oraclecorp.com/region-build/capabilities?owner=oracle-kubernetes-engine) to automate the above


# Creating the application releases CM

## Initialising the template to generate a JSON which is used for releases, deployment and CM creation

```
cd release-tooling
sheepy init -t templates/cpo-release/meta.py -m "app=true" --output-file app.json 
```

## Create Application Releases

```
sheepy deploy -d releases/cpo-release/app.json create --all 
```

## Create CM

```
sheepy cm -d releases/cpo-release/app.json create
```

# Creating the infrastructure releases CM

## Initialising the template to generate a JSON which is used for releases, deployment and CM creation

```
<change the config ID in meta.py if you plan to create infra releases with a different config ID>
sheepy init -t templates/cpo-release/meta.py -m "infra=true" --output-file infra.json 
```

## Create Application Releases

```
sheepy deploy -d releases/cpo-release/infra.json create --all 
```

## Create CM

```
sheepy cm -d releases/cpo-release/infra.json create
```
