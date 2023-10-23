# Backlink
[![Go Reference](https://pkg.go.dev/badge/github.com/TheGejr/backlink.svg)](https://pkg.go.dev/github.com/TheGejr/backlink) ![GitHub Workflow Status (with event)](https://img.shields.io/github/actions/workflow/status/TheGejr/backlink/release.yml) ![GitHub](https://img.shields.io/github/license/TheGejr/backlink) ![GitHub release (with filter)](https://img.shields.io/github/v/release/TheGejr/backlink)





This simple program makes a list of backlinks from a website. It reports on both external and internal backlinks.
It can be used for a quick overview of a website.

## Installation and Usage
Install this program:
```
$ go install github.com/TheGejr/backlink@latest
```

```
USAGE: backlink [OPTION...] DOMAIN

$ backlink --help
```

## Example usage
```
$ backlink https://gejr.dk
https://gejr.dk/
https://gejr.dk/about/
https://gejr.dk/blog/
https://twitter.com/Gejr_sec
https://github.com/TheGejr
https://linkedin.com/in/gejr
```

```
$ backlink https://gejr.dk -o output.txt
$ cat output.txt
https://gejr.dk/
https://gejr.dk/about/
https://gejr.dk/blog/
https://twitter.com/Gejr_sec
https://github.com/TheGejr
https://linkedin.com/in/gejr
```