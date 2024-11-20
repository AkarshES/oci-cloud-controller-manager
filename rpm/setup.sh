#!/bin/bash -x

INSTALL=""
for pkg in go-toolset tar gzip sudo jq libblkid-devel rpm-build rpmdevtools; do
    rpm -q ${pkg}
    if [ $? -ne 0 ]; then
        INSTALL+="${pkg} "
    fi
done

if [ ! "x${INSTALL}" = "x" ]; then
    set +e
    dnf -y install ${INSTALL}
fi
