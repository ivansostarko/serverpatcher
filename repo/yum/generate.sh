#!/usr/bin/env bash
set -euo pipefail

YUM_ARCH="${YUM_ARCH:-x86_64}"

rm -rf "public/yum/${YUM_ARCH}"
mkdir -p "public/yum/${YUM_ARCH}"

cp -f dist/*.rpm "public/yum/${YUM_ARCH}/"

createrepo_c "public/yum/${YUM_ARCH}"

if [[ -n "${GPG_KEY_ID:-}" ]]; then
  pushd "public/yum/${YUM_ARCH}/repodata" >/dev/null
  gpg --batch --yes --default-key "${GPG_KEY_ID}" --detach-sign --armor -o repomd.xml.asc repomd.xml
  popd >/dev/null
fi

echo "YUM repo generated at public/yum/${YUM_ARCH}"
