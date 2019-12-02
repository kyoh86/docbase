# docbase

A CLI tool to make the docbase more convenience!

[![Go Report Card](https://goreportcard.com/badge/github.com/kyoh86/docbase)](https://goreportcard.com/report/github.com/kyoh86/docbase)
[![Coverage Status](https://img.shields.io/codecov/c/github/kyoh86/docbase.svg)](https://codecov.io/gh/kyoh86/docbase)

## Install

```
go get github.com/kyoh86/docbase
```

## Usage

```
docbase --help
```

Required: 
Pass a API token with `--token=TOKEN` flag or an environment variable `DOCBASE_API_TOKEN`, and domain (i.e. **DOMAIN**.docbase.io) with `--domain=DOMAIN` flag or the envar `DOCBASE_DOMAIN`.

## Functions

* List tags

```
docbase --token=TOKEN --domain=DOMAIN tags list
```

* Edit tags (i.e. all `howto` tags to `manual` and `foo` to `bar`)

```
docbase --token=TOKEN --domain=DOMAIN tags edit howto:manual foo:bar
```

This tool is still incomplete...

Please contribute to create new functions.

# LICENSE

[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg)](http://www.opensource.org/licenses/MIT)

This is distributed under the [MIT License](http://www.opensource.org/licenses/MIT).
