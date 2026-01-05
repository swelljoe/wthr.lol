Name:       wthr
Version:    %{_version}
Release:    1%{?dist}
Summary:    No-BS Weather App
License:    MIT
URL:        https://wthr.lol
Source0:    %{name}-%{version}.tar.gz

BuildRequires: golang
BuildRequires: systemd-rpm-macros

%description
A lightweight, no-nonsense weather application written in Go.

%prep
%setup -q

%build
go build -v -o bin/wthr ./cmd/wthr

%install
# Binary
install -D -m 0755 bin/wthr %{buildroot}%{_bindir}/wthr

# Static Assets & Templates
mkdir -p %{buildroot}%{_datadir}/wthr
cp -r static %{buildroot}%{_datadir}/wthr/
cp -r templates %{buildroot}%{_datadir}/wthr/

# Database Seed
# loading wthr.db from source root to be the seed
install -D -m 0644 wthr.db %{buildroot}%{_datadir}/wthr/wthr.db.seed

# Systemd Unit
install -D -m 0644 packaging/systemd/wthr.service %{buildroot}%{_unitdir}/wthr.service

# Data Directory
mkdir -p %{buildroot}%{_sharedstatedir}/wthr

%pre
# Create 'wthr' user/group if they don't exist
getent group wthr >/dev/null || groupadd -r wthr
getent passwd wthr >/dev/null || \
    useradd -r -g wthr -d %{_sharedstatedir}/wthr -s /sbin/nologin \
    -c "wthr.lol Service User" wthr
exit 0

%post
%systemd_post wthr.service

# Initialize DB from seed if missing
if [ ! -f %{_sharedstatedir}/wthr/wthr.db ]; then
    cp %{_datadir}/wthr/wthr.db.seed %{_sharedstatedir}/wthr/wthr.db
    chown wthr:wthr %{_sharedstatedir}/wthr/wthr.db
    chmod 640 %{_sharedstatedir}/wthr/wthr.db
fi

# Ensure permissions on data dir (in case it existed but permissions were wrong)
chown wthr:wthr %{_sharedstatedir}/wthr

%preun
%systemd_preun wthr.service

%postun
%systemd_postun_with_restart wthr.service

%files
%{_bindir}/wthr
%{_unitdir}/wthr.service
%{_datadir}/wthr
%dir %attr(0750,wthr,wthr) %{_sharedstatedir}/wthr
