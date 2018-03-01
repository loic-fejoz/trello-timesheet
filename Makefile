all: build

cmd/trello2csv/trello2csv: cmd/trello2csv/main.go
	cd cmd/trello2csv && go build

build:
	cd cmd/trello2csv && go build
	cd cmd/csv2projects && go build

build-dep:
	go get ./...

clean:
	rm -f ./cmd/trello2csv/trello2csv

%.lst: cmd/trello2csv/trello2csv
	./cmd/trello2csv/trello2csv --appKey=`cat key.txt` --token=`cat token.txt` --board="Br3L4U2M" --finish="Fini" > $@

%.csv: %.lst cmd/csv2projects/csv2projects
	./cmd/csv2projects/csv2projects --filename $<

test: loic.csv
