package common


// Vector
// 異なるデータ型*Token, *IRなどをスライスとして扱うための構造体(wrapperみたいなものか?)
type Vector struct {
    Data []interface{}
    Capacity int
    Len int
}


type Map struct {
    Keys *Vector
    Vals *Vector
}

// token.go

const (
    TK_NUM = iota + 256 // Number Literal
    TK_IDENT // Identifier
    TK_RETURN
    TK_EOF
)

// Tokenizer
type Token struct {
    Ty int // token type
    Val int // number literal
    Name string // identifier
    Input string // token string
}


// parse.go

const (
    ND_NUM = iota + 256 // number literal
    ND_IDENT // identifier
    ND_RETURN // Return statement
    ND_COMP_STMT // Compound statement
    ND_EXPR_STMT // Expression statement
)

type Node struct {
    Ty int // node type
    Lhs *Node // left-hand side
    Rhs *Node // right-hand side
    Val int // number literal
    Name string // Identifier
    Expr *Node // "return" or Expression statement
    Stmts *Vector // Compound statement
}


// ir.go

const (
    IR_IMM = iota // immediate value
    IR_MOV
    IR_RETURN
    IR_ALLOCA
    IR_LOAD
    IR_STORE
    IR_KILL
    IR_NOP
)

type IR struct {
    Op int
    Lhs int
    Rhs int
}
