toylang: parser/*.go ast/*.go main.go vm/*.go builtins/*.go
	go build

run: toylang
	./toylang

test: toylang
	go test

parser/nstrm.peg.go: parser/nstrm.peg
	(cd parser;make)
