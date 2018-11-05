package parse

import (
    . "g9cc/common"
    . "g9cc/util"
    "fmt"
)

var tokens *Vector
// "tokens.Data[]" array's index
var pos int = 0

var int_ty Type = Type{Ty: INT, Ptr_of: nil}

func ptr_of(base *Type) *Type {
    ty := new(Type)
    ty.Ty = PTR
    ty.Ptr_of = base
    return ty
}

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

func is_typename() bool {
    t := tokens.Data[pos].(*Token)
    return t.Ty == TK_INT
}

func new_node(op int, lhs *Node, rhs *Node) *Node {
    n := new(Node) // new()関数でNode型のポインタを返す
    n.Op = op
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
        node.Ty = &int_ty
        node.Op = ND_NUM
        node.Val = t.Val
        return node
    }

    if t.Ty == TK_IDENT {
        node.Name = t.Name

        if !consume('(') {
            // 識別子
            // '('の場合関数呼び出しとみなされ、pos++となり、このif文の条件はfalseとなる
            node.Op = ND_IDENT
            return node
        }

        // 関数呼び出し
        node.Op = ND_CALL
        node.Args = New_vec()
        if consume(')') {
            // 関数に引数がない場合
            return node
        }

        // 引数がある場合
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

func unary() *Node {
    if consume('*') {
        node := new(Node)
        node.Op = ND_DEREF
        node.Expr = mul()
        return node
    }

    return term()
}

func mul() *Node {
    var lhs *Node = unary()

    for true {
        t := tokens.Data[pos].(*Token)

        if !(t.Ty == '*' || t.Ty == '/') {
            return lhs
        }
        // t.Tyが * または　/ の場合
        pos++
        lhs = new_node(t.Ty, lhs, unary())
    }

    // ここには通常到達しない
    var err *Node
    return err
}

func add() *Node {

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
        t := tokens.Data[pos].(*Token)

        if !(t.Ty == '+' || t.Ty == '-') {
            return lhs
        }
        pos++
        lhs = new_node(t.Ty, lhs, mul())
    }

    // 通常ここには到達しない
    var err *Node
    return err
}

// 四則演算(mul(), add())が終わったところでrelを呼び,不等号のチェックを行う
func rel() *Node {
    var lhs *Node = add()
    for true {
        t := tokens.Data[pos].(*Token)
        if t.Ty == '<' {
            pos++
            lhs = new_node('<', lhs, add())
            continue
        }
        if t.Ty == '>' {
            pos++
            lhs = new_node('<', add(), lhs)
            continue
        }

        return lhs
    }

    err := new(Node)
    return err
}

func logand() *Node {
    var lhs *Node = rel()
    for true {
        t := tokens.Data[pos].(*Token)
        if t.Ty != TK_LOGAND {
            return lhs
        }
        pos++
        lhs = new_node(ND_LOGAND, lhs, rel())
    }

    err := new(Node)
    return err
}

func logor() *Node {
    var lhs *Node = logand()
    for true {
        t := tokens.Data[pos].(*Token)
        if t.Ty != TK_LOGOR {
            return lhs
        }
        pos++
        lhs = new_node(ND_LOGOR, lhs, logand())
    }

    err := new(Node)
    return err
}

func assign() *Node {
    lhs := logor()
    if consume('=') {
        // =文の場合
        return new_node('=', lhs, logor())
    }
    // =文でない場合
    return lhs
}

// typeは予約語ゆえ
func type_() *Type {
    t := tokens.Data[pos].(*Token)
    if t.Ty != TK_INT {
        Error(fmt.Sprintf("typename expected, but got %s", t.Input))
    }
    pos++

    ty := &int_ty
    for consume('*') {
        ty = ptr_of(ty)
    }
    return ty
}

func decl() *Node {
    node := new(Node)
    node.Op = ND_VARDEF
    node.Ty = type_()

    t := tokens.Data[pos].(*Token)
    if t.Ty != TK_IDENT {
        Error(fmt.Sprintf("variable name expected, but got %s", t.Input))
    }
    node.Name = t.Name
    pos++

    if consume('=') {
        node.Init = assign()
    }
    expect(';')
    return node
}

func param() *Node {
    node := new(Node)
    node.Op = ND_VARDEF
    node.Ty = type_()

    t := tokens.Data[pos].(*Token)
    if t.Ty != TK_IDENT {
        Error(fmt.Sprintf("parameter name expected, but got %s", t.Input))
    }
    node.Name = t.Name
    pos++
    return node
}

func expr_stmt() *Node {
    node := new(Node)
    node.Op = ND_EXPR_STMT
    node.Expr = assign()
    expect(';')
    return node
}

func stmt() *Node {
    node := new(Node)
    t := tokens.Data[pos].(*Token)

    switch t.Ty {
    case TK_INT:
        return decl()
    case TK_IF:
        pos++
        node.Op = ND_IF
        expect('(')
        node.Cond = assign()
        expect(')')

        node.Then = stmt()

        if consume(TK_ELSE) {
            node.Els = stmt()
        }
        return node
    case TK_FOR:
        pos++
        node.Op = ND_FOR
        expect('(')
        // node.Init = assign()
        // expect(';')
        if is_typename() {
            node.Init = decl()
        } else {
            node.Init = expr_stmt()
        }
        node.Cond = assign()
        expect(';')
        node.Inc = assign()
        expect(')')
        node.Body = stmt()
        return node
    case TK_RETURN:
        pos++
        node.Op = ND_RETURN
        node.Expr = assign()
        expect(';')
        return node
    case '{':
        pos++
        node.Op = ND_COMP_STMT
        node.Stmts = New_vec()
        for !consume('}') {
            Vec_push(node.Stmts, stmt())
        }
        return node
    default:
        return expr_stmt()
    }

    err := new(Node)
    return err
}

func compound_stmt() *Node {
    // ASTのroot node
    node := new(Node)
    node.Op = ND_COMP_STMT
    node.Stmts = New_vec()

    for !consume('}') {
        // 関数の終わり"}"まで
        // 一文(;で終わる文)づつparseし、node.Stmtsにpushする
        Vec_push(node.Stmts, stmt())
    }

    return node
}

func function() *Node {
    node := new(Node)
    node.Op = ND_FUNC
    node.Args = New_vec()

    t := tokens.Data[pos].(*Token)

    if t.Ty != TK_INT {
        Error(fmt.Sprintf("function return type expected, but got %s", t.Input))
    }
    pos++;

    t = tokens.Data[pos].(*Token)
    if t.Ty != TK_IDENT {
        Error(fmt.Sprintf("function name expected, but got %s", t.Input))
    }
    node.Name = t.Name
    pos++

    expect('(')
    if !consume(')') {
        // 引数が存在した場合
        // param()を呼び出すこと引数の変数宣言(ex. add(int x, int y){})を認めた
        Vec_push(node.Args, param())
        for consume(',') {
            Vec_push(node.Args, param())
        }
        expect(')')
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
