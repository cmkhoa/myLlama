# Releases and compatibility

## Version policy

Start at **1.0.0-rc.1**, then promote to **1.0.0** after the roadmap's real-model
and clean-install gates pass. A release candidate is explicitly a prerelease,
not a stability claim. SemVer reserves 1.0.0 for a defined public contract:
https://semver.org/.

`VERSION` is the single source for release numbers. Git tags must equal `v` plus
that file, for example `v1.0.0-rc.1`. Source builds identify as `dev`; packaged
binaries embed version, source commit, build date and Go version.

After 1.0:

- Patch: compatible fixes and maintenance.
- Minor: compatible features and additive fields/endpoints.
- Major: breaking behavior, removed fields, renamed commands/settings or changed
  defaults that break documented workflows. Document migration before release.

The supported contract includes documented HTTP endpoints/error codes, CLI
commands/flags, configuration variables, profile fields and JSON output shapes.
Clients must ignore unknown response fields. Human-readable text and internal Go
packages are not stable interfaces. Source imports remain internal to this repo.

## What CI verifies

Every branch push/PR and explicit CI dispatch runs:

1. Formatting, unit/integration tests and vet on Linux/macOS, minimum Go 1.24 and
   the current stable Go release. The minimum is a compatibility test, not the
   compiler used to distribute releases.
2. Linux race checks and `govulncheck` with current stable Go.
3. Python SDK 2.8.1 -> actual myLlama binary -> managed mock runner on Linux/macOS.
   Also exercises CLI JSON, idle shutdown, signal cleanup and log privacy.
4. Four cross-compiled archives, checksums and native Linux installer/version
   smoke tests. The macOS SDK job separately tests native Go binaries; clean
   downloaded-archive testing remains a release promotion gate.

Official Actions are pinned to revisions. Dependabot proposes Actions and SDK
updates. It does not merge them automatically. No cloud API key, model download
or GPU is needed. CI artifacts expire after 14 days; published releases persist.

## Build and inspect locally

Requirements: Go 1.24+; Python 3.12+ for packaging/SDK tests; `make` and standard
POSIX tools. Use a current supported Go toolchain for release builds.

```sh
make verify
make race
make build
python3 -m venv .venv
.venv/bin/pip install -r tests/compat/requirements.txt
.venv/bin/python tests/compat/test_sdk.py
make release
python3 scripts/verify_release.py
```

`dist/` must be empty or absent before a release build. This prevents stale assets
from being published. To keep a previous build, choose a fresh output directory:
`python3 scripts/release.py --output /tmp/myLlama-release-check`.

Archives include both commands, installer, examples and operating docs. No runner,
model, local profile or log is bundled. Supported archive targets: Linux and macOS,
each amd64 and arm64. CPU/GPU inference support is a separate runtime concern.

## Publish a version

1. Update `VERSION` and add a matching curated `CHANGELOG.md` section in a small PR.
2. Review compatibility and complete the appropriate roadmap gates. Merge only
   after CI is green. Use the exact merged commit for the release.
3. From an up-to-date checkout of `main`, create and push the explicit tag:

   ```sh
   git tag -a v1.0.0-rc.1 -m "myLlama 1.0.0-rc.1"
   git push origin v1.0.0-rc.1
   ```

4. The Release workflow reruns all CI, verifies the tag/version and that the commit
   belongs to main, checks downloaded assets, generates artifact attestations,
   uploads everything to a draft release, then publishes it. RC tags are marked
   prerelease and never become Latest; stable tags become Latest.
5. Verify a downloaded archive/checksum and its provenance, then perform the
   manual real-model smoke test with a recorded runtime/model revision.

No branch push, dependency update or pull request automatically increments a
version or publishes a release. Repeated pipeline runs may repair a draft but
refuse to overwrite an already published release. Never move a published tag:
ship a new patch/RC number. A failed publish may leave a draft for inspection.

The release job uses the built-in `GITHUB_TOKEN` with job-scoped contents,
attestation, artifact metadata and OIDC permissions; no personal token or signing
secret is required. Artifact attestation availability depends on repository
visibility/plan; this repository is public.

## Repository setup and rollback

After the workflow is merged, configure a main-branch ruleset requiring the CI
checks and PR review to match the project's contributor model. Do not guess check
names before the first run or silently lock out the sole maintainer. Tag
restrictions and reviewer requirements are repository-admin decisions. Ordinary
CI remains read-only; the tag publish job has the release permissions.

If a release is broken, mark its notes clearly, restore the prior binaries from
the previous release, and issue a new fixed version. Configuration formats must
have documented migration/backup steps before a breaking change. myLlama does
not auto-update installed binaries.

## Verify downloads

Download an archive and `SHA256SUMS` from the same release, then check its hash
(`sha256sum --check SHA256SUMS` on Linux; on macOS, compare `shasum -a 256 ARCHIVE`
with its entry). The full checksum file lists all four archives; checking the
whole file requires downloading all four.

```sh
gh attestation verify myLlama_1.0.0-rc.1_darwin_arm64.tar.gz --repo cmkhoa/myLlama
```

Checksums detect mismatches; attestations identify the GitHub build workflow and
source revision. Verify both before installing.
