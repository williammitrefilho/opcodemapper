package amd64

import(

	"testing"
)

func TestProcessor(t *testing.T){
	
	p := NewProcessor()
	
	instructionBytes := []byte{
		
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
		0x48, 0x8B, 0x5C, 0x24,
	}
	
	p.ReadInstruction(instructionBytes)
	if p.immediateVal != 0x0{
		
		t.Fatalf("imm32 = %x", p.immediateVal)
	}
	if p.nFedBytes != 3{
		
		t.Fatalf("%v fed bytes (%x)", p.nFedBytes, p.rEXPrefix)
	}
	if p.opcode_str != "MOV"{
		
		t.Fatalf("opcode_str is %v", p.opcode_str)
	}
	if p.displacementVal != 0x0{
		
		t.Fatalf("displacementVal is %x", p.displacementVal)
	}
	if p.dstOper != nil{
		
		if *p.dstOper != "RBX"{
			t.Fatalf("dstOper = %v", *p.dstOper)
		}
	}
	if p.srcOper != nil{
		
		if *p.srcOper != "[RBP]+0x18"{
			t.Fatalf("srcOper = %v, sib:%v", *p.srcOper, p.sibByte)
		}
	}
}

func TestFeedByte(t *testing.T){
	
	p:= NewProcessor()
	
	pr := p.feedByte(0x48)
	p.feedByte(0x83)
	r := p.feedByte(0xEC)

	if !p.finished{
		
		t.Fatalf("p is not finished. opcode: %v modrm: %v, hasModRM:%v hasOpcode:%v, pr:%v r:%v", p.opcode, p.mod_r_m_Status, p.hasModRM, p.hasOpcode, pr, r)
	}
	if p.immediateBytes != 1{
		
		t.Fatalf("immediateBytes is %v. ModRM.reg:%v", r, p.mod_r_m_Status.reg)
	}
	if p.opcode_str != "SUB"{
		
		t.Fatalf("opcode_str is %v", p.opcode_str)
	}
	if *p.dstOper != "RSP"{
		
		t.Fatalf("r/m is set to %v", p.mod_r_m_Status.r_m_str)
	}
	if *p.srcOper != ""{
		
		t.Fatalf("p.srcOper is set to %v", *p.srcOper)
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