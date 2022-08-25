# Caldera Security Tests

[![License](http://img.shields.io/:license-mit-blue.svg)](https://github.com/l50/caldera-security-tests/blob/master/LICENSE)
[![🚨 Semgrep Analysis](https://github.com/l50/caldera-security-tests/actions/workflows/semgrep.yaml/badge.svg)](https://github.com/l50/caldera-security-tools/actions/workflows/semgrep.yaml)
[![goreleaser](https://github.com/l50/caldera-security-tests/actions/workflows/goreleaser.yml/badge.svg)](https://github.com/l50/caldera-security-tests/actions/workflows/goreleaser.yml)

Execute two Stored XSS vulnerabilities that were found in
[MITRE Caldera](https://github.com/mitre/caldera) by [Jayson Grace](https://techvomit.net)
from the Meta Purple Team.

## Table of Contents

- [Report of Findings](docs/REPORT.md)
- [Usage](#usage)
- [Development](#development)
  - [Dependencies](#dependencies)
  - [Developer Environment Setup](#developer-environment-setup)

---

## Usage

Create test environment:

```bash
git clone https://github.com/mitre/caldera.git
git clone https://github.com/l50/caldera-security-tests
cd caldera-security-tests
# Environment for the first XSS
./bin/"cst-$(uname)" TestEnv --uno
# Environment for the second XSS
./bin/"cst-$(uname)" TestEnv --dos
```

Create test environment, run the first XSS,
and tear the test environment down:

```bash
./bin/cst-darwin TestEnv -1
./bin/"cst-$(uname)" StoredXSSUno
./bin/"cst-$(uname)" TestEnv -d
```

Create test environment, run the second XSS,
and tear the test environment down:

```bash
./bin/cst-darwin TestEnv -2
./bin/"cst-$(uname)" StoredXSSDos
./bin/"cst-$(uname)" TestEnv -d
```

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
   git clone https://github.com/l50/caldera-security-tests.git
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
   mage -d .mage/ -compile ../magefile
   ```

4. Install pre-commit hooks and dependencies:

   ```bash
   ./magefile installPreCommitHooks
   ```

5. Update and run pre-commit hooks locally:

   ```bash
   ./magefile runPreCommit
   ```

6. Compile binary:

   ```bash
   ./magefile compile $uname
   ```
