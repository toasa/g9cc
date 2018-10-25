package ir

import (
    . "g9cc/common"
    . "g9cc/util"
    "strings"
    "fmt"
)

var code *Vector
var regno int
var basereg int

var vars map[string]interface{}
var bpoff int

func add(op int, lhs int, rhs int) *IR {
    var ir *IR = new(IR)
    ir.Op = op
    ir.Lhs = lhs
    ir.Rhs = rhs
    Vec_push(code, ir)
    return ir
}

func add_imm(op, lhs, imm int) *IR {
    var ir *IR = new(IR)
    ir.Op = op
    ir.Lhs = lhs
    ir.Has_imm = true
    ir.Imm = imm
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

    add(IR_MOV, r, basereg)
    add_imm('+', r, off)
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

    var alloca *IR = add(IR_ALLOCA, basereg, -1)
    gen_stmt(node)
    alloca.Rhs = bpoff
    add(IR_KILL, basereg, -1)
    return code
}
