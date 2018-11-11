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

func Node2bool(n *Node) bool {
    if n == nil {
        return false
    } else {
        return true
    }
}

func Format(fmts ...string) string {
    var str string
    for _, fmt := range fmts {
        str += fmt + "\000"
    }
    return str
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
    case *Node:
        for i := 0; i < v.Len; i++ {
            fmt.Printf("%+v\n", v.Data[i])
        }
        fmt.Printf("=== END OF PRINT NODE ===\n\n")
    case *IR:
        for i := 0; i < v.Len; i++ {
            fmt.Printf("%+v\n", v.Data[i])
        }
        fmt.Printf("=== END OF PRINT IR ===\n\n")
    case *Function:
        for i := 0; i < v.Len; i++ {
            fmt.Printf("%+v\n", v.Data[i])
        }
        fmt.Printf("=== END OF PRINT FUNCTION ===\n\n")
    }
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


// StringBuilder
func New_sb() *StringBuilder {
    sb := new(StringBuilder)
    sb.Capacity = 8
    sb.Len = 0
    return sb
}

func Sb_grow(sb *StringBuilder, len int) {
    if sb.Len + len <= sb.Capacity {
        return
    }

    for sb.Len + len > sb.Capacity {
        sb.Capacity *= 2
    }
}

func Sb_append(sb *StringBuilder, s string) {
    Sb_grow(sb, len(s))
    sb.Data = s
    sb.Len += len(s)
}

func Sb_get(sb *StringBuilder) string {
    Sb_grow(sb, 1)
    //sb.Data[sb.Len] = '\000'
    return sb.Data
}

func Ptr_of(base *Type) *Type {
    ty := new(Type)
    ty.Ty = PTR
    ty.Ptr_of = base
    return ty
}

func Ary_of(base *Type, len_ int) *Type{
    ty := new(Type)
    ty.Ty = ARY
    ty.Ary_of = base
    ty.Len = len_
    return ty
}

func Size_of(ty *Type) int {
    if ty.Ty == CHAR {
        return 1
    }
    if ty.Ty == INT {
        return 4
    }
    if ty.Ty == PTR {
        return 8
    }
    Assert(ty.Ty == ARY, "ty.Ty is not ARY")
    return Size_of(ty.Ary_of) * ty.Len
}
