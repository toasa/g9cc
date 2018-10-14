package main

import (
    "fmt"
    "os"
    "strconv"
)

func main() {
    if len(os.Args) != 2 {
        fmt.Println("Usaga: g9cc <code>")
    }

    i, _ := strconv.Atoi(os.Args[1])

    fmt.Println("	.section	__TEXT,__text,regular,pure_instructions")
    fmt.Println("	.macosx_version_min 10, 12")
    fmt.Println("	.globl	_main")
    fmt.Println("	.p2align	4, 0x90")
    fmt.Println("_main:                                  ## @main")
    fmt.Println(".cfi_startproc")
    fmt.Println("## BB#0:")
    fmt.Println("	pushq	%rbp")
    fmt.Println("Ltmp0:")
    fmt.Println("	.cfi_def_cfa_offset 16")
    fmt.Println("Ltmp1:")
    fmt.Println("	.cfi_offset %rbp, -16")
    fmt.Println("	movq	%rsp, %rbp")
    fmt.Println("Ltmp2:")
    fmt.Println("	.cfi_def_cfa_register %rbp")
    fmt.Printf("	movl	$%d, %%eax             ## imm = 0x4B0", i)
    fmt.Println("	movl	$0, -4(%rbp)")
    fmt.Println("	popq	%rbp")
    fmt.Println("	retq")
    fmt.Println("	.cfi_endproc")
    fmt.Printf("\n\n")
    fmt.Println(".subsections_via_symbols")

}
