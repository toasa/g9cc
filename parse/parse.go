package parse

import (
    . "g9cc/common"
    "fmt"
)

var tokens *Vector
// "tokens" array's index
var pos int = 0

func new_node(op int, lhs *Node, rhs *Node) *Node {
    n := new(Node) // new()関数でNode型のポインタを返す
    n.Ty = op
    n.Lhs = lhs
    n.Rhs = rhs

    return n
}

func new_node_num(val int) *Node {
    n := new(Node)
    n.Ty = TK_NUM
    n.Val = val
    return n
}

func number() *Node {
    t, ok := tokens.Data[pos].(*Token)
    if !ok {
        Error("Not *Token type is in tokens.data[]")
    }

    if t.Ty != TK_NUM {
        Error(fmt.Sprintf("number expected, but got %s", t.Input))
    }
    pos++

    return new_node_num(t.Val)
}

// expression
func expr() *Node {
    var lhs *Node = number()

    // この文はすごい
    // 2 + 3 - 4から以下の構文木を作成した
    //
    //       ---(-)---
    //       |       |
    //    --(+)--    4
    //    |     |
    //    2     3

    for true {
        t, ok := tokens.Data[pos].(*Token)
        if !ok {
            Error("Not *Token type is in tokens.data[]")
        }

        op := t.Ty
        if !(op == '+' || op == '-') {
            break
        }
        pos++
        lhs = new_node(op, lhs, number())
    }

    t, ok := tokens.Data[pos].(*Token)
    if !ok {
        Error("Not *Token type is in tokens.data[]")
    }

    if t.Ty != TK_EOF {
        Error(fmt.Sprintf("stray token: %s", t.Input))
    }

    return lhs
}

func Parse(v *Vector) *Node {
    tokens = v
    pos = 0
    return expr()
}
