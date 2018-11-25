g9cc: main.go

test: g9cc
	./test.sh

test2: g9cc
	./test2.sh

clean:
	rm -f tmp*
