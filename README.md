check [![Build Status](https://travis-ci.org/opennota/check.png?branch=master)](https://travis-ci.org/opennota/check)
=======

A set of utilities for checking Go sources.

## Installation

    $ go get github.com/opennota/check/cmd/defercheck
    $ go get github.com/opennota/check/cmd/structcheck
    $ go get github.com/opennota/check/cmd/varcheck

## Usage

Find repeating `defer`s.

```
$ defercheck go/parser
Repeating defer p.closeScope() inside function parseSwitchStmt
```

Find unused struct fields.

```
$ structcheck --help
Usage of structcheck:
  -a=false: Count assignments only
    -n=1: Minimum use count

$ structcheck fmt
pp.n
ssave.argLimit
ssave.limit
ssave.maxWid
ssave.nlIsEnd
ssave.nlIsSpace
```

Find unused global variables and constants.

```
$ varcheck --help
Usage of varcheck:
  -e=false: Report exported variables and constants

$ varcheck image/jpeg
huffIndexLuminanceDC
huffIndexChrominanceDC
huffIndexLuminanceAC
maxV
huffIndexChrominanceAC
quantIndexLuminance
maxH
quantIndexChrominance
```

## Known limitations

structcheck doesn't handle embedded structs yet.

## License

GNU GPL v3+

