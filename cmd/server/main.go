package main

import (
	"log"
	"os"
	"schedule/internal/handler"
	"schedule/internal/repository"
	"schedule/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "schedule:schedule@tcp(localhost:3306)/schedule_db?parseTime=true"
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	repo := repository.New(db)
	svc := service.New(repo)
	h := handler.New(svc)

	r := gin.Default()

	r.GET("/quotes", h.GetUnscheduledQuotes)
	r.POST("/jobs", h.CreateJob)
	r.PATCH("/jobs/:id/complete", h.CompleteJob)
	r.GET("/technicians/:id/notifications", h.GetTechnicianNotifications)
	r.GET("/managers/:id/notifications", h.GetManagerNotifications)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
