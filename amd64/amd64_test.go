package amd64

import(

	"testing"
)

func TestProcessor(t *testing.T){
	
	p := NewProcessor()
	
	instructionBytes := []byte{
		
		0x78, 0x12, 0x40, 0x00,
	}
	
	p.ReadInstruction(instructionBytes)
	if p.immediateVal != 0x12{
		
		t.Fatalf("imm32 = %x", p.immediateVal)
	}
	if p.nFedBytes != 1{
		
		t.Fatalf("%v fed bytes", p.nFedBytes)
	}
}

func TestFeedByte(t *testing.T){
	
	p:= NewProcessor()
	
	p.feedByte(0xFF)
	r := p.feedByte(0x25)
	if !p.finished{
		
		t.Fatalf("p is not finished. opcode: %v modrm: %v, hasModRM:%v hasOpcode:%v", p.opcode, p.mod_r_m_Status, p.hasModRM, p.hasOpcode)
	}
	if r != 0{
		
		t.Fatalf("r is %v", r)
	}
}

func TestSetModRM_SIB(t *testing.T){
	
	p := NewProcessor()
	p.opcode = 0xFF
	p.hasOpcode = true
	p.setOpcode(0x25)
	if(p.mod_r_m_Status.reg != 4){
		
		t.Fatalf("ModRM.reg = %v", p.mod_r_m_Status)
	}
}