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
    if t.Ty == ty {
        pos++
        return true
    } else {
        return false
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
    t, _ := tokens.Data[pos].(*Token)
    pos++

    if t.Ty == '(' {
        var node *Node = assign()
        expect(')')
        return node
    }

    node := new(Node)

    if t.Ty == TK_NUM {
        node.Ty = ND_NUM
        node.Val = t.Val
        return node
    }

    if t.Ty == TK_IDENT {
        node.Name = t.Name

        if !consume('(') {
            // 識別子
            node.Ty = ND_IDENT
            return node
        }

        // 関数呼び出し
        node.Ty = ND_CALL
        node.Args = New_vec()
        if consume(')') {
            return node
        }

        Vec_push(node.Args, assign())
        for consume(',') {
            Vec_push(node.Args, assign())
        }
        expect(')')
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
    node := new(Node)
    t, _ := tokens.Data[pos].(*Token)

    switch t.Ty {
    case TK_IF:
        pos++
        node.Ty = ND_IF
        expect('(')
        node.Cond = assign()
        expect(')')

        node.Then = stmt()

        if consume(TK_ELSE) {
            node.Els = stmt()
        }
        return node
    case TK_RETURN:
        pos++
        node.Ty = ND_RETURN
        node.Expr = assign()
        expect(';')
        return node
    default:
        node.Ty = ND_EXPR_STMT
        node.Expr = assign()
        expect(';')
        return node
    }

    err := new(Node)
    return err
}

func compound_stmt() *Node {
    // ASTのroot node
    node := new(Node)
    node.Ty = ND_COMP_STMT
    node.Stmts = New_vec()

    for !consume('}') {
        Vec_push(node.Stmts, stmt())
    }

    return node
}

func function() *Node {
    node := new(Node)
    node.Ty = ND_FUNC
    node.Args = New_vec()

    t := tokens.Data[pos].(*Token)
    if t.Ty != TK_IDENT {
        Error(fmt.Sprintf("function name expected, but got %s", t.Input))
    }
    node.Name = t.Name
    pos++

    expect('(')
    for !consume(')') {
        Vec_push(node.Args, term())
    }
    expect('{')
    node.Body = compound_stmt()
    return node
}

func Parse(tokens_ *Vector) *Vector {
    tokens = tokens_

    pos = 0
    v := New_vec()

    for t := tokens.Data[pos].(*Token); t.Ty != TK_EOF; t = tokens.Data[pos].(*Token) {
        Vec_push(v, function())
    }
    return v
}
