<h2 align="center">PFD - network realy tools</h2>

## Introduction

fpd is a simple, high performance relay server written in golang.

## Features
- Zero configuration. Setup and run in one command.
- Concurrency. Bidirectional concurrent traffic leads to high performance.
- Low resources cost.

## Usage

An example to listen on port 30000 and forwarding traffic to example.com:12345 is as follows.
```
pfd -l 127.0.0.1:30000 -r example.com:12345
```