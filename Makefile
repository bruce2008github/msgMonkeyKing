all: msgMonkeyKing.go
	
	gofmt -w=true msgMonkeyKing.go

	go build -o msgMonkeyKing -gcflags "-N -l" msgMonkeyKing.go

clean:
	rm -f msgMonkeyKing
