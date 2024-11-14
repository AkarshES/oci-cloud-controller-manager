VERSION         ?= $(shell cat ocibuild.conf | grep source_version: | cut -d' ' -f2 | sed 's/"//g' )
NAME            ?= ignition
BLD_NUMBER      ?= 0
RELEASE         ?= $(shell cat ocibuild.conf | grep minor_version: | cut -d' ' -f2 | sed 's/"//g' )
BLD_RELEASE     ?= $(RELEASE).$(BLD_NUMBER)$(subst -,,$(BLD_BRANCH_SUFFIX))
BLD_COMMIT_HASH ?= 01234567
BLD_ARCH        ?= x86_64
DIST            ?= $(error DIST not set!)

.PHONY: PACKAGE_TARGET PKG_SOURCE
PKG_TARGET := rpmbuild/RPMS/$(BLD_ARCH)/$(NAME)-$(VERSION)-$(BLD_RELEASE).el8.$(BLD_ARCH).rpm
PKG_SOURCE := rpmbuild/SOURCES/$(NAME)-$(BLD_VERSION).tar.gz
PKG_SPEC   := rpmbuild/SPECS/$(NAME).spec

.DEFAULTTARGET: rpm

.PHONY: setup
setup:
	./setup.sh
	/usr/bin/go version

.PHONY: rpm
rpm: setup $(PKG_TARGET)

$(PKG_SOURCE): $(BINARY_PATH) | rpmbuild
	mkdir -p $(dir $@)
	tar -czvf $@ -C $(dir $(BINARY_PATH)) $(NAME)

$(PKG_SPEC): rpmbuild
	cp -a rpm/specs/* rpmbuild/SPECS/

rpmbuild:
	mkdir -p $(CURDIR)/rpmbuild/{BUILD,BUILDROOT,RPMS,SOURCES,SPECS,SRPMS}

$(PKG_TARGET): $(PKG_SOURCE) $(PKG_SPEC)
	rpmbuild -bb \
		--define "name $(NAME)" \
		--define "_version $(VERSION)" \
		--define "_topdir $(CURDIR)/rpmbuild" \
		--define "release $(subst -,,$(RELEASE).$(BLD_NUMBER))$(subst -,,$(BLD_BRANCH_SUFFIX))" rpmbuild/SPECS/$(NAME).spec

clean:
	rm -r rpmbuild

include-test:
	echo "Uneet Patel FVD"

