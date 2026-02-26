# Copyright 2018 Oracle and/or its affiliates. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# temp directory for which we pull repos to.
export TMP_DEP_DIR		:= ${PWD}/temp_repos

CWD := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

include rpm/fvd_rpm.mk

# ioke/screts repository settings (clone-secrets))
SECRETS_REPO	:= ssh://git@bitbucket.oci.oraclecorp.com:7999/okei/secrets.git

ifeq ($(TC_BUILD),0)
export SECRETS_LOCAL	?= ${TMP_DEP_DIR}/secrets
export SECRETS_DIR		:= ${SECRETS_LOCAL}/k8-infra/${REGION_SECRETS}
export KUBECONFIG		:= ${SECRETS_DIR}/kubeconfig.TNL
else
export SECRETS_LOCAL	:= /secrets
export SECRETS_DIR      := ${SECRETS_LOCAL}/k8-infra/${REGION_SECRETS}
export KUBECONFIG		:= ${SECRETS_DIR}/kubeconfig.TNL
endif

PKG := github.com/oracle/oci-cloud-controller-manager
REGISTRY ?= odo-docker-signed-local.artifactory.oci.oraclecorp.com
IMAGE ?= $(REGISTRY)/oke-public-cloud-provider-oci
COMPONENT ?= oci-cloud-controller-manager oci-volume-provisioner oci-flexvolume-driver cloud-provider-oci oci-csi-controller-driver oci-csi-node-driver
OCI_CLI_VERSION ?= master
KUBECTL_VERSION ?= 1.11.0

DEPLOYMENT_PIPELINE_ID := "ocid1.devopsbuildpipeline.oc1.phx.amaaaaaa7aeocuiaa54z5gqks6ewodlvvs2iuyc3chxqfyfu2g6hi2zxgasq"

GIT_COMMIT := $(shell GCOMMIT=`git rev-parse --short HEAD`; if [ -n "`git status . --porcelain`" ]; then echo "$$GCOMMIT-dirty"; else echo $$GCOMMIT; fi)
DOCKER_REPO_ROOT?=/go/src/github.com/oracle/oci-cloud-controller-manager
# Allow overriding for release versions else just equal the build (git hash)
ifeq "$(BUILD_NUMBER)" ""
    VERSION_SUFFIX   ?= $(GIT_COMMIT)
else
    VERSION_SUFFIX   ?= $(GIT_COMMIT)-$(BUILD_NUMBER)
endif

K8S_VERSION := $(shell cat VERSION)
VERSION ?= oke-$(K8S_VERSION)-$(VERSION_SUFFIX)
BUILD = $(VERSION)

GOOS ?= linux
ARCH ?= amd64

SRC_DIRS := cmd controllers pkg # directories which hold app source (not vendored)

# Allows overriding where the CCM should look for the cloud provider config
# when running via make run-dev.
CLOUD_PROVIDER_CFG ?= $$(pwd)/cloud-provider.yaml

RETURN_CODE := $(shell sed --version >/dev/null 2>&1; echo $$?)
ifeq ($(RETURN_CODE),1)
    SED_INPLACE = -i ''
else
    SED_INPLACE = -i
endif

.PHONY: all
all: check test build

.PHONY: gofmt
gofmt:
	@./hack/check-gofmt.sh $(SRC_DIRS)

.PHONY: golint
golint:
	@./hack/check-golint.sh $(SRC_DIRS)

.PHONY: govet
govet:
	@./hack/check-govet.sh $(SRC_DIRS)

.PHONY: check
check: gofmt govet golint

.PHONY: build-dirs
build-dirs:
	@mkdir -p dist/
	@mkdir -p dist/arm/

.PHONY: build
build: build-dirs
	@for component in $(COMPONENT); do \
		GOOS=$(GOOS) GOARCH=$(ARCH) CGO_ENABLED=1 go build -mod vendor -o dist/$$component -ldflags "-X main.version=$(VERSION) -X main.build=$(BUILD)" ./cmd/$$component ; \
    done

.PHONY: build-arm
build-arm: build-dirs
	@for component in $(COMPONENT); do \
		GOOS=$(GOOS) GOARCH=$(ARCH) CGO_ENABLED=1 go build -mod vendor -o dist/arm/$$component -ldflags "-X main.version=$(VERSION) -X main.build=$(BUILD)" ./cmd/$$component ; \
    done

.PHONY: manifests
manifests: build-dirs
	@cp -a manifests/**/*.yaml dist
	@sed $(SED_INPLACE)                         \
	  's#${IMAGE}:latest#${IMAGE}:${VERSION}#g' \
	  dist/*.yaml

.PHONY: vendor
vendor:
	@GO111MODULE=on go mod vendor -v

.PHONY: test
test:
	@./hack/test.sh $(SRC_DIRS)

.PHONY: coverage
coverage: test
	GO111MODULE=off go tool cover -html=coverage.out -o coverage.html
	GO111MODULE=off go tool cover -func=coverage.out > coverage.txt

# Run the canary tests - in single run mode.
.PHONY: canary-run-once
canary-run-once:
	@./hack/test-canary.sh run-once

# Run the canary tests - in monitor (infinite loop) mode.
.PHONY: canary-monitor
canary-monitor:
	@./hack/test-canary.sh monitor

# Validate the generated canary test image. Runs test once
# and monitors from sidecar.
.PHONY: validate-canary
validate-canary:
	@./hack/validate-canary.sh run

.PHONY: clean
clean:
	@rm -rf dist

.PHONY: run-ccm-dev
run-ccm-dev:
	@go run cmd/oci-cloud-controller-manager/main.go  \
	  --kubeconfig=$(KUBECONFIG)                      \
	  --cloud-config=$(CLOUD_PROVIDER_CFG)            \
	  --cluster-cidr=10.244.0.0/16                    \
	  --leader-elect-resource-lock=configmaps         \
	  --cloud-provider=oci                            \
	  -v=4

.PHONY: run-volume-provisioner-dev
run-volume-provisioner-dev:
	@NODE_NAME=$(shell hostname)                      \
	CONFIG_YAML_FILENAME=cloud-provider.yaml          \
	go run cmd/oci-volume-provisioner/main.go         \
	    --kubeconfig=$(KUBECONFIG)                    \
	    -v=4

.PHONY: image
BUILD_ARGS = --no-cache --pull --build-arg COMPONENT="$(COMPONENT)"
image:
	docker build $(BUILD_ARGS) \
		-t $(IMAGE):$(VERSION) .
	docker build $(BUILD_ARGS) \
		-t $(IMAGE)-arm:$(VERSION) -f Dockerfile_arm .

.PHONY: push
push: image
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE)-arm:$(VERSION)

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: build-local
build-local: build-dirs
	@docker run --rm \
		   --privileged \
			 -w $(DOCKER_REPO_ROOT) \
			 -v $(PWD):$(DOCKER_REPO_ROOT) \
			 -e COMPONENT="$(COMPONENT)" \
			 -e GOPATH=/go/ \
			odo-docker-signed-local.artifactory.oci.oraclecorp.com/oke-golang-fips:go1.25.0-ol8-102 /bin/bash -c \
			'for component in ${COMPONENT}; do \
				echo building $$component && GOOS=$(GOOS) GOARCH=$(ARCH) CGO_ENABLED=1 go build -mod vendor -buildvcs=false -o dist/$$component -ldflags "-X main.version=$(VERSION) -X main.build=$(BUILD)" ./cmd/$$component ; \
			 done'

.PHONY: test-local
test-local: build-dirs
	@docker run --rm \
		   --privileged \
			 -w $(DOCKER_REPO_ROOT) \
			 -v $(PWD):$(DOCKER_REPO_ROOT) \
			 -e COMPONENT="$(COMPONENT)" \
			 -e GOPATH=/go/ \
			odo-docker-signed-local.artifactory.oci.oraclecorp.com/odx-oke/oke/k8-manager-base:go1.20.8-1.0.19-165 \
			make coverage image

.PHONY: run-ccm-e2e-tests-local
run-ccm-e2e-tests-local:
	echo "Start run-ccm-e2e-tests-local"
	set -x && ./hack/run_e2e_test.sh

# make temporary depenancy base dir
create-deps-dir:
	@if [ ! -d $(TMP_DEP_DIR) ]; then \
		mkdir $(TMP_DEP_DIR); \
	fi

# clone or update the secrets repository
clone-secrets: create-deps-dir
	@if [ ! -d ${SECRETS_LOCAL} ]; then \
		git clone ${SECRETS_REPO} ${SECRETS_LOCAL}; \
	else \
		cd ${SECRETS_LOCAL}; \
		git pull; \
		cd ..; \
	fi

images/oke-ccm-e2e-tests-pop-image.tar.gz:
	mkdir -p images/e2e-tests
	rm -rf ${TMP_DEP_DIR}/secrets.tar.gz
	tar zcf ${TMP_DEP_DIR}/secrets.tar.gz -C ${SECRETS_LOCAL} .
	sops -i -e --oci-kms https://avnzdivwaadfa-crypto.kms.us-phoenix-1.oraclecloud.com/ocid1.key.oc1.phx.avnzdivwaadfa.abyhqljrlxrkhc2g3wokrgishtxzt7ztxilatsvmshwk6w2yr75pfgadenlq --oci-profile SOPS ${TMP_DEP_DIR}/secrets.tar.gz
	docker build --rm --build-arg https_proxy="$$https_proxy" --build-arg OCI_CLI_VERSION="$(OCI_CLI_VERSION)" --build-arg KUBECTL_VERSION="$(KUBECTL_VERSION)" -t oke-ccm-e2e-tests-pop -f images/e2e-tests/Dockerfile .
	rm -rf ${TMP_DEP_DIR}/secrets.tar.gz
	docker save oke-ccm-e2e-tests-pop | gzip > images/oke-ccm-e2e-tests-pop-image.tar.gz

out/oke-ccm-tests-pop-$(BUILD_NUMBER).tar.gz: run-command/validate/* images/oke-ccm-e2e-tests-pop-image.tar.gz
	mkdir -p out
	rm -f out/*
	tar -czvf out/oke-ccm-e2e-tests-pop-$(BUILD_NUMBER).tar.gz images/oke-ccm-e2e-tests-pop-image.tar.gz run-command

.PHONY: create-pop
create-pop: clone-secrets out/oke-ccm-tests-pop-$(BUILD_NUMBER).tar.gz

# NPN
.PHONY: install-controller-runtime
install-controller-runtime:
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.7.0

.PHONY: npn-generate
npn-generate:
	$(GOPATH)/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./api/..."

.PHONY: rpm-test
rpm-test:
	ls -lart
	pwd
	find . -print | grep -v "vendor"

.pl-patched:
	echo "Executing Patch to fix upstream changes"
	git apply --apply patches/0001-Modify-sidecar-upstream-to-use-versiond-feature-gate.patch
	echo "Verify patch changes"
	git diff

.PHONY: extract-multiarch-sha
extract-multiarch-sha:
	curl https://artifactory-builds.oci.oraclecorp.com/odo-docker-signed-local/oke-public-cloud-provider-oci/$(IMAGE_TAG)/list.manifest.json?properties


.PHONY: onboard-cicd
onboard-cicd:
	@printf "\033[1;32mCongrats! You’re just a code push away!\033[0m\n"
	$(eval TARGET_BRANCH_PREFIX := $(shell git rev-parse --abbrev-ref HEAD))
	@# Determine DEPLOY_MODE with validation and default
	@if [ -z "$(MODE)" ]; then \
	  echo "No MODE provided, defaulting to 'PUSH_ONLY'"; \
	  DEPLOY_MODE="PUSH_ONLY"; \
	elif [ "$(MODE)" = "DEFAULT" ] || [ "$(MODE)" = "PUSH_ONLY" ]; then \
	  DEPLOY_MODE="$(MODE)"; \
	else \
	  echo "Invalid MODE: must be 'DEFAULT' or 'PUSH_ONLY'"; \
	  exit 1; \
	fi; \
	echo "Using current branch: $(TARGET_BRANCH_PREFIX)"; \
	echo "Using deploy mode: $$DEPLOY_MODE"; \
	awk -v pid="$(DEPLOYMENT_PIPELINE_ID)" \
	    -v branch="$(TARGET_BRANCH_PREFIX)" \
	    -v tenancy="\"CNP_DEV\"" \
	    -v deploy_mode="\"$$DEPLOY_MODE\"" '\
		/^releaseBranches:.*\]$$/ {gsub(/\]$$/, ", \"" branch "*\"]")} \
		/^triggerOnCommitBranches:.*\]$$/ {gsub(/\]$$/, ", \"" branch "*\"]")} \
		/^bitbucketTag:/ { \
			print $$0; \
			print "pipelineId: \"" pid "\""; \
			print "exportVariables = {"; \
			print "  CCM_ARTIFACT_VERSION = $${version}"; \
			print "  DEPLOY_MODE = " deploy_mode; \
			print "  TRIGGER_E2ES = \"false\""; \
			print "}"; \
			next \
		} \
		{print} \
	' ocibuild.conf > temp && mv temp ocibuild.conf

	@printf "\n\033[1;42;30m  CONFIGURATION APPLIED TO ocibuild.conf  \033[0m\n\n"
	@printf "\033[1;34mParameter Customization Guide:\033[0m\n"
	@printf "  \033[1;33mDEPLOY_MODE\033[0m\n"
	@printf "    - DEFAULT   : Image push + mapping update\n"
	@printf "    - PUSH_ONLY : Image push only\n\n"
	@printf "  \033[1;33mTRIGGER_E2ES\033[0m\n"
	@printf "    - true  : triggers E2E tests after deployment\n"
	@printf "    - false : skips E2E triggers\n\n"
	@printf "  \033[1;33mTARGET_TENANCY\033[0m\n"
	@printf "    - CNP_DEV  : deploys to oci_cnp_dev tenancy\n"
	@printf "    - ODX_MOCK : deploys to odx-mock tenancy\n"
	@printf "    - ALL      : deploys to entire integ environments\n\n"
	@printf "Pipeline notifications will be available in the \033[1;4m#cpo-devops\033[0m channel\n"
	@printf "\nMode documentation can be found in the \033]8;;file://$(PWD)/docs/Development.md\adocs/Development.md\033]8;;\033\\ file\n"

.PHONY: deboard-cicd
deboard-cicd:
	@printf "\033[1;33mRemoving CI/CD configuration from ocibuild.conf...\033[0m\n"
	$(eval TARGET_BRANCH_PREFIX := $(shell git rev-parse --abbrev-ref HEAD))
	@# Remove branch from releaseBranches and triggerOnCommitBranches arrays
	@# Remove pipelineId and exportVariables block
	@awk -v branch="$(TARGET_BRANCH_PREFIX)" '\
	   /^releaseBranches:/ { \
	      gsub(", \"" branch "\\*\"", ""); \
	      gsub("\"" branch "\\*\", ", ""); \
	      gsub("\"" branch "\\*\"", ""); \
	   } \
	   /^triggerOnCommitBranches:/ { \
	      gsub(", \"" branch "\\*\"", ""); \
	      gsub("\"" branch "\\*\", ", ""); \
	      gsub("\"" branch "\\*\"", ""); \
	   } \
	   /^pipelineId:/ { next } \
	   /^exportVariables = \{/ { \
	      while (getline > 0 && $$0 !~ /^\}$$/) continue; \
	      next \
	   } \
	   {print} \
	' ocibuild.conf > temp && mv temp ocibuild.conf
	@printf "\033[1;32mCI/CD configuration removed successfully!\033[0m\n"
	@printf "Branch '\033[1;33m$(TARGET_BRANCH_PREFIX)\033[0m' has been deboarded from CI/CD pipeline.\n"
