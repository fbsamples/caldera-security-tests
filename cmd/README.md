# caldera-security-tests/cmd

The `cmd` package is a collection of utility functions
designed to simplify common cmd tasks.

---

## Table of contents

- [Functions](#functions)
- [Installation](#installation)
- [Usage](#usage)
- [Tests](#tests)
- [Contributing](#contributing)
- [License](#license)

---

## Functions

### Execute()

```go
Execute()
```

Execute adds child commands to the root
command and sets flags appropriately.

---

### Wait(float64)

```go
Wait(float64) time.Duration
```

Wait is used to wait for a period
of time.

---

## Installation

To use the caldera-security-tests/cmd package, you first need to install it.
Follow the steps below to install via go get.

```bash
go get github.com/l50/goutils/v2/cmd
```

---

## Usage

After installation, you can import the package in your Go project
using the following import statement:

```go
import "github.com/l50/goutils/v2/cmd"
```

---

## Tests

To ensure the package is working correctly, run the following
command to execute the tests for `caldera-security-tests/cmd`:

```bash
go test -v
```

---

## Contributing

Pull requests are welcome. For major changes,
please open an issue first to discuss what
you would like to change.

---

## License

This project is licensed under the MIT
License - see the [LICENSE](../LICENSE)
file for details.
