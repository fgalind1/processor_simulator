
# Processor Simulator

 - [How-To](#how-to)
   - [Commands](#commands)
     - [Run a program](#run-a-program)
   - [Debugging](#debugging)
     - [Registers and/or memory data](#registers-and-or-memory-data)
 - [Registers](#registers)
 - [Architecture](#architecture)
   - [Status Register](#status-register)
 - [Instructions](#instructions)
   - [Instruction Flow](#instruction-flow)
   - [Instruction Set](#instruction-set)
      - [Instruction Formats](#instruction-formats)
      - [List of Instructions](#list-of-instructions)
         - [Arithmetic](#list-of-instructions)
         - [Data Transfer](#data-transfer)
         - [Control-Flow](#control-flow)

## How-To

### Commands

#### Run a program

```
run <assembly-filename>
    [-s, --step-by-step](bool)        (run interactively step by step. default: false)
    [-d, --data-memory](string)       (filename where to save data memory once the program has finished)
    [-r, --registers-memory](string)  (filename where to save registers memory once the program has finished)
```
Sample: `run samples/sample.txt --step-by-step --data-memory samples/sample.dat --registers-memory samples/sample.reg`

### Debugging

#### Registers and/or memory data

If the program is executed using the flag `-s` or `--step-by-step` you will be able to see the state of registers and/or data memory at the end of every step executed otherwise if the flag is not provided you will be able to see the final state at the end.

The following menu will be presented:

```
Press the desired key and then hit [ENTER]...
 - (R) to see registers memory
 - (D) to see data memory
 - (E) to exit and quit
 - (*) Any other key to continue
```

If selected `R` or `D`, the data will be displayed in the following format:

```
           0x00            0x04            0x08            0x0C
0x00    0x0000000A      0x0010000A      0x000C0000      0x00000000
0x10    0x000100E8      0x00000008      0x00000012      0x00000087
0x20    0x0000FF00      0x00000000      0x00D00068      0x002000A8
0x30    0x000000E8      0x0000C008      0x00000012      0x00000087
0x40    0x00000012      0x00000000      0x00100000      0x00000000
....
```

This format applies also if the flags for saving the memory into files by using  `--data-memory` or `--registers-memory`

## Architecture

The current pipeline architecture is based on a 3-stages pipeline:
 * Fetch
 * Decode
 * Execute
 
The following diagram shows the components of the architecture along with the pipeline and the interaction with its components:

![Processor architecture](/img/architecture.png)

## Registers

   Name                 |  Description                        |
------------------------|-------------------------------------|
 [R0](#status-register) | [Status Register](#status-register) |
 R1 - R31               | Genral Purpose                      |

### Status Register

The status register corresponds to register `R0`

 Bit   | Label |  Description   |
-------|-------|----------------|
 0     |  --   |  Not used      |
 1     |  --   |  Not used      |
 2     |  PF   |  Parity flag   |
 3     |  --   |  Not used      |
 4     |  --   |  Not used      |
 5     |  --   |  Not used      |
 6     |  ZF   |  Zero flag     |
 7     |  SF   |  Sign flag     |
 8     |  --   |  Not used      |
 9     |  --   |  Not used      |
 10    |  --   |  Not used      |
 11    |  OF   |  Overflow flag |
 12-31 |  --   |  Not used      |

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
add   Rd,Rs,Rt | Rd = Rs + Rt   |  R   | 000000 | Rd | Rs | Rt  |  -  |  -  | with overflow              |
addu  Rd,Rs,Rt | Rd = Rs + Rt   |  R   | 000001 | Rd | Rs | Rt  |  -  |  -  | without overflow           |
sub   Rd,Rs,Rt | Rd = Rs - Rt   |  R   | 000010 | Rd | Rs | Rt  |  -  |  -  | with overflow              |
subu  Rd,Rs,Rt | Rd = Rs - Rt   |  R   | 000011 | Rd | Rs | Rt  |  -  |  -  | without overflow           |
addi  Rd,Rs,C  | Rd = Rs + C    |  I   | 000100 | Rd | Rs | Immediate (16)||| immediate with overflow    |
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
j    C         | jump to address |  J   | 100100 |        Target (26)     ||||| load upper immediate | (To be implemented)
