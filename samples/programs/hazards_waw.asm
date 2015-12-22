;
; Data Hazards
;  - WAW Dependency
;

; Create a RAW first so that the second ADDI tries first to be executed
LLI     R2, 10

ADDI 	R2, R10, 1
ADDI	R2, R20, 1