package util

import (
    "fmt"
    "os"
    . "g9cc/common"
)


// An error reporting function.
func Error(msgs ...string) {
    for _, msg := range msgs {
        fmt.Println(msg)
    }
    os.Exit(1)
}

func Assert(b bool, msg string) {
    if !b {
        Error(msg)
    }
}

func Int2bool(n int) bool {
    if n == 0 {
        return false
    } else {
        return true
    }
}


// Vector

func New_vec() *Vector {
    var v *Vector = new(Vector)
    v.Capacity = 16
    v.Len = 0
    v.Data = make([]interface{}, v.Capacity)
    return v
}

func Vec_push(v *Vector, elem interface{}) {
    if v.Len == v.Capacity {
        v.Capacity *= 2
        // v.dataの容量を増やすための処理
        for i := 0; i < v.Capacity; i++ {
            var a interface{}
            v.Data = append(v.Data, a)
        }
    }
    v.Data[v.Len] = elem
    v.Len++
}

func PrintVector(v *Vector) {
    switch v.Data[0].(type) {
    case *Token:
        for i := 0; i < v.Len; i++ {
            fmt.Printf("%+v\n", v.Data[i])
        }
        fmt.Printf("=== END OF PRINT TOKEN ===\n\n")

    case *IR:
        for i := 0; i < v.Len; i++ {
            fmt.Printf("%+v\n", v.Data[i])
        }
        fmt.Printf("=== END OF PRINT IR ===\n\n")
    }
}

// Node(未完)
func PrintAST(node *Node) {

    for i := 0; i < node.Stmts.Len; i++ {
        n, _ := node.Stmts.Data[i].(*Node)
        fmt.Printf("%+v\n", n)
    }

    fmt.Printf("=== END OF PRINT AST ===\n\n")
}


// Map
// Goの標準機能のmapを使用しているため、以下の関数は不要

func New_map() *Map {
    m := new(Map)
    m.Keys = New_vec()
    m.Vals = New_vec()

    return m
}

func Map_put(m *Map, key string, val interface{}) {
    Vec_push(m.Keys, key)
    Vec_push(m.Vals, val)
}

func Map_get(m *Map, key string) interface{} {
    for i := m.Keys.Len - 1; i >= 0; i-- {
        // str, _ := mKeys.Data[i].(string)
        if m.Keys.Data[i] == key {
            return m.Vals.Data[i]
        }
    }

    return nil
}

func Map_exists(m *Map, key string) bool {
    for i := 0; i < m.Keys.Len; i++ {
        if m.Keys.Data[i] == key {
            return true
        }
    }
    return false
}
