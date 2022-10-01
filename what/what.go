package what

const(
	
	TYPE_MODRM int = 1
	TYPE_OPCODE = 2
	TYPE_EXT_PREFIX = 3
	
	OP_TYPE_IMM = 1
	OP_TYPE_REG = 2
	OP_TYPE_MEM = 3
)

type InstructionByte struct{
	
	Value byte
	Type int
	Instructions []*Instruction
}

type Instruction struct{
	
	Bytes []*InstructionByte
	OperandSpec1 *OperandSpec
	OperandSpec2 *OperandSpec
}

type OperandSpec struct{
	
	Type int
	Size int
}

type ByteValue struct{
	
	InstructionName string
	NextByteValue *OpcodeMapper
}

type InstructionByteMap map[uint8]InstructionByte

type InstructionByteMapper struct{
	
	InsMap InstructionByteMap
}

func New() *InstructionByteMapper{
	
	Mb := new(InstructionByteMapper)
	Mb.OpMap = make(InstructionByteMap)
	return Mb
}

func NewInstruction() *Instruction{
	
	return new(Instruction)
}

func (m *InstructionByteMapper)AddEntry(instrByte InstructionByte, instruction Instruction){
	
	presentInstr := Lookup(instrByte)
	if(presentInstr == nil){
		
		m.InsMap[instrByte.Value] = instrByte
	}
}

func (m *InstructionByteMapper)Lookup(instrByte InstructionByte) InstructionByte{
	
	return m.InsMap[instrByte.Value]
}