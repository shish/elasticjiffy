%define debug_package %{nil}
%global commit c5a4525bfa3bd9997834d0603c40093e50e3fd19
%global shortcommit %(c=%{commit}; echo ${c:0:7})

Name:           elasticjiffy
Version:        0.1
Release:        1%{?dist}
Summary:        A simple web server for logging JiffyWeb data to ElasticSearch

License:        MIT
URL:            https://github.com/shish/%{name}
Source0:        https://github.com/shish/%{name}/archive/%{name}-%{version}-%{shortcommit}.tgz
# BuildRoot:      %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)

# BuildRequires:  
# Requires:       

%description


%prep
%setup -qn %{name}-%{commit}


%build
/usr/local/go/bin/go build elasticjiffy.go


%install
rm -rf $RPM_BUILD_ROOT
mkdir -p $RPM_BUILD_ROOT/usr/bin/
mkdir -p $RPM_BUILD_ROOT/etc/init.d/
cp elasticjiffy $RPM_BUILD_ROOT/usr/bin/
cp elasticjiffy.init $RPM_BUILD_ROOT/etc/init.d/elasticjiffy


%files
%defattr(-,root,root,-)
/usr/bin/elasticjiffy
/etc/init.d/elasticjiffy
# %doc


%changelog
