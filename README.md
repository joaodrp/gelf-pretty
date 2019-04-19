[![Release](https://img.shields.io/github/release/joaodrp/gelf-pretty.svg?style=flat-square)](https://github.com/joaodrp/gelf-pretty/releases/latest)
[![Travis](https://img.shields.io/travis/joaodrp/gelf-pretty.svg?style=flat-square)](https://travis-ci.org/joaodrp/gelf-pretty)
[![Coverage Status](https://img.shields.io/codecov/c/github/joaodrp/gelf-pretty/master.svg?style=flat-square)](https://codecov.io/gh/joaodrp/gelf-pretty)
[![Go Report Card](https://goreportcard.com/badge/github.com/joaodrp/gelf-pretty?style=flat-square)](https://goreportcard.com/report/github.com/joaodrp/gelf-pretty)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/joaodrp/gelf-pretty)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square)](LICENSE)
[![SayThanks.io](https://img.shields.io/badge/Say%20Thanks-!-1EAEDB.svg?style=flat-square)](https://saythanks.io/to/joaodrp)

# GELF Pretty

Binary to pretty-print [Graylog Extended Log Format (GELF)](http://docs.graylog.org/en/latest/pages/gelf.html) messages. Simply read lines from `stdin` and pretty-print them to `stdout`.

This project adheres to the Contributor Covenant [code of conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please refer to our [contributing guide](CONTRIBUTING.md) for further information.

## Installation

You can install `gelf-pretty` using one of the following options:
 
- Pre-built packages for macOS and Linux (easiest);
- Pre-compiled binaries for macOS, Linux and Windows;
- From source.

### Pre-built packages
#### macOS

Install via [Homebrew](https://brew.sh/):

```bash
$ brew install joaodrp/tap/gelf-pretty
```

#### Linux

Install via [Snapcraft](https://snapcraft.io/gelf-pretty): 

```bash
$ snap install gelf-pretty
```

You can also download `.deb` or `.rpm` packages from the [releases page](https://github.com/joaodrp/gelf-pretty/releases) and install with `dpkg -i` or `rpm -i` respectively.

### Pre-compiled binaries

Download the correct archive for your platform from the [releases page](https://github.com/joaodrp/gelf-pretty/releases) and extract the `gelf-pretty` binary to a directory included in your `$PATH`/`Path`.

### From source

```bash
$ go get -u github.com/joaodrp/gelf-pretty
```
Make sure that the `$GOPATH/bin` folder is in your `$PATH`.

## Output Format

GELF messages are pretty-printed in the following format:

```text
[<timestamp>] <level>: <_app>/<_logger> on <host>: <short_message> <_*>=<value>\n
    <full_message>
```

### Description
- `<timestamp>` is the value of the standard GELF unix `timestamp` field, formatted as `2006-01-02 15:04:05.000`;

- `<level>` is the value of the standard GELF log `level` field, formatted in a human-readable form (e.g. `DEBUG` instead of `7`);

- `<_app>` is an optional *reserved* additional field. It can be used to identify the name of the application emitting the logs. If not provided, the forward slash that follows it is omitted;

- `<_logger>` is an optional *reserved* additional field. It can be used to identify the specific application module or logger instance that is emitting a given log line;

- `<host>` is the value of the standard GELF `host` field;

- `<short_message>` is the value of the standard GELF `short_message` field;

- `<_*>=<value>` is any number of GELF additional fields (`_*`), formatted as `key=value` pairs separated by a whitespace. The keys leading underscore is omitted for readability;

- `<full_message>` is the value of the standard GELF `full_message` field (usually used for exception backtraces). It is preceded by a new line and indented with four spaces.

### Colors

`gelf-pretty` automatically detects if the output stream is a `TTY` or not. If (and only if) it is, the output will be formatted by default with ANSI colors for improved readability.

## Usage

To pretty-print GELF logs from your application simply pipe its output to `gelf-pretty`:

```bash
$ app | gelf-pretty
```

Run `gelf-pretty --help` for a list of available options:

```bash
$ gelf-pretty --help
Usage of gelf-pretty:
  --no-color
        Disable color output
  --version
        Show version information
```

### Capture stderr

If your application writes to the `stderr` stream you will need to pipe it  along with `stdout`:

```bash
$ app 2>&1 | gelf-pretty
```

### Disable colors

To disable colored output (even if the output stream is a `TTY`) use the `--no-color` option:

```bash
$ app | gelf-pretty --no-color
```

## FAQ

### My logs are not formatted, why?

`gelf-pretty` validates each input line. If a line (delimited by `\n`) is not a valid JSON string or is invalid according to the [GELF specification](http://docs.graylog.org/en/latest/pages/gelf.html#gelf-payload-specification), `gelf-pretty` will simply echo it back to the `stdout` without any modification (silently, with no error messages).

If you believe that your log messages are valid, please [open a new issue](https://github.com/joaodrp/gelf-pretty/issues/new) and let us know.
