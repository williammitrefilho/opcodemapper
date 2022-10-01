package disassembler

import(
	
	"testing"
)
func TestToBytes(t *testing.T){
	
	bytes := toBytes("abcd")
	tBytes := []byte{
		0xab, 0xcd,
	}
	if bytes[0] != tBytes[0] || bytes[1] != tBytes[1]{
		
		t.Fatalf("byteFail")
	}
}