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

func find_typedef(name string) *Type {
    for e := env; e != nil; e = e.next {
        _, ok := e.typedefs[name]
        if ok {
            return e.typedefs[name].(*Type)
        }
    }
    return nil
}

func find_tag(name string) *Type {
    for e := env; e != nil; e = e.next {
        _, ok := e.tags[name]
        if ok {
            return e.tags[name].(*Type)
        }
    }
    return nil
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
        return find_typedef(t.Name) != nil
    }
    return t.Ty == TK_INT || t.Ty == TK_CHAR ||
    t.Ty == TK_STRUCT || t.Ty == TK_VOID
}

func add_members(ty *Type, members *Vector) {
    off := 0
    for i := 0; i < members.Len; i++ {
        node := members.Data[i].(*Node)
        Assert(node.Op == ND_VARDEF, "node.Op is not ND_VARDEF")

        t := node.Ty
        off = Roundup(off, t.Align)
        t.Offset = off
        off += t.Size

        if ty.Align < node.Ty.Align {
            ty.Align = node.Ty.Align
        }
    }

    ty.Members = members
    ty.Size = Roundup(off, ty.Align)
}

func read_type() *Type {
    t := tokens.Data[pos].(*Token)
    pos++

    if t.Ty == TK_IDENT {
        ty := find_typedef(t.Name)
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

        var ty *Type = nil
        if tag != "" && members == nil {
            ty = find_tag(tag)
        }

        if ty == nil {
            ty = new(Type)
            ty.Ty = STRUCT
        }

        if members != nil {
            add_members(ty, members)
            if tag != "" {
                env.tags[tag] = ty
            }
        }
        return ty
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

// 演算子優先順位というものがあり、
// primaty()が最も高く、expr()が最も低い

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
        var node *Node = expr()
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
        if consume(TK_INC) {
            lhs = new_expr(ND_POST_INC, lhs)
            continue
        }

        if consume(TK_DEC) {
            lhs = new_expr(ND_POST_DEC, lhs)
            continue
        }
        if consume('.') {
            lhs = new_expr(ND_DOT, lhs)
            lhs.Name = ident()
            continue
        }

        if consume(TK_ARROW) {
            lhs = new_expr(ND_DOT, new_expr(ND_DEREF, lhs))
            lhs.Name = ident()
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
    if consume('-') {
        return new_expr(ND_NEG, unary())
    }
    if consume('*') {
        return new_expr(ND_DEREF, unary())
    }
    if consume('&') {
        return new_expr(ND_ADDR, unary())
    }
    if consume('!') {
        return new_expr('!', unary())
    }
    if consume(TK_INC) {
        return new_expr(ND_PRE_INC, unary())
    }
    if consume(TK_DEC) {
        return new_expr(ND_PRE_DEC, unary())
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
    for {
        if consume('*') {
            lhs = new_binop('*', lhs, unary())
        } else if consume('/') {
            lhs = new_binop('/', lhs, unary())
        } else if consume('%') {
            lhs = new_binop('%', lhs, unary())
        } else {
            return lhs
        }
    }
}

func add() *Node {

    lhs := mul()
    for {
        if consume('+') {
            lhs = new_binop('+', lhs, mul())
        } else if consume('-') {
            lhs = new_binop('-', lhs, mul())
        } else {
            return lhs
        }
    }
}

func shift() *Node {
    lhs := add()
    for {
        if consume(TK_SHL) {
            lhs = new_binop(ND_SHL, lhs, add())
        } else if consume(TK_SHR) {
            lhs = new_binop(ND_SHR, lhs, add())
        } else {
            return lhs
        }
    }
}

// 四則演算(mul(), add())が終わったところでrelを呼び,不等号のチェックを行う
func relational() *Node {
    var lhs *Node = shift()
    for {
        if consume('<') {
            lhs = new_binop('<', lhs, shift())
        } else if consume('>') {
            lhs = new_binop('<', shift(), lhs)
        } else if consume(TK_LE) {
            lhs = new_binop(ND_LE, lhs, shift())
        } else if consume(TK_GE) {
            lhs = new_binop(ND_LE, shift(), lhs)
        } else {
            return lhs
        }
    }
}

func equality() *Node {
    lhs := relational()
    for {
        if consume(TK_EQ) {
            lhs = new_binop(ND_EQ, lhs, relational())
        } else if consume(TK_NE) {
            lhs = new_binop(ND_NE, lhs, relational())
        } else {
            return lhs
        }
    }
}

func bit_and() *Node {
    lhs := equality()
    for consume('&') {
        lhs = new_binop('&', lhs, equality())
    }
    return lhs
}

func bit_xor() *Node {
    lhs := bit_and()
    for consume('^') {
        lhs = new_binop('^', lhs, bit_and())
    }
    return lhs
}

func bit_or() *Node {
    lhs := bit_xor()
    for consume('|') {
        lhs = new_binop('|', lhs, bit_xor())
    }
    return lhs
}

func logand() *Node {
    var lhs *Node = bit_or()
    for consume(TK_LOGAND) {
        lhs = new_binop(ND_LOGAND, lhs, bit_or())
    }
    return lhs
}

func logor() *Node {
    var lhs *Node = logand()
    for consume(TK_LOGOR) {
        lhs = new_binop(ND_LOGOR, lhs, logand())
    }
    return lhs
}

func conditional() *Node {
    cond := logor()
    if !consume('?') {
        return cond
    }

    node := new(Node)
    node.Op = '?'
    node.Cond = cond
    node.Then = expr()
    expect(':')
    node.Els = conditional()
    return node
}

// '='を処理する
func assign() *Node {
    lhs := conditional()
    if !consume('=') {
        return lhs
    }
    // '='文の場合
    return new_binop('=', lhs, conditional())
}

// 演算子優先順位の最下位
func expr() *Node {
    lhs := assign()
    if !consume(',') {
        return lhs
    }
    return new_binop(',', lhs, expr())
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
        len_ := expr()
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
    node := new_expr(ND_EXPR_STMT, expr())
    expect(';')
    return node
}

func stmt() *Node {
    node := new(Node)
    t := tokens.Data[pos].(*Token)
    pos++

    switch t.Ty {
    case TK_TYPEDEF:
        node := decl()
        Assert(node.Name != "", "")
        env.typedefs[node.Name] = node.Ty
        return &null_stmt
    // case TK_INT, TK_CHAR, TK_STRUCT:
    //     return decl()
    case TK_IF:
        node.Op = ND_IF
        expect('(')
        node.Cond = expr()
        expect(')')

        node.Then = stmt()

        if consume(TK_ELSE) {
            node.Els = stmt()
        }
        return node
    case TK_FOR:
        node.Op = ND_FOR
        expect('(')
        if is_typename() {
            node.Init = decl()
        } else {
            node.Init = expr_stmt()
        }
        node.Cond = expr()
        expect(';')
        node.Inc = new_expr(ND_EXPR_STMT, expr())
        expect(')')
        node.Body = stmt()
        return node
    case TK_WHILE:
        // while文はfor文の初期化とインクリメントがないものとして扱っている
        node.Op = ND_FOR
        node.Init = &null_stmt
        node.Inc = &null_stmt
        expect('(')
        node.Cond = expr()
        expect(')')
        node.Body = stmt()
        return node
    case TK_DO:
        node.Op = ND_DO_WHILE
        node.Body = stmt()
        expect(TK_WHILE)
        expect('(')
        node.Cond = expr()
        expect(')')
        expect(';')
        return node
    case TK_RETURN:
        node.Op = ND_RETURN
        node.Expr = expr()
        expect(';')
        return node
    case '{':
        node.Op = ND_COMP_STMT
        node.Stmts = New_vec()
        for !consume('}') {
            Vec_push(node.Stmts, stmt())
        }
        return node
    case ';':
        return &null_stmt
    default:
        pos--;
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
    is_typedef := consume(TK_TYPEDEF)
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
        if is_typedef {
            Error(fmt.Sprintf("typedef %s has function definition", name))
        }
        node.Body = compound_stmt()
        return node
    }

    ty = read_array(ty)
    expect(';')

    if is_typedef {
        env.typedefs[name] = ty
        return nil
    }

    // Global variable
    node := new(Node)
    node.Op = ND_VARDEF
    node.Ty = ty
    node.Name = name

    node.Is_extern = is_extern
    if !is_extern {
        // node.Data = ""
        node.Len = node.Ty.Size
    }
    return node
}

func Parse(tokens_ *Vector) *Vector {
    tokens = tokens_

    pos = 0
    env = new_env(env)
    v := New_vec()

    // for t := tokens.Data[pos].(*Token); t.Ty != TK_EOF; t = tokens.Data[pos].(*Token) {
    //     Vec_push(v, toplevel())
    // }
    // return v

    for {
        t := tokens.Data[pos].(*Token)
        if t.Ty == TK_EOF {
            return v
        }
        node := toplevel()
        if node != nil {
            Vec_push(v, node)
        }
    }
}
