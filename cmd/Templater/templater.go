package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/sa-kemper/golangGetTextTest/internal/LogHelp"
	"github.com/sa-kemper/golangGetTextTest/internal/Response"
)

var config struct {
	BindAddress string
}
var HandleUtility *Response.Utility

func init() {
	flag.StringVar(&config.BindAddress, "bind-address", "127.0.0.1:8081", "Bind address")
}

func main() {
	flag.Parse()
	server := http.Server{
		Addr:                         config.BindAddress,
		Handler:                      nil,
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		ReadTimeout:                  0,
		ReadHeaderTimeout:            0,
		WriteTimeout:                 0,
		IdleTimeout:                  0,
		MaxHeaderBytes:               0,
		TLSNextProto:                 nil,
		ConnState:                    nil,
		ErrorLog:                     log.Default(),
		BaseContext:                  nil,
		ConnContext:                  nil,
		Protocols:                    nil,
	}
	HandleUtility = &Response.Utility{
		Template: Templates,
	}

	for path, function := range routingTable {
		http.HandleFunc(path, function)
	}

	if LogHelp.PrintableLogLevel >= 3 {

		fmt.Print(`|                           _____                                                   
|                          / ____|						             |                         
|                         | |  __  ___						         |                         
|                         | | |_ |/ _ \						         |                        
|                         | |__| | (_) |					         |	                       
|                          \_____|\___/_   _            _	         |		   			   
|                          / ____|    | | | |          | |	         |	   				  
|                         | |  __  ___| |_| |_ _____  _| |_	         |   					 
|                         | | |_ |/ _ \ __| __/ _ \ \/ / __|         |   						
|                         | |__| |  __/ |_| ||  __/>  <| |_	         |   					 
|                          \_____|\___|\__|\__\___/_/\_\\__|         |   						
|                         |__   __|      | |				         |		                   
|                            | | ___  ___| |_				         |		                  
|                            | |/ _ \/ __| __|				         |		                 
|                            | |  __/\__ \ |_				         |		                  
|                            |_|\___||___/\__|				         |		                 
                                   `)
		fmt.Printf("\nListening on http://%s\n", config.BindAddress)
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("err")
		panic(err)
	}
}
