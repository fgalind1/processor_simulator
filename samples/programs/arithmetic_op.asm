;
; Example of bubble sort algorithm
;

; ALU operations

LLI    R0, 107
LLI    R1, 20  
LLI    R2, 5
LLI    R3, 13

ADD    R10, R0, R1  ; Expected 0x7F
ADDI   R11, R0, 20  ; Expected 0x7F
SUB    R12, R0, R1  ; Expected 0x57
SUBI   R13, R0, 20  ; Expected 0x57

MUL    R14, R0, R1  ; Expected 0x85C

SHL    R15, R0, R2  ; Expected 0xD60
SHLI   R16, R0, 5   ; Expected 0xD60
SHR    R17, R0, R2  ; Expected 0x03
SHRI   R18, R0, 5   ; Expected 0x03

AND    R19, R0, R3  ; Expected 0x09
ANDI   R20, R0, 13  ; Expected 0x09
OR     R21, R0, R3  ; Expected 0x6F
ORI    R22, R0, 13  ; Expected 0x6F

; FPU operations

; Load in R24 (135.154 = 0x4307276D)
LUI    R24, 17159         ; 0x4307----
ADDI   R24, R24, 10093    ; 0x----276D

; Load in R25 (5.4512 = 0x40AE703B)
LUI    R25, 16558         ; 0x40AE----
ADDI   R25, R25, 28731    ; 0x----703B

FADD   R26, R24, R25      ; Expected 140.6052 = 0x430C9AEE
FSUB   R27, R24, R25      ; Expected 129.7028 = 0x4301B3EB
FMUL   R28, R24, R25      ; Expected 736.7514848 = 0x44383018
FDIV   R29, R24, R25      ; Expected 24.7934398 = 0x41C658F7

;              Expected Registers Memory
; 
;     	   0x00		   0x04		   0x08		   0x0C
; ...	   ...         ...         ...         ...     
; 0x20	   ...         ...    	0x0000007F	0x0000007F
; 0x30	0x00000057	0x00000057	0x0000085C	0x00000D60
; 0x40	0x00000D60	0x00000003	0x00000003	0x00000009
; 0x50	0x00000009	0x0000006F	0x0000006F	   ...     

; 0x60	0x4307276D	0x40AE703B	0x430C9AEF  0x4301B3EB 
; 0x70	0x44383019	0x44383018	   ...         ...     
; ...	   ...         ...         ...         ...     