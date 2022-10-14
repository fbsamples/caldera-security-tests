# CALDERA Security Regression Pipeline

[![License](https://img.shields.io/github/license/l50/goutils?label=License&style=flat&color=blue&logo=github)](https://github.com/fbsamples/caldera-security-tests/blob/main/LICENSE)
[![🚨 Semgrep Analysis](https://github.com/fbsamples/caldera-security-tests/actions/workflows/semgrep.yaml/badge.svg)](https://github.com/fbsamples/caldera-security-tests/actions/workflows/semgrep.yaml)
[![goreleaser](https://github.com/fbsamples/caldera-security-tests/actions/workflows/goreleaser.yaml/badge.svg)](https://github.com/fbsamples/caldera-security-tests/actions/workflows/goreleaser.yaml)
[![Baseline Tests](https://github.com/fbsamples/caldera-security-tests/actions/workflows/baseline.yaml/badge.svg)](https://github.com/fbsamples/caldera-security-tests/actions/workflows/baseline.yaml)
[![Security Regression Pipeline](https://github.com/fbsamples/caldera-security-tests/actions/workflows/srp.yaml/badge.svg)](https://github.com/fbsamples/caldera-security-tests/actions/workflows/srp.yaml)

This project was created to provide a proof of concept example of a
Security Regression Pipeline for vulnerabilities that were discovered
in [MITRE CALDERA](https://github.com/mitre/caldera)
by [Jayson Grace](https://techvomit.net) from Meta's Purple Team.

The attacks are run against a fresh test environment with the latest
MITRE CALDERA on a weekly basis using
[Github Actions](https://github.com/features/actions). Because patches
have been created for all of the discovered
vulnerabilities, the attacks are expected to fail.

If any of the vulnerabilities are successful during one of these runs,
an issue is automatically created noting the regression.

Ideally this should be run as part of a CI/CD pipeline gating commits,
but it can also work as a standalone entity for Purple Team
engagements, pentests, etc.

---

## Table of Contents

- [Setup](#setup)
  - [Apple Silicon users](#apple-silicon-users)
  - [Test Environment Preparation](#test-environment-preparation)
- [Running the tests as a github action](#running-the-tests-as-a-github-action)
- [Running the tests locally](#running-the-tests-locally)
- [Hacking on the Project](#hacking-on-the-project)
  - [Dependencies](#dependencies)
  - [Developer Environment Setup](#developer-environment-setup)

---

## Setup

### Apple Silicon users

Run this command:

```bash
export DOCKER_DEFAULT_PLATFORM=linux/amd64
```

### Test Environment Preparation

1. Download the release binary from github
   and drop it in a directory:

   ```bash
   mkdir bin && cd $_
   # Put downloaded binary here
   ```

2. Clone the caldera repo:

   ```bash
   cd ../ && git clone https://github.com/mitre/caldera.git
   ```

---

## Running the tests locally

Create vulnerable test environment, run the [first XSS](https://github.com/metaredteam/external-disclosures/security/advisories/GHSA-5m86-x5ph-jc47),
and tear the test environment down:

```bash
./bin/cst-darwin TestEnv -v
export OS="$(uname | python3 -c "print(open(0).read().lower().strip())")"
./bin/"cst-${OS}" StoredXSSUno
./bin/"cst-${OS}" TestEnv -d
```

Create vulnerable test environment, run the [second XSS](https://github.com/metaredteam/external-disclosures/security/advisories/GHSA-2gjc-v4hv-m4p9),
and tear the test environment down:

```bash
./bin/cst-darwin TestEnv -v
./bin/"cst-$(uname)" StoredXSSDos
./bin/"cst-$(uname)" TestEnv -d
```

Create test environment using the most recent commit
to the default CALDERA branch, try running all attacks,
and tear the test environment down:

```bash
./bin/cst-darwin TestEnv -r
./bin/"cst-$(uname)" StoredXSSUno
./bin/"cst-$(uname)" StoredXSSDos
./bin/"cst-$(uname)" TestEnv -d
```

Parameters for the tests can be modified
in the generated `config/config.yaml` file.
This file is created as soon as the `TestEnv`
command in the above example is run.

---

### Running the tests as a github action

You can incorporate the CALDERA SRP into your CALDERA fork
by creating `.github/workflows/srp.yaml` and populating
it with the following contents:

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

You can use the outcomes of these workflow runs to gate
updates for your CALDERA deployments if a security regression
in the latest CALDERA release is detected.

---

## Hacking on the Project

### Dependencies

- [Install homebrew](https://brew.sh/):

  ```bash
  # Linux
  sudo apt-get update
  sudo apt-get install -y build-essential procps curl file git
  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
  eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"

  # macOS
  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
  ```

- [Install ruby](https://www.ruby-lang.org/en/):

  ```bash
  brew install ruby
  ```

- [Install gvm](https://github.com/moovweb/gvm):

  ```bash
  bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
  source "${GVM_BIN}"
  ```

- [Install golang](https://go.dev/):

  ```bash
  gvm install go1.18
  ```

- [Install pre-commit](https://pre-commit.com/):

  ```bash
  brew install pre-commit
  ```

- [Install Mage](https://magefile.org/):

  ```bash
  go install github.com/magefile/mage@latest
  ```

### Developer Environment Setup

0. [Fork this project](https://docs.github.com/en/get-started/quickstart/fork-a-repo)

1. Clone your forked repo and caldera:

   ```bash
   git clone https://github.com/fbsamples/caldera-security-tests.git
   git clone https://github.com/mitre/caldera.git
   ```

2. (Optional) If you installed gvm, create golang pkgset specifically for this project:

   ```bash
   VERSION='1.18'
   PROJECT=caldera-security-tests

   gvm install "go${VERSION}"
   gvm use "go${VERSION}"
   gvm pkgset create "${PROJECT}"
   gvm pkgset use "${PROJECT}"
   ```

3. Install dependencies:

   ```bash
   mage installDeps
   ```

4. Install pre-commit hooks:

   ```bash
   mage installPreCommitHooks
   ```

5. Update and run pre-commit hooks locally:

   ```bash
   mage runPreCommit
   ```

6. Compile binary:

   ```bash
   export OS="$(uname | python3 -c "print(open(0).read().lower().strip())")"
   mage compile "${OS}"
   ```
