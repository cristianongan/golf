package main

import (
	"flag"
	"log"
	"start/config"
	"start/server"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	environment := flag.String("ENV", "local", "descriptions") // -ENV is option for command line
	isTesting := flag.Bool("TEST", false, "run with db test")  // -TEST
	flag.Parse()
	config.ReadConfigFile(*environment)
	// envConfig := config.GetConfig()

	log.Println("Env", *environment)
	log.Println("Is Testing", *isTesting)

	// ============ Server start
	server.Init()
}
