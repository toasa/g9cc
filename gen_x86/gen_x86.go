package codegen

import (
    . "g9cc/common"
    . "g9cc/util"
    . "g9cc/regs"
    "fmt"
)

// Code generator
// irv.data[]内の各ir(中間表現)に対し、ir.opから機械的にアセンブリを生成していく

// cdecl(C Declaration)
// ・関数の引数は右から左の順にスタックに積まれる
// ・呼び出された側の関数ではeax, ecx, edxのレジスタのもとの値を保存することなく使用して良い
//      → 呼び出し側の関数では(必要なら)呼び出す前にそれらのレジスタをスタック上に保持する
// ・スタックポインタの処理は呼び出し側で行う
// ・スタック上の引数データの削除は呼び出し側で行う

// x64 calling convention(cc)
// 整数・ポインタ引数: rdi, rsi, rdx, rcx, r8, r9
// 戻り値: rax
// システムコール: rcx
// レジスタのみでは引数の数が不足する場合、スタックを使用

// x86 calling convention
// 関数呼び出す側のcc
// ・関数の引数は右側からpushする
// ・call func
// ・引数の数×4、espを増やす
// ・関数の戻り値はraxにいれる
//
// ex.
// int func(int a, int b);
// のとき
// push b
// push a
// call func
// add esp 3×4

// 関数を呼び出される側のcc
// ・帰り値はeaxに代入する
// ・汎用レジスタの内、ebx, esi, edi, ebp, espの値は関数呼び出し時の値に戻す
//      →(ecx, edxの値を保存する必要はない)
// ・ret命令で関数呼び出しから戻る
//      この時引数は、esp+(左から引数が何番目にあるか)

var label int

// 関数の引数の値を代入するためのレジスタ
var argreg []string = []string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}

func gen(fn *Function) {

    var ret string = fmt.Sprintf(".Lend%d", label)
    label++

    fmt.Printf(".globl _%s\n", fn.Name)
    fmt.Printf("_%s:\n", fn.Name)

    // 関数呼び出しの先頭で以下の２行を行う
    // 呼び出し元のrbpをスタックにpushする。そしてrbpにrspを代入する(呼び出し先の関数の基点となるアドレスを作る)
    // rbp: 関数内においてスタック領域を扱う処理の基準となるメモリアドレス
    fmt.Printf("    push rbp\n")
    fmt.Printf("    mov rbp, rsp\n")

    fmt.Printf("    sub rsp, %d\n", fn.Stacksize)
    fmt.Printf("    push r12\n")
    fmt.Printf("    push r13\n")
    fmt.Printf("    push r14\n")
    fmt.Printf("    push r15\n")

    for i := 0; i < fn.Ir.Len; i++ {
        ir := fn.Ir.Data[i].(*IR)

        switch ir.Op {
        case IR_IMM:
            fmt.Printf("    mov %s, %d\n", Regs[ir.Lhs], ir.Rhs)
        case IR_SUB_IMM:
            fmt.Printf("    sub %s, %d\n", Regs[ir.Lhs], ir.Rhs)
        case IR_MOV:
            fmt.Printf("    mov %s, %s\n", Regs[ir.Lhs], Regs[ir.Rhs])
        case IR_RETURN: // lhsに格納された値をアキュムレータに渡し、戻り値とする
            fmt.Printf("    mov rax, %s\n", Regs[ir.Lhs])
            fmt.Printf("    jmp %s\n", ret)
        case IR_CALL:
            for i := 0; i < ir.Nargs; i++ {
                fmt.Printf("    mov %s, %s\n", argreg[i], Regs[ir.Args[i]])
            }

            fmt.Printf("    push r10\n")
            fmt.Printf("    push r11\n")
            fmt.Printf("    mov rax, 0\n")
            // 関数名の先頭に"_"が必要
            // callの次のアドレスをスタックに積んで、ラベル_%sを実行する
            fmt.Printf("    call _%s\n", ir.Name)
            fmt.Printf("    pop r11\n")
            fmt.Printf("    pop r10\n")

            fmt.Printf("    mov %s, rax\n", Regs[ir.Lhs])
        case IR_LABEL:
            fmt.Printf(".L%d:\n", ir.Lhs)
        case IR_JMP:
            fmt.Printf("    jmp .L%d\n", ir.Lhs)
        case IR_UNLESS:
            // 今の所, lhsの(レジスタの)値が0ならラベルに飛ぶ
            fmt.Printf("    cmp %s, 0\n", Regs[ir.Lhs])
            fmt.Printf("    je .L%d\n", ir.Rhs)
        case IR_LOAD:
            fmt.Printf("    mov %s, [%s]\n", Regs[ir.Lhs], Regs[ir.Rhs])
        case IR_STORE:
            fmt.Printf("    mov [%s], %s\n", Regs[ir.Lhs], Regs[ir.Rhs])
        case IR_SAVE_ARGS:
            for i := 0; i < ir.Lhs; i++ {
                fmt.Printf("    mov [rbp-%d], %s\n", (i + 1) * 8, argreg[i])
            }
        case IR_ADD:
            fmt.Printf("    add %s, %s\n", Regs[ir.Lhs], Regs[ir.Rhs])
        case IR_SUB:
            fmt.Printf("    sub %s, %s\n", Regs[ir.Lhs], Regs[ir.Rhs])
        case IR_MUL:
            fmt.Printf("    mov rax, %s\n", Regs[ir.Rhs])
            // mul reg: 予めrax(アキュムレータ)に格納された値と
            //          regに格納された値の掛け算を行い,結果をraxに格納する
            fmt.Printf("    mul %s\n", Regs[ir.Lhs])
            // 掛け算の結果を汎用レジスタに格納する
            fmt.Printf("    mov %s, rax\n", Regs[ir.Lhs])
        case IR_DIV:
            // raxに左辺値を代入
            fmt.Printf("    mov rax, %s\n", Regs[ir.Lhs])
            // convert quadword to octaword: 符号拡張(アキュムレータを拡張する)
            // wordは4byteのこと? => ocは32byte?
            fmt.Printf("    cqo\n")
            fmt.Printf("    div %s\n", Regs[ir.Rhs])
            fmt.Printf("    mov %s, rax\n", Regs[ir.Lhs])
        case IR_NOP:

        default:
            Error("unknown operator")
        }
    }

    fmt.Printf("%s:\n", ret)
    // 関数の最後で以下の処理を行う
    // rspにrbpを代入する(ローカル変数の確保などで、rspを更新した場合rbpに戻すための処理)
    // rbpには関数呼び出し直後のスタックトップのアドレスが格納されていたので、そこまでrspを戻すことができる
    // popしてrbpを呼び、関数呼び出しもとでの値へrbpをもどす
    fmt.Printf("    pop r15\n")
    fmt.Printf("    pop r14\n")
    fmt.Printf("    pop r13\n")
    fmt.Printf("    pop r12\n")
    fmt.Printf("    mov rsp, rbp\n")
    fmt.Printf("    pop rbp\n")
    fmt.Printf("    ret\n")
}

func Gen_x86(fns *Vector) {
    fmt.Printf(".intel_syntax noprefix\n")

    for i := 0; i < fns.Len; i++ {
        fn := fns.Data[i].(*Function)
        gen(fn)
    }
}