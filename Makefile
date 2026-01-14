all:
	go build -o bin/indexer ./cmd/indexer
	go build -o bin/search ./cmd/search
	go build -o bin/tlgviewer ./cmd/tlgviewer
	go build -o bin/readauth ./cmd/readauth
	go build -o bin/lemmata ./cmd/lemmata

	cp scripts/linux/* bin/

	cd dependencies && ../bin/indexer
index:
	bin/indexer dependencies/grc.lsj.xml

clean:
	rm -rf bin/

