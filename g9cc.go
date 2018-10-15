// アセンブリ言語の構文には, AT&TのものとIntelのものがある。
// レジスタ名に%がつくのがAT&T, そうでないものがIntel

package main

import (
    "fmt"
    "os"
    // "strconv"
)

func main() {
    if len(os.Args) != 2 {
        fmt.Println("Usage: g9cc <code>")

        return
    }

    // argv := []rune(os.Args[1])
    argv := os.Args[1]
    argv += "\000"


    // fmt.Println("	.section	__TEXT,__text,regular,pure_instructions")
    // fmt.Println("	.macosx_version_min 10, 12")
    fmt.Println("	.globl	_main")
    // fmt.Println("	.p2align	4, 0x90")
    fmt.Println("_main:                                  ## @main")
    // fmt.Println(".cfi_startproc")
    // fmt.Println("## BB#0:")
    // fmt.Println("	pushq	%rbp")
    // fmt.Println("Ltmp0:")
    // fmt.Println("	.cfi_def_cfa_offset 16")
    // fmt.Println("Ltmp1:")
    // fmt.Println("	.cfi_offset %rbp, -16")
    // fmt.Println("	movq	%rsp, %rbp")
    // fmt.Println("Ltmp2:")
    // fmt.Println("	.cfi_def_cfa_register %rbp")
    //
    var i int = 0;
    var num int32;

    if '0' <= argv[i] && argv[i] <= '9' {
        num = int32(argv[i] - '0')
        i++
        for ; '0' <= argv[i] && argv[i] <= '9'; i++ {
            num = num * 10 + int32(argv[i] - '0')
        }

        fmt.Printf("	movl	$%d, %%eax    ## imm = 0x4B0\n", num)
    }

    for i < len(argv) - 1{

        if argv[i] == '+' {
            i++;

            if '0' <= argv[i] && argv[i] <= '9' {
                num = int32(argv[i] - '0')
                i++
                for ; '0' <= argv[i] && argv[i] <= '9'; i++ {
                    num = num * 10 + int32(argv[i] - '0')
                }
            }

            fmt.Printf("	add	   $%d, %%eax\n", num);
            continue;
        }

        if argv[i] == '-' {
            i++;

            if '0' <= argv[i] && argv[i] <= '9' {
                num = int32(argv[i] - '0')
                i++
                for ; '0' <= argv[i] && argv[i] <= '9'; i++ {
                    num = num * 10 + int32(argv[i] - '0')
                }
            }

            fmt.Printf("	sub	   $%d, %%eax\n", argv[i]);
            continue;
        }

        // 終端文字
        if argv[i] == '\000' {
            return
        }

        fmt.Printf("unexpected character: %c\n", argv[i]);
    }

    // fmt.Println("	movl	$0, -4(%rbp)")
    // fmt.Println("	popq	%rbp")
    fmt.Println("	retq")
    // fmt.Println("	.cfi_endproc")
    // fmt.Printf("\n\n")
    // fmt.Println(".subsections_via_symbols")

}
