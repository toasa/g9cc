package parse

import (
    . "g9cc/common"
    . "g9cc/util"
    "fmt"
    // "github.com/k0kubun/pp"
)

type Env struct {
    typedefs map[string]interface{}
    tags map[string]interface{}
    next *Env
}

var tokens *Vector
// "tokens.Data[]" array's index
var pos int = 0
var env *Env
var null_stmt Node = Node{Op: ND_NULL}

func new_env(next *Env) *Env {
    env := new(Env)
    env.typedefs = make(map[string]interface{})
    env.tags = make(map[string]interface{})
    env.next = next
    return env
}

func expect(ty int) {
    t, _ := tokens.Data[pos].(*Token)
    if t.Ty != ty {
        Error(fmt.Sprintf("%c (%d) expected, but got %c (%d)", ty, ty, t.Ty, t.Ty))
    }
    pos++
}

func new_prim_ty(ty, size int) *Type {
    ret := new(Type)
    ret.Ty = ty
    ret.Size = size
    ret.Align = size
    return ret
}

func void_ty() *Type {
    return new_prim_ty(VOID, 0)
}

func char_ty() *Type {
    return new_prim_ty(CHAR, 1)
}

func int_ty() *Type {
    return new_prim_ty(INT, 4)
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
    if t.Ty == TK_IDENT {
        _, ok := env.typedefs[t.Name]
        return ok
    }
    return t.Ty == TK_INT || t.Ty == TK_CHAR ||
    t.Ty == TK_STRUCT || t.Ty == TK_VOID
}

func read_type() *Type {
    t := tokens.Data[pos].(*Token)
    pos++

    if t.Ty == TK_IDENT {
        ty := env.typedefs[t.Name].(*Type)
        if ty == nil {
            pos--
        }
        return ty
    }
    if t.Ty == TK_INT {
        return int_ty()
    }
    if t.Ty == TK_CHAR {
        return char_ty()
    }
    if t.Ty == TK_VOID {
        return void_ty()
    }
    if t.Ty == TK_STRUCT {
        var tag string = ""
        t := tokens.Data[pos].(*Token)
        if t.Ty == TK_IDENT {
            pos++
            tag = t.Name
        }

        var members *Vector = nil
        if consume('{') {
            members = New_vec()
            for !consume('}') {
                Vec_push(members, decl())
            }
        }

        if tag == "" && members == nil {
            Error("bad struct definition")
        }

        if tag != "" && members != nil {
            env.tags[tag] = members
        } else if ( tag != "" && members == nil) {
            members = env.tags[tag].(*Vector)
            if members == nil {
                Error(fmt.Sprintf("incomplete type: %s", tag))
            }
        }

        return Struct_of(members)
    }

    pos--
    return nil
}

func new_binop(op int, lhs *Node, rhs *Node) *Node {
    n := new(Node) // new()関数でNode型のポインタを返す
    n.Op = op
    n.Lhs = lhs
    n.Rhs = rhs
    return n
}

func new_expr(op int, expr *Node) *Node {
    node := new(Node)
    node.Op = op
    node.Expr = expr
    return node
}

func ident() string {
    t := tokens.Data[pos].(*Token)
    pos++
    if t.Ty != TK_IDENT {
        Error(fmt.Sprintf("identifier expected, but got %s", t.Input))
    }
    return t.Name
}

func primary() *Node {
    t := tokens.Data[pos].(*Token)
    pos++

    if t.Ty == '(' {
        if consume('{') {
            node := new(Node)
            node.Op = ND_STMT_EXPR
            node.Body = compound_stmt()
            expect(')')
            return node
        }
        var node *Node = assign()
        expect(')')
        return node
    }

    node := new(Node)

    if t.Ty == TK_NUM {
        node.Ty = int_ty()
        node.Op = ND_NUM
        node.Val = t.Val
        return node
    }

    if t.Ty == TK_STR {
        // 文字列はchar型の配列として扱う
        node.Ty = Ary_of(char_ty(), len(t.Str))
        node.Op = ND_STR
        node.Data = t.Str
        node.Len = t.Len
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

func postfix() *Node {
    lhs := primary()

    for {
        if consume('.') {
            node := new(Node)
            node.Op = ND_DOT
            node.Expr = lhs
            node.Name = ident()
            lhs = node
            continue
        }

        if consume(TK_ARROW) {
            node := new(Node)
            node.Op = ND_DOT
            node.Expr = new_expr(ND_DEREF, lhs)
            node.Name = ident()
            lhs = node
            continue
        }

        if consume('[') {
            lhs = new_expr(ND_DEREF, new_binop('+', lhs, assign()))
            expect(']')
            continue
        }
        return lhs
    }
}

// 識別子の先頭につく'*' or '&'を読み取る
func unary() *Node {
    if consume('*') {
        return new_expr(ND_DEREF, mul())
    }
    if consume('&') {
        return new_expr(ND_ADDR, mul())
    }
    if consume(TK_SIZEOF) {
        return new_expr(ND_SIZEOF, unary())
    }
    if consume(TK_ALIGNOF) {
        return new_expr(ND_ALIGNOF, unary())
    }
    return postfix()
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
        lhs = new_binop(t.Ty, lhs, unary())
    }

    // ここには通常到達しない
    var err *Node
    return err
}

func add() *Node {

    var lhs *Node = mul()

    for true {
        t := tokens.Data[pos].(*Token)

        if !(t.Ty == '+' || t.Ty == '-') {
            return lhs
        }
        pos++
        lhs = new_binop(t.Ty, lhs, mul())
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
            lhs = new_binop('<', lhs, add())
            continue
        }
        if t.Ty == '>' {
            pos++
            lhs = new_binop('<', add(), lhs)
            continue
        }

        return lhs
    }

    err := new(Node)
    return err
}

func equality() *Node {
    lhs := rel()
    for true {
        t := tokens.Data[pos].(*Token)
        if t.Ty == TK_EQ {
            pos++
            lhs = new_binop(ND_EQ, lhs, rel())
            continue
        }
        if t.Ty == TK_NE {
            pos++
            lhs = new_binop(ND_NE, lhs, rel())
            continue
        }
        return lhs
    }

    err := new(Node)
    return err
}

func logand() *Node {
    var lhs *Node = equality()
    for true {
        t := tokens.Data[pos].(*Token)
        if t.Ty != TK_LOGAND {
            return lhs
        }
        pos++
        lhs = new_binop(ND_LOGAND, lhs, equality())
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
        lhs = new_binop(ND_LOGOR, lhs, logand())
    }

    err := new(Node)
    return err
}

// '='を処理する
func assign() *Node {
    lhs := logor()
    if consume('=') {
        // =文の場合
        return new_binop('=', lhs, logor())
    }
    // =文でない場合
    return lhs
}

// typeは予約語ゆえ
// 型宣言を読み取る. ex. int a, int **b,...
func type_() *Type {
    t := tokens.Data[pos].(*Token)
    ty := read_type()
    if ty == nil  {
        Error(fmt.Sprintf("typename expected, but got %s", t.Input))
    }

    for consume('*') {
        ty = Ptr_to(ty)
    }
    return ty
}

func read_array(ty *Type) *Type {
    v := New_vec()
    for consume('[') {
        len_ := primary()
        if len_.Op != ND_NUM {
            Error("number expected")
        }
        Vec_push(v, len_)
        expect(']')
    }
    for i := v.Len - 1; i >= 0; i-- {
        len_ := v.Data[i].(*Node)
        ty = Ary_of(ty, len_.Val)
    }
    return ty
}

func decl() *Node {
    node := new(Node)
    node.Op = ND_VARDEF

    // Read the first half of type name (e.g. `int *`)
    node.Ty = type_()

    // Read an indentifier
    node.Name = ident()

    // Read the second half of type name (e.g. `[3][5]`)
    node.Ty = read_array(node.Ty)
    if node.Ty.Ty == VOID {
        Error(fmt.Sprintf("void variable: %s", node.Name))
    }

    // Read an initializer
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
    node.Name = ident()
    return node
}

func expr_stmt() *Node {
    node := new_expr(ND_EXPR_STMT, assign())
    expect(';')
    return node
}

func stmt() *Node {
    node := new(Node)
    t := tokens.Data[pos].(*Token)

    switch t.Ty {
    case TK_TYPEDEF:
        pos++
        node := decl()
        Assert(node.Name != "", "")
        env.typedefs[node.Name] = node.Ty
        return &null_stmt
    case TK_INT, TK_CHAR, TK_STRUCT:
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
        if is_typename() {
            node.Init = decl()
        } else {
            node.Init = expr_stmt()
        }
        node.Cond = assign()
        expect(';')
        node.Inc = new_expr(ND_EXPR_STMT, assign())
        expect(')')
        node.Body = stmt()
        return node
    case TK_WHILE:
        pos++
        // while文はfor文の初期化とインクリメントがないものとして扱っている
        node.Op = ND_FOR
        node.Init = &null_stmt
        node.Inc = &null_stmt
        expect('(')
        node.Cond = assign()
        expect(')')
        node.Body = stmt()
        return node
    case TK_DO:
        pos++
        node.Op = ND_DO_WHILE
        node.Body = stmt()
        expect(TK_WHILE)
        expect('(')
        node.Cond = assign()
        expect(')')
        expect(';')
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
    case ';':
        pos++
        return &null_stmt
    default:
        if is_typename() {
            return decl()
        }
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

    env := new_env(env)
    for !consume('}') {
        // 関数の終わり"}"まで
        // 一文(;で終わる文)づつparseし、node.Stmtsにpushする
        Vec_push(node.Stmts, stmt())
    }
    env = env.next

    return node
}

func toplevel() *Node {
    is_extern := consume(TK_EXTERN)
    ty := type_()
    if ty == nil {
        t := tokens.Data[pos].(*Token)
        Error(fmt.Sprintf("typename expected, but got %s", t.Input))
    }

    name := ident()

    // Function
    if consume('(') {
        node := new(Node)
        node.Op = ND_FUNC
        node.Ty = ty
        node.Name = name
        node.Args = New_vec()

        if !consume(')') {
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

    // Global variable
    node := new(Node)
    node.Op = ND_VARDEF
    node.Ty = read_array(ty)
    node.Name = name

    if is_extern {
        node.Is_extern = true
    } else {
        // node.Data = ""
        node.Len = node.Ty.Size
    }

    expect(';')
    return node
}

func Parse(tokens_ *Vector) *Vector {
    tokens = tokens_

    pos = 0
    env = new_env(env)
    v := New_vec()

    for t := tokens.Data[pos].(*Token); t.Ty != TK_EOF; t = tokens.Data[pos].(*Token) {
        Vec_push(v, toplevel())
    }
    return v
}
