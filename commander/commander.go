package commander

import(
//	"fmt"
//	"modgo.com/what"
	"log"
	"regexp"
	"modgo.com/disassembler"
)

type Commander struct{
	
	Code string
}

type CommanderError struct{
	
	Message string
}


type CommanderResponse struct{

	ResultClass string
	Value any
}

func Error(Message string) *CommanderError{
	
	cmdErr := new(CommanderError)
	cmdErr.Message = Message
	return cmdErr
}

func (c *Commander) EvaluateCode() *CommanderError{
	
	anyIntruder, err := regexp.MatchString("[^0-9A-Fa-f]", c.Code)
	if err != nil{
		
		log.Fatal(err)
	}
	if anyIntruder{
		
		return Error("There is an intruder")
	}
	if (len(c.Code) % 2) > 0{
		
		return Error("Could you please 8-bit align?")
	}
	disassembler.Disassemble(c.Code)
	return nil
}

func (c *Commander) Run() *CommanderResponse{
	
	if c.Code != ""{
		
		cmdErr := c.EvaluateCode()
		cr := new(CommanderResponse)
		if(cmdErr != nil){
			
			cr.ResultClass = "error"
			cr.Value = cmdErr
			return cr
		}
		cr.ResultClass = "OK"
		return cr
	}
	
	cr := new(CommanderResponse)
	cr.ResultClass = "error"
	return cr
}

func New() *Commander{
	
	return new(Commander)
}