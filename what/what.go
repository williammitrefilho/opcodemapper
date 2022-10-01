package what

type Error struct{
	
	Message string
}
type InstructionByteStream struct{
	
	bytes []byte
	index int
}

func (istr *InstructionByteStream) ReadBytes(n int) ([]byte, *Error){
	
	if (istr.index + n) > len(istr.bytes){
		
		return nil, NewError("Out of bounds")
	}
	rBytes := istr.bytes[istr.index:(istr.index + n)]
	return rBytes, nil
}

func NewError(msg string) *Error{
	
	err := new(Error)
	err.Message = msg
	return err
}

func NewInstructionByteStream(bytes []byte) *InstructionByteStream{
	
	istr := new(InstructionByteStream)
	istr.index = 0
	istr.bytes = bytes
	
	return istr
}