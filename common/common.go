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

type StringBuilder struct {
    Data string
    Capacity int
    Len int
}

// token.go

const (
    TK_NUM = iota + 256 // Number Literal
    TK_IDENT // Identifier
    TK_IF // "if"
    TK_ELSE // "else"
    TK_RETURN // "return"
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
    ND_IF // "if"
    ND_RETURN // "return"
    ND_CALL // Function call
    ND_FUNC // Function definition
    ND_COMP_STMT // Compound statement
    ND_EXPR_STMT // Expression statement
)

type Node struct {
    Ty int // node type
    Lhs *Node // left-hand side
    Rhs *Node // right-hand side
    Val int // number literal
    Expr *Node // "return" or Expression statement
    Stmts *Vector // Compound statement

    Name string // Identifier

    // "if"
    Cond *Node // condtion in IF stmt
    Then *Node
    Els *Node

    // Function definition
    Body *Node

    // Function call
    Args *Vector
}


// ir.go

const (
    IR_ADD = iota
    IR_SUB
    IR_MUL
    IR_DIV
    IR_IMM
    IR_SUB_IMM
    IR_MOV
    IR_RETURN
    IR_CALL
    IR_LABEL
    IR_JMP
    IR_UNLESS
    IR_LOAD
    IR_STORE
    IR_KILL
    IR_SAVE_ARGS
    IR_NOP
)

type IR struct {
    Op int
    Lhs int
    Rhs int

    // Function call
    Name string
    Nargs int
    Args [6]int
}

const (
    IR_TY_NOARG = iota
    IR_TY_REG
    IR_TY_IMM
    IR_TY_LABEL
    IR_TY_REG_REG
    IR_TY_REG_IMM
    IR_TY_REG_LABEL
    IR_TY_CALL
)

type IRInfo struct {
    Op int
    Name string
    Ty int
}

type Function struct {
    Name string
    // Args [6]int
    Stacksize int
    Ir *Vector
}
