all:
	go build -o bin/indexer ./cmd/indexer
	go build -o bin/search ./cmd/search
	go build -o bin/tlgviewer ./cmd/tlgviewer
	go build -o bin/readauth ./cmd/readauth
	go build -o bin/lemmata ./cmd/lemmata
	cp scripts/linux/* bin/
	cd dependencies && ../bin/indexer -f grc.lsj.xml -o lsj.idt && ../bin/indexer -f lat.ls.perseus-eng1.xml -o ls.idt

index:
	cd dependencies && ../bin/indexer -f grc.lsj.xml -o lsj.idt && ../bin/indexer -f lat.ls.perseus-eng1.xml -o ls.idt

clean:
	rm -rf bin/

