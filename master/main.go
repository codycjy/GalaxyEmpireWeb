package main

import "GalaxyEmpireWeb/routes"

func main() {
	r := routes.RegisterRoutes()
	r.Run(":9333")
}
