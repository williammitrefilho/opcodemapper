package whatserver

import(
	"net/http"
	"log"
	"encoding/json"
	"os"
	"fmt"
	"modgo.com/commander"
)

func Handler(Writer http.ResponseWriter, Request *http.Request){
	
	if(Request.Method != "POST"){//Whatever be the request method, besides POST, serve the HTML form
		
		responseBytes, err := os.ReadFile("C:/opcodemapper/whatserver/page_to_serve.html")
		if err != nil{
			
			log.Fatal(err)
		} else {
			
			Writer.Write(responseBytes)
		}
	} else {
			
		fmt.Printf("Content length:%v\n", Request.ContentLength)
		Commander := commander.New()
		
		if Request.ContentLength > 7{     //    {"":""} 
			
			buffer := make([]byte, Request.ContentLength)
				
			nReadBytes, err := Request.Body.Read(buffer)
			if err != nil{
			
				log.Fatal(err)
			}
			fmt.Printf("%v read bytes:", nReadBytes)
			os.Stdout.Write(buffer)
			fmt.Println()
			
			err = json.Unmarshal(buffer, Commander);
			if(err != nil){
			
				log.Fatal(err)
			}
		}
		result := Commander.Run() // Commander do his thing
		resultStr, err := json.Marshal(result)
		if(err != nil){
			
			log.Fatal(err)
		}
		os.Stdout.Write(resultStr)
	}
}