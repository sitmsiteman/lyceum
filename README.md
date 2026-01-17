# TLG Reader for Unix & Plan 9

## Features

- Reads TLG/PHI files and `authtab.dir` (sourced from legacy TLG CD-ROMs).
- Searches Greek words in LSJ (supports Ancient Greek and Beta Code).
- Searches Latin words in Lewis & Short.
- Performs morphological analysis (using `diogenes` data).

## Usage

It is recommended to use the helper scripts located in `bin/`.

### Browsing Files

To browse `authtab.dir`:

    readauth -f path/to/authtab.dir

To list available works:

    tlgviewer -f path/to/tlg[0000-9999].txt -list

To read a full text (use `more` or `less` for paging on Unix):

    tlgviewer -f path/to/tlg[0000-9999].txt -w n

### Searching Dictionaries

To search for Greek words:

    search -w γένος -dic grc.lsj.xml -dicidt lsj.idt \
       -idt greek-analyses.idt -a greek-analyses.txt

    search -w γένος

Beta Code is also supported:

    search -w ge/nos

To search for Latin words:

    search -lat -w logos -dic lat.ls.perseus-eng1.xml -dicidt ls.idt \
        -idt latin-analyses.idt -a latin-analyses.txt

    search -lat -w logos

For full usage details, use the `--help` flag.

### Plumber Integration

Add the following rule to your `lib/plumbing`.

    type is text
    data matches '([Ά-ώἀ-ῼ]+)'
    plumb to none
    plumb start window rc -c '/bin/grdic '$0'; hold'

### Acme Integration

Use `TLG9` for acme integration.

Add the folling rule to your `lib/plumbing`.

    # Open TLG worklist in acme
    type is text
    data matches 'TLG[0-9]+'
    plumb to none
    plumb start window /bin/LHelper $0

    # Open TLG text in acme
    type is text
    data matches 'ID:([0-9]+)'
    plumb to none
    plumb start window /bin/WHelper $1

## Dependencies

This project relies on prebuilt data files from [Diogenes](https://d.iogen.es/d) and the Perseus Project.

To use `search`, you need the following files (obtainable from the prebuilt `diogenes` package):

- `dependencies/grc.lsj.xml`
- `dependencies/lat.ls.perseus-eng1.xml`
- `dependencies/greek-analyses.idt`
- `dependencies/greek-analyses.txt`
- `dependencies/latin-analyses.idt`
- `dependencies/latin-analyses.txt`

For morphology analysis only, you need:

- `dependencies/greek-lemmata.txt`
- `dependencies/latin-lemmata.txt`

## Build

### Unix

Run `make`.

### 9front

Run `build.rc`.

## Caveats & Bugs

- TLG/PHI files must have lowercase filenames (including extensions).
- Test coverage is currently limited.
- Citations may occasionally contain a redundant "1.1." prefix.

## Links

- [Diogenes Desktop](https://d.iogen.es/d)
- [Perseus Project](https://www.perseus.tufts.edu/)
- [Source Repository](https://github.com/sitmsiteman/tlgread-go)

