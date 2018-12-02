package gen_ir

import (
    . "g9cc/common"
    . "g9cc/util"
    "fmt"
    // "github.com/k0kubun/pp"
)

// Compile AST to intermediate code that has infinite number of registers.
// Base pointer is always assigned to r0(notation of -dump-ir).

var code *Vector
// 汎用レジスタの番号
var nreg int = 1
var nlabel int = 1
var return_label int
var return_reg int

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

func choose_insn(node *Node, op8, op32, op64 int) int {

    if node.Ty.Size == 1 {
        return op8
    } else if node.Ty.Size == 4 {
        return op32
    }
    Assert(node.Ty.Size == 8, "unmatched size")
    return op64
}

func load_insn(node *Node) int {
    return choose_insn(node, IR_LOAD8, IR_LOAD32, IR_LOAD64)
}

func store_insn(node *Node) int  {
    return choose_insn(node, IR_STORE8, IR_STORE32, IR_STORE64)
}

func store_arg_insn(node *Node) int {
    return choose_insn(node, IR_STORE8_ARG, IR_STORE32_ARG, IR_STORE64_ARG)
}

func gen_lval(node *Node) int {

    if node.Op == ND_DEREF {
        return gen_expr(node.Expr)
    }

    if node.Op == ND_DOT {
        r1 := gen_lval(node.Expr)
        r2 := nreg
        nreg++
        add(IR_IMM, r2, node.Offset)
        add(IR_ADD, r1, r2)
        kill(r2)
        return r1
    }

    if node.Op == ND_LVAR {
        r := nreg
        nreg++
        add(IR_BPREL, r, node.Offset)
        return r
    }

    Assert(node.Op == ND_GVAR, "not an global variety")

    r := nreg
    nreg++

    ir := add(IR_LABEL_ADDR, r, -1)
    ir.Name = node.Name
    return r
}

func gen_binop(ty int, node *Node) int {
    lhs := gen_expr(node.Lhs)
    rhs := gen_expr(node.Rhs)
    add(ty, lhs, rhs)
    kill(rhs)
    return lhs
}

func gen_expr(node *Node) int {
    switch node.Op {
    case ND_NUM:
        r := nreg
        nreg++
        add(IR_IMM, r, node.Val)
        return r
    case ND_EQ:
        return gen_binop(IR_EQ, node)
    case ND_NE:
        return gen_binop(IR_NE, node)
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
    case ND_GVAR, ND_LVAR, ND_DOT:
        var r int = gen_lval(node)
        add(load_insn(node), r, r)
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
        add(load_insn(node), r, r)
        return r
    case ND_STMT_EXPR:
        orig_label := return_label
        orig_reg := return_reg
        return_label = nlabel
        nlabel++
        r := nreg
        nreg++
        return_reg = r

        gen_stmt(node.Body)
        label(return_label)

        return_label = orig_label
        return_reg = orig_reg
        return r
    case '=':
        var rhs int = gen_expr(node.Rhs)
        // lhsはメモリへstoreするためのアドレスが格納されたレジスタ(の番号)が入っている
        var lhs int = gen_lval(node.Lhs)
        add(store_insn(node), lhs, rhs)
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
            return gen_binop(insn, node)
        }

        rhs := gen_expr(node.Rhs)

        r := nreg
        nreg++

        add(IR_IMM, r, node.Lhs.Ty.Ptr_to.Size)
        add(IR_MUL, rhs, r)
        kill(r)

        lhs := gen_expr(node.Lhs)

        add(insn, lhs, rhs)
        kill(rhs)

        return lhs
    case '*':
        return gen_binop(IR_MUL, node)
    case '/':
        return gen_binop(IR_DIV, node)
    case '<':
        return gen_binop(IR_LT, node)
    case ND_LE:
        return gen_binop(IR_LE, node)
    case '&':
        return gen_binop(IR_AND, node)
    case '|':
        return gen_binop(IR_OR, node)
    case '^':
        return gen_binop(IR_XOR, node)
    case ',':
        kill(gen_expr(node.Lhs))
        return gen_expr(node.Rhs)
    case '?':
        x := nlabel
        nlabel++
        y := nlabel
        nlabel++
        r := gen_expr(node.Cond)

        add(IR_UNLESS, r, x)
        r2 := gen_expr(node.Then)
        add(IR_MOV, r, r2)
        kill(r2)
        add(IR_JMP, y, -1)

        label(x)
        r3 := gen_expr(node.Els)
        add(IR_MOV, r, r3)
        kill(r2)
        label(y)
        return r
    case '!':
        lhs := gen_expr(node.Expr)
        rhs := nreg
        nreg++
        add(IR_IMM, rhs, 0)
        add(IR_EQ, lhs, rhs)
        kill(rhs)
        return lhs
    default:
        Assert(false, "unknown AST type")
    }

    err := 0
    return err
}

func gen_stmt(node *Node) {
    if node.Op == ND_NULL {
        return
    }

    if node.Op == ND_VARDEF {

        if node.Init == nil {
            return
        }

        var rhs int = gen_expr(node.Init)
        var lhs int = nreg
        nreg++
        // ベースレジスタから、変数のオフセット分引く
        add(IR_BPREL, lhs, node.Offset)
        // メモリ上のスタックで、左辺値(lhs)に対し、右辺値(rhs)を代入する
        add(store_insn(node), lhs, rhs)
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
            return
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
        gen_stmt(node.Inc)
        add(IR_JMP, x, -1)
        label(y)
        return
    }

    if node.Op == ND_DO_WHILE {
        x := nlabel
        nlabel++
        label(x)
        gen_stmt(node.Body)
        r := gen_expr(node.Cond)
        add(IR_IF, r, x)
        kill(r)
        return
    }

    if node.Op == ND_RETURN {
        r := gen_expr(node.Expr)

        // Statement expression (GNU extension)
        if return_label != 0 {
            add(IR_MOV, return_reg, r)
            kill(r)
            add(IR_JMP, return_label, -1)
            return
        }

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

        if node.Op == ND_VARDEF {
            continue
        }

        Assert(node.Op == ND_FUNC, "Type of root node is not ND_FUNC")

        // fn.Irに使用
        code = New_vec()
        // 各関数ごとにregsiter numとstacksizeを初期化している.
        // nregが1からはじまるのは、レジスタの配列Regsの0番目にrbpがあるから.
        nreg = 1

        for i := 0; i < node.Args.Len; i++ {
            arg := node.Args.Data[i].(*Node)
            add(store_arg_insn(arg), arg.Offset, i)
        }

        gen_stmt(node.Body)

        fn := new(Function)
        fn.Name = node.Name
        fn.Stacksize = node.Stacksize
        fn.Ir = code
        fn.Globals = node.Globals
        Vec_push(v, fn)
    }
    return v
}
