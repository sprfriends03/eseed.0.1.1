package main

import (
	"app/route"
	"app/store"

	"github.com/sirupsen/logrus"
)

// @title                       Document APIs
// @version                     1.0
// @description                 This is the API documentation for the Eseed service.
// @termsOfService              http://swagger.io/terms/
// @contact.name                API Support
// @contact.url                 http://www.swagger.io/support
// @contact.email               support@swagger.io
// @license.name                Apache 2.0
// @license.url                 http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath                    /
// @securityDefinitions.basic   BasicAuth
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @tag.name Auth
// @tag.description Operations related to general user authentication
// @tag.name Member Auth
// @tag.description Operations related to member authentication and authorization
// @tag.name Cms
// @tag.description Operations related to Content Management System
func main() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.Fatalln(route.Bootstrap(store.New()))
}
