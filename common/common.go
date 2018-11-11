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

type Type struct {
    Ty int
    Ptr_of *Type // Pointer
    Ary_of *Type // Array
    Len int
}

// token.go

const (
    TK_NUM = iota + 256 // Number Literal
    TK_IDENT // Identifier
    TK_INT // "int"
    TK_CHAR // "char"
    TK_IF // "if"
    TK_ELSE // "else"
    TK_FOR // "for"
    TK_LOGOR // ||
    TK_LOGAND // &&
    TK_RETURN // "return"
    TK_SIZEOF // "sizeof"
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
    ND_VARDEF // Variable definition
    ND_LVAR // Variable reference
    ND_IF // "if"
    ND_FOR // "for"
    ND_ADDR // address operator ("&")
    ND_DEREF // pointer dereference ("*")
    ND_LOGAND // &&
    ND_LOGOR // ||
    ND_RETURN // "return"
    ND_SIZEOF // "sizeof"
    ND_CALL // Function call
    ND_FUNC // Function definition
    ND_COMP_STMT // Compound statement
    ND_EXPR_STMT // Expression statement
)

const (
    INT = iota
    CHAR
    PTR
    ARY
)

type Node struct {
    Op int // Node type
    Ty *Type // C type
    Lhs *Node // left-hand side
    Rhs *Node // right-hand side
    Val int // number literal
    Expr *Node // "return" or Expression statement
    Stmts *Vector // Compound statement

    Name string // Identifier

    // "if" (cond) then "else" els
    // "for" (init; cond; inc) body
    Cond *Node // condtion in IF stmt
    Then *Node
    Els *Node

    Init *Node
    Inc *Node
    Body *Node

    // Function definition
    Stacksize int

    // Local variable
    Offset int

    // Function call
    Args *Vector
}


// gen_ir.go

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
    IR_LT
    IR_JMP
    IR_UNLESS
    IR_LOAD8
    IR_LOAD32
    IR_LOAD64
    IR_STORE8
    IR_STORE32
    IR_STORE64
    IR_STORE8_ARG
    IR_STORE32_ARG
    IR_STORE64_ARG
    IR_KILL
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
    IR_TY_JMP
    IR_TY_LABEL
    IR_TY_REG_REG
    IR_TY_REG_IMM
    IR_TY_IMM_IMM
    IR_TY_REG_LABEL
    IR_TY_CALL
)

type IRInfo struct {
    Name string
    Ty int
}

type Function struct {
    Name string
    // Args [6]int
    Stacksize int
    Ir *Vector
}
