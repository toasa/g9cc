package regs

var Regs []string = []string{"rbp", "r10", "r11", "rbx", "r12", "r13", "r14", "r15"}
var Regs8[]string = []string{"bpl", "r10b", "r11b", "bl", "r12b", "r13b", "r14b", "r15b"}
var Regs32[]string = []string{"ebp", "r10d", "r11d", "ebx", "r12d", "r13d", "r14d", "r15d"}
var Len_Regs int = len(Regs)

// 関数の引数の値を代入するためのレジスタ
var Argregs []string = []string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}
var Argregs8 []string = []string{"dil", "sil", "dl", "cl", "r8b", "r9b"}
var Argregs32 []string = []string{"edi", "esi", "edx", "ecx", "r8d", "r9d"}
