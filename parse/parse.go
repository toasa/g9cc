package parse

import (
    . "g9cc/common"
    . "g9cc/util"
    "fmt"
)

var tokens *Vector
// "tokens.Data[]" array's index
var pos int = 0

func expect(ty int) {
    t, _ := tokens.Data[pos].(*Token)
    if t.Ty != ty {
        Error(fmt.Sprintf("%c (%d) expected, but got %c (%d)", ty, ty, t.Ty, t.Ty))
    }
    pos++
}

func consume(ty int) bool {
    t, _ := tokens.Data[pos].(*Token)
    if t.Ty != ty {
        return false
    } else {
        pos++
        return true
    }
}

func new_node(op int, lhs *Node, rhs *Node) *Node {
    n := new(Node) // new()関数でNode型のポインタを返す
    n.Ty = op
    n.Lhs = lhs
    n.Rhs = rhs

    return n
}

func term() *Node {
    node := new(Node)
    t, _ := tokens.Data[pos].(*Token)
    pos++

    if t.Ty == TK_NUM {
        node.Ty = ND_NUM
        node.Val = t.Val
        return node
    }

    if t.Ty == TK_IDENT {
        node.Ty = ND_IDENT
        node.Name = t.Name
        return node
    }

    Error(fmt.Sprintf("number expected, but got %s", t.Input))
    err := new(Node)
    return err
}

func mul() *Node {
    var lhs *Node = term()

    for true {
        t, ok := tokens.Data[pos].(*Token)
        if !ok {
            Error("Not *Token type is in tokens.data[]")
        }

        op := t.Ty
        if !(op == '*' || op == '/') {
            return lhs
        }
        pos++
        lhs = new_node(op, lhs, term())
    }

    // ここには通常到達しない
    var err *Node
    return err
}

// expression
func expr() *Node {

    // この文はすごい
    // 2 * 3 + 4, 2 + 3 * 4 から以下の構文木を作成した
    // 正しく、掛け算が優先されている
    //
    //       ---(+)---        ---(+)---
    //       |       |        |       |
    //    --(*)--    4        2    --(*)--
    //    |     |                  |     |
    //    2     3                  3     4

    var lhs *Node = mul()

    for true {
        t, ok := tokens.Data[pos].(*Token)
        if !ok {
            Error("Not *Token type is in tokens.data[]")
        }

        op := t.Ty
        if !(op == '+' || op == '-') {
            // 数値のみの式の場合この分岐となる。
            // 通常の計算式の場合はここには到達しない
            return lhs
        }
        pos++
        lhs = new_node(op, lhs, mul())
    }

    // 通常ここには到達しない
    var err *Node
    return err
}

func assign() *Node {
    lhs := expr()
    if consume('=') {
        return new_node('=', lhs, expr())
    }
    return lhs
}

func stmt() *Node {
    // ASTのroot node
    node := new(Node)
    node.Ty = ND_COMP_STMT
    node.Stmts = New_vec()

    for true {
        t, _ := tokens.Data[pos].(*Token)
        if t.Ty == TK_EOF {
            return node
        }

        // expression?
        e := new(Node)

        if t.Ty == TK_RETURN {
            pos++
            e.Ty = ND_RETURN
            e.Expr = assign()
        } else {
            e.Ty = ND_EXPR_STMT
            e.Expr = assign()
        }

        Vec_push(node.Stmts, e)
        expect(';')
    }

    err := new(Node)
    return err
}

func Parse(v *Vector) *Node {
    tokens = v

    return stmt()
}
