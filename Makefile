# make g9cc はなにも起こらない(現在のディレクトリにmain.goが存在することを確かめる)
g9cc: main.go

# make testの前に, make g9ccを実行する
test: g9cc
	./test.sh

clean:
	rm -f tmp*
