# A time to think, and turn the whole thing around
When the project came up, we believed we would, within a single battery cycle, be able to implement at least the primary map from [The Manual](https://www.amd.com/system/files/TechDocs/24594.pdf).

Not giving ourselves too much time to think away from the text editor screen, our preferred approach to implementing the AMD64 summed up to:

1. Writing three methods, namely **setOpcode**, **setModRM** and **setSIBByte**.
2. Go filling it with "if" statements until inside the body there could be found a block for every possible instruction byte (!).

Roughly 30% of the way into implementing the primary opcode map, the three methods started becoming, even for our bright minds, extremely difficult to read.

So we saw ourselves in need of putting the laptop away for a while, and do some thinking.

And there came a more fundamental approach.

The opcodes are now **opcode** structs, which can in turn, be mapped to values from 0-255.

And in **Go**, the opcode byte map could finally be itself, a```go map[byte]*opcode```.

## 1. The opcode struct

The opcode struct is intended to be the primary building block of the opcode map. It contains informations about mnemonic, operand size, whether it needs extra bytes for arriving at the complete instruction, such as, in **i386** and **AMD64**, are the instruction prefixes, the **ModRM**, **SIB** bytes, 1-4 displacement and/or 1-8 immediate bytes. The fields in the struct are thus named accordingly.

```go
type opcode struct{
	
	opcode byte
	immediateBytes uint8
	displacementBytes uint8
	
	operandSize uint8
	
	dstOperand uint8// whether r/m, reg or imm
	srcOperand uint8// whether r/m, mem or imm
	
	mnemonic string
	
	r_m_str string
	reg_str string
	imm_str string
	
	r_m uint8
	mod uint8
	scale uint8
	index uint8
}
```