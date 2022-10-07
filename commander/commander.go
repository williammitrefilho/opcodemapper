package commander

import(
//	"fmt"
//	"modgo.com/what"
	"log"
	"regexp"
	"modgo.com/disassembler"
	"modgo.com/amd64"
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

func (c *Commander) EvaluateCode() ([]*amd64.Instruction, *CommanderError){
	
	anyIntruder, err := regexp.MatchString("[^0-9A-Fa-f]", c.Code)
	if err != nil{
		
		log.Fatal(err)
	}
	if anyIntruder{
		
		return nil, Error("There is an intruder")
	}
	if (len(c.Code) % 2) > 0{
		
		return nil, Error("Could you please 8-bit align?")
	}
	instrs, err := disassembler.Disassemble(c.Code)
	if err != nil{
		
		return nil, Error(err.Error())
	}
	return instrs, nil
}

func (c *Commander) Run() *CommanderResponse{
	
	if c.Code != ""{
		
		instrs, cmdErr := c.EvaluateCode()
		cr := new(CommanderResponse)
		if(cmdErr != nil){
			
			cr.ResultClass = "error"
			cr.Value = cmdErr
			return cr
		}
		cr.ResultClass = "OK"
		cr.Value = instrs
		return cr
	}
	
	cr := new(CommanderResponse)
	cr.ResultClass = "error"
	return cr
}

func New() *Commander{
	
	return new(Commander)
}