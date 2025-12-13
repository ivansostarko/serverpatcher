#!/usr/bin/env bash
set -euo pipefail

APT_DIST="${APT_DIST:-stable}"
APT_COMPONENT="${APT_COMPONENT:-main}"
APT_ARCH="${APT_ARCH:-amd64}"

rm -rf public/apt
mkdir -p "public/apt/pool/${APT_COMPONENT}"
mkdir -p "public/apt/dists/${APT_DIST}/${APT_COMPONENT}/binary-${APT_ARCH}"

cp -f dist/*.deb "public/apt/pool/${APT_COMPONENT}/"

pushd public/apt >/dev/null

dpkg-scanpackages -m "pool/${APT_COMPONENT}" /dev/null > "dists/${APT_DIST}/${APT_COMPONENT}/binary-${APT_ARCH}/Packages"
gzip -9c "dists/${APT_DIST}/${APT_COMPONENT}/binary-${APT_ARCH}/Packages" > "dists/${APT_DIST}/${APT_COMPONENT}/binary-${APT_ARCH}/Packages.gz"

apt-ftparchive release "dists/${APT_DIST}" > "dists/${APT_DIST}/Release"

if [[ -n "${GPG_KEY_ID:-}" ]]; then
  gpg --batch --yes --default-key "${GPG_KEY_ID}" --clearsign -o "dists/${APT_DIST}/InRelease" "dists/${APT_DIST}/Release"
  gpg --batch --yes --default-key "${GPG_KEY_ID}" -abs -o "dists/${APT_DIST}/Release.gpg" "dists/${APT_DIST}/Release"
fi

popd >/dev/null
echo "APT repo generated at public/apt"
