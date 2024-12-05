SHELL := /bin/bash

.PHONY: build
build: out/infra-release-validator-ccm-csi-${VERSION}.tar.gz

out/infra-release-validator-ccm-csi-${VERSION}.tar.gz: run-command/**
	mkdir -p out
	rm -f out/*
	tar -czvf out/infra-release-validator-ccm-csi-${VERSION}.tar.gz run-command
