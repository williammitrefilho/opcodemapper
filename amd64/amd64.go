package amd64

import(
	"fmt"
)

var GPRNames = []string{
	
	"RAX","RCX","RDX","RBX","RSP","RBP","RSI","RDI",
}
type modRMStatus struct{
	
	mod byte
	reg byte
	r_m byte
	r_m_str string
	reg_str string
}

type sIBByte struct{
	
	scale byte
	index byte
	base byte
}

type Processor struct{
	
	fedBytes [256]byte
	nFedBytes int
	
	rEXPrefix byte
	mnemonicRoot string
	finished bool
	
	opcode byte
	hasOpcode bool
	extendedOpcode bool
	
	mod_r_m_Status *modRMStatus
	sibByte *sIBByte
	hasModRM bool
	
	srcOper string
	dstOper string
	
	immediateVal int32
}

func NewProcessor() *Processor{
	
	p := new(Processor)
	p.mod_r_m_Status = new(modRMStatus)
	return p
}

func (p *Processor) feedByte(b byte) int{
	
	p.fedBytes[p.nFedBytes] = b
	p.nFedBytes++
	if !p.isPrefix(b){// check whether b is a prefix or an opcode
	
		return p.setOpcode(b)
	}
	return 0
}

func (p *Processor) ReadInstruction(instructionBytes []byte){
	
	intBytes := 0
	index := 0

	for !p.finished{

		intBytes = p.feedByte(instructionBytes[index])
		index++
		if intBytes == 0{
			
			break
		}
	}
	
	if intBytes > 0 && index+intBytes <= len(instructionBytes){
		
		var imm int32 = 0
		o := 0
		for i := index; i < index+intBytes; i++{
			
			imm |= (int32)(instructionBytes[i]) << (8*o)
			o++
		}
		p.immediateVal = imm
	}
}

func (p *Processor) setOpcode(b byte) int{
	
	if p.hasOpcode{
		
		highNibble := p.opcode>>4; lowNibble := p.opcode&0xF
		if highNibble < 4{
			
			if lowNibble < 6 || (lowNibble > 0x7 && lowNibble < 0xE){
				
				return p.setModRM_SIB(b)
			}
		} else if highNibble == 0xF{
			
			if lowNibble > 0xD{
				
				return p.setModRM_SIB(b)
			}
		}
	
	} else {
		
		highNibble := b>>4; lowNibble := b&0xF
		p.opcode = b
		p.hasOpcode = true
		if highNibble < 4{
			
			if lowNibble < 6 || (lowNibble > 0x7 && lowNibble < 0xE){
				
				return 1
			}
		} else if highNibble == 5{// POP or PUSH, no more bytes needed
			
			p.finished = true
			return 0
		} else if highNibble == 6{// POP or PUSH with more bytes needed
			
			if lowNibble > 7 && lowNibble < 0xC{
				
				p.finished = true
				if lowNibble % 2 == 0{
					
					if lowNibble == 8{
						
						return 4
					} else if lowNibble == 0xA{
						
						return 1
					}
				} else {//modRM byte will be needed, so just ask for one more
					
					return 1
				}
			}
		} else if highNibble == 7{// Jump instructions that need another single byte
			
			p.finished = true
			return 1
		} else if highNibble == 8{
			
			if lowNibble < 4{
				
				return 1
			}
		} else if highNibble == 9{
			
			if lowNibble < 8{
				
				p.finished = true
				return 0
			}
		} else if highNibble == 0xE{
			
			if lowNibble == 8{
				
				p.finished = true
				return 4
			}
		} else if highNibble == 0xF{
			
			if lowNibble > 0xD{// opcodes which are extended by the ModRM byte
				
				return 1
			}
		}
	}
	return 0
}	

func (p *Processor) setModRM_SIB(b byte) int{
	
	if p.hasModRM{
		
		return p.setSIBByte(b)
	}
	p.hasModRM = true
	p.mod_r_m_Status.mod = b >> 6
	p.mod_r_m_Status.reg = (b&0x38) >> 3
	p.mod_r_m_Status.r_m = (b&0x7)
	
	p.mod_r_m_Status.reg_str = GPRNames[p.mod_r_m_Status.reg]
	
	if p.opcode > 0xFD{// opcode is extended by the ModRM.reg field
		
		if p.mod_r_m_Status.reg == 4{
			
			if p.mod_r_m_Status.r_m == 5{
				
				if p.mod_r_m_Status.mod == 0{
					
					
				}
			}
			p.finished = true
			return 0
		}
	}
	if p.mod_r_m_Status.mod == 3 {
		
		p.mod_r_m_Status.r_m_str = GPRNames[p.mod_r_m_Status.r_m]
	} else {
		
		if p.mod_r_m_Status.r_m != 4{
			
			p.mod_r_m_Status.r_m_str = fmt.Sprintf("[%v]", GPRNames[p.mod_r_m_Status.r_m])
		} else {
			// the only case when we we will need another byte, which is the SIB byte
			return 1
		}
	}
	p.finished = true
	return 0
}

func (p *Processor) setSIBByte(b byte) int{
	
	p.sibByte.scale = b >> 6
	p.sibByte.index = (b&0x38) >> 3
	p.sibByte.base = (b&0x7)
	
	scale := 1 << p.sibByte.scale
	var scaled_index, base string
	moreBytes := 0
	if p.sibByte.index == 4{
		
		scaled_index = ""
	} else{
		
		scaled_index = fmt.Sprintf("%v*[%v]", scale, GPRNames[p.sibByte.index])
	}
	if p.sibByte.base == 5{
		
		if p.mod_r_m_Status.mod == 0{
			
			base = "disp32"
			moreBytes = 4
		} else if p.mod_r_m_Status.mod == 1{
			
			base = "[RBP] + disp8"
			moreBytes = 1
		} else {
			
			base = "[RBP] + disp32"
			moreBytes = 4
		}
	}
	sibOper := fmt.Sprintf("%v + %v", scaled_index, base)
	p.mod_r_m_Status.r_m_str = sibOper
	
	p.finished = true
	
	return moreBytes
}

func (p *Processor) isPrefix(b byte) bool{
	
	if p.hasOpcode{
		
		return false
	}
	
	if (b&0xF0) == 0x40{//REX prefix
		
		p.rEXPrefix = b
		return true
	}
	return b == 0x66 || b == 0x67 || b == 0x2E || b == 0x3E || b == 0x26 || b == 0x64 || b == 0x65 || b == 0x36 || b == 0xF0 || b == 0xF3 || b == 0xF2
}