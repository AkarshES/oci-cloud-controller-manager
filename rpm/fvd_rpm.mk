VERSION         ?= $(shell cat ocibuild.conf | grep ccmVersion: | cut -d' ' -f2 | sed 's/"//g' )
NAME            ?= oci-flexvolume-driver
BLD_ARCH        ?= x86_64
FVD_BINARY_PATH ?= /spart/input/dist
WORK_DIR 		?= /sparta/input

.PHONY: PACKAGE_TARGET
PKG_TARGET := $(WORK_DIR)/rpmbuild/RPMS/$(BLD_ARCH)/$(NAME)-$(VERSION).$(BLD_ARCH).rpm
PKG_SOURCE := $(WORK_DIR)/rpmbuild/SOURCES/$(NAME)-$(VERSION).tar.gz
PKG_SPEC   := $(WORK_DIR)/rpmbuild/SPECS/$(NAME).spec

.DEFAULTTARGET: rpm

.PHONY: setup
setup:
	./setup.sh


.PHONY: rpm
rpm: $(PKG_TARGET)

$(PKG_TARGET): $(PKG_SPEC) $(PKG_SOURCE)
	echo "Done Packing"
	find . | sed -e "s/[^-][^\/]*\// |/g" -e "s/|\([^ ]\)/|-\1/"
	rpmbuild -bb \
		--define "name $(NAME)" \
		--define "_version $(VERSION)" \
		--define "_topdir $(WORK_DIR)/rpmbuild" \
		--define "_flexvolume_install_path $(WORK_DIR)/installtest" \
		--define "_release 1" $(WORK_DIR)/rpmbuild/SPECS/fvd.spec
	echo "Done Building"
	find . | sed -e "s/[^-][^\/]*\// |/g" -e "s/|\([^ ]\)/|-\1/"

$(PKG_SOURCE): $(FVD_BINARY_PATH) | rpmbuild
	echo $@
	echo "Printing"
	echo $(dir $@)
	mkdir -p $(dir $@)
	echo "Made directory, now making zip"
	tar -czvf $@ -C $(FVD_BINARY_PATH) $(NAME)


$(PKG_SPEC): rpmbuild
	echo $(FVD_BINARY_PATH)
	echo $(WORK_DIR)
	cp -a $(WORK_DIR)/rpm/specs/* $(WORK_DIR)/rpmbuild/SPECS/

rpmbuild:
	mkdir -p $(WORK_DIR)/rpmbuild/{BUILD,BUILDROOT,RPMS,SOURCES,SPECS,SRPMS}

clean:
	rm -r rpmbuild

tree:
	find . | sed -e "s/[^-][^\/]*\// |/g" -e "s/|\([^ ]\)/|-\1/"
