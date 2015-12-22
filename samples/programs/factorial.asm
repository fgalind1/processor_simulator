;
; Example of factorial algorithm
;

; Input factorial
LLI     R10, 12                            ; Factorial of 12

; -----------------------------------------------------
;             Function factorial(R10(number))
; -----------------------------------------------------
;   - Input: R10
;   - Output R11
; -----------------------------------------------------

LLI     R12, 1                            ; Temporal register equal to 1
ADDI    R11, R10, 0                       ; Output starting equal to R10
J       FACTORIAL                         ; Call factorial function    

; function factorial (R10) R11 {

    FACTORIAL:

    SUB     R10, R10, R12                 ; Subtract 1 from R10
    MUL     R11, R11, R10                 ; X * (X-1)

    ; if N > 1 return factorial(N-1)
    BGT     R10, R12, FACTORIAL           ; Recursively call factorial function if R10 > 1 (base case)

    ; return
; }


; -------------------- Output -------------------------
;
; R11 = factorial (R10)
; e.g:  7! = 7 x 6 x 5 x 4 x 3 x 2 x 1 = 5040 = 13B0
;
; -----------------------------------------------------