Name:           oci
Version:        %{_version}
Release:        %{_release}%{?dist}
Summary:        OCI's Flex Volume Driver Binary

License:        ASL 2.0
URL:            https://bitbucket.oci.oraclecorp.com/projects/OKE/repos/oci-cloud-controller-manager
Source0:        %{name}-%{version}.tar.gz

# Use passed value as an argument or default to a standard path
# This will dictate where the binary should be installed into the user's system
# Default path can be overridden with arguments
%{!?_flexvolume_install_path: %define _flexvolume_install_path /etc/libexec/kubernetes/kubelet-plugins/volume/exec/oracle~oci/}

%description
OCI's flex volume driver binary

%prep
tar -xvzf %{SOURCE0}

%build
# No build necessary for precompiled binary

%install
rm -rf %{buildroot}
mkdir -p %{buildroot}%{_flexvolume_install_path}
install -m 0755 %{name} %{buildroot}%{_flexvolume_install_path}

%files
%{_flexvolume_install_path}/oci

%changelog
* Thu Nov 20 2024 Uneet <uneet.patel@oracle.com> - 1.0.0-1
- Initial package release for fvd binary
