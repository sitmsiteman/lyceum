# Lyceum: A TLG/PHI reader for Plan 9 (or \*nix)

## Features

- Reads TLG/PHI files and `authtab.dir` (sourced from legacy TLG CD-ROMs).
- Searches Greek words in LSJ (supports Ancient Greek and Beta Code).
- Searches Latin words in Lewis & Short.
- Performs morphological analysis (using `diogenes` data).

## Usage

It is recommended to use `lyceum` with `acme(1)`.

There is also a primitive frontend, `lyceum/reader`, included in the Plan 9 installation.

### Browsing TLG/PHI (in Plan 9)

To browse `authtab.dir`:

	% lyceum/readauth -f path/to/authtab.dir

To list available works:

	% lyceum/tlgviewer -f path/to/tlg[0000-9999].txt -list

To read a full text (use `more` or `less` for paging on Unix. On Plan 9, use `p`.):

	% lyceum/tlgviewer -f path/to/tlg[0000-9999].txt -w n

### Searching Dictionaries

To search for Greek words:

	% lyceum/search -w γένος -dic grc.lsj.xml -dicidt lsj.idt \
		-idt greek-analyses.idt -a greek-analyses.txt

or

	% lyceum/search -w γένος

Beta Code is also supported:

	% lyceum/search -w ge/nos

To search for Latin words:

	% lyceum/search -lat -w logos -dic lat.ls.perseus-eng1.xml -dicidt ls.idt \
        -idt latin-analyses.idt -a latin-analyses.txt

or

	% lyceum/search -lat -w logos

For full usage details, use the `--help` flag.

### Plumber Integration

Add the following rule to your `lib/plumbing`.

	type is text
	data matches '([Ά-ώἀ-ῼ]+)'
	plumb to none
	plumb start window /bin/lyceum/grdic ''$1''; hold

### Acme Integration

You can use `lyceum/TLG` or `lyceum/PHI` in `acme(1)` with some plumbing rules.

Add the folling rule to your `lib/plumbing`.

	# Open TLG worklist in acme
	type is text
	data matches 'TLG[0-9]+'
	plumb to none
	plumb start window /bin/lyceum/LHelper $0

	# Open TLG text in acme
	type is text
	data matches 'ID:([0-9]+)'
	plumb to none
	plumb start window /bin/lyceum/WHelper $1


## Dependencies

### Build

- Go
- On Unix/Linux, `curl` is also required.

### Runtime

- Polytonic Greek fonts: e.g. [Gentium](https://software.sil.org/gentium) (A modified version is included in `fonts/` directory).

## Build

### Unix

Run `make`.

### 9front

Run `install.rc` and move `TLG-E`, `PHI-5` directories to `/sys/lib/lyceum`.

## Caveats & Bugs

- Plan9/Unix scripts contain hard-coded paths.
- TLG/PHI files must have lowercase filenames (including extensions).

## Links

- [Diogenes Desktop](https://d.iogen.es/d)
- [Perseus Project](https://www.perseus.tufts.edu/)
- [Source Repository](https://github.com/sitmsiteman/lyceum)

