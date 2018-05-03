package main

import (
	"net/http"
	"google.golang.org/appengine"
	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
	"dnaController"
)


func main() {
	appengine.Main()
}

func init() {
	router := gin.New()
	router.Use(cors.Default())
	v1 := router.Group("/v1")
	v1.POST("/mutant", dnaController.IsMutant)
	v1.GET("/stats", dnaController.Statistics)
	v1.DELETE("/db", dnaController.ClearDB)

	http.Handle("/", router)
}
