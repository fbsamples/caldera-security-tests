# CALDERA Security Regression Pipeline PoC

[![License](https://img.shields.io/github/license/l50/goutils?label=License&style=flat&color=blue&logo=github)](https://github.com/fbsamples/caldera-security-tests/blob/main/LICENSE)
[![🚨 Semgrep Analysis](https://github.com/fbsamples/caldera-security-tests/actions/workflows/semgrep.yaml/badge.svg)](https://github.com/fbsamples/caldera-security-tests/actions/workflows/semgrep.yaml)
[![goreleaser](https://github.com/fbsamples/caldera-security-tests/actions/workflows/goreleaser.yaml/badge.svg)](https://github.com/fbsamples/caldera-security-tests/actions/workflows/goreleaser.yaml)
[![Baseline Tests](https://github.com/fbsamples/caldera-security-tests/actions/workflows/baseline.yaml/badge.svg)](https://github.com/fbsamples/caldera-security-tests/actions/workflows/baseline.yaml)
[![Security Regression Pipeline](https://github.com/fbsamples/caldera-security-tests/actions/workflows/srp.yml/badge.svg)](https://github.com/fbsamples/caldera-security-tests/actions/workflows/srp.yml)

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

- [Usage](#usage)
- [Hacking on the Project](#hacking-on-the-project)
  - [Dependencies](#dependencies)
  - [Developer Environment Setup](#developer-environment-setup)

---

## Usage

---

### Apple Silicon users

Run this command:

```bash
export DOCKER_DEFAULT_PLATFORM=linux/amd64
```

---

Create test environment:

```bash
git clone https://github.com/mitre/caldera.git
git clone https://github.com/fbsamples/caldera-security-tests
cd caldera-security-tests
# Download the release binary and drop it in ./bin/
# from the root of the repo.
```

Create vulnerable test environment, run the first XSS,
and tear the test environment down:

```bash
./bin/cst-darwin TestEnv -v
export OS="$(uname | python3 -c "print(open(0).read().lower().strip())")"
./bin/"cst-${OS}" StoredXSSUno
./bin/"cst-${OS}" TestEnv -d
```

Create vulnerable test environment, run the second XSS,
and tear the test environment down:

```bash
./bin/cst-darwin TestEnv -v
./bin/"cst-$(uname)" StoredXSSDos
./bin/"cst-$(uname)" TestEnv -d
```

Create test environment using the most recent commit
to the default CALDERA branch, try running both attacks,
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

3. Generate the `magefile` binary:

   ```bash
   mage -d .mage/ installDeps
   mage -d -compile ../magefile
   ```

4. Install pre-commit hooks and dependencies:

   ```bash
   mage -d .mage/ installDeps
   mage -d .mage/ installPreCommitHooks
   ```

5. Update and run pre-commit hooks locally:

   ```bash
   ./magefile runPreCommit
   ```

6. Compile binary:

   ```bash
   export OS="$(uname | python3 -c "print(open(0).read().lower().strip())")"
   ./magefile compile ${OS}
   ```
