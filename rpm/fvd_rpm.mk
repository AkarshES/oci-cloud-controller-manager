VERSION          ?= $(shell cat ocibuild.conf | grep ccmVersion: | cut -d' ' -f2 | sed 's/"//g' )
VERSION_V_REMOVED ?= $(shell echo $(VERSION) | sed 's/^v//')
FVD_INSTALL_NAME ?= oci
BLD_ARCH         ?= x86
FVD_BINARY_PATH  ?= ""
WORK_DIR 		 ?= ~/sparta/input
RPM_INSTALL_PATH ?= "/etc/libexec/kubernetes/kubelet-plugins/volume/exec/oracle~oci/"
INPUT_FVD_BINARY_NAME ?= "oci-flexvolume-driver"
RPM_VERSION ?= $(subst -,_,$(VERSION_V_REMOVED))
RELEASE ?= 1

.PHONY: PACKAGE_TARGET
PKG_TARGET := $(WORK_DIR)/rpmbuild/RPMS/$(BLD_ARCH)/$(FVD_INSTALL_NAME)-$(RPM_VERSION).$(BLD_ARCH).rpm
PKG_SOURCE := $(WORK_DIR)/rpmbuild/SOURCES/$(FVD_INSTALL_NAME)-$(RPM_VERSION).tar.gz
PKG_SPEC   := $(WORK_DIR)/rpmbuild/SPECS/$(FVD_INSTALL_NAME).spec

.DEFAULTTARGET: rpm

.PHONY: setup
setup:
	./setup.sh

.PHONY: rpm
rpm: $(PKG_TARGET)

$(PKG_TARGET): $(PKG_SPEC) $(PKG_SOURCE)
	rpmbuild -v -bb \
		--define "_version $(RPM_VERSION)" \
		--define "_topdir $(WORK_DIR)/rpmbuild" \
		--define "_flexvolume_install_path $(RPM_INSTALL_PATH)" \
		--define "_flexvolume_install_name $(FVD_INSTALL_NAME)" \
		--define "_release $(RELEASE)" $(WORK_DIR)/rpmbuild/SPECS/fvd.spec

$(PKG_SOURCE): $(FVD_BINARY_PATH) | rpmbuild
	mkdir -p $(dir $@)
	cp $(FVD_BINARY_PATH)/$(INPUT_FVD_BINARY_NAME) $(FVD_BINARY_PATH)/$(FVD_INSTALL_NAME)
	tar -czvf $@ -C $(FVD_BINARY_PATH) $(FVD_INSTALL_NAME)


$(PKG_SPEC): rpmbuild
	cp -a $(WORK_DIR)/rpm/specs/* $(WORK_DIR)/rpmbuild/SPECS/

rpmbuild:
	mkdir -p $(WORK_DIR)/rpmbuild/{BUILD,BUILDROOT,RPMS,SOURCES,SPECS,SRPMS}

clean-rpm:
	rm -r rpmbuild

tree:
	find . | sed -e "s/[^-][^\/]*\// |/g" -e "s/|\([^ ]\)/|-\1/"
