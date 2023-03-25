# CALDERA Security Regression Pipeline

[![License](https://img.shields.io/github/license/l50/goutils?label=License&style=flat&color=blue&logo=github)](https://github.com/fbsamples/caldera-security-tests/blob/main/LICENSE)
[![🚨 Semgrep Analysis](https://github.com/fbsamples/caldera-security-tests/actions/workflows/semgrep.yaml/badge.svg)](https://github.com/fbsamples/caldera-security-tests/actions/workflows/semgrep.yaml)
[![goreleaser](https://github.com/fbsamples/caldera-security-tests/actions/workflows/goreleaser.yaml/badge.svg)](https://github.com/fbsamples/caldera-security-tests/actions/workflows/goreleaser.yaml)
[![Baseline Tests](https://github.com/fbsamples/caldera-security-tests/actions/workflows/baseline.yaml/badge.svg)](https://github.com/fbsamples/caldera-security-tests/actions/workflows/baseline.yaml)
[![Security Regression Pipeline](https://github.com/fbsamples/caldera-security-tests/actions/workflows/srp.yaml/badge.svg)](https://github.com/fbsamples/caldera-security-tests/actions/workflows/srp.yaml)

This project was created to provide an example of a TTP Runner
and accompanying Security Regression Pipeline (SRP) for vulnerabilities
that were discovered in [MITRE CALDERA](https://github.com/mitre/caldera)
by [Jayson Grace](https://techvomit.net) from Meta's Purple Team.

The attacks that are automated using the TTP Runner are
run regularly against a fresh test environment with the latest
MITRE CALDERA on a weekly basis using
[Github Actions](https://github.com/features/actions). Because patches
have been created for all of the discovered
vulnerabilities, these attacks are expected to fail.

If any of the attacks land successfully during one of these runs,
an issue is automatically created noting the regression.

Ideally this should be run as part of a CALDERA IaC deployment
pipeline to gate commits. However, it can also be used as a
standalone tool for Purple Team engagements, pentests, etc.
that include CALDERA in the scope.

---

## Table of Contents

- [Getting Started](#getting-started)
  - [Test Environment Preparation](#test-environment-preparation)
- [Execution](#execution)
  - [Execute TTP Runner in SRP](#execute-ttp-runner-in-srp)
  - [Execute TTP Runner Locally](#execute-ttp-runner-locally)

---

## Getting Started

### Test Environment Preparation

- Run this command if on an ARM-based macOS system:

  ```bash
  export ARCH="$(uname -a | awk '{ print $NF }')"
  if [[ $ARCH == "arm64" ]]; then
      export DOCKER_DEFAULT_PLATFORM=linux/amd64
  fi
  ```

- Download the latest caldera-security-tests release from github or run this:

  ```bash
  export ARCH="$(uname -a | awk '{ print $NF }')"
  export OS="$(uname | python3 -c 'print(open(0).read().lower().strip())')"
  gh release download -p "*${OS}_${ARCH}.tar.gz"
  tar -xvf *tar.gz
  ```

- Clone the caldera repo:

  ```bash
  # From the caldera-security-tests repo root
  pushd ../ && git clone https://github.com/mitre/caldera.git && popd
  ```

---

## Execution

### Execute TTP Runner in SRP

You can incorporate the CALDERA SRP into your CALDERA fork
by creating `.github/workflows/srp.yaml` and populating it with the following contents:

```yaml
name: CALDERA Security Regression Pipeline
on:
  pull_request:
  push:
    branches: [master]

  # Run once a week (see https://crontab.guru)
  schedule:
    - cron: "0 0 * * 0"

  # Allows you to run this workflow manually from the Actions tab
  workflow_dispatch:

jobs:
  tests:
    uses: fbsamples/caldera-security-tests/.github/workflows/srp.yaml@main
```

The outcomes of these workflow runs can
be used to gate updates for your CALDERA deployments if a security regression is
detected in the latest CALDERA release.

### Execute TTP Runner Locally

Create vulnerable test environment, run the [first XSS](https://github.com/metaredteam/external-disclosures/security/advisories/GHSA-5m86-x5ph-jc47),
and tear the test environment down:

```bash
./caldera-security-tests testEnv -v
./caldera-security-tests storedXSSUno
./caldera-security-tests testEnv -d
```

Create vulnerable test environment, run the [second XSS](https://github.com/metaredteam/external-disclosures/security/advisories/GHSA-2gjc-v4hv-m4p9),
and tear the test environment down:

```bash
./caldera-security-tests testEnv -v
./caldera-security-tests storedXSSDos
./caldera-security-tests testEnv -d
```

Create vulnerable test environment, run the [third XSS](https://github.com/metaredteam/external-disclosures/security/advisories/GHSA-7344-4pg9-qf45),
and tear the test environment down:

```bash
./caldera-security-tests testEnv -v
./caldera-security-tests storedXSSTres
./caldera-security-tests testEnv -d
```

Create test environment using the most recent commit
to the default CALDERA branch, try running all attacks,
and tear the test environment down:

```bash
./caldera-security-tests testEnv -r
./caldera-security-tests storedXSSUno
./caldera-security-tests storedXSSDos
./caldera-security-tests storedXSSTres
./caldera-security-tests testEnv -d
```

Parameters for the tests can be modified
in the generated `config/config.yaml` file.
This file is created as soon as the `testEnv`
command in the above example is run.
