# Repo signing keys

Publish your public key files under `public/keys/` so clients can install them.

Recommended:
- APT: sign `InRelease` and `Release.gpg`
- YUM/DNF: sign `repomd.xml` (repo_gpgcheck=1) and/or sign RPM packages.
