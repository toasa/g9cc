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
    TK_EOF
)

// Tokenizer
type Token struct {
    Ty int // token type
    Val int // number literal
    Input string // token string
}


// parse.go

const (
    ND_NUM = 256
)

type Node struct {
    Ty int // node type
    Lhs *Node // left-hand side
    Rhs *Node // right-hand side
    Val int // number literal
}


// ir.go

const (
    IR_IMM = iota // immediate value
    IR_MOV
    IR_RETURN
    IR_KILL
    IR_NOP
)

type IR struct {
    Op int
    Lhs int
    Rhs int
}
