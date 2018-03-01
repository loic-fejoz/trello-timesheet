all: build

cmd/trello2csv/trello2csv: cmd/trello2csv/main.go
	cd cmd/trello2csv && go build

build:
	cd cmd/trello2csv && go build

build-dep:
	go get ./...

clean:
	rm -f ./cmd/trello2csv/trello2csv

test: cmd/trello2csv/trello2csv
	./cmd/trello2csv/trello2csv --appKey=`cat key.txt` --token=`cat token.txt` --board="Br3L4U2M" --finish="Fini"
