# Maintainer guide: proper APT + YUM repositories (Option C)

This is the distribution model you asked for. It’s also where cutting corners becomes a supply-chain incident.

## What this bundle gives you
- .deb + .rpm packaging via nfpm
- APT repo generation (dpkg-scanpackages + apt-ftparchive)
- YUM repo generation (createrepo_c)
- GitHub Actions workflow that publishes `public/` to `gh-pages`

## Setup: GitHub Pages
In GitHub:
- Settings → Pages
- Source: Deploy from a branch
- Branch: `gh-pages` (root)

Repo base URL becomes:
`https://<your-org>.github.io/<repo-name>/`

## Signing (recommended)
Create a dedicated signing key:
```bash
gpg --full-generate-key
gpg --list-secret-keys --keyid-format=long
```

Store in GitHub Secrets:
- `GPG_PRIVATE_KEY` (ASCII armored private key export)
- `GPG_PASSPHRASE`
- `GPG_KEY_ID`

Publish your public key under `public/keys/` (committed or generated).

## Release flow
Tag and push:
```bash
git tag v0.1.0
git push origin v0.1.0
```

Workflow runs:
- builds packages
- generates repos under `public/`
- publishes to `gh-pages`

## Policy decision: auto-enable timer
The postinstall script enables the timer by default. If you do not want that:
- remove `systemctl enable --now serverpatcher.timer` from `packaging/scripts/postinstall.sh`
