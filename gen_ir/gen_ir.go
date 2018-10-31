package ir

import (
    . "g9cc/common"
    . "g9cc/util"
    "strings"
    "fmt"
    "os"
)

// Compile AST to intermediate code that has infinite number of registers.
// Base pointer is always assigned to r0(notation of -dump-ir).

// 中の要素の順序はcommon.go内のirのconstと一致させる
var Irinfo_arr []IRInfo = []IRInfo{
    // name, ty
    {"ADD", IR_TY_REG_REG},
    {"SUB", IR_TY_REG_REG},
    {"MUL", IR_TY_REG_REG},
    {"DIV", IR_TY_REG_REG},
    {"MOV", IR_TY_REG_IMM},
    {"SUB", IR_TY_REG_IMM},
    {"MOV", IR_TY_REG_REG},
    {"RET", IR_TY_REG},
    {"CALL", IR_TY_CALL},
    {"", IR_TY_LABEL},
    {"JMP", IR_TY_LABEL},
    {"UNLESS", IR_TY_REG_LABEL},
    {"LOAD", IR_TY_REG_REG},
    {"STORE", IR_TY_REG_REG},
    {"KILL", IR_TY_REG},
    {"SAVE_ARGS", IR_TY_IMM},
    {"NOP", IR_TY_NOARG},
}

var code *Vector

// 各識別子のrbpからのoffsetを登録するための辞書
var vars map[string]interface{}
// 汎用レジスタの番号
var regno int

var stacksize int

var label int

func tostr(ir *IR) string {
    info := Irinfo_arr[ir.Op]
    switch info.Ty {
    case IR_TY_LABEL:
        return fmt.Sprintf(".L%d:", ir.Lhs)
    case IR_TY_IMM:
        return fmt.Sprintf("%s %d", info.Name, ir.Lhs)
    case IR_TY_REG:
        return fmt.Sprintf("%s r%d", info.Name, ir.Lhs)
    case IR_TY_REG_REG:
        return fmt.Sprintf("%s r%d, r%d", info.Name, ir.Lhs, ir.Rhs)
    case IR_TY_REG_IMM:
        return fmt.Sprintf("%s r%d, %d", info.Name, ir.Lhs, ir.Rhs)
    case IR_TY_REG_LABEL:
        return fmt.Sprintf("%s r%d, .L%d", info.Name, ir.Lhs, ir.Rhs)
    case IR_TY_CALL:
        var sb string
        sb = fmt.Sprintf("r%d = %s(", ir.Lhs, ir.Name)
        for i := 0; i < ir.Nargs; i++ {
            sb += fmt.Sprintf("r%d, ", ir.Args[i])
        }
        sb += ")"
        return sb + "\000"
    default:
        Assert(info.Ty == IR_TY_NOARG, "not IR_TY_NOARG")
        return fmt.Sprintf("%s", info.Name)
    }
}

func Dump_ir(irv *Vector) {
    for i := 0; i < irv.Len; i++ {
        fn, _ := irv.Data[i].(*Function)

        fmt.Fprintf(os.Stderr, "%s():\n", fn.Name);
        for j := 0; j < fn.Ir.Len; j++ {
            ir := fn.Ir.Data[j].(*IR)
            fmt.Fprintf(os.Stderr, "  %s\n", tostr(ir))
        }
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

// 左辺値には識別子がくる
func gen_lval(node *Node) int {
    if node.Ty != ND_IDENT {
        Error("not an lvalue")
    }

    _, ok := vars[node.Name]

    if !ok {
        // varsに識別子の登録がされていない場合
        // 識別子をメモリ上へstoreしたり、メモリからloadする時、
        // base pointerからの距離を, map varsに格納しておく。
        stacksize += 8
        vars[node.Name] = stacksize
    }

    var r int = regno
    regno++
    off, _ := vars[node.Name].(int)

    // r(現在の汎用レジスタ)にrbpを代入する
    add(IR_MOV, r, 0)
    // r(現在はrbpの値)からoffset(メモリ上にある識別子が, [rbp]からどれほど離れているか)
    // だけ減算する
    add(IR_SUB_IMM, r, off)
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

    if node.Ty == ND_CALL {
        var args [6]int
        for i := 0; i < node.Args.Len; i++ {
            // 関数に引数がある場合
            arg, _ := node.Args.Data[i].(*Node)
            args[i] = gen_expr(arg)
        }

        r := regno
        regno++

        var ir *IR = add(IR_CALL, r, -1)
        ir.Name = node.Name
        ir.Nargs = node.Args.Len
        ir.Args = args

        for i := 0; i < ir.Nargs; i++ {
            add(IR_KILL, ir.Args[i], -1)
        }
        return r
    }

    if node.Ty == '=' {
        var rhs int = gen_expr(node.Rhs)
        // lhsはメモリへstoreするためのアドレスが格納されたレジスタ(の番号)が入っている
        var lhs int = gen_lval(node.Lhs)
        add(IR_STORE, lhs, rhs)
        add(IR_KILL, rhs, -1)
        return lhs
    }

    Assert((strings.Contains("+-*/", string(node.Ty))), "operator expected")

    var ty int
    if node.Ty == '+' {
        ty = IR_ADD
    } else if node.Ty == '-' {
        ty = IR_SUB
    } else if node.Ty == '*' {
        ty = IR_MUL
    } else if node.Ty == '/' {
        ty = IR_DIV
    }

    // lhs, rhsどちらも数値が格納されている
    var lhs int = gen_expr(node.Lhs)
    var rhs int = gen_expr(node.Rhs)

    add(ty, lhs, rhs)
    add(IR_KILL, rhs, -1)

    return lhs
}

func gen_stmt(node *Node) {

    if node.Ty == ND_IF {
        r := gen_expr(node.Cond)
        x := label
        label++

        // jmp命令での飛び先のラベルxをRhsに格納する
        add(IR_UNLESS, r, x)
        add(IR_KILL, r, -1)

        gen_stmt(node.Then)

        if !(Node2bool(node.Els)) {
            // else文がない場合
            add(IR_LABEL, x, -1)
            return
        }

        y := label
        label++
        add(IR_JMP, y, -1)
        add(IR_LABEL, x, -1)
        gen_stmt(node.Els)
        add(IR_LABEL, y, -1)
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

func gen_args(nodes *Vector) {
    if nodes.Len == 0 {
        return
    }

    add(IR_SAVE_ARGS, nodes.Len, -1)

    // varsに各識別子のoffsetを登録する処理
    for i := 0; i < nodes.Len; i++ {
        node := nodes.Data[i].(*Node)
        if node.Ty != ND_IDENT {
            Error("bad parameter")
        }

        stacksize += 8
        vars[node.Name] = stacksize
    }
}

func Gen_ir(nodes *Vector) *Vector{

    v := New_vec()

    // ===変数名(型)===
    // v(*Vector)
    // | - Data([]*Function)-
    // |                    | - Name(string)
    // |                    | - Args([6]int)
    // |                    | - Ir(*Vector) -
    // |                                    | - code(*Vector)-
    // |                                                     | - Data([]*IR): ここに関数の中身の
    // |                                                                      中間表現が格納される
    // |
    // |


    for i := 0; i < nodes.Len; i++ {
        node := nodes.Data[i].(*Node)
        Assert(node.Ty == ND_FUNC, "Type of root node is not ND_FUNC")

        // fn.Irに使用
        code = New_vec()
        vars = make(map[string]interface{})
        // 各関数ごとにregsiter numとstacksizeを初期化している.
        // regnoが1からはじまるのは、レジスタの配列Regsの0番目にrbpがあるから.
        regno = 1
        stacksize = 0

        gen_args(node.Args)
        gen_stmt(node.Body)

        fn := new(Function)
        fn.Name = node.Name
        fn.Stacksize = stacksize
        fn.Ir = code
        Vec_push(v, fn)
    }
    return v
}
