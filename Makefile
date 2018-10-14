g9cc: g9cc.go

test: g9cc
	./test.sh

clean:
	rm -f tmp*
