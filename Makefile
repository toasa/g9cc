g9cc: main.go

test: g9cc
	./test.sh

clean:
	rm -f tmp*
