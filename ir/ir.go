package ir

import (
    . "g9cc/common"
    . "g9cc/util"
    "strings"
    "fmt"
)

var irinfo []IRInfo = []IRInfo{
    // op, name, ty
    {'+', "+", IR_TY_REG_REG},
    {'-', "-", IR_TY_REG_REG},
    {'*', "*", IR_TY_REG_REG},
    {'/', "/", IR_TY_REG_REG},
    {IR_IMM, "MOV", IR_TY_REG_IMM},
    {IR_ADD_IMM, "ADD", IR_TY_REG_IMM},
    {IR_MOV, "MOV", IR_TY_REG_REG},
    {IR_LABEL, "", IR_TY_LABEL},
    {IR_UNLESS, "UNLESS", IR_TY_REG_LABEL},
    {IR_RETURN, "RET", IR_TY_REG},
    {IR_ALLOCA, "ALLOCA", IR_TY_REG_IMM},
    {IR_LOAD, "LOAD", IR_TY_REG_REG},
    {IR_STORE, "STORE", IR_TY_REG_REG},
    {IR_KILL, "KILL", IR_TY_REG},
    {IR_NOP, "NOP", IR_TY_NOARG},
    {0, "", 0},
}

var code *Vector
var regno int
var basereg int

var vars map[string]interface{}
var bpoff int

var label int

func Get_irinfo(ir *IR) *IRInfo {
    for _, info := range irinfo {
        if info.Op == ir.Op {
            return &info
        }
    }
    Error("invalid instruction")

    err := new(IRInfo)
    return err
}

func tostr(ir *IR) string {
    info := Get_irinfo(ir)
    switch info.Ty {
    case IR_TY_LABEL:
        return fmt.Sprintf(".L%d:", ir.Lhs)
    case IR_TY_REG:
        return fmt.Sprintf("%s r%d", info.Name, ir.Lhs)
    case IR_TY_REG_REG:
        return fmt.Sprintf("%s r%d, r%d", info.Name, ir.Lhs, ir.Rhs)
    case IR_TY_REG_IMM:
        return fmt.Sprintf("%s r%d, %d", info.Name, ir.Lhs, ir.Rhs)
    case IR_TY_REG_LABEL:
        return fmt.Sprintf("%s r%d, .L%d", info.Name, ir.Lhs, ir.Rhs)
    default:
        Assert(info.Ty == IR_TY_NOARG, "not IR_TY_NOARG")
        return fmt.Sprintf("%s", info.Name)
    }
}

func Dump_ir(irv *Vector) {
    for i := 0; i < irv.Len; i++ {
        ir, _ := irv.Data[i].(*IR)
        fmt.Printf("%s\n", tostr(ir))
    }
}

func add(op int, lhs int, rhs int) *IR {
    var ir *IR = new(IR)
    ir.Op = op
    ir.Lhs = lhs
    ir.Rhs = rhs
    Vec_push(code, ir)
    return ir
}

func gen_lval(node *Node) int {
    if node.Ty != ND_IDENT {
        Error("not an lvalue")
    }

    _, ok := vars[node.Name]
    // varsに識別子の登録がされていない場合、bp(ベースポインタ)のオフセットを8上げる
    if !ok {
        vars[node.Name] = bpoff
        bpoff += 8
    }

    var r int = regno
    regno++
    off, _ := vars[node.Name].(int)

    // r(現在の汎用レジスタ)にbaseregを代入する
    add(IR_MOV, r, basereg)
    // r(現在の汎用レジスタ)に値offを加算する
    add(IR_ADD_IMM, r, off)
    return r
}

func gen_expr(node *Node) int {

    if node.Ty == ND_NUM {
        var r int = regno
        regno++
        add(IR_IMM, r, node.Val)
        return r;
    }

    if node.Ty == ND_IDENT {
        var r int = gen_lval(node)
        add(IR_LOAD, r, r)
        return r
    }

    if node.Ty == '=' {
        var rhs int = gen_expr(node.Rhs)
        var lhs int = gen_lval(node.Lhs)
        add(IR_STORE, lhs, rhs)
        add(IR_KILL, rhs, -1)
        return lhs
    }

    Assert((strings.Contains("+-*/", string(node.Ty))), "operator expected")

    // lhs, rhsどちらも数値が格納されている
    var lhs int = gen_expr(node.Lhs)
    var rhs int = gen_expr(node.Rhs)

    add(node.Ty, lhs, rhs)
    add(IR_KILL, rhs, -1)

    return lhs
}

func gen_stmt(node *Node) {

    if node.Ty == ND_IF {
        r := gen_expr(node.Cond)
        x := label
        label++
        add(IR_UNLESS, r, x)
        add(IR_KILL, r, -1)
        gen_stmt(node.Then)
        add(IR_LABEL, x, -1)
        return
    }

    if node.Ty == ND_RETURN {
        r := gen_expr(node.Expr)
        add(IR_RETURN, r, -1)
        add(IR_KILL, r, -1)
        return
    }

    if node.Ty == ND_EXPR_STMT {
        r := gen_expr(node.Expr)
        add(IR_KILL, r, -1)
        return
    }

    if node.Ty == ND_COMP_STMT {
        for i := 0; i < node.Stmts.Len; i++ {
            n, _ := node.Stmts.Data[i].(*Node)
            gen_stmt(n)
        }
        return
    }

    Error(fmt.Sprintf("unknown node: %d", node.Ty))
}

// ASTを引数にとり、中間表現を返す
// irは {op: , lhs: , rhs : }からなる
// op = numのとき => {IR_IMM, register_index, num_value}
// op = '+'のとき => {'+', lhsの値が格納されたregisterのindex, rhsの値が格納されたregisterのindex}
// opが'+'or'-'の直後 => {IR_KILL, rhsの値が格納されたregisterのindex, 0}
// ここで決めたregisterのindexは確定ではなく, alloc_regs()で配列insを一つひとつ読みながら
// 最終的なregister を決定する
func Gen_ir(node *Node) *Vector{
    Assert(node.Ty == ND_COMP_STMT, "Type of root node is not ND_COMP_STMT")
    code = New_vec()
    regno = 1
    basereg = 0
    vars = make(map[string]interface{})
    bpoff = 0
    label = 0

    var alloca *IR = add(IR_ALLOCA, basereg, -1)
    gen_stmt(node)
    alloca.Rhs = bpoff
    add(IR_KILL, basereg, -1)
    return code
}
