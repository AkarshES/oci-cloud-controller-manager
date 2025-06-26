# Development documentation

The CCM has a simple build system based on `make`. Dependencies are managed
using [`go mod`][2].

## Setup
 1. Ensure you have the aforementioned development tools installed as well as the release 1.15.x of [Go][3].

 2. Clone the repo under `/src/github.com/oracle/`
 
 3. make vendor
 
## Build locally
 1. `make build-local $COMPONENT`
 2. `make image $COMPONENT `
 Note: `$COMPONENT` is optional. If you don't specify all components image will be built and placed under dist folder.

## DaemonSet manifests

You can template `manifests/cloud-controller-manager/oci-cloud-controller-manager.yaml`
using `make manifests`. This enables running specific versions of the CCM with
the proviso that the version has been pushed to Github, the CI pipeline has
passed, and `HEAD` is pointed to the commit in question. You can then execute
the following to run the CCM as a DaemonSet (RBAC optional):

```console
$ kubectl apply -f dist/oci-cloud-controller-manager.yaml
$ kubectl apply -f dist/oci-cloud-controller-manager-rbac.yaml
```
## Get the multi-arch manifest
`make IMAGE_TAG='v1.30-fd9fd08ee7f-5037' extract-multiarch-sha`

## Onboard to automated release pipelines

You can update the `ocibuild.conf` file to include the current Git branch and pipeline metadata using:

```bash
make modify-ocibuild MODE=DEFAULT
```

### Arguments
- `MODE=DEFAULT` *(default)* – Add configuration to perform image push and mapping updates deployments
- `MODE=PUSH_ONLY` – Add configuration to perform image push only deployments for the branch

> ℹ️ If `MODE` is not provided, it defaults to `DEFAULT`.
>
> ℹ️ The pipeline works for integ phoenix environment and triggering E2Es is only supported in oci-cnp-dev tenancy is not supported as of now

### This command will:
Trigger the following sequence of pipelienes
1. [CCM/CSI | Build & Trigger Integ Deployment](https://devops.oci.oraclecorp.com/devops-build/projects/ocid1.devopsproject.oc1.phx.amaaaaaa7aeocuiaaajogoeyfulynhzq5fy5d7qvmf23rahiugqiusk2kfra/build-pipelines/ocid1.devopsbuildpipeline.oc1.phx.amaaaaaa7aeocuiaa54z5gqks6ewodlvvs2iuyc3chxqfyfu2g6hi2zxgasq?_ctx=us-phoenix-1%2Coke-nodes-devopspipelines)
2. [CCM/CSI Deploy | PHX | INTEG](https://devops.oci.oraclecorp.com/devops-build/projects/ocid1.devopsproject.oc1.phx.amaaaaaa7aeocuiaaajogoeyfulynhzq5fy5d7qvmf23rahiugqiusk2kfra/build-pipelines/ocid1.devopsbuildpipeline.oc1.phx.amaaaaaa7aeocuiadl4lfbd3btwvy7b4ys724pbwkpw7c7u3thy4brdoptzq?_ctx=us-phoenix-1%2Coke-nodes-devopspipelines)

### Parameter Customization Guide

- **DEPLOY_MODE**
    - `DEFAULT`   → Image push + mapping update
    - `PUSH_ONLY` → Image push only

- **TRIGGER_E2ES**
    - `true`  → triggers E2E tests after deployment
    - `false` → skips E2E triggers

- **TARGET_TENANCY**
    - `CNP_DEV`  → deploys to oci_cnp_dev tenancy
    - `ODX_MOCK` → deploys to odx-mock tenancy
    - `ALL`      → deploys to entire integ environment

## Deboarding from automated release pipelines

You can update the `ocibuild.conf` manually to undo the changes from `onboard-cicd` command or run:

```bash
make deboard-cicd
```

## Running the e2e tests
See [README.md](../test/e2e/cloud-controller-manager/README.md)

[1]: https://www.docker.com/
[2]: https://github.com/golang/go/wiki/Modules
[3]: https://golang.org/

