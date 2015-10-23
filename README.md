
# Processor Simulator

 - [Instructions](#instructions)
   - [Instruction Flow](#instruction-flow)
   - [Instruction Set](#instruction-set)
      - [Instruction Formats](#instruction-formats)
      - [List of Instructions](#list-of-instructions)
         - [Arithmetic](#list-of-instructions)
         - [Data Transfer](#data-transfer)
         - [Control-Flow](#control-flow)

## Instructions

### Instruction Flow

```
   Human Readable  - it is the basic instruction in a assembly file) 
        |          - e.g ADD 14, 18, 14
        V  
     Go Object     - the assembly instruction is translated into a
        |            struct/object to be used by the processor
        V 
   Binary (32-bit) - then it is translated into a 32-bit integer accordignly to formats
        |          - this instruction will be stored in the instructions memory
        V 
     Go Object     - CPU will decode the instruction into an object again and it will
                     execute the instruction accordignly to the definitions
```

### Instruction Set

#### Instruction Formats

 The next tables shows the format structure of the instructions accordingly to the different types: `R, I, J`

 Type | Format (32 bits)||||||
------|------------|--------|--------|--------|----------|----------|
  R   | Opcode (6) | Rd (5) | Rs (5) | Rt (5) | Shmt (5) | Func (6) |
  
 Type | Format (32 bits)||||
 -----|------------|--------|--------|----------------------------------------------------|
  I   | Opcode (6) | Rd (5) | Rs (5) | - I m m e d i a t e (1 6 b i t s) - |
  
 Type | Format (32 bits)||
 -----|------------|----||
  J   | Opcode (6) | - - - - - - - - - - A d d r e s s (2 6 b i t s ) - - - - - - - - - - |

   - All instructions are `32-bit` long (`1 word`)
   - `Rs`, `Rt`, and `Rd` are general purpose registers
   - `PC` stands for the program counter address
   - `C` denotes a constant (immediate)
   - `-` denotes that those values do not care

#### List of Instructions

##### Aritmetic
 - From Opcode **00**0000 to **00**1111
 
    Syntax     |  Description   | Type | 31              |||||            0 |         Notes              |
---------------|----------------|------|--------|----|----|-----|-----|-----|----------------------------|
add   Rd,Rs,Rt | Rd = Rs + Rt   |  R   | 000000 | Rd | Rs | Rt  |  -  |  -  | with overflow              | TBI (Overflow)
addu  Rd,Rs,Rt | Rd = Rs + Rt   |  R   | 000001 | Rd | Rs | Rt  |  -  |  -  | without overflow           |
sub   Rd,Rs,Rt | Rd = Rs - Rt   |  R   | 000010 | Rd | Rs | Rt  |  -  |  -  | with overflow              | TBI (Overflow)
subu  Rd,Rs,Rt | Rd = Rs - Rt   |  R   | 000011 | Rd | Rs | Rt  |  -  |  -  | without overflow           |
addi  Rd,Rs,C  | Rd = Rs + C    |  I   | 000100 | Rd | Rs | Immediate (16)||| immediate with overflow    | TBI (Overflow)
addiu Rd,Rs,C  | Rd = Rs + C    |  I   | 000101 | Rd | Rs | Immediate (16)||| immediate without overflow |
cmp   Rd,Rs,Rt | Rd = Rs <=> Rt |  R   | 000110 | Rd | Rs | Rt  |  -  |  -  | 1 (s<t), 2 (=), 4 (s>t)    |
 

##### Data Transfer
 - From Opcode **01**0000 to **01**1111

    Syntax     |  Description   | Type | 31              |||||             0 |         Notes           |
---------------|----------------|------|--------|----|----|------|-----|-----|-------------------------|
lw    Rd,Rs,C  | Rd = M[Rs + C] |  I   | 010000 | Rd | Rs | Offset (16)    ||| load M[Rs + C] into Rd  |
sw    Rd,Rs,C  | M[Rs + C] = Rd |  I   | 010001 | Rd | Rs | Offset (16)    ||| store Rd into M[Rs + C] |
lli   Rd,C     | Rd = C         |  I   | 010010 | Rd | -  | Immediate (16) ||| load lower immediate    |
sli   Rd,C     | M[Rd] = C      |  I   | 010011 | Rd | -  | Immediate (16) ||| store lower immediate   |
lui   Rd,C     | Rd = C << 16   |  I   | 010100 | Rd | -  | Immediate (16) ||| load upper immediate    |
sui   Rd,C     | M[Rd] = C << 16|  I   | 010101 | Rd | -  | Immediate (16) ||| store upper immediate   |

##### Control-Flow 
 - From Opcode **10**0000 to **10**1111
 
    Syntax     |   Description   | Type | 31              |||||             0 |          Notes       |
---------------|-----------------|------|--------|----|----|------|-----|-----|----------------------|
beq  Rd,Rs,C   | br on equal     |  I   | 100000 | Rd | Rs | Immediate (16) ||| PC = PC + 4 + 4*C    |
bne  Rd,Rs,C   | br on not equal |  I   | 100001 | Rd | Rs | Immediate (16) ||| PC = PC + 4 + 4*C    |
blt  Rd,Rs,C   | br on less      |  I   | 100010 | Rd | Rs | Immediate (16) ||| PC = PC + 4 + 4*C    |
bgt  Rd,Rs,C   | br on greater   |  I   | 100011 | Rd | Rs | Immediate (16) ||| PC = PC + 4 + 4*C    |
j    C         | jump to address |  J   | 100100 |        Target (26)     ||||| load upper immediate | TBI
