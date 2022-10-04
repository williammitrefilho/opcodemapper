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
	finished bool
	
	opcode byte
	hasOpcode bool
	extendedOpcode bool
	opcode_str string
	
	mod_r_m_Status *modRMStatus
	sibByte *sIBByte
	hasModRM bool
	
	operandSize uint8
	
	srcOper *string
	dstOper *string
	
	immediateBytes uint8
	displacementBytes uint8
	
	immediateVal int64
	immediateVal_str string
	displacementVal int64
}

func NewProcessor() *Processor{
	
	p := new(Processor)
	p.mod_r_m_Status = new(modRMStatus)
	p.sibByte = new(sIBByte)
	return p
}

func (p *Processor) feedByte(b byte) int{
	
	p.fedBytes[p.nFedBytes] = b
	p.nFedBytes++
	if !p.isPrefix(b){// check whether b is a prefix or an opcode
	
		return p.setOpcode(b)
	}
	return 1
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
	
	boundary := index+(int)(p.displacementBytes)
	if p.displacementBytes > 0 && boundary <= len(instructionBytes){
		
		var imm int64 = 0
		o := 0
		for i := index; i < boundary; i++{
			
			imm |= (int64)(instructionBytes[i]) << (8*o)
			o++
		}
		p.displacementVal = imm
	}
	boundary += (int)(p.immediateBytes)
	index += (int)(p.immediateBytes)
	if p.immediateBytes > 0 && boundary <= len(instructionBytes){
		
		var imm int64 = 0
		o := 0
		for i := index; i < boundary; i++{
			
			imm |= (int64)(instructionBytes[i]) << (8*o)
			o++
		}
		p.immediateVal = imm
	}
	if p.displacementVal > 0{
		
		p.mod_r_m_Status.r_m_str = fmt.Sprintf("%v+0x%x", p.mod_r_m_Status.r_m_str, p.displacementVal)
	}
	p.immediateVal_str = fmt.Sprintf("0x%x", p.immediateVal)
}

func (p *Processor) setOpcode(b byte) int{
	
	if p.hasOpcode{
		
		p.setModRM_SIB(b)
/*		
		highNibble := p.opcode>>4; lowNibble := p.opcode&0xF
		if highNibble < 4{
			
			if lowNibble < 6 || (lowNibble > 0x7 && lowNibble < 0xE){
				
				return p.setModRM_SIB(b)
			}
		} else if highNibble == 0x8{
			
			if lowNibble < 0x4 || (lowNibble > 7 && lowNibble < 0xC){
				
				return p.setModRM_SIB(b)
			}
		} else if highNibble == 0xF{
			
			if lowNibble > 0xD{
				
				return p.setModRM_SIB(b)
			}
		}
*/		
	
	} else {
		
		highNibble := b>>4; lowNibble := b&0xF
		p.opcode = b
		p.hasOpcode = true
		if highNibble == 3{
			
			if lowNibble > 7 && lowNibble < 0xE{
				
				p.opcode_str = "CMP"
				if lowNibble < 0xC{
					
					if lowNibble%2 == 0{
						
//						p.immediateBytes = 1
					} else {
						
						if p.rEXPrefix & 0x08 != 0{
					
//							p.immediateBytes = 8
						} else{
							
//							p.immediateBytes = 4
						}
					}
					if lowNibble < 0xA{
						
						p.srcOper = &p.mod_r_m_Status.reg_str
						p.dstOper = &p.mod_r_m_Status.r_m_str
					} else {
						
						p.srcOper = &p.mod_r_m_Status.reg_str
						p.dstOper = &p.mod_r_m_Status.r_m_str
					}
					return 1
				}
			}
		} else if highNibble == 5{// POP or PUSH, no more bytes needed
			
			p.dstOper = &GPRNames[lowNibble]
			if lowNibble < 8{
				
				p.opcode_str = "PUSH"
			} else {
				
				p.opcode_str = "POP"
			}
			p.finished = true
			return 0
		} else if highNibble == 6{// POP or PUSH with more bytes needed
			
			if lowNibble > 7 && lowNibble < 0xC{
				
				p.finished = true
				if lowNibble % 2 == 0{
					
					if lowNibble == 8{
						
						p.immediateBytes = 4
						return 4
					} else if lowNibble == 0xA{
						
						p.immediateBytes = 1
						return 1
					}
				} else {//modRM byte will be needed, so just ask for one more
					
					return 1
				}
			}
		} else if highNibble == 7{// Jump instructions that need another single byte
			
			p.finished = true
			p.immediateBytes = 1
			if lowNibble == 5{
				
				p.opcode_str = "JNZ"
			}
			return 1
		} else if highNibble == 8{
			
			if lowNibble < 4{// extended opcodes in which the instruction is specified my the ModRM byte
				
				return 1
			} else if lowNibble > 0x7 && lowNibble < 0xC{// MOV instructions that require the ModRM byte
				
				p.opcode_str = "MOV"
				if lowNibble%2 == 0{
					
					p.operandSize = 1
				} else{
					
					if p.rEXPrefix & 0x08 != 0{
						
						p.operandSize = 8
					} else {
						
						p.operandSize = 4
					}
				}
				if lowNibble < 0xA{
					
					p.dstOper = &p.mod_r_m_Status.r_m_str
					p.srcOper = &p.mod_r_m_Status.reg_str
				} else{
					
					p.dstOper = &p.mod_r_m_Status.reg_str
					p.srcOper = &p.mod_r_m_Status.r_m_str
				}
				return 1
			}
		} else if highNibble == 9{
			
			if lowNibble < 8{
				
				p.finished = true
				return 0
			}
		} else if highNibble == 0xB{
			
			p.opcode_str = "MOV"
			if lowNibble > 0x7{
				
				p.srcOper = &p.immediateVal_str
				p.dstOper = &GPRNames[lowNibble-8]
				if p.rEXPrefix & 0x08 != 0{
					
					p.immediateBytes = 8
				} else{
					
					p.immediateBytes = 4
				}
			}
		} else if highNibble == 0xE{
			
			if lowNibble == 8{
				
				p.opcode_str = "CALL"
				p.finished = true
				p.immediateBytes = 4
				p.dstOper = &p.immediateVal_str
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
	moreBytes := 0
	if p.opcode > 0xFD || (p.opcode >= 0x80 && p.opcode <= 0x83){// opcode is extended by the ModRM.reg field
		
		p.mod_r_m_Status.reg_str = ""
		if p.opcode > 0xFD{
			
			if p.mod_r_m_Status.reg == 4{
				
				if p.mod_r_m_Status.r_m == 5{
					
					if p.mod_r_m_Status.mod == 0{
						
						
					}
				}
				p.finished = true
			}
		}else {
		
			if p.opcode == 0x83{
				
				if p.mod_r_m_Status.reg == 5{
					
					p.opcode_str = "SUB"
					p.dstOper = &p.mod_r_m_Status.r_m_str
					p.srcOper = &p.immediateVal_str
					p.immediateBytes = 1
				} else if p.mod_r_m_Status.reg == 4{
					
					p.opcode_str = "AND"
					p.dstOper = &p.mod_r_m_Status.r_m_str
					p.srcOper = &p.immediateVal_str
					p.immediateBytes = 1
				}
			}
		}
	}
	if p.mod_r_m_Status.mod == 3 {
		
		p.mod_r_m_Status.r_m_str = GPRNames[p.mod_r_m_Status.r_m]
	} else if p.mod_r_m_Status.r_m_str == ""{
		
		if p.mod_r_m_Status.r_m == 5{
			
			if p.mod_r_m_Status.mod == 0{
				
				p.mod_r_m_Status.r_m_str = ""
				p.displacementBytes = 4
			} else if p.mod_r_m_Status.mod == 1{
				
				p.mod_r_m_Status.r_m_str = "[RBP]"
				p.displacementBytes = 1
			}
		} else if p.mod_r_m_Status.r_m != 4{
			
			p.mod_r_m_Status.r_m_str = fmt.Sprintf("[%v]", GPRNames[p.mod_r_m_Status.r_m])
		} else {
			// the only case in which we will need another byte, that being the SIB byte
			moreBytes++
			return moreBytes
		}
	}
	p.finished = true
	return moreBytes
}

func (p *Processor) setSIBByte(b byte) int{
	
	p.sibByte.scale = b >> 6
	p.sibByte.index = (b&0x38) >> 3
	p.sibByte.base = (b&0x7)
	
	scale := 1 << p.sibByte.scale
	var scaled_index, base string
	var moreBytes byte = 0
	if p.sibByte.index == 4{
		
		scaled_index = ""
	} else{
		
		scaled_index = fmt.Sprintf("+%v*[%v]", scale, GPRNames[p.sibByte.index])
	}
	if p.sibByte.base == 5{
		
		if p.mod_r_m_Status.mod == 0{
			
			base = ""
			moreBytes = 4
		} else if p.mod_r_m_Status.mod == 1{
			
			base = "[RBP]"
			moreBytes = 1
		} else {
			
			base = "[RBP]"
			moreBytes = 4
		}
	} else {
		
		if p.sibByte.base == 4{
		
			base = "[RSP]"
		}
		
		if p.mod_r_m_Status.mod == 0{
			
			moreBytes = 0
		} else if p.mod_r_m_Status.mod == 1{
			
			moreBytes = 1
		} else{
			
			moreBytes = 4
		}
	}
	sibOper := fmt.Sprintf("%v %v", base, scaled_index)
	p.mod_r_m_Status.r_m_str = sibOper
	
	p.finished = true
	p.displacementBytes = moreBytes
	return (int)(moreBytes)
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