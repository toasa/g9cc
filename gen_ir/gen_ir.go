package ir

import (
    . "g9cc/common"
    . "g9cc/util"
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
    {"LT", IR_TY_REG_REG},
    {"JMP", IR_TY_JMP},
    {"UNLESS", IR_TY_REG_LABEL},
    {"LOAD", IR_TY_REG_REG},
    {"STORE", IR_TY_REG_REG},
    {"KILL", IR_TY_REG},
    {"SAVE_ARGS", IR_TY_IMM},
    {"NOP", IR_TY_NOARG},
}

var code *Vector
//
// // 各識別子のrbpからのoffsetを登録するための辞書
// var vars map[string]interface{}
// // 汎用レジスタの番号
var regno int
//
// var stacksize int
//
var label int

func tostr(ir *IR) string {
    info := Irinfo_arr[ir.Op]
    switch info.Ty {
    case IR_TY_LABEL:
        return fmt.Sprintf(".L%d:", ir.Lhs)
    case IR_TY_IMM:
        return fmt.Sprintf("  %s %d", info.Name, ir.Lhs)
    case IR_TY_REG:
        return fmt.Sprintf("  %s r%d", info.Name, ir.Lhs)
    case IR_TY_JMP:
        return fmt.Sprintf("  %s .L%d", info.Name, ir.Lhs)
    case IR_TY_REG_REG:
        return fmt.Sprintf("  %s r%d, r%d", info.Name, ir.Lhs, ir.Rhs)
    case IR_TY_REG_IMM:
        return fmt.Sprintf("  %s r%d, %d", info.Name, ir.Lhs, ir.Rhs)
    case IR_TY_REG_LABEL:
        return fmt.Sprintf("  %s r%d, .L%d", info.Name, ir.Lhs, ir.Rhs)
    case IR_TY_CALL:
        var sb string
        sb = fmt.Sprintf("  r%d = %s(", ir.Lhs, ir.Name)
        for i := 0; i < ir.Nargs; i++ {
            sb += fmt.Sprintf("r%d, ", ir.Args[i])
        }
        sb += ")"
        return sb + "\000"
    default:
        Assert(info.Ty == IR_TY_NOARG, "not IR_TY_NOARG")
        return fmt.Sprintf("  %s", info.Name)
    }
}

func Dump_ir(irv *Vector) {
    for i := 0; i < irv.Len; i++ {
        fn, _ := irv.Data[i].(*Function)

        fmt.Fprintf(os.Stderr, "%s():\n", fn.Name);
        for j := 0; j < fn.Ir.Len; j++ {
            ir := fn.Ir.Data[j].(*IR)
            fmt.Fprintf(os.Stderr, "%s\n", tostr(ir))
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

// cにおいて代入文の
func gen_lval(node *Node) int {
    // if node.Ty != ND_IDENT {
    //     Error("not an lvalue")
    // }
    //
    // _, ok := vars[node.Name]
    //
    // if !ok {
    //     Error(fmt.Sprintf("undefined variable: %s", node.Name))
    // }

    if node.Ty != ND_LVAR {
        Error(fmt.Sprintf("not lvalue: %d (%s)", node.Ty, node.Name))
    }
    var r int = regno
    regno++

    // r(現在の汎用レジスタ)にrbpを代入する
    add(IR_MOV, r, 0)
    // r(現在はrbpの値)からoffset(メモリ上にある識別子が, [rbp]からどれほど離れているか)
    // だけ減算する
    add(IR_SUB_IMM, r, node.Offset)
    return r
}

func gen_binop(ty int, lhs *Node, rhs *Node) int {
    r1 := gen_expr(lhs)
    r2 := gen_expr(rhs)
    add(ty, r1, r2)
    add(IR_KILL, r2, -1)
    return r1
}

func gen_expr(node *Node) int {

    switch node.Ty {
    case ND_NUM:
        r := regno
        regno++
        add(IR_IMM, r, node.Val)
        return r
    case ND_LOGAND:
        // return処理を行うラベルx
        x := label
        label++

        var r1 int = gen_expr(node.Lhs)
        // レジスタr1の値がfalse(0)ならばラベルxへ飛ぶ
        add(IR_UNLESS, r1, x)
        var r2 int = gen_expr(node.Rhs)
        add(IR_MOV, r1, r2)
        add(IR_KILL, r2, -1)
        // r2の値が0でもラベルxへjmp
        add(IR_UNLESS, r1, x)
        // r1, r2ともにfalse(0)ではなかったため、
        // 戻り値としてtrue(1)を返すためr1に1を代入する
        add(IR_IMM, r1, 1)
        add(IR_LABEL, x, -1)
        return r1
    case ND_LOGOR:
        // 中継先のラベル
        x := label
        label++
        // 最終的な行き先のラベル
        y := label
        label++

        var r1 int = gen_expr(node.Lhs)
        add(IR_UNLESS, r1, x)
        // r1値がfalse(0)でない場合, r1にtrue(1)を代入し,
        // y(最終的なラベル)に飛ぶ
        add(IR_IMM, r1, 1)
        add(IR_JMP, y, -1)

        // 中継先ラベルxの用意
        add(IR_LABEL, x, -1)

        var r2 int = gen_expr(node.Rhs)
        add(IR_MOV, r1, r2)
        add(IR_KILL, r2, -1)
        // r1値がtrue(0)でかつ, r2値がfalse(0)のときラベルyへ飛ぶ
        add(IR_UNLESS, r1, y)
        // r1値がtrue(0)でかつ, r2値もtrue(1)のため,
        // 戻り値としてtrue(1)を返す。そのためにr1に値1を代入する
        add(IR_IMM, r1, 1)
        add(IR_LABEL, y, -1)
        return r1
    case ND_LVAR:
        var r int = gen_lval(node)
        add(IR_LOAD, r, r)
        return r
    case ND_CALL:
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
    case '=':
        var rhs int = gen_expr(node.Rhs)
        // lhsはメモリへstoreするためのアドレスが格納されたレジスタ(の番号)が入っている
        var lhs int = gen_lval(node.Lhs)
        add(IR_STORE, lhs, rhs)
        add(IR_KILL, rhs, -1)
        return lhs
    case '+':
        return gen_binop(IR_ADD, node.Lhs, node.Rhs)
    case '-':
        return gen_binop(IR_SUB, node.Lhs, node.Rhs)
    case '*':
        return gen_binop(IR_MUL, node.Lhs, node.Rhs)
    case '/':
        return gen_binop(IR_DIV, node.Lhs, node.Rhs)
    case '<':
        return gen_binop(IR_LT, node.Lhs, node.Rhs)
    default:
        Assert(false, "unknown AST type")
    }

    err := 0
    return err
}

func gen_stmt(node *Node) {
    if node.Ty == ND_VARDEF {

        if node.Init == nil {
            return
        }

        var rhs int = gen_expr(node.Init)
        var lhs int = regno
        regno++
        // lhsにベースレジスタのアドレスを代入
        add(IR_MOV, lhs, 0)
        // ベースレジスタから、変数のオフセット分引く
        add(IR_SUB_IMM, lhs, node.Offset)
        // メモリ上のスタックで、左辺値(lhs)に対し、右辺値(rhs)を代入する
        add(IR_STORE, lhs, rhs)
        add(IR_KILL, lhs, -1)
        add(IR_KILL, rhs, -1)

        return
    }

    if node.Ty == ND_IF {
        if Node2bool(node.Els) {
            // else文がある場合
            x := label
            label++
            y := label
            label++
            r := gen_expr(node.Cond)
            // レジスタrの値がfalse(0)の場合, ラベルxへジャンプする
            add(IR_UNLESS, r, x)
            add(IR_KILL, r, -1)

            gen_stmt(node.Then)
            add(IR_JMP, y, -1)
            add(IR_LABEL, x, -1)

            gen_stmt(node.Els)
            add(IR_LABEL, y, -1)
        }

        x := label
        label++
        r := gen_expr(node.Cond)

        // レジスタrの値がfalse(0)ならばラベルxへ飛ぶ
        add(IR_UNLESS, r, x)
        add(IR_KILL, r, -1)

        gen_stmt(node.Then)

        add(IR_LABEL, x, -1)
        return
    }

    if node.Ty == ND_FOR {
        x := label
        label++
        y := label
        label++

        gen_stmt(node.Init)
        add(IR_LABEL, x, -1)
        r := gen_expr(node.Cond)
        add(IR_UNLESS, r, y)
        add(IR_KILL, r, -1)
        gen_stmt(node.Body)
        add(IR_KILL, gen_expr(node.Inc), -1)
        add(IR_JMP, x, -1)
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

// func gen_args(nodes *Vector) {
//     if nodes.Len == 0 {
//         return
//     }
//
//     add(IR_SAVE_ARGS, nodes.Len, -1)
//
//     // varsに各識別子のoffsetを登録する処理
//     for i := 0; i < nodes.Len; i++ {
//         node := nodes.Data[i].(*Node)
//         if node.Ty != ND_IDENT {
//             Error("bad parameter")
//         }
//
//         stacksize += 8
//         vars[node.Name] = stacksize
//     }
// }

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
        // 各関数ごとにregsiter numとstacksizeを初期化している.
        // regnoが1からはじまるのは、レジスタの配列Regsの0番目にrbpがあるから.
        regno = 1

        if nodes.Len > 0 {
            add(IR_SAVE_ARGS, node.Args.Len, -1)
        }
        gen_stmt(node.Body)

        fn := new(Function)
        fn.Name = node.Name
        fn.Stacksize = node.Stacksize
        fn.Ir = code
        Vec_push(v, fn)
    }
    return v
}
