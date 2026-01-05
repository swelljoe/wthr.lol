#!/bin/bash
set -ex

echo "=== STARTING BUILD SCRIPT ==="

# Install build dependencies
dnf install -y rpm-build golang systemd-rpm-macros tar git wget

# Setup Build Environment
mkdir -p $HOME/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}

# Fix git ownership issue since we are mounting the runner's workspace
git config --global --add safe.directory /src

echo "=== BUILDING IMPORT-GEO ==="
go build -v -o bin/import-geo ./cmd/import-geo || { echo "go build failed"; exit 1; }

echo "=== POPULATING DB ==="
./bin/import-geo || { echo "import-geo failed"; exit 1; }
ls -lh wthr.db

echo "=== CREATING TARBALL ==="
# Create Source Tarball
tar --exclude-vcs -czf $HOME/rpmbuild/SOURCES/wthr-$VERSION.tar.gz --transform "s,^,wthr-$VERSION/," .

echo "=== BUILDING RPM ==="
rpmbuild -bb packaging/rpm/wthr.spec \
  --define "_version $VERSION" \
  --define "_topdir $HOME/rpmbuild" || { echo "rpmbuild failed"; exit 1; }

echo "=== COPYING ARTIFACTS ==="
# Check if RPMs exist
find $HOME/rpmbuild/RPMS -type f
cp $HOME/rpmbuild/RPMS/*/*.rpm /src/ || { echo "Copy failed - no RPMs found"; exit 1; }

echo "=== BUILD COMPLETE ==="
