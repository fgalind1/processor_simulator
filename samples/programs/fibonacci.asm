;
; Example of fibonacci algorithm
;

; Input fibonacci

LLI     R1, 15                                ; Fibonnaci of 15  

; -----------------------------------------------------
;             Function fibonacci(R10(number))
; -----------------------------------------------------
;   - Input: R1
;   - Output R2
; -----------------------------------------------------

; function factorial (R10) R11 {
    LLI     R2, 1                             ; Output
    LLI     R12, 1                            ; k (loop index)
    LLI     R13, 0                            ; i variable

    ; for (k = 1; k <= n; k+=1) {
        FIBONACCI:

        BGT     R12, R1, END_FIBONACCI        ; break if k > n
        ADDI    R12, R12, 1                   ; k += 1 

        ADD     R14, R13, R2                  ; temp = i + output;
        ADDI    R13, R2, 0                    ; i = output;
        ADDI    R2, R14, 0                    ; output = temp;

        J       FIBONACCI
    ; }

    END_FIBONACCI:
; }

; Store result in MEM(0x00)
LLI    R14, 0
SW     R14, R2, 0

; -------------------- Output --------------------------
;
;    R2 = fibonacci(R10) = MEM(0x00)
;    
;     N    |  0  1  2  3  4  5  6   7   8   9  ...  15
;  -----------------------------------------------------
;   fib(N) |  1  1  2  3  5  8  13  21  34  55 ...  987
;   fib(N) |  1  1  2  3  5  8  0D  15  22  37 ...  3DB
;   
; ------------------------------------------------------