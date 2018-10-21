package util

import (
    "fmt"
    "runtime"
    . "g9cc/common"
)

func expect(line, expected, actual int) {
    if expected == actual {
        return
    }
    Error(fmt.Sprintf("%d: %d expected, but got %d\n", line, expected, actual))
}

func Vec_test() {
    var vec *Vector = New_vec()
    _, _, line, ok := runtime.Caller(0)
    if !ok {
        fmt.Println("fail to get a stacktrace")
        return
    }
    expect(line, 0, vec.Len)

    for i := 0; i < 100; i++ {
        Vec_push(vec, i)
    }
    expect(line, 100, vec.Len)
    d, _ := vec.Data[0].(int)
    expect(line, 0, d)
    d, _ = vec.Data[50].(int)
    expect(line, 50, d)
    d, _ = vec.Data[99].(int)
    expect(line, 99, d)
}

func Map_test() {
    var m *Map = New_map()

    _, _, line, ok := runtime.Caller(0)
    if !ok {
        fmt.Println("fail to get a stacktrace")
        return
    }

    i, _ := Map_get(m, "foo").(int)
    expect(line, 0, i)

    Map_put(m, "foo", 2)
    i, _ = Map_get(m, "foo").(int)
    expect(line, 2, i)

    Map_put(m, "bar", 4)
    i, _ = Map_get(m, "bar").(int)
    expect(line, 4, i)

    Map_put(m, "foo", 6)
    i, _ = Map_get(m, "foo").(int)
    expect(line, 6, i)
}

func Util_test() {
    // Vec_test()
    Map_test()
}
