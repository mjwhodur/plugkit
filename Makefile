0-plugin-test:
	go build -o host ./examples/0-plugin-test/client
	go build -o plugin ./examples/0-plugin-test/plug
	./host