package main

import (
	"flag"
	"github.com/harranali/authority"
	"log"
	"start/cmd/seed"
	"start/config"
	"start/datasources"
)

func main() {
	environment := flag.String("ENV", "local", "descriptions")
	config.ReadConfigFile(*environment)
	datasources.MySqlConnect()

	auth := authority.New(authority.Options{
		TablesPrefix: "auth_",
		DB:           datasources.GetDatabase(),
	})

	// create roles
	for _, authoritySeed := range (seed.AuthoritySeed{}).GetCreateRoles() {
		if err := authoritySeed.Run(auth); err != nil {
			log.Println("[DEBUG]", "Running seed:", authoritySeed.Name, ", Err:", err)
		}
	}

	// assign permissions
	for _, authoritySeed := range (seed.AuthoritySeed{}).GetAssignPermissions() {
		if err := authoritySeed.Run(auth); err != nil {
			log.Println("[DEBUG]", "Running seed:", authoritySeed.Name+",", "Err:", err)
		}
	}
}
