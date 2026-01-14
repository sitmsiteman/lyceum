# TLG Reader in Go

## Features

- Reads TLG files and `authtab.dir`.
- Searches Greek words (Ancient Greek/Beta Code).
- Performs morphological analysis (with `diogenes` dependencies).

## Usage

To browse `authtab.dir`, use:

`readauth -f path/to/authtab.dir`

To list available works, use:

`tlgviewer -f path/to/tlg[0000-9999].txt -list`

To read a full text, use:

`tlgviewer -f path/to/tlg[0000-9999].txt -w n`

If you need a pager, you can use `more` or `less`.

To search for Greek words, use:

`search -w γένος -lsj grc.lsj.xml -lsjidt lsj.idt -idt greek-analyses.idt -a greek-analyses.txt`

or

`search -w γένος`

## Dependencies

This project relies on data files from `diogenes`.

To use `search`, you need:

```
dependencies/grc.lsj.xml
dependencies/greek-analyses.idt
dependencies/greek-analyses.txt
```

For morphology, you only need: (WIP)

```
dependencies/greek-lemmata.txt
```

For faster search speeds, run `make index` before using `search`.

## Build

Run `make all` and `make index`.

## Caveats & Bugs

- TLG directory should have lowercase filenames (including extensions).

- Only tested against Aristotle and Plato datasets.

- Bekker pagination output may contain a leading "1.1." prefix.

- Some of the entries in Aristotle are skipped when use `readauth`.
