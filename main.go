package main

import (
	"app/route"
	"app/store"

	"github.com/sirupsen/logrus"
)

// @title                       Document APIs
// @securityDefinitions.basic   BasicAuth
// @securityDefinitions.apikey  BearerAuth
// @in                          Header
// @name                        Authorization
// @BasePath                    /
func main() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.Fatalln(route.Bootstrap(store.New()))
}
