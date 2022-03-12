package main

import (
	"os"
	"os/signal"
	"syscall"
	"fmt"
	fiber "github.com/gofiber/fiber/v2"
	server "analyticstracker/v1/server"
	utils "analyticstracker/v1/utils"
)

var s *fiber.App

func SetupCloseHandler() {
	c := make( chan os.Signal )
	signal.Notify( c , os.Interrupt , syscall.SIGTERM , syscall.SIGINT )
	go func() {
		<-c
		fmt.Println( "\r- Ctrl+C pressed in Terminal" )
		s.Shutdown()
		os.Exit( 0 )
	}()
}

func main() {
	SetupCloseHandler()
	config := utils.ReadConfig( "./config.json" )
	fmt.Println( config )
	s = server.New( &config )
	fmt.Println( s )
	fmt.Printf( "Listening on %s\n" , "9337" )
	result := s.Listen( ":9337" )
	fmt.Println( result )
}