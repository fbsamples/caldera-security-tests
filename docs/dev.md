# Dev

If you're planning to develop/make changes to the `caldera-security-tests` project,
you're in the right place!

## Environment Setup

1. [Fork this project](https://docs.github.com/en/get-started/quickstart/fork-a-repo)

1. [Install homebrew](https://brew.sh/):

```bash
# Linux
sudo apt-get update
sudo apt-get install -y build-essential procps curl file git
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"

# macOS
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

1. Download and install the [gh cli tool](https://cli.github.com/).

1. Clone the repo:

   ```bash
   gh repo clone fbsamples/caldera-security-tests
   cd caldera-security-tests
   ```

1. [Install dependencies with brew](https://www.ruby-lang.org/en/):

   ```bash
   brew install ruby pre-commit
   ```

1. [Install gvm](https://github.com/moovweb/gvm):

   ```bash
   bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
   ```

1. [Install golang](https://go.dev/):

   ```bash
   source .gvm
   ```

1. [Install Mage](https://magefile.org/):

   ```bash
   go install github.com/magefile/mage@latest
   ```

1. Install and run pre-commit hooks:

   ```bash
   mage runPreCommit
   ```

1. Install dependencies:

   ```bash
   mage installDeps
   ```

1. Clone your forked repo and caldera:

   ```bash
   git clone https://github.com/your/caldera-security-tests.git
   git clone https://github.com/mitre/caldera.git
   ```

1. Compile binary:

   ```bash
   export OS="$(uname | python3 -c "print(open(0).read().lower().strip())")"
   mage compile "${OS}"
   ```
