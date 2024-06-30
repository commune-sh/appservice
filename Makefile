all: clean build 
build: 
	cd cmd/commune;go build -o ../../bin/commune
vendor: clean vendorbuild 
vendorbuild:
	go build -mod=vendor -o bin/commune cmd/commune/main.go
clean: 
	rm -f bin/commune;
deps:
	-go install github.com/cortesi/modd/cmd/modd@latest;
