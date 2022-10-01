package commander

import(
//	"fmt"
)

type Commander struct{
	
	Code string
}

type CommanderResponse struct{

	resultClass string
	values map[string]string
}

func (c *Commander) Run() *CommanderResponse{
	
	return new(CommanderResponse)
}

func New() *Commander{
	
	return new(Commander)
}