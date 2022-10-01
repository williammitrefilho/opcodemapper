package server

import(
	"crypto/tls"
	"log"
	"net/http"
	"time"
	"os"
)

func Listen(handler func(http.ResponseWriter, *http.Request)){
	
	certPem, err := os.ReadFile("C:/opcodemapper/whatserver/O=Acme Co.crt")
	if err != nil{
		
		log.Fatal(err)
	}
	keyPem, err := os.ReadFile("C:/opcodemapper/whatserver/key.k")
	if err != nil{
		
		log.Fatal(err)
	}
	cert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		log.Fatal(err)
	}
	cfg := &tls.Config{Certificates: []tls.Certificate{cert}}
	srv := &http.Server{
		TLSConfig:    cfg,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}
	http.HandleFunc("/", handler)
	log.Fatal(srv.ListenAndServeTLS("", ""))
}