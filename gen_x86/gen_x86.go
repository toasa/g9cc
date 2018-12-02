package codegen

import (
    . "g9cc/common"
    . "g9cc/util"
    . "g9cc/regs"
    "fmt"
    "strings"
    "strconv"
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
// 関数の引数、整数・ポインタ引数: rdi, rsi, rdx, rcx, r8, r9
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
var argreg8 []string = []string{"dil", "sil", "dl", "cl", "r8b", "r9b"}
var argreg32 []string = []string{"edi", "esi", "edx", "ecx", "r8d", "r9d"}
var argreg64 []string = []string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}

func escape(s string, len int) string {
    var escaped[256]int32
    escaped['\b'] = 'b'
    escaped['\f'] = 'f'
    escaped['\n'] = 'n'
    escaped['\r'] = 'r'
    escaped['\t'] = 't'
    escaped['\\'] = '\\'
    escaped['\''] = '\''
    escaped['"'] = '"'

    var buf string

    for s_i := 0; s_i < len; s_i++ {
        esc := escaped[int32(s[s_i])]
        if esc != 0 {
            buf += "\\"
            buf += strconv.Itoa(int(esc))
        } else if (Is_graph(s[s_i]) || s[s_i] == ' ') {
            buf += string(s[s_i])
        } else {
            //buf += "\u0000"
            //buf += fmt.Sprintf("\\%03o", s[s_i])
        }
    }
    //buf += "\u0000"
    return buf
}

func emit_cmp(ir *IR, insn string) {
    // 左右のレジスタを比較. 比較結果はフラグレジスタ(x86-64の場合、RFLAGS)
    // に格納される
    fmt.Printf("    cmp %s, %s\n", Regs[ir.Lhs], Regs[ir.Rhs])
    fmt.Printf("    %s %s\n", insn, Regs8[ir.Lhs])

    // 9cc には movzb と記載, だがアセンブリ時
    // error: invalid instruction mnemonic 'movzb'　となった
    // movzbl: move zero extended byte to long

    // こちらはうまく行った
    // movzx: move with zero extention

    // 8bitレジスタ, 例えばAL(アキュムレータ・ロー)に結果をセットした場合
    // RAXの値も変わっている(ALはRAXの下位8bit故). だが、RAXの上位56bit
    // はもとの値のままなので、ゼロを入れクリアする必要がある. それを行うのが、
    // movzx
    fmt.Printf("    movzx %s, %s\n", Regs[ir.Lhs], Regs8[ir.Lhs])
}

func gen(fn *Function) {

    var ret string = fmt.Sprintf(".Lend%d", label)
    label++

    fmt.Printf("    .globl _%s\n", fn.Name)
    fmt.Printf("_%s:\n", fn.Name)

    // 関数呼び出しの先頭で以下の２行を行う
    // 呼び出し元のrbpをスタックにpushする。そしてrbpにrspを代入する(呼び出し先の関数の基点となるアドレスを作る)
    // rbp: 関数内においてスタック領域を扱う処理の基準となるメモリアドレス
    fmt.Printf("    push rbp\n")
    fmt.Printf("    mov rbp, rsp\n")

    fmt.Printf("    sub rsp, %d\n", Roundup(fn.Stacksize, 16))
    fmt.Printf("    push r12\n")
    fmt.Printf("    push r13\n")
    fmt.Printf("    push r14\n")
    fmt.Printf("    push r15\n")

    for i := 0; i < fn.Ir.Len; i++ {
        ir := fn.Ir.Data[i].(*IR)

        switch ir.Op {
        case IR_IMM:
            fmt.Printf("    mov %s, %d\n", Regs[ir.Lhs], ir.Rhs)
        case IR_BPREL:
            fmt.Printf("    lea %s, [rbp - %d]\n", Regs[ir.Lhs], ir.Rhs)
        case IR_MOV:
            fmt.Printf("    mov %s, %s\n", Regs[ir.Lhs], Regs[ir.Rhs])
        case IR_RETURN: // lhsに格納された値をアキュムレータに渡し、戻り値とする
            fmt.Printf("    mov rax, %s\n", Regs[ir.Lhs])
            fmt.Printf("    jmp %s\n", ret)
        case IR_CALL:
            for i := 0; i < ir.Nargs; i++ {
                fmt.Printf("    mov %s, %s\n", argreg64[i], Regs[ir.Args[i]])
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
        case IR_LABEL_ADDR:
            // load effective address
            // 第２オペランドの実行アドレスを計算し、第１オペランドに格納する
            // 第２オペランドが格納されたアドレスはripによっても変化する？
            fmt.Printf("    lea %s, [rip + %s]\n", Regs[ir.Lhs], ir.Name)
        case IR_EQ:
            // ZFフラグの値をオペランドへ格納
            emit_cmp(ir, "sete")
        case IR_NE:
            // ZFフラグと逆の値をオペランドへ格納
            emit_cmp(ir, "setne")
        case IR_LT:
            // フラグレジスタの値をオペランドに格納
            emit_cmp(ir, "setl")
        case IR_LE:
            emit_cmp(ir, "setle")
        case IR_AND:
            fmt.Printf("    and %s, %s\n", Regs[ir.Lhs], Regs[ir.Rhs])
        case IR_OR:
            fmt.Printf("    or %s, %s\n", Regs[ir.Lhs], Regs[ir.Rhs])
        case IR_XOR:
            fmt.Printf("    xor %s, %s\n", Regs[ir.Lhs], Regs[ir.Rhs])
        case IR_JMP:
            fmt.Printf("    jmp .L%d\n", ir.Lhs)
        case IR_IF:
            // 左右のオペランドを引き算し、フラグレジスタに結果を格納
            fmt.Printf("    cmp %s, 0\n", Regs[ir.Lhs])
            // jump if not equal
            // ZF(ゼロ・フラグ)が0のとき(cmpを行い左右のオペランドが等しくないとき),
            // オペランドのラベルへジャンプ
            fmt.Printf("    jne .L%d\n", ir.Rhs)
        case IR_UNLESS:
            // 今の所, lhsの(レジスタの)値が0ならラベルに飛ぶ
            fmt.Printf("    cmp %s, 0\n", Regs[ir.Lhs])
            fmt.Printf("    je .L%d\n", ir.Rhs)
        case IR_LOAD8:
            fmt.Printf("    mov %s, [%s]\n", Regs8[ir.Lhs], Regs[ir.Rhs])
            fmt.Printf("    movzx %s, %s\n", Regs[ir.Lhs], Regs8[ir.Rhs])
        case IR_LOAD32:
            fmt.Printf("    mov %s, [%s]\n", Regs32[ir.Lhs], Regs[ir.Rhs])
        case IR_LOAD64:
            fmt.Printf("    mov %s, [%s]\n", Regs[ir.Lhs], Regs[ir.Rhs])
        case IR_STORE8:
            fmt.Printf("    mov [%s], %s\n", Regs[ir.Lhs], Regs8[ir.Rhs])
        case IR_STORE32:
            fmt.Printf("    mov [%s], %s\n", Regs[ir.Lhs], Regs32[ir.Rhs])
        case IR_STORE64:
            fmt.Printf("    mov [%s], %s\n", Regs[ir.Lhs], Regs[ir.Rhs])
        case IR_STORE8_ARG:
            fmt.Printf("    mov [rbp-%d], %s\n", ir.Lhs, argreg8[ir.Rhs])
        case IR_STORE32_ARG:
            fmt.Printf("    mov [rbp-%d], %s\n", ir.Lhs, argreg32[ir.Rhs])
        case IR_STORE64_ARG:
            fmt.Printf("    mov [rbp-%d], %s\n", ir.Lhs, argreg64[ir.Rhs])
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

func Gen_x86(globals *Vector, fns *Vector) {
    fmt.Printf("    .intel_syntax noprefix\n")

    // .data以下をデータセグメントに配置する命令
    fmt.Printf("    .data\n")
    for i := 0; i < globals.Len; i++ {
        var_ := globals.Data[i].(*Var)
        if var_.Is_extern {
            continue
        }
        fmt.Printf("%s:\n", var_.Name)

        if len(var_.Data + "\u0000") == var_.Len {
            fmt.Printf("    .ascii \"%s\"\n", escape(var_.Data + "\u0000", var_.Len))
        } else {
            fmt.Printf("    .ascii \"%s\"\n", strings.Repeat("\\000", var_.Len))
        }
    }

    // .text以下をテキストセグメントに配置する命令
    fmt.Printf("    .text\n")

    for i := 0; i < fns.Len; i++ {
        fn := fns.Data[i].(*Function)
        gen(fn)
    }

    func_alloc()
}

func func_alloc() {

    // this is assembly language below C functions,
    // assembled in intel_syntax.
    // this code is generated by
    // `gcc -S -masm=intel alloc.c`
    // command

    // **************************************
    //
    // int *alloc1(int x, int y) {
    //     static int arr[2];
    //     arr[0] = x;
    //     arr[1] = y;
    //     return arr;
    // }
    //
    // int *alloc2(int x, int y) {
    //     static int arr[2];
    //     arr[0] = x;
    //     arr[1] = y;
    //     return arr + 1;
    // }
    //
    // int **alloc_ptr_ptr(int x) {
    //     static int **p;
    //     static int *q;
    //     static int r;
    //     r = x;
    //     q = &r;
    //     p = &q;
    //     return p;
    // }
    //
    // **************************************


    fmt.Println("    .globl _alloc1")
    fmt.Println("_alloc1:")
    fmt.Println("    push rbp")
    fmt.Println("    mov rbp, rsp")
    fmt.Println("    lea rax, [rip + _alloc1.arr]")
    fmt.Println("    mov dword ptr [rbp - 4], edi")
    fmt.Println("    mov dword ptr [rbp - 8], esi")
    fmt.Println("    mov esi, dword ptr [rbp - 4]")
    fmt.Println("    mov dword ptr [rip + _alloc1.arr], esi")
    fmt.Println("    mov esi, dword ptr [rbp - 8]")
    fmt.Println("    mov dword ptr [rip + _alloc1.arr+4], esi")
    fmt.Println("    pop rbp")
    fmt.Println("    ret")
    fmt.Println("")

    fmt.Println("    .globl _alloc2")
    fmt.Println("_alloc2:")
    fmt.Println("    push rbp")
    fmt.Println("    mov rbp, rsp")
    fmt.Println("    lea rax, [rip + _alloc2.arr]")
    fmt.Println("    add rax, 4")
    fmt.Println("    mov dword ptr [rbp - 4], edi")
    fmt.Println("    mov dword ptr [rbp - 8], esi")
    fmt.Println("    mov esi, dword ptr [rbp - 4]")
    fmt.Println("    mov dword ptr [rip + _alloc2.arr], esi")
    fmt.Println("    mov esi, dword ptr [rbp - 8]")
    fmt.Println("    mov dword ptr [rip + _alloc2.arr+4], esi")
    fmt.Println("    pop rbp")
    fmt.Println("    ret")
    fmt.Println("")

    fmt.Println("    .globl _alloc_ptr_ptr")
    fmt.Println("_alloc_ptr_ptr:")
    fmt.Println("    push rbp")
    fmt.Println("    mov rbp, rsp")
    fmt.Println("    lea rax, [rip + _alloc_ptr_ptr.q]")
    fmt.Println("    lea rcx, [rip + _alloc_ptr_ptr.r]")
    fmt.Println("    mov dword ptr [rbp - 4], edi")
    fmt.Println("    mov edi, dword ptr [rbp - 4]")
    fmt.Println("    mov dword ptr [rip + _alloc_ptr_ptr.r], edi")
    fmt.Println("    mov qword ptr [rip + _alloc_ptr_ptr.q], rcx")
    fmt.Println("    mov qword ptr [rip + _alloc_ptr_ptr.p], rax")
    fmt.Println("    mov rax, qword ptr [rip + _alloc_ptr_ptr.p]")
    fmt.Println("    pop rbp")
    fmt.Println("    ret")
    fmt.Println("")
    fmt.Println("    .zerofill __DATA,__bss,_alloc1.arr,8,2  ## @alloc1.arr")
    fmt.Println("    .zerofill __DATA,__bss,_alloc2.arr,8,2  ## @alloc2.arr")
    fmt.Println("    .zerofill __DATA,__bss,_alloc_ptr_ptr.p,8,3 ## @alloc_ptr_ptr.p")
    fmt.Println("    .zerofill __DATA,__bss,_alloc_ptr_ptr.q,8,3 ## @alloc_ptr_ptr.q")
    fmt.Println("    .zerofill __DATA,__bss,_alloc_ptr_ptr.r,4,2 ## @alloc_ptr_ptr.r")
    fmt.Println("")

    fmt.Println("    .globl _global_arr")
    fmt.Println("global_arr:")
    fmt.Println("    .long 5")
    fmt.Println("")
}
