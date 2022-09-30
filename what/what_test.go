package what

import(
	"testing"
)

func TestAddEntry(t *testing.T){
	
	Mapper := New()
	Mapper.AddEntry(0xFF, ByteValue{"final", nil})
	if Mapper.OpMap[0xFF].InstructionName != "final"{
			
		t.Fatalf("InstructionName deu %v\n", Mapper.OpMap[0xFF].InstructionName)
	}
}

func TestLookup(t *testing.T){
	
	Mapper := New()
	Mapper.AddEntry(0xFF, ByteValue{"final", nil})
	Result := Mapper.Lookup([]byte{0xFF})
	if(Result.InstructionName != "final"){
			
		t.Fatalf("InstructionName deu %v\n", Result.InstructionName)
	}
}