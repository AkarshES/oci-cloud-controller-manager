Name:           yourapp
Version:        1.0.0
Release:        1%{?dist}
Summary:        Go application example

License:        MIT
URL:            http://example.com
Source0:        %{name}-%{version}.tar.gz

%description
This is a Go-based command-line tool.

%prep
%setup -q

%build
# No build necessary for precompiled binary

%install
mkdir -p %{buildroot}/usr/local/bin
install -m 0755 yourapp %{buildroot}/usr/local/bin/yourapp

%files
/usr/local/bin/yourapp

%changelog
* Thu Nov 13 2024 Uneet <uneet.patel@oracle.com> - 1.0.0-1
- Initial package release
