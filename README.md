# firmdiff

Firmware diffing for engineers who need answers, not just bytes.

[![CI](https://github.com/lostbinarylabs/firmdiff/actions/workflows/ci.yml/badge.svg)](https://github.com/lostbinarylabs/firmdiff/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/lostbinarylabs/firmdiff)](https://goreportcard.com/report/github.com/lostbinarylabs/firmdiff)
[![License](https://img.shields.io/github/license/lostbinarylabs/firmdiff)](LICENSE)

firmdiff helps engineers compare firmware builds and explain **what actually changed**.

It is designed for embedded systems engineers dealing with:

- non-reproducible builds
- unexplained binary differences
- mysterious firmware regressions
- toolchain instability
- build system drift

Instead of diffing hex dumps, **firmdiff explains the difference**.

---

# Why firmdiff exists

Embedded firmware builds often change even when the source code does not.

Engineers frequently discover:

- identical builds produce different binaries
- compiler upgrades change output
- link ordering changes addresses
- toolchains inject timestamps
- build flags drift over time

firmdiff helps you detect and explain these problems.

---

# Installation

```bash
go install github.com/lostbinarylabs/firmdiff/cmd/firmdiff@latest

or build locally

git clone https://github.com/lostbinarylabs/firmdiff
cd firmdiff
go build ./cmd/firmdiff

Quick start

Run firmdiff on a firmware project:

firmdiff run --src examples/hello_firmdiff

Example output:

Running build A...
Running build B...

Comparing firmware...

Binary difference detected

Possible causes:
 - timestamp embedded in binary
 - object file order change
 - linker script variation
Commands
Command	Description
firmdiff run	perform firmware build comparison
firmdiff explain	analyse and explain binary differences
firmdiff report	produce markdown diff report
firmdiff sweep	run extended analysis
Example projects

See the examples/ directory for runnable demos.

Documentation

Full documentation lives in the docs/ folder.

Getting Started

Command Guide

Architecture

Troubleshooting

Roadmap

Planned features include:

ELF symbol diffing

linker map analysis

reproducible build verification

CI integration

Contributing

Contributions are welcome.

Please read:

CONTRIBUTING.md