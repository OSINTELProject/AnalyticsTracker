package utils

import (
	"fmt"
	"encoding/json"
	"io/ioutil"
	server "analyticstracker/v1/server"
)

func ReadConfig( config_path string ) ( config server.ServerConfig ) {
	content , err := ioutil.ReadFile( config_path )
	if err != nil { fmt.Println( "Error when opening file: " , err ) }
	err = json.Unmarshal( content , &config )
	if err != nil { fmt.Println( "Error during Unmarshal(): " , err ) }
	return
}