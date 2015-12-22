;
; Example of Inner product algorithm
;

; Macros to pre-fill memory
@0x0040: 37 15 24 12 24 27 25 69 45 21 54 32 14 25 14 26 58 12 35 68 37 15 24 12 24 27 25 69 45 21 54 32 14 25 14 26 58 12 35 68
@0x0240: 16 85 34 25 75 21 66 85 48 98 32 15 25 65 48 41 52 69 57 18 16 85 34 25 75 21 66 85 48 98 32 15 25 65 48 41 52 69 57 18

LLI    R10, 64                                ; array A address (0x0040)
LLI    R11, 576                               ; array B address (0x0240)
LLI    R12, 40                                ; arrays length

; function innerProduct (R10, R11) R1 {
    LLI     R1, 0                             ; output
    LLI     R14, 0                            ; i loop variable

    ; for (i = 0; i < n; i+=1) {
        FOR:

        BEQ     R14, R12, END_FOR             ; break if k > n
        ADDI    R14, R14, 1                   ; k += 1

        ; Load A[i] and B[i] from memory
        LW      R15, R10, 0                   ; R15 = A[i]
        LW      R16, R11, 0                   ; R16 = B[i]
        MUL     R17, R15, R16                 ; R17 = R15 * R16
        ADD     R1, R1, R17                   ; R1 += R18;

        ; Increment array index 
        ADDI    R10, R10, 4                   ; A += 4 
        ADDI    R11, R11, 4                   ; B += 4 

        J       FOR
    ; }

    END_FOR:
; }

; Store result in MEM(0x00)
LLI    R14, 0
SW     R14, R1, 0