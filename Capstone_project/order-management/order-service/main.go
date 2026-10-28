package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/graphql-go/handler"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"order-management/order-service/client"
	"order-management/order-service/config"
	"order-management/order-service/controller"
	"order-management/order-service/database"
	"order-management/order-service/graphql"
	"order-management/order-service/middleware"
	"order-management/order-service/repository"
	"order-management/order-service/service"
	"os"
)

// @title Production-Ready Order Management API
// @version 1.0
// @description Order management REST API with product stock reservation and resilience.
// @host localhost:8080
// @BasePath /
func main() {
	c := config.Load()
	ctx := context.Background()
	db, err := database.Connect(ctx, c)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	middleware.InitMetrics()
	r := repository.NewOrderRepository(db)
	pc := client.NewProductClient(c.ProductServiceURL)
	s := service.NewOrderService(r, pc)
	ctl := controller.NewOrderController(s)
	g := gin.New()
	g.Use(middleware.RequestID(), middleware.Headers(), middleware.Logging(), middleware.Recovery(), middleware.Metrics())
	g.GET("/health/live", func(c *gin.Context) { c.JSON(200, gin.H{"status": "UP"}) })
	g.GET("/health/ready", func(c *gin.Context) {
		if err := db.Ping(c); err != nil {
			c.JSON(503, gin.H{"status": "DOWN"})
			return
		}
		c.JSON(200, gin.H{"status": "UP"})
	})
	g.GET("/metrics", gin.WrapH(promhttp.Handler()))
	api := g.Group("/api/v1/orders")
	api.POST("", ctl.Create)
	api.GET("", ctl.List)
	api.GET("/:id", ctl.Get)
	api.PUT("/:id/status", ctl.Status)
	schema, err := graphql.NewSchema(s)
	if err != nil {
		log.Fatal(err)
	}
	h := handler.New(&handler.Config{Schema: &schema, Pretty: true, GraphiQL: true})
	g.Any("/graphql", gin.WrapH(h))
	addr := ":" + c.AppPort
	log.Println("order API listening on", addr)
	if err := http.ListenAndServe(addr, g); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
