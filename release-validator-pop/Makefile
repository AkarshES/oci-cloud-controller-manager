SHELL := /bin/bash

.PHONY: build
build: out/release-validator-ccm-csi-${VERSION}.tar.gz

out/release-validator-ccm-csi-${VERSION}.tar.gz: run-command/** image_versions.json
	cp -r ./../shepherd/limits/flock_structure/image_versions.json image_versions.json
	mkdir -p out
	rm -f out/*
	tar -czvf out/release-validator-ccm-csi-${VERSION}.tar.gz run-command image_versions.json
