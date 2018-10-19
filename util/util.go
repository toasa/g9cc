package common

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

// Vector
// 異なるデータ型*Token, *IRなどをスライスとして扱うための構造体(wrapperみたいなものか?)
type Vector struct {
    Data []interface{}
    Capacity int
    Len int
}

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
