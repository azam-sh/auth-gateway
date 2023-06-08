package main

import (
	"authgateway/initializers"
	"authgateway/routes"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDB()
}

func main() {
	routes.StartRoutes()
}
