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
    {"LOAD8", IR_TY_REG_REG},
    {"LOAD32", IR_TY_REG_REG},
    {"LOAD64", IR_TY_REG_REG},
    {"STORE8", IR_TY_REG_REG},
    {"STORE32", IR_TY_REG_REG},
    {"STORE64", IR_TY_REG_REG},
    {"STORE8_ARG", IR_TY_IMM_IMM},
    {"STORE32_ARG", IR_TY_IMM_IMM},
    {"STORE64_ARG", IR_TY_IMM_IMM},
    {"KILL", IR_TY_REG},
    {"SAVE_ARGS", IR_TY_IMM},
    {"NOP", IR_TY_NOARG},
}

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
    case IR_TY_IMM_IMM:
        return fmt.Sprintf("  %s %d, %d", info.Name, ir.Lhs, ir.Rhs)
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

var code *Vector
// 汎用レジスタの番号
var nreg int
var nlabel int

func add(op int, lhs int, rhs int) *IR {
    var ir *IR = new(IR)
    ir.Op = op
    ir.Lhs = lhs
    ir.Rhs = rhs
    Vec_push(code, ir)
    return ir
}

func kill(r int) {
    add(IR_KILL, r, -1)
}

func label(x int) {
    add(IR_LABEL, x, -1)
}

func gen_lval(node *Node) int {

    if node.Op == ND_DEREF {
        return gen_expr(node.Expr)
    }

    Assert(node.Op == ND_LVAR, "not an lvalue")
    r := nreg
    nreg++
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
    kill(r2)
    return r1
}

func gen_expr(node *Node) int {
    switch node.Op {
    case ND_NUM:
        r := nreg
        nreg++
        add(IR_IMM, r, node.Val)
        return r
    case ND_LOGAND:
        // return処理を行うラベルx
        x := nlabel
        nlabel++

        var r1 int = gen_expr(node.Lhs)
        // レジスタr1の値がfalse(0)ならばラベルxへ飛ぶ
        add(IR_UNLESS, r1, x)
        var r2 int = gen_expr(node.Rhs)
        add(IR_MOV, r1, r2)
        kill(r2)
        // r2の値が0でもラベルxへjmp
        add(IR_UNLESS, r1, x)
        // r1, r2ともにfalse(0)ではなかったため、
        // 戻り値としてtrue(1)を返すためr1に1を代入する
        add(IR_IMM, r1, 1)
        label(x)
        return r1
    case ND_LOGOR:
        // 中継先のラベル
        x := nlabel
        nlabel++
        // 最終的な行き先のラベル
        y := nlabel
        nlabel++

        var r1 int = gen_expr(node.Lhs)
        add(IR_UNLESS, r1, x)
        // r1値がfalse(0)でない場合, r1にtrue(1)を代入し,
        // y(最終的なラベル)に飛ぶ
        add(IR_IMM, r1, 1)
        add(IR_JMP, y, -1)

        // 中継先ラベルxの用意
        label(x)

        var r2 int = gen_expr(node.Rhs)
        add(IR_MOV, r1, r2)
        kill(r2)
        // r1値がtrue(0)でかつ, r2値がfalse(0)のときラベルyへ飛ぶ
        add(IR_UNLESS, r1, y)
        // r1値がtrue(0)でかつ, r2値もtrue(1)のため,
        // 戻り値としてtrue(1)を返す。そのためにr1に値1を代入する
        add(IR_IMM, r1, 1)
        label(y)
        return r1
    case ND_LVAR:
        var r int = gen_lval(node)
        if node.Ty.Ty == CHAR {
            add(IR_LOAD8, r, r)
        } else if node.Ty.Ty == INT {
            add(IR_LOAD32, r, r)
        } else {
            add(IR_LOAD64, r, r)
        }
        return r
    case ND_CALL:
        var args [6]int
        for i := 0; i < node.Args.Len; i++ {
            // 関数に引数がある場合
            arg, _ := node.Args.Data[i].(*Node)
            args[i] = gen_expr(arg)
        }

        r := nreg
        nreg++

        var ir *IR = add(IR_CALL, r, -1)
        ir.Name = node.Name
        ir.Nargs = node.Args.Len
        ir.Args = args

        for i := 0; i < ir.Nargs; i++ {
            kill(ir.Args[i])
        }
        return r
    case ND_ADDR:
        return gen_lval(node.Expr)
    case ND_DEREF:
        r := gen_expr(node.Expr)
        // 間接参照(int型のポインタが指すメモリを参照する)のでload命令
        add(IR_LOAD64, r, r)
        return r
    case '=':
        var rhs int = gen_expr(node.Rhs)
        // lhsはメモリへstoreするためのアドレスが格納されたレジスタ(の番号)が入っている
        var lhs int = gen_lval(node.Lhs)
        if node.Lhs.Ty.Ty == CHAR {
            add(IR_STORE8, lhs, rhs)
        } else if node.Lhs.Ty.Ty == INT {
            add(IR_STORE32, lhs, rhs)
        } else {
            add(IR_STORE64, lhs, rhs)
        }
        kill(rhs)
        return lhs
    case '+', '-':
        var insn int
        // Goには三項演算子がない
        if node.Op == '+' {
            insn = IR_ADD
        } else {
            insn = IR_SUB
        }

        if node.Lhs.Ty.Ty != PTR {
            return gen_binop(insn, node.Lhs, node.Rhs)
        }

        rhs := gen_expr(node.Rhs)

        r := nreg
        nreg++

        add(IR_IMM, r, Size_of(node.Lhs.Ty.Ptr_of))
        add(IR_MUL, rhs, r)
        kill(r)

        lhs := gen_expr(node.Lhs)

        add(insn, lhs, rhs)
        kill(rhs)

        return lhs
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
    if node.Op == ND_VARDEF {

        if node.Init == nil {
            return
        }

        var rhs int = gen_expr(node.Init)
        var lhs int = nreg
        nreg++
        // lhsにベースレジスタのアドレスを代入
        add(IR_MOV, lhs, 0)
        // ベースレジスタから、変数のオフセット分引く
        add(IR_SUB_IMM, lhs, node.Offset)
        // メモリ上のスタックで、左辺値(lhs)に対し、右辺値(rhs)を代入する
        if node.Ty.Ty == CHAR {
            add(IR_STORE8, lhs, rhs)
        } else if node.Ty.Ty == INT {
            add(IR_STORE32, lhs, rhs)
        } else {
            add(IR_STORE64, lhs, rhs)
        }
        kill(lhs)
        kill(rhs)
        return
    }

    if node.Op == ND_IF {

        if Node2bool(node.Els) {
            // else文がある場合
            x := nlabel
            nlabel++
            y := nlabel
            nlabel++
            r := gen_expr(node.Cond)
            // レジスタrの値がfalse(0)の場合, ラベルxへジャンプする
            add(IR_UNLESS, r, x)
            kill(r)

            gen_stmt(node.Then)
            add(IR_JMP, y, -1)
            label(x)

            gen_stmt(node.Els)
            label(y)
        }

        x := nlabel
        nlabel++
        r := gen_expr(node.Cond)

        // レジスタrの値がfalse(0)ならばラベルxへ飛ぶ
        add(IR_UNLESS, r, x)
        kill(r)

        gen_stmt(node.Then)

        label(x)
        return
    }

    if node.Op == ND_FOR {
        x := nlabel
        nlabel++
        y := nlabel
        nlabel++

        gen_stmt(node.Init)
        label(x)
        r := gen_expr(node.Cond)
        add(IR_UNLESS, r, y)
        kill(r)
        gen_stmt(node.Body)
        kill(gen_expr(node.Inc))
        add(IR_JMP, x, -1)
        label(y)
        return
    }

    if node.Op == ND_RETURN {
        r := gen_expr(node.Expr)
        add(IR_RETURN, r, -1)
        kill(r)
        return
    }

    if node.Op == ND_EXPR_STMT {
        kill(gen_expr(node.Expr))
        return
    }

    if node.Op == ND_COMP_STMT {
        for i := 0; i < node.Stmts.Len; i++ {
            n, _ := node.Stmts.Data[i].(*Node)
            gen_stmt(n)
        }
        return
    }

    Error(fmt.Sprintf("unknown node: %d", node.Op))
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
        Assert(node.Op == ND_FUNC, "Type of root node is not ND_FUNC")

        // fn.Irに使用
        code = New_vec()
        // 各関数ごとにregsiter numとstacksizeを初期化している.
        // nregが1からはじまるのは、レジスタの配列Regsの0番目にrbpがあるから.
        nreg = 1

        for i := 0; i < node.Args.Len; i++ {
            arg := node.Args.Data[i].(*Node)

            if arg.Ty.Ty == CHAR {
                add(IR_STORE8_ARG, arg.Offset, i)
            } else if arg.Ty.Ty == INT {
                add(IR_STORE32_ARG, arg.Offset, i)
            } else {
                add(IR_STORE64_ARG, arg.Offset, i)
            }
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
