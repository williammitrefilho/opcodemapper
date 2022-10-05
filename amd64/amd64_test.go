package amd64_1

import(
	"testing"
)

func TestNeedsModRM(t *testing.T){
	
	instructionBytes := []byte{
//Let's see if we can correctly disassemble 64-bit chrome.exe for x86-64, one instruction at a time
		
//		0x48, 0x89, 0x5C, 0x24, 0x20, 0x55,
//		0xE8, 0x0B, 0x00, 0x00, 0x00,
//		0x48, 0x89, 0x5C, 0x24, 0x20,
//		0x55,
//		0x48, 0x8B, 0xEC,
//		0x48, 0x83, 0xEC, 0x20,
//		0x48, 0x8B, 0x05, 0xB0, 0x9D, 0x10, 0x00,
//		0x48, 0xBB, 0x32, 0xA2, 0xDF, 0x2D, 0x99, 0x2B, 0x00, 0x00,
//		0x48, 0x3B, 0xC3
//		0x75, 0x74,
//		0x48, 0x8B, 0x5C, 0x24, 0x48,
		0x48, 0xF7, 0xD0,
	}
	if !opcodeMap[instructionBytes[1]].needsModRM(){
		
		t.Fatalf("said that byte 0x%x didn't need ModRM", instructionBytes[1])
	}
}

func TestSetModRM(t *testing.T){
	
	instructionBytes := []byte{
//Let's see if we can correctly disassemble 64-bit chrome.exe for x86-64, one instruction at a time
		
//		0x48, 0x89, 0x5C, 0x24, 0x20, 0x55,
//		0xE8, 0x0B, 0x00, 0x00, 0x00,
//		0x48, 0x89, 0x5C, 0x24, 0x20,
//		0x55,
//		0x48, 0x8B, 0xEC,
//		0x48, 0x83, 0xEC, 0x20,
//		0x48, 0x8B, 0x05, 0xB0, 0x9D, 0x10, 0x00,
//		0x48, 0xBB, 0x32, 0xA2, 0xDF, 0x2D, 0x99, 0x2B, 0x00, 0x00,
//		0x48, 0x3B, 0xC3
//		0x75, 0x74,
		0x48, 0x8B, 0x5C, 0x24, 0x48,
//		0x48, 0xF7, 0xD0,
	}
	
	var opcode opcode = *opcodeMap[instructionBytes[1]]
	if !opcode.needsModRM(){
		
		t.Fatalf("said that byte 0x%x didn't need ModRM", instructionBytes[1])
	}
	opcode.setModRM(instructionBytes[2])
	
	if opcode.mnemonic != "MOV"{
		
		t.Fatalf("opcode mnemonic: %v", opcode.mnemonic)
	}
	if opcode.r_m != SIB{
		
		t.Fatalf("opcode r_m:%v", opcode.r_m)
	}
	opcode.setSIBByte(instructionBytes[3])
	if opcode.scale != 1{
		
		t.Fatalf("opcode SIB.scale is %v", opcode.scale)
	}
	if opcode.r_m != RSP{
		
		t.Fatalf("opcode.r_m is %v", opcode.r_m)
	}
	if opcode.displacementBytes != 1{
		
		t.Fatalf("opcode displacementBytes is %v", opcode.displacementBytes)
	}
}

func TestFeedByte(t *testing.T){
	instructionBytes := []byte{
//Let's see if we can correctly disassemble 64-bit chrome.exe for x86-64, one instruction at a time
		
//		0x48, 0x89, 0x5C, 0x24, 0x20, 0x55,
//		0xE8, 0x0B, 0x00, 0x00, 0x00,
//		0x48, 0x89, 0x5C, 0x24, 0x20,
//		0x55,
//		0x48, 0x8B, 0xEC,
//		0x48, 0x83, 0xEC, 0x20,
//		0x48, 0x8B, 0x05, 0xB0, 0x9D, 0x10, 0x00,
//		0x48, 0xBB, 0x32, 0xA2, 0xDF, 0x2D, 0x99, 0x2B, 0x00, 0x00,
//		0x48, 0x3B, 0xC3
//		0x75, 0x74,
		0x48, 0x8B, 0x5C, 0x24, 0x48,
//		0x48, 0xF7, 0xD0,
	}
	
	instr := new(instruction)
	r := instr.feedByte(instructionBytes[0])
	if r != 1{
		
		t.Fatalf("r is %v", r)
	}
	r = instr.feedByte(instructionBytes[1])
	if r != 1{
		
		t.Fatalf("r is %v", r)
	}
	if instr.opcode == nil{
		
		t.Fatalf("opcode is nil")
	}
	r = instr.feedByte(instructionBytes[2])
	if r != 1{
		
		t.Fatalf("r is %v", r)
	}
	if instr.opcode.needsModRM(){
		
		t.Fatalf("instr still asking for ModRM")
	}
	if instr.opcode.r_m != SIB{
		
		t.Fatalf("instr's r_m field is %v", instr.opcode.r_m)
	}
	r = instr.feedByte(instructionBytes[3])
	if r != 0{
		
		t.Fatalf("r is %v", r)
	}
	if !instr.finished{
		
		t.Fatalf("instr is not finished")
	}
	if instr.opcode.r_m == SIB{
		
		t.Fatalf("instr's r_m field is still SIB")
	}
	if instr.opcode.displacementBytes != 1{
		
		t.Fatalf("instr's displacementBytes is %v", instr.opcode.displacementBytes)
	}
	if instr.opcode.immediateBytes != 0{
		
		t.Fatalf("instr's immediateBytes is %v", instr.opcode.immediateBytes)
	}
}

func TestInstructionFromBytes(t *testing.T){
	
	instructionBytes := []byte{
//Let's see if we can correctly disassemble 64-bit chrome.exe for x86-64, one instruction at a time
		
//		0x48, 0x89, 0x5C, 0x24, 0x20, 0x55,
//		0xE8, 0x0B, 0x00, 0x00, 0x00,
//		0x48, 0x89, 0x5C, 0x24, 0x20,
//		0x55,
//		0x48, 0x8B, 0xEC,
//		0x48, 0x83, 0xEC, 0x20,
//		0x48, 0x8B, 0x05, 0xB0, 0x9D, 0x10, 0x00,
		0x48, 0xBB, 0x32, 0xA2, 0xDF, 0x2D, 0x99, 0x2B, 0x00, 0x00,
//		0x48, 0x3B, 0xC3
//		0x75, 0x74,
//		0x48, 0x8B, 0x5C, 0x24, 0x48,
//		0x48, 0xF7, 0xD0,
	}
	instr := InstructionFromBytes(instructionBytes)
	if instr == nil{
		
		t.Fatalf("instruction is nil")
	}
	if instr.opcode.mnemonic != "MOV"{
		
		t.Fatalf("mnemonic is %v", instr.opcode.mnemonic)
	}
	if instr.opcode.immediateBytes != 8{
		
		t.Fatalf("immediateBytes is %v", instr.opcode.immediateBytes)
	}
	if instr.immediate != 0x2b992ddfa232{
		
		t.Fatalf("immediate is %x", instr.immediate)
	}
}