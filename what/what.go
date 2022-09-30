package what

type ByteValue struct{
	
	InstructionName string
	NextByteValue *OpcodeMapper
}

type OpcodeMap map[uint8]ByteValue

type OpcodeMapper struct{
	
	OpMap OpcodeMap
}

func New() *OpcodeMapper{
	
	Mb := new(OpcodeMapper)
	Mb.OpMap = make(OpcodeMap)
	return Mb
}

func (m *OpcodeMapper)AddEntry(Opcode uint8, Value ByteValue){
	
	m.OpMap[Opcode] = Value
}

func (m *OpcodeMapper)Lookup(Bytes[]byte) (ByteValue){
	
	return m.OpMap[Bytes[0]]
}