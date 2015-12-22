;
; Example of bubble sort algorithm
;    (forward jumps flavor)
;

; -----------------------------------------------------
;    Input  = [ 9, D, 23, A, 7, 18, C, 6, 15, F]
;    Output = [ 6, 7, 9, A, C, D, F, 15, 18, 23]
; -----------------------------------------------------

@0x20: 09 0D 23 0A 07 18 0C 06 15 0F

LLI    R10, 32                        ; array address
LLI    R11, 10                        ; array lenght

; 21 operations

; -----------------------------------------------------
;             Function bubbleSort(array, lengt)
; -----------------------------------------------------
;   - Input values (R10(arrayAddress), R11(lenght))
;   - Constant values (R21 = 1)
;   - Temporal registers (R2 - R7)
; -----------------------------------------------------

; function bubbleSort(R10(arrayAddress), R11(lenght) {

LLI    R21, 1                       ; R21 = 1 (constant)
LLI    R3, 1                        ; swapped = true
ADDI   R2, R11, 0                   ; i += lenght

; for(R2(i) = lenght - 1; i > 0 && swapped; i--) {

    FOR_I:

    SUB    R2, R2, R21              ; i = i - 1
    BLT    R2, R21, END_FOR_I       ; continue if i >= 1, break if i < 1
    BNE    R3, R21, END_FOR_I       ; continue if swapped = true
    
    ; Start instructions
    LLI    R3, 0                    ; swapped = false
    LLI    R4, 0                    ; j = 0
    ADDI   R5, R10, 0               ; R5 = address of index 0

    ; for(R4(j) = 0; j < i; j++) {

        FOR_J:

        BEQ    R2, R4, END_FOR_J    ; break if j == i (forward jump)

        ; Start instructions
        LW     R6, R5, 0            ; R6 = data[j]
        LW     R7, R5, 4            ; R6 = data[j+1]

        ; if (data[j] > data[j+1]) {

            BGT    R6, R7, DO_SWAP  ; swap if data[j] > data[i+1]
            J      END_IF_SWAP      ; break from if

            DO_SWAP:
            ; Start instructions
            SW     R5, R7, 0        ; data[j] = data[j+1]
            SW     R5, R6, 4        ; data[j+1] = data[j]
            LLI    R3, 1            ; swapped = true
        ; }    
        END_IF_SWAP:

        ADDI   R4, R4, 1            ; j = j + 1
        ADDI   R5, R5, 4            ; address += 4 bytes

        J      FOR_J
    ; }
    END_FOR_J:

    J      FOR_I
; }
END_FOR_I:

; -----------------------------------------------------
;        End function bubbleSort(array, lengt)
; -----------------------------------------------------
