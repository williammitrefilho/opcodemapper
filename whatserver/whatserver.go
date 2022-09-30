package whatserver

import(
	"net/http"
//	"modgo.com/what"
	"log"
//	"encoding/json"
)

func Handler(Writer http.ResponseWriter, Request *http.Request){
	
	if(Request.Method != "POST"){
		
		log.Println("Some other method")
	} else {
		
		if(Request.ContentLength == 0){
			
			log.Println("Empty request")
		} else {
			
			
		}
	}
}