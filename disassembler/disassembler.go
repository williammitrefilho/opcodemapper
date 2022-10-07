package disassembler

import (
//	"modgo.com/what"
	"strings"
//	"fmt"
	"modgo.com/amd64"
//	"errors"
)

type Instruction amd64.Instruction

func toBytes(String string)[]byte{
	
	charMap := map[byte]byte{
		0x30:0x00, 0x31:0x01, 0x32:0x02, 0x33:0x03, 0x34:0x04, 0x35:0x05, 0x36:0x06, 0x37:0x07, 0x38:0x08, 0x39:0x09,
		0x41:0x0A, 0x42:0x0B, 0x43:0x0C, 0x44:0x0D, 0x45:0x0E, 0x46:0x0F,
		0x61:0x0A, 0x62:0x0B, 0x63:0x0C, 0x64:0x0D, 0x65:0x0E, 0x66:0x0F,
	}
	var bytes [128]byte
	nBytes := 0
	strReader := strings.NewReader(String)
	
	for strReader.Len() > 0{
		
		bByte, _ := strReader.ReadByte()
		bByte2, _ := strReader.ReadByte()
		
		opVal := charMap[bByte] << 4 | charMap[bByte2]
		bytes[nBytes] = opVal
		nBytes++
	}
	return bytes[:nBytes]
}

func Disassemble(code string) ([]*amd64.Instruction, error){
	
	p := amd64.New()
	p.LoadCode(toBytes(code))
	instrs, err := p.Run()
	if err != nil{
		
		return nil, err
	}
	return instrs, nil
	
//	fmt.Printf("%v\n", istr)
}