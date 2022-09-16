# Caldera Security Tests

[![License](http://img.shields.io/:license-mit-blue.svg)](https://github.com/fbsamples/caldera-security-tests/blob/main/LICENSE)
[![🚨 Semgrep Analysis](https://github.com/fbsamples/caldera-security-tests/actions/workflows/semgrep.yaml/badge.svg)](https://github.com/fbsamples/caldera-security-tests/actions/workflows/semgrep.yaml)
[![goreleaser](https://github.com/fbsamples/caldera-security-tests/actions/workflows/goreleaser.yml/badge.svg)](https://github.com/fbsamples/caldera-security-tests/actions/workflows/goreleaser.yml)
[![Tests](https://github.com/fbsamples/caldera-security-tests/actions/workflows/tests.yaml/badge.svg)](https://github.com/fbsamples/caldera-security-tests/actions/workflows/tests.yaml)

Execute two Stored XSS vulnerabilities that were found in
[MITRE Caldera](https://github.com/mitre/caldera) by [Jayson Grace](https://techvomit.net)
from the Meta Purple Team.

## Table of Contents

- [Usage](#usage)
- [Development](#development)
  - [Dependencies](#dependencies)
  - [Developer Environment Setup](#developer-environment-setup)

---

## Usage

Create test environment:

```bash
git clone https://github.com/mitre/caldera.git
git clone https://github.com/fbsamples/caldera-security-tests
cd caldera-security-tests
# Download the release binary and drop it in ./bin/
# from the root of the repo.
```

Create first test environment, run the first XSS,
and tear the test environment down:

```bash
./bin/cst-darwin TestEnv -1
export OS="$(uname | python3 -c "print(open(0).read().lower().strip())")"
./bin/"cst-${OS}" StoredXSSUno
./bin/"cst-${OS}" TestEnv -d
```

Create second test environment, run the second XSS,
and tear the test environment down:

```bash
./bin/cst-darwin TestEnv -2
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
