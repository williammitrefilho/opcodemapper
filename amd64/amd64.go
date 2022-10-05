package amd64_1

import(
	
//	"fmt"
//	"log"
//	"errors"
)

const(
	
	SIB = 0x19
	MOD_R_M = 0x18
	R_M = 0x17
	REG = 0x16
	IMM = 0x15
	NONE = 0x14
	
	AH = 0x13
	BH = 0x12
	DH = 0x11
	CH = 0x10
	BL = 0x0F
	DL = 0x0E
	CL = 0x0D
	RFLAGS = 0x0C
	DX = 0x0B
	ES_SI = 0x0A
	ES_DI = 9
	AL = 8
		
	GPRNames = "RAXRCXRDXRBXRSPRBPRSIRDI_ALMDIMSI_DXFLA_CL_DL_BL_CH_DH_BH_AH"
	
	RAX = 0
	RCX = 1
	RDX = 2
	RBX = 3
	RSP = 4
	RBP = 5
	RSI = 6
	RDI = 7
)

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

func regStr(reg byte) string{
	
	if reg > AH{
		
		return ""
	}
	pos := 3*reg
	return GPRNames[pos:pos+3]
}

func (opc *opcode) setModRM(b byte){
	
	mod := b>>6
	reg := (b&0x38) >> 3
	r_m := b&0x7

	if opc.mnemonic == ""{//The operation will be defined by the ModRM.reg field
		
		extendedOp := extendedOpcodeMap[opc.opcode][reg]
		old_opc := opc.opcode
		*opc = *extendedOp
		opc.opcode = old_opc
		opc.mnemonic = extendedOp.mnemonic
		opc.dstOperand = extendedOp.dstOperand
		opc.srcOperand = extendedOp.srcOperand

	} else {
		
		opc.reg_str = regStr(reg)
	}
// now let's move on to the ModRM.r/m field
	r_m_info := modRM_map[mod][r_m]
	opc.displacementBytes = r_m_info.displacementBytes
	opc.r_m = r_m_info.r_m
	opc.mod = mod
	
	if(opc.dstOperand == R_M){
		
		opc.dstOperand = opc.r_m
	} else if opc.dstOperand == REG{
		
		opc.dstOperand = reg
	}
	if(opc.srcOperand == R_M){
		
		opc.srcOperand = opc.r_m
	} else if opc.srcOperand == REG{
		
		opc.srcOperand = reg
	}
}

func (opc *opcode) needsModRM() bool{
	
	return opc.dstOperand == R_M || opc.srcOperand == R_M || opc.dstOperand == REG || opc.srcOperand == REG || opc.mnemonic == ""
}

func (opc *opcode) setSIBByte(b byte){
	
	scale := b>>6
	index := (b&0x38)>>3
	base := b&0x7
	
	opc.scale = 1<<scale
	if index == 4{
		
		opc.index = NONE
	} else{
		
		opc.index = index
	}
	
	if base == 5 && opc.mod == 0{
		
		opc.r_m = NONE
	} else{
		
		opc.r_m = base
	}
}

type instruction struct{
	
	rEXPrefix byte
	bytes [32]byte
	nBytes uint8
	opcode *opcode
	finished bool
	
	immediate int64
	displacement int64
}

func FromSlice(bytes []byte) int64{
	
	var num int64 = 0
	for i:=0; i < len(bytes); i++{
		
		num |= (int64)(bytes[i]) << (8*i)
	}
	return num
}

func (i *instruction) feedByte(b byte) uint8{
	
	if i.nBytes == 32{
		
		return 0
	}
	i.bytes[i.nBytes] = b
	i.nBytes++
	if i.opcode == nil{
		
		opc := opcodeMap[b]
		if opc != nil{
			
			i.opcode = new(opcode)
			*i.opcode = *opc

			if i.opcode.immediateBytes == 9{
				
				if i.rEXPrefix&0x8 > 0{
					
					i.opcode.immediateBytes = 8
				} else{
					
					i.opcode.immediateBytes = 4
				}
			}
			if i.opcode.needsModRM(){
				
				return 1
			}
			i.finished = true
			return 0
		} else if prefixMap[b]{
			
			if (b&0xF0) == 0x40{
				
				i.rEXPrefix = b
			}
			return 1
		}
	} else {
		
		if i.opcode.needsModRM(){
			
			i.opcode.setModRM(b)
			if i.opcode.r_m == SIB{
				
				return 1
			}
			i.finished = true
			return 0
		} else {
			
			if i.opcode.r_m == SIB{
				
				i.opcode.setSIBByte(b)
				i.finished = true
				return 0
			}
		}
	}
	return 0
}

func InstructionFromBytes(bytes []byte) *instruction{
	
	instr := new(instruction)
	i := 0
	for i < len(bytes){
		
		r := instr.feedByte(bytes[i])
		i++
		if instr.finished || r == 0{
			
			break
		}
	}
	if !instr.finished{
		
		return nil
	}
	if instr.opcode.displacementBytes > 0 {
		
		if (i + (int)(instr.opcode.displacementBytes)) > len(bytes){
			
			return nil
		}
		instr.displacement = FromSlice(bytes[i:i+(int)(instr.opcode.displacementBytes)])
		i += (int)(instr.opcode.displacementBytes)
	}
	if instr.opcode.immediateBytes > 0 {
		
		if (i + (int)(instr.opcode.immediateBytes)) > len(bytes){
			
			return nil
		}
		instr.immediate = FromSlice(bytes[i:i+(int)(instr.opcode.immediateBytes)])
	}
	
	return instr
}

type modRM_mapEntry struct{
	
	r_m uint8
	displacementBytes uint8
}

type opcode_map map[byte]*opcode

var modRM_map0, modRM_map1, modRM_map2, modRM_map3 = []modRM_mapEntry{
	{RAX, 0}, {RCX, 0}, {RDX, 0}, {RBX, 0}, {SIB, 0}, {NONE, 4}, {RSI, 0}, {RDI, 0},
}, []modRM_mapEntry{
	{RAX, 1}, {RCX, 1}, {RDX, 1}, {RBX, 1}, {SIB, 1}, {RBP, 1}, {RSI, 1}, {RDI, 1},
}, []modRM_mapEntry{
	{RAX, 4}, {RCX, 4}, {RDX, 4}, {RBX, 4}, {SIB, 4}, {RBP, 4}, {RSI, 4}, {RDI, 4},
}, []modRM_mapEntry{
	{RAX, 0}, {RCX, 0}, {RDX, 0}, {RBX, 0}, {RSP, 0}, {RBP, 4}, {RSI, 0}, {RDI, 0},
}

var modRM_map = [][]modRM_mapEntry{modRM_map0, modRM_map1, modRM_map2, modRM_map3,}

var prefixMap = map[byte]bool{
	
	0x66:true, 0x67:true, 0x2E:true, 0x3E:true, 0x26:true, 0x64:true, 0x65:true,
	0x36:true, 0xF0:true, 0xF3:true, 0xF2:true,
	0x40:true, 0x41:true, 0x42:true, 0x43:true, 0x44:true, 0x45:true, 0x46:true,
	0x47:true, 0x48:true, 0x49:true, 0x4A:true, 0x4B:true, 0x4C:true, 0x4D:true,
	0x4E:true, 0x4F:true,
}
var opcodeMap = opcode_map{
	
	0x00:{0x00, 0, 0, 1, R_M, REG, "ADD", "", "", "", 0, 0, 0, 0},
	0x01:{0x01, 0, 0, 9, R_M, REG, "ADD", "", "", "", 0, 0, 0, 0},
	0x02:{0x02, 0, 0, 1, REG, R_M, "ADD", "", "", "", 0, 0, 0, 0},
	0x03:{0x03, 0, 0, 9, REG, R_M, "ADD", "", "", "", 0, 0, 0, 0},
	0x04:{0x04, 1, 0, 1, AL, NONE, "ADD", "", "", "", 0, 0, 0, 0},
	0x05:{0x05, 1, 0, 9, RAX, NONE, "ADD", "", "", "", 0, 0, 0, 0},
//	0x06:{0x06, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},Invalid in 64-bit mode
//	0x07:{0x07, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},Invalid in 64-bit mode
	0x08:{0x08, 0, 0, 1, R_M, REG, "OR", "", "", "", 0, 0, 0, 0},
	0x09:{0x09, 0, 0, 9, R_M, REG, "OR", "", "", "", 0, 0, 0, 0},
	0x0A:{0x0A, 0, 0, 1, REG, R_M, "OR", "", "", "", 0, 0, 0, 0},
	0x0B:{0x0B, 0, 0, 9, REG, R_M, "OR", "", "", "", 0, 0, 0, 0},
	0x0C:{0x0C, 1, 0, 1, AL, NONE, "OR", "", "", "", 0, 0, 0, 0},
	0x0D:{0x0D, 1, 0, 9, RAX, NONE, "OR", "", "", "", 0, 0, 0, 0},
//	0x0E:{0x0E, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},Invalid in 64-bit mode
//	0x0F:{0x0F, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},Invalid in 64-bit mode
	0x10:{0x10, 0, 0, 1, R_M, REG, "ADC", "", "", "", 0, 0, 0, 0},
	0x11:{0x11, 0, 0, 1, R_M, REG, "ADC", "", "", "", 0, 0, 0, 0},
	0x12:{0x12, 0, 0, 1, R_M, REG, "ADC", "", "", "", 0, 0, 0, 0},
	0x13:{0x13, 0, 0, 1, R_M, REG, "ADC", "", "", "", 0, 0, 0, 0},
	0x14:{0x14, 0, 0, 1, R_M, REG, "ADC", "", "", "", 0, 0, 0, 0},
	0x15:{0x15, 0, 0, 1, R_M, REG, "ADC", "", "", "", 0, 0, 0, 0},
//	0x16:{0x16, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},Invalid in 64-bit mode
//	0x17:{0x17, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},Invalid in 64-bit mode
	0x18:{0x18, 0, 0, 1, R_M, REG, "SBB", "", "", "", 0, 0, 0, 0},
	0x19:{0x19, 0, 0, 9, R_M, REG, "SBB", "", "", "", 0, 0, 0, 0},
	0x1A:{0x1A, 0, 0, 1, REG, R_M, "SBB", "", "", "", 0, 0, 0, 0},
	0x1B:{0x1B, 0, 0, 9, REG, R_M, "SBB", "", "", "", 0, 0, 0, 0},
	0x1C:{0x1C, 1, 0, 1, AL, NONE, "SBB", "", "", "", 0, 0, 0, 0},
	0x1D:{0x1D, 4, 0, 9, RAX, NONE, "SBB", "", "", "", 0, 0, 0, 0},
//	0x1E:{0x1E, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},Invalid in 64-bit mode
//	0x1F:{0x1F, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},Escape to secondary opcode map[not yet implemented]
	0x20:{0x20, 0, 0, 1, R_M, REG, "AND", "", "", "", 0, 0, 0, 0},
	0x21:{0x21, 0, 0, 1, R_M, REG, "AND", "", "", "", 0, 0, 0, 0},
	0x22:{0x22, 0, 0, 1, REG, R_M, "AND", "", "", "", 0, 0, 0, 0},
	0x23:{0x23, 0, 0, 1, REG, R_M, "AND", "", "", "", 0, 0, 0, 0},
	0x24:{0x24, 0, 0, 1, AL, NONE, "AND", "", "", "", 0, 0, 0, 0},
	0x25:{0x25, 0, 0, 1, RAX, NONE, "AND", "", "", "", 0, 0, 0, 0},
//	0x26:{0x26, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},seg ES
//	0x27:{0x27, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},Invalid in 64-bit mode
	0x28:{0x28, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
	0x29:{0x29, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
	0x2A:{0x2A, 0, 0, 1, REG, R_M, "", "", "", "", 0, 0, 0, 0},
	0x2B:{0x2B, 0, 0, 1, REG, R_M, "", "", "", "", 0, 0, 0, 0},
	0x2C:{0x2C, 0, 0, 1, AL, NONE, "", "", "", "", 0, 0, 0, 0},
	0x2D:{0x2D, 0, 0, 1, RAX, NONE, "", "", "", "", 0, 0, 0, 0},
//	0x2E:{0x2E, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0}, seg CS
//	0x2F:{0x2F, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},Invalid in 64-bit mode
	0x30:{0x30, 0, 0, 1, R_M, REG, "XOR", "", "", "", 0, 0, 0, 0},
	0x31:{0x31, 0, 0, 1, R_M, REG, "XOR", "", "", "", 0, 0, 0, 0},
	0x32:{0x32, 0, 0, 1, REG, R_M, "XOR", "", "", "", 0, 0, 0, 0},
	0x33:{0x33, 0, 0, 1, REG, R_M, "XOR", "", "", "", 0, 0, 0, 0},
	0x34:{0x34, 0, 0, 1, AL, NONE, "XOR", "", "", "", 0, 0, 0, 0},
	0x35:{0x35, 0, 0, 1, RAX, NONE, "XOR", "", "", "", 0, 0, 0, 0},
//	0x36:{0x36, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},seg SS
//	0x37:{0x37, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},Invalid in 64-bit mode
	0x38:{0x38, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
	0x39:{0x39, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
	0x3A:{0x3A, 0, 0, 1, REG, R_M, "", "", "", "", 0, 0, 0, 0},
	0x3B:{0x3B, 0, 0, 1, REG, R_M, "", "", "", "", 0, 0, 0, 0},
	0x3C:{0x3C, 0, 0, 1, AL, NONE, "", "", "", "", 0, 0, 0, 0},
	0x3D:{0x3D, 0, 0, 1, RAX, NONE, "", "", "", "", 0, 0, 0, 0},
//	0x3E:{0x3E, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},seg DS
//	0x3F:{0x3F, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},Invalid in 64-bit mode
//	0x40:{0x40, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},0x4_ bytes are used as
//	0x41:{0x41, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},REX prefixes in 64-bit mode
//	0x42:{0x42, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
//	0x43:{0x43, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
//	0x44:{0x44, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
//	0x45:{0x45, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
//	0x46:{0x46, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
//	0x47:{0x47, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
//	0x48:{0x48, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
//	0x49:{0x49, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
//	0x4A:{0x4A, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
//	0x4B:{0x4B, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
//	0x4C:{0x4C, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
//	0x4D:{0x4D, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
//	0x4E:{0x4E, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
//	0x4F:{0x4F, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
	0x50:{0x50, 0, 0, 9, RAX, NONE, "PUSH", "", "", "", 0, 0, 0, 0},
	0x51:{0x51, 0, 0, 9, RCX, NONE, "PUSH", "", "", "", 0, 0, 0, 0},
	0x52:{0x52, 0, 0, 9, RDX, NONE, "PUSH", "", "", "", 0, 0, 0, 0},
	0x53:{0x53, 0, 0, 9, RBX, NONE, "PUSH", "", "", "", 0, 0, 0, 0},
	0x54:{0x54, 0, 0, 9, RSP, NONE, "PUSH", "", "", "", 0, 0, 0, 0},
	0x55:{0x55, 0, 0, 9, RBP, NONE, "PUSH", "", "", "", 0, 0, 0, 0},
	0x56:{0x56, 0, 0, 9, RSI, NONE, "PUSH", "", "", "", 0, 0, 0, 0},
	0x57:{0x57, 0, 0, 9, RDI, NONE, "PUSH", "", "", "", 0, 0, 0, 0},
	0x58:{0x58, 0, 0, 9, RAX, NONE, "", "", "", "", 0, 0, 0, 0},
	0x59:{0x59, 0, 0, 9, RCX, NONE, "", "", "", "", 0, 0, 0, 0},
	0x5A:{0x5A, 0, 0, 9, RDX, NONE, "", "", "", "", 0, 0, 0, 0},
	0x5B:{0x5B, 0, 0, 9, RBX, NONE, "", "", "", "", 0, 0, 0, 0},
	0x5C:{0x5C, 0, 0, 9, RSP, NONE, "", "", "", "", 0, 0, 0, 0},
	0x5D:{0x5D, 0, 0, 9, RBP, NONE, "", "", "", "", 0, 0, 0, 0},
	0x5E:{0x5E, 0, 0, 9, RSI, NONE, "", "", "", "", 0, 0, 0, 0},
	0x5F:{0x5F, 0, 0, 9, RDI, NONE, "", "", "", "", 0, 0, 0, 0},
//	0x60:{0x00, 0, 0, 1, NONE, NONE, "PUSHA", "", "", "", 0, 0, 0, 0},Invalid in 64-bit mode
//	0x61:{0x61, 0, 0, 1, R_M, REG, "POPA", "", "", "", 0, 0, 0, 0},Invalid in 64-bit mode
//	0x62:{0x62, 0, 0, 1, R_M, REG, "BOUND", "", "", "", 0, 0, 0, 0},Invalid in 64-bit mode
	0x63:{0x63, 0, 0, 9, REG, R_M, "MOVSXD", "", "", "", 0, 0, 0, 0},
//	0x64:{0x64, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},seg FS prefix
//	0x65:{0x65, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},seg GS prefix
//	0x66:{0x66, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},operand size override prefix
//	0x67:{0x67, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},address size override prefix
	0x68:{0x68, 4, 0, 0, NONE, NONE, "PUSH", "", "", "", 0, 0, 0, 0},
	0x69:{0x69, 4, 0, 9, REG, R_M, "IMUL", "", "", "", 0, 0, 0, 0},
	0x6A:{0x6A, 1, 0, 0, NONE, NONE, "PUSH", "", "", "", 0, 0, 0, 0},
	0x6B:{0x6B, 1, 0, 9, REG, R_M, "IMUL", "", "", "", 0, 0, 0, 0},
	0x6C:{0x6C, 0, 0, 1, ES_DI, DX, "INSB", "", "", "", 0, 0, 0, 0},
	0x6D:{0x6D, 0, 0, 9, ES_DI, DX, "INSW/D", "", "", "", 0, 0, 0, 0},
	0x6E:{0x6E, 0, 0, 1, DX, ES_SI, "OUTS/OUTSB", "", "", "", 0, 0, 0, 0},
	0x6F:{0x6F, 0, 0, 9, DX, ES_SI, "OUTS/OUTSW/D", "", "", "", 0, 0, 0, 0},
	0x70:{0x70, 1, 0, 0, NONE, NONE, "JO", "", "", "", 0, 0, 0, 0},
	0x71:{0x71, 1, 0, 0, NONE, NONE, "JNO", "", "", "", 0, 0, 0, 0},
	0x72:{0x72, 1, 0, 0, NONE, NONE, "JB", "", "", "", 0, 0, 0, 0},
	0x73:{0x73, 1, 0, 0, NONE, NONE, "JNB", "", "", "", 0, 0, 0, 0},
	0x74:{0x74, 1, 0, 0, NONE, NONE, "JZ", "", "", "", 0, 0, 0, 0},
	0x75:{0x75, 1, 0, 0, NONE, NONE, "JNZ", "", "", "", 0, 0, 0, 0},
	0x76:{0x76, 1, 0, 0, NONE, NONE, "JBE", "", "", "", 0, 0, 0, 0},
	0x77:{0x77, 1, 0, 0, NONE, NONE, "JNBE", "", "", "", 0, 0, 0, 0},
	0x78:{0x78, 1, 0, 0, NONE, NONE, "JS", "", "", "", 0, 0, 0, 0},
	0x79:{0x79, 1, 0, 0, NONE, NONE, "JNS", "", "", "", 0, 0, 0, 0},
	0x7A:{0x7A, 1, 0, 0, NONE, NONE, "JP", "", "", "", 0, 0, 0, 0},
	0x7B:{0x7B, 1, 0, 0, NONE, NONE, "JNP", "", "", "", 0, 0, 0, 0},
	0x7C:{0x7C, 1, 0, 0, NONE, NONE, "JL", "", "", "", 0, 0, 0, 0},
	0x7D:{0x7D, 1, 0, 0, NONE, NONE, "JNL", "", "", "", 0, 0, 0, 0},
	0x7E:{0x7E, 1, 0, 0, NONE, NONE, "JLE", "", "", "", 0, 0, 0, 0},
	0x7F:{0x7F, 1, 0, 0, NONE, NONE, "JNLE", "", "", "", 0, 0, 0, 0},
	0x80:{0x80, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},//Group 1
	0x81:{0x81, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},//Group 1
//	0x82:{0x82, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},//Group 1, but I.I.64-B.M.
	0x83:{0x83, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},//Group 1
	0x84:{0x84, 0, 0, 1, R_M, REG, "TEST", "", "", "", 0, 0, 0, 0},
	0x85:{0x85, 0, 0, 1, R_M, REG, "TEST", "", "", "", 0, 0, 0, 0},
	0x86:{0x86, 0, 0, 1, R_M, REG, "XCHG", "", "", "", 0, 0, 0, 0},
	0x87:{0x87, 0, 0, 1, R_M, REG, "XCHG", "", "", "", 0, 0, 0, 0},
	0x88:{0x88, 0, 0, 1, R_M, REG, "MOV", "", "", "", 0, 0, 0, 0},
	0x89:{0x89, 0, 0, 9, R_M, REG, "MOV", "", "", "", 0, 0, 0, 0},
	0x8A:{0x8A, 0, 0, 1, R_M, REG, "MOV", "", "", "", 0, 0, 0, 0},
	0x8B:{0x8B, 0, 0, 9, R_M, REG, "MOV", "", "", "", 0, 0, 0, 0},
//	0x8C:{0x8C, 0, 0, 1, MOD_R_M, REG, "MOV", "", "", "", 0, 0, 0, 0},To be mplemented
	0x8D:{0x8D, 0, 0, 9, REG, MOD_R_M, "LEA", "", "", "", 0, 0, 0, 0},
	0x8E:{0x8E, 0, 0, 4, REG, R_M, "MOV", "", "", "", 0, 0, 0, 0},
//	0x8F:{0x8F, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},XOP escape prefix
//	0x90:{0x90, 0, 0, 1, R_M, REG, "XCHG", "", "", "", 0, 0, 0, 0},T.B.I.
	0x91:{0x91, 0, 0, 9, RCX, RAX, "XCHG", "", "", "", 0, 0, 0, 0},
	0x92:{0x92, 0, 0, 9, RDX, RAX, "XCHG", "", "", "", 0, 0, 0, 0},
	0x93:{0x93, 0, 0, 9, RBX, RAX, "XCHG", "", "", "", 0, 0, 0, 0},
	0x94:{0x94, 0, 0, 9, RSP, RAX, "XCHG", "", "", "", 0, 0, 0, 0},
	0x95:{0x95, 0, 0, 9, RBP, RAX, "XCHG", "", "", "", 0, 0, 0, 0},
	0x96:{0x96, 0, 0, 9, RSI, RAX, "XCHG", "", "", "", 0, 0, 0, 0},
	0x97:{0x97, 0, 0, 0, NONE, NONE, "CBW", "", "", "", 0, 0, 0, 0},
	0x99:{0x99, 0, 0, 0, NONE, NONE, "CWD", "", "", "", 0, 0, 0, 0},
//	0x9A:{0x9A, 0, 0, 1, R_M, RAX, "", "", "", "", 0, 0, 0, 0},I.I.64-B.M.
	0x9B:{0x9B, 0, 0, 0, NONE, NONE, "WAIT", "", "", "", 0, 0, 0, 0},
	0x9C:{0x9C, 0, 0, 9, RFLAGS, NONE, "PUSHF", "", "", "", 0, 0, 0, 0},
	0x9D:{0x9D, 0, 0, 9, RFLAGS, NONE, "POPF", "", "", "", 0, 0, 0, 0},
	0x9E:{0x9E, 0, 0, 0, NONE, NONE, "SAHF", "", "", "", 0, 0, 0, 0},
	0x9F:{0x9F, 0, 0, 0, NONE, NONE, "LAHF", "", "", "", 0, 0, 0, 0},
	0xA0:{0xA0, 1, 0, 1, AL, NONE, "MOV", "", "", "", 0, 0, 0, 0},
	0xA1:{0xA1, 9, 0, 9, RAX, NONE, "MOV", "", "", "", 0, 0, 0, 0},
	0xA3:{0xA3, 1, 0, 1, NONE, AL, "MOV", "", "", "", 0, 0, 0, 0},
	0xA4:{0xA4, 0, 0, 1, ES_DI, ES_SI, "MOVSB", "", "", "", 0, 0, 0, 0},
	0xA5:{0xA5, 0, 0, 9, ES_DI, ES_SI, "MOVSW/D/Q", "", "", "", 0, 0, 0, 0},
	0xA6:{0xA6, 0, 0, 1, ES_SI, ES_DI, "CMPSB", "", "", "", 0, 0, 0, 0},
	0xA7:{0xA7, 0, 0, 9, ES_SI, ES_DI, "CMPSW/D/Q", "", "", "", 0, 0, 0, 0},
	0xA8:{0xA8, 1, 0, 1, AL, NONE, "TEST", "", "", "", 0, 0, 0, 0},
	0xA9:{0xA9, 4, 0, 9, RAX, NONE, "TEST", "", "", "", 0, 0, 0, 0},
	0xAA:{0xAA, 0, 0, 1, ES_DI, AL, "STOSB", "", "", "", 0, 0, 0, 0},
	0xAB:{0xAB, 0, 0, 9, ES_DI, RAX, "STOSW", "", "", "", 0, 0, 0, 0},
	0xAC:{0xAC, 0, 0, 1, AL, ES_SI, "LODSB", "", "", "", 0, 0, 0, 0},
	0xAD:{0xAD, 0, 0, 9, RAX, ES_SI, "LODSW/D/Q", "", "", "", 0, 0, 0, 0},
	0xAE:{0xAE, 0, 0, 1, AL, ES_DI, "SCASB", "", "", "", 0, 0, 0, 0},
	0xAF:{0xAF, 0, 0, 1, RAX, ES_DI, "SCASW/D/Q", "", "", "", 0, 0, 0, 0},
	0xB0:{0xB0, 1, 0, 1, AL, NONE, "MOV", "", "", "", 0, 0, 0, 0},
	0xB1:{0xB1, 1, 0, 1, CL, NONE, "MOV", "", "", "", 0, 0, 0, 0},
	0xB2:{0xB2, 1, 0, 1, DL, NONE, "MOV", "", "", "", 0, 0, 0, 0},
	0xB3:{0xB3, 1, 0, 1, BL, NONE, "MOV", "", "", "", 0, 0, 0, 0},
	0xB4:{0xB4, 1, 0, 1, AH, NONE, "MOV", "", "", "", 0, 0, 0, 0},
	0xB5:{0xB5, 1, 0, 1, CH, NONE, "MOV", "", "", "", 0, 0, 0, 0},
	0xB6:{0xB6, 1, 0, 1, DH, NONE, "MOV", "", "", "", 0, 0, 0, 0},
	0xB7:{0xB7, 1, 0, 1, BH, NONE, "MOV", "", "", "", 0, 0, 0, 0},
	0xB8:{0xB8, 9, 0, 9, RAX, NONE, "MOV", "", "", "", 0, 0, 0, 0},
	0xB9:{0xB9, 9, 0, 9, RCX, NONE, "MOV", "", "", "", 0, 0, 0, 0},
	0xBA:{0xBA, 9, 0, 9, RDX, NONE, "MOV", "", "", "", 0, 0, 0, 0},
	0xBB:{0xBB, 9, 0, 9, RBX, NONE, "MOV", "", "", "", 0, 0, 0, 0},
	0xBC:{0xBC, 9, 0, 9, RSP, NONE, "MOV", "", "", "", 0, 0, 0, 0},
	0xBD:{0xBD, 9, 0, 9, RBP, NONE, "MOV", "", "", "", 0, 0, 0, 0},
	0xBE:{0xBE, 9, 0, 9, RSI, NONE, "MOV", "", "", "", 0, 0, 0, 0},
	0xBF:{0xBF, 9, 0, 9, RDI, NONE, "MOV", "", "", "", 0, 0, 0, 0},
	0xC0:{0xC0, 0, 0, 0, NONE, NONE, "", "", "", "", 0, 0, 0, 0},//Group 2
	0xC1:{0xC1, 0, 0, 0, NONE, NONE, "", "", "", "", 0, 0, 0, 0},//Group 2
	0xC2:{0xC2, 2, 0, 0, NONE, NONE, "RET", "", "", "", 0, 0, 0, 0},
	0xC3:{0xC3, 0, 0, 0, NONE, NONE, "RET", "", "", "", 0, 0, 0, 0},
//	0xC4:{0xC4, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},Invalid in 64-bit mode
//	0xC5:{0xC5, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},Invalid in 64-bit mode
	0xC6:{0xC6, 0, 0, 0, NONE, NONE, "", "", "", "", 0, 0, 0, 0},//Group 11
	0xC7:{0xC7, 0, 0, 0, NONE, NONE, "", "", "", "", 0, 0, 0, 0},//Group 11
	0xC8:{0xC8, 1, 0, 1, NONE, NONE, "ENTER", "", "", "", 0, 0, 0, 0},
	0xC9:{0x09, 0, 0, 0, NONE, NONE, "LEAVE", "", "", "", 0, 0, 0, 0},
	0xCA:{0xCA, 2, 0, 2, NONE, NONE, "RET", "", "", "", 0, 0, 0, 0},
	0xCB:{0xCB, 0, 0, 0, NONE, NONE, "RET", "", "", "", 0, 0, 0, 0},
	0xCC:{0xCC, 0, 0, 0, NONE, NONE, "INT3", "", "", "", 0, 0, 0, 0},
	0xCD:{0xCD, 1, 0, 0, NONE, NONE, "INT", "", "", "", 0, 0, 0, 0},
//	0xCE:{0xCE, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},I.I.64-B.M.
	0xCF:{0xCF, 0, 0, 9, NONE, NONE, "IRET", "", "", "", 0, 0, 0, 0},
	0xD0:{0xD0, 0, 0, 0, NONE, NONE, "", "", "", "", 0, 0, 0, 0},//Group 2
	0xD1:{0xD1, 0, 0, 0, NONE, NONE, "", "", "", "", 0, 0, 0, 0},//Group 2
	0xD2:{0xD2, 0, 0, 0, NONE, NONE, "", "", "", "", 0, 0, 0, 0},//Group 2
	0xD3:{0xD3, 0, 0, 0, NONE, NONE, "", "", "", "", 0, 0, 0, 0},//Group 2
//	0xD4:{0xD4, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},I.I.64-B.M.
//	0xD5:{0xD5, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},I.I.64-B.M.
//	0xD6:{0xD6, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},// Invalid
	0xD7:{0xD7, 0, 0, 0, NONE, NONE, "XLAT", "", "", "", 0, 0, 0, 0},
//	0xD8:{0xD8, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
//	0xD9:{0xD9, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
//	0xDA:{0xDA, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},0xD8 thru 0xDF
//	0xDB:{0xDB, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},are x87 instructions,
//	0xDC:{0xDC, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},yet to be implemented
//	0xDD:{0xDD, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
//	0xDE:{0xDE, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
//	0xDF:{0xDF, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0},
	0xE0:{0xE0, 1, 0, 1, NONE, NONE, "LOOPNE/NZ", "", "", "", 0, 0, 0, 0},
	0xE1:{0xE1, 1, 0, 1, NONE, NONE, "LOOPE/Z", "", "", "", 0, 0, 0, 0},
	0xE2:{0xE2, 1, 0, 1, NONE, NONE, "LOOP", "", "", "", 0, 0, 0, 0},
	0xE3:{0xE3, 1, 0, 1, NONE, NONE, "JrCXZ", "", "", "", 0, 0, 0, 0},
	0xE4:{0xE4, 1, 0, 1, AL, NONE, "IN", "", "", "", 0, 0, 0, 0},
	0xE5:{0xE5, 1, 0, 1, RAX, NONE, "IN", "", "", "", 0, 0, 0, 0},
	0xE6:{0xE6, 1, 0, 1, AL, NONE, "OUT", "", "", "", 0, 0, 0, 0},
	0xE7:{0xE7, 1, 0, 1, RAX, NONE, "OUT", "", "", "", 0, 0, 0, 0},
	0xE8:{0xE8, 4, 0, 0, NONE, NONE, "CALL", "", "", "", 0, 0, 0, 0},
	0xE9:{0xE9, 4, 0, 0, NONE, NONE, "JMP", "", "", "", 0, 0, 0, 0},
//	0xEA:{0xEA, 0, 0, 1, R_M, REG, "JMP", "", "", "", 0, 0, 0, 0},I.I.64-B.M.
	0xEB:{0xEB, 1, 0, 0, NONE, NONE, "JMP", "", "", "", 0, 0, 0, 0},
	0xEC:{0xEC, 0, 0, 9, AL, DX, "IN", "", "", "", 0, 0, 0, 0},
	0xED:{0xED, 0, 0, 9, RAX, DX, "IN", "", "", "", 0, 0, 0, 0},
	0xEE:{0xEE, 0, 0, 9, DX, AL, "OUT", "", "", "", 0, 0, 0, 0},
	0xEF:{0xEF, 0, 0, 9, DX, RAX, "OUT", "", "", "", 0, 0, 0, 0},
//	0xF0:{0xF0, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0}, LOCK Prefix
	0xF1:{0xF1, 0, 0, 1, R_M, REG, "INT1", "", "", "", 0, 0, 0, 0},
//	0xF2:{0xF2, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0}, REPNE Prefix
//	0xF3:{0xF3, 0, 0, 1, R_M, REG, "", "", "", "", 0, 0, 0, 0}, REP/REPE Prefix
	0xF4:{0xF4, 0, 0, 0, NONE, NONE, "HLT", "", "", "", 0, 0, 0, 0},
	0xF5:{0xF5, 0, 0, 0, NONE, NONE, "CMC", "", "", "", 0, 0, 0, 0},
	0xF6:{0xF6, 0, 0, 0, NONE, NONE, "", "", "", "", 0, 0, 0, 0},//Group 3
	0xF7:{0xF7, 0, 0, 0, NONE, NONE, "", "", "", "", 0, 0, 0, 0},//Group 3
	0xF8:{0xF8, 0, 0, 0, NONE, NONE, "CLC", "", "", "", 0, 0, 0, 0},
	0xF9:{0xF9, 0, 0, 0, NONE, NONE, "STC", "", "", "", 0, 0, 0, 0},
	0xFA:{0xFA, 0, 0, 0, NONE, NONE, "CLI", "", "", "", 0, 0, 0, 0},
	0xFB:{0xFB, 0, 0, 0, NONE, NONE, "STI", "", "", "", 0, 0, 0, 0},
	0xFC:{0xFC, 0, 0, 0, NONE, NONE, "CLD", "", "", "", 0, 0, 0, 0},
	0xFD:{0xFD, 0, 0, 0, NONE, NONE, "STD", "", "", "", 0, 0, 0, 0},
	0xFE:{0xFE, 0, 0, 0, NONE, NONE, "", "", "", "", 0, 0, 0, 0},//Group 4
	0xFF:{0xFF, 0, 0, 0, NONE, NONE, "", "", "", "", 0, 0, 0, 0},//Group 4
}

var extendedOpcodeMap = map[byte]opcode_map{
	
	0x80:{
		
		0:{0x80, 1, 0, 1, R_M, NONE, "ADD", "", "", "", 0, 0, 0, 0},
		1:{0x80, 1, 0, 1, R_M, NONE, "OR", "", "", "", 0, 0, 0, 0},
		2:{0x80, 1, 0, 1, R_M, NONE, "ADC", "", "", "", 0, 0, 0, 0},
		3:{0x80, 1, 0, 1, R_M, NONE, "SBB", "", "", "", 0, 0, 0, 0},
		4:{0x80, 1, 0, 1, R_M, NONE, "AND", "", "", "", 0, 0, 0, 0},
		5:{0x80, 1, 0, 1, R_M, NONE, "SUB", "", "", "", 0, 0, 0, 0},
		6:{0x80, 1, 0, 1, R_M, NONE, "XOR", "", "", "", 0, 0, 0, 0},
		7:{0x80, 1, 0, 1, R_M, NONE, "CMP", "", "", "", 0, 0, 0, 0},
	},
	0x81:{
		
		0:{0x81, 4, 0, 9, R_M, NONE, "ADD", "", "", "", 0, 0, 0, 0},
		1:{0x81, 4, 0, 9, R_M, NONE, "OR", "", "", "", 0, 0, 0, 0},
		2:{0x81, 4, 0, 9, R_M, NONE, "ADC", "", "", "", 0, 0, 0, 0},
		3:{0x81, 4, 0, 9, R_M, NONE, "SBB", "", "", "", 0, 0, 0, 0},
		4:{0x81, 4, 0, 9, R_M, NONE, "AND", "", "", "", 0, 0, 0, 0},
		5:{0x81, 4, 0, 9, R_M, NONE, "SUB", "", "", "", 0, 0, 0, 0},
		6:{0x81, 4, 0, 9, R_M, NONE, "XOR", "", "", "", 0, 0, 0, 0},
		7:{0x81, 4, 0, 9, R_M, NONE, "CMP", "", "", "", 0, 0, 0, 0},
	},
	0x82:{
		
		0:{0x80, 1, 0, 1, R_M, NONE, "ADD", "", "", "", 0, 0, 0, 0},
		1:{0x80, 0, 0, 1, R_M, REG, "OR", "", "", "", 0, 0, 0, 0},
		2:{0x80, 0, 0, 1, R_M, REG, "ADC", "", "", "", 0, 0, 0, 0},
		3:{0x80, 0, 0, 1, R_M, REG, "SBB", "", "", "", 0, 0, 0, 0},
		4:{0x80, 0, 0, 1, R_M, REG, "AND", "", "", "", 0, 0, 0, 0},
		5:{0x80, 0, 0, 1, R_M, REG, "SUB", "", "", "", 0, 0, 0, 0},
		6:{0x80, 0, 0, 1, R_M, REG, "XOR", "", "", "", 0, 0, 0, 0},
		7:{0x80, 0, 0, 1, R_M, REG, "CMP", "", "", "", 0, 0, 0, 0},
	},
	0x83:{
		
		0:{0x80, 1, 0, 9, R_M, NONE, "ADD", "", "", "", 0, 0, 0, 0},
		1:{0x80, 1, 0, 9, R_M, NONE, "OR", "", "", "", 0, 0, 0, 0},
		2:{0x80, 1, 0, 9, R_M, NONE, "ADC", "", "", "", 0, 0, 0, 0},
		3:{0x80, 1, 0, 9, R_M, NONE, "SBB", "", "", "", 0, 0, 0, 0},
		4:{0x80, 1, 0, 9, R_M, NONE, "AND", "", "", "", 0, 0, 0, 0},
		5:{0x80, 1, 0, 9, R_M, NONE, "SUB", "", "", "", 0, 0, 0, 0},
		6:{0x80, 1, 0, 9, R_M, NONE, "XOR", "", "", "", 0, 0, 0, 0},
		7:{0x80, 1, 0, 9, R_M, NONE, "CMP", "", "", "", 0, 0, 0, 0},
	},
	0x8F:{
		
		0:{0x8F, 0, 0, 9, R_M, NONE, "POP", "", "", "", 0, 0, 0, 0},
	},
	0xC0:{
		
		0:{0x80, 1, 0, 1, R_M, NONE, "ROL", "", "", "", 0, 0, 0, 0},
		1:{0x80, 1, 0, 1, R_M, NONE, "ROR", "", "", "", 0, 0, 0, 0},
		2:{0x80, 1, 0, 1, R_M, NONE, "RCL", "", "", "", 0, 0, 0, 0},
		3:{0x80, 1, 0, 1, R_M, NONE, "RCR", "", "", "", 0, 0, 0, 0},
		4:{0x80, 1, 0, 1, R_M, NONE, "SHL/SAL", "", "", "", 0, 0, 0, 0},
		5:{0x80, 1, 0, 1, R_M, NONE, "SHR", "", "", "", 0, 0, 0, 0},
		6:{0x80, 1, 0, 1, R_M, NONE, "SHL/SAL", "", "", "", 0, 0, 0, 0},
		7:{0x80, 1, 0, 1, R_M, NONE, "SAR", "", "", "", 0, 0, 0, 0},
	},
	0xC1:{
		
		0:{0x80, 1, 0, 9, R_M, NONE, "ROL", "", "", "", 0, 0, 0, 0},
		1:{0x80, 1, 0, 9, R_M, NONE, "ROR", "", "", "", 0, 0, 0, 0},
		2:{0x80, 1, 0, 9, R_M, NONE, "RCL", "", "", "", 0, 0, 0, 0},
		3:{0x80, 1, 0, 9, R_M, NONE, "RCR", "", "", "", 0, 0, 0, 0},
		4:{0x80, 1, 0, 9, R_M, NONE, "SHL/SAL", "", "", "", 0, 0, 0, 0},
		5:{0x80, 1, 0, 9, R_M, NONE, "SHR", "", "", "", 0, 0, 0, 0},
		6:{0x80, 1, 0, 9, R_M, NONE, "SHL/SAL", "", "", "", 0, 0, 0, 0},
		7:{0x80, 1, 0, 9, R_M, NONE, "SAR", "", "", "", 0, 0, 0, 0},
	},
	0xC6:{
		
		0:{0x80, 1, 0, 1, R_M, NONE, "MOV", "", "", "", 0, 0, 0, 0},
	},
	0xC7:{
		
		0:{0x80, 4, 0, 9, R_M, NONE, "MOV", "", "", "", 0, 0, 0, 0},
	},
	0xD0:{
		
		0:{0x80, 0, 0, 1, R_M, NONE, "ROL", "", "", "", 0, 0, 0, 0},
		1:{0x80, 0, 0, 1, R_M, NONE, "ROR", "", "", "", 0, 0, 0, 0},
		2:{0x80, 0, 0, 1, R_M, NONE, "RCL", "", "", "", 0, 0, 0, 0},
		3:{0x80, 0, 0, 1, R_M, NONE, "RCR", "", "", "", 0, 0, 0, 0},
		4:{0x80, 0, 0, 1, R_M, NONE, "SHL/SAL", "", "", "", 0, 0, 0, 0},
		5:{0x80, 0, 0, 1, R_M, NONE, "SHR", "", "", "", 0, 0, 0, 0},
		6:{0x80, 0, 0, 1, R_M, NONE, "SHL/SAL", "", "", "", 0, 0, 0, 0},
		7:{0x80, 0, 0, 1, R_M, NONE, "SAR", "", "", "", 0, 0, 0, 0},
	},
	0xD1:{
		
		0:{0x80, 0, 0, 9, R_M, NONE, "ROL", "", "", "", 0, 0, 0, 0},
		1:{0x80, 0, 0, 9, R_M, NONE, "ROR", "", "", "", 0, 0, 0, 0},
		2:{0x80, 0, 0, 9, R_M, NONE, "RCL", "", "", "", 0, 0, 0, 0},
		3:{0x80, 0, 0, 9, R_M, NONE, "RCR", "", "", "", 0, 0, 0, 0},
		4:{0x80, 0, 0, 9, R_M, NONE, "SHL/SAL", "", "", "", 0, 0, 0, 0},
		5:{0x80, 0, 0, 9, R_M, NONE, "SHR", "", "", "", 0, 0, 0, 0},
		6:{0x80, 0, 0, 9, R_M, NONE, "SHL/SAL", "", "", "", 0, 0, 0, 0},
		7:{0x80, 0, 0, 9, R_M, NONE, "SAR", "", "", "", 0, 0, 0, 0},
	},
	0xD2:{
		
		0:{0x80, 0, 0, 1, R_M, CL, "ROL", "", "", "", 0, 0, 0, 0},
		1:{0x80, 0, 0, 1, R_M, CL, "ROR", "", "", "", 0, 0, 0, 0},
		2:{0x80, 0, 0, 1, R_M, CL, "RCL", "", "", "", 0, 0, 0, 0},
		3:{0x80, 0, 0, 1, R_M, CL, "RCR", "", "", "", 0, 0, 0, 0},
		4:{0x80, 0, 0, 1, R_M, CL, "SHL/SAL", "", "", "", 0, 0, 0, 0},
		5:{0x80, 0, 0, 1, R_M, CL, "SHR", "", "", "", 0, 0, 0, 0},
		6:{0x80, 0, 0, 1, R_M, CL, "SHL/SAL", "", "", "", 0, 0, 0, 0},
		7:{0x80, 0, 0, 1, R_M, CL, "SAR", "", "", "", 0, 0, 0, 0},
	},
	0xD3:{
		
		0:{0x80, 0, 0, 9, R_M, CL, "ROL", "", "", "", 0, 0, 0, 0},
		1:{0x80, 0, 0, 9, R_M, CL, "ROR", "", "", "", 0, 0, 0, 0},
		2:{0x80, 0, 0, 9, R_M, CL, "RCL", "", "", "", 0, 0, 0, 0},
		3:{0x80, 0, 0, 9, R_M, CL, "RCR", "", "", "", 0, 0, 0, 0},
		4:{0x80, 0, 0, 9, R_M, CL, "SHL/SAL", "", "", "", 0, 0, 0, 0},
		5:{0x80, 0, 0, 9, R_M, CL, "SHR", "", "", "", 0, 0, 0, 0},
		6:{0x80, 0, 0, 9, R_M, CL, "SHL/SAL", "", "", "", 0, 0, 0, 0},
		7:{0x80, 0, 0, 9, R_M, CL, "SAR", "", "", "", 0, 0, 0, 0},
	},
	0xF6:{
		
		0:{0x80, 1, 0, 1, R_M, NONE, "TEST", "", "", "", 0, 0, 0, 0},
		1:{0x80, 1, 0, 1, R_M, NONE, "TEST", "", "", "", 0, 0, 0, 0},
		2:{0x80, 0, 0, 1, R_M, NONE, "NOT", "", "", "", 0, 0, 0, 0},
		3:{0x80, 0, 0, 1, R_M, NONE, "NEG", "", "", "", 0, 0, 0, 0},
		4:{0x80, 0, 0, 1, R_M, NONE, "MUL", "", "", "", 0, 0, 0, 0},
		5:{0x80, 0, 0, 1, R_M, NONE, "IMUL", "", "", "", 0, 0, 0, 0},
		6:{0x80, 0, 0, 1, R_M, NONE, "DIV", "", "", "", 0, 0, 0, 0},
		7:{0x80, 0, 0, 1, R_M, NONE, "IDIV", "", "", "", 0, 0, 0, 0},
	},
	0xF7:{
		
		0:{0x80, 4, 0, 9, R_M, NONE, "TEST", "", "", "", 0, 0, 0, 0},
		1:{0x80, 4, 0, 9, R_M, NONE, "TEST", "", "", "", 0, 0, 0, 0},
		2:{0x80, 0, 0, 9, R_M, NONE, "NOT", "", "", "", 0, 0, 0, 0},
		3:{0x80, 0, 0, 9, R_M, NONE, "NEG", "", "", "", 0, 0, 0, 0},
		4:{0x80, 0, 0, 9, R_M, NONE, "MUL", "", "", "", 0, 0, 0, 0},
		5:{0x80, 0, 0, 9, R_M, NONE, "IMUL", "", "", "", 0, 0, 0, 0},
		6:{0x80, 0, 0, 9, R_M, NONE, "DIV", "", "", "", 0, 0, 0, 0},
		7:{0x80, 0, 0, 9, R_M, NONE, "IDIV", "", "", "", 0, 0, 0, 0},
	},
	0xFE:{
		
		0:{0x80, 0, 0, 1, R_M, NONE, "INC", "", "", "", 0, 0, 0, 0},
		1:{0x80, 0, 0, 1, R_M, NONE, "DEC", "", "", "", 0, 0, 0, 0},
	},
	0xFF:{
		
		0:{0x80, 0, 0, 1, R_M, NONE, "INC", "", "", "", 0, 0, 0, 0},
		1:{0x80, 0, 0, 1, R_M, NONE, "DEC", "", "", "", 0, 0, 0, 0},
		2:{0x80, 0, 0, 1, R_M, NONE, "CALL", "", "", "", 0, 0, 0, 0},
		3:{0x80, 0, 0, 1, R_M, NONE, "CALL", "", "", "", 0, 0, 0, 0},
		4:{0x80, 0, 0, 1, R_M, NONE, "JMP", "", "", "", 0, 0, 0, 0},
		5:{0x80, 0, 0, 1, R_M, NONE, "JMP", "", "", "", 0, 0, 0, 0},
		6:{0x80, 0, 0, 1, R_M, NONE, "PUSH", "", "", "", 0, 0, 0, 0},
	},
}