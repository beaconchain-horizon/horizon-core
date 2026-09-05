package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	onlineDB  *gorm.DB
	offlineDB *gorm.DB
)

type Transaction struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Amount    int64     `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func connectOnlineDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgresql://root:Ck925kQ1Qe3ypJvDyLIDoX3g@horizon:5432/postgres?sslmode=disable"
	}
	var err error
	if strings.HasPrefix(dsn, "sqlite://") {
		dsn = strings.TrimPrefix(dsn, "sqlite://")
		onlineDB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	} else {
		onlineDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}
	if err != nil {
		log.Printf("Error connecting to primary DB: %v", err)
		return
	}
	onlineDB.AutoMigrate(&Transaction{})
	log.Println("Connected to primary DB")
}

func connectOfflineDB() {
	var err error
	offlineDB, err = gorm.Open(sqlite.Open("horizon_offline.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Error connecting to SQLite:", err)
	}
	offlineDB.AutoMigrate(&Transaction{})
	log.Println("Connected to SQLite")
}

func syncTransactions() {
	if onlineDB == nil {
		return
	}
	var offlineTransactions []Transaction
	offlineDB.Where("status = ?", "offline").Find(&offlineTransactions)
	for _, tx := range offlineTransactions {
		tx.Status = "confirmed"
		onlineDB.Create(&tx)
	}
	offlineDB.Delete(&Transaction{}, "status = ?", "offline")
}

func main() {
	godotenv.Load()
	connectOnlineDB()
	connectOfflineDB()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "X-API-Key", "X-Admin-Key"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "online"})
	})
	r.POST("/api/v1/transactions", func(c *gin.Context) {
		var tx Transaction
		c.ShouldBindJSON(&tx)
		if onlineDB == nil {
			tx.Status = "offline"
			offlineDB.Create(&tx)
			c.JSON(http.StatusCreated, gin.H{"message": "Saved offline", "transaction": tx})
			return
		}
		tx.Status = "confirmed"
		onlineDB.Create(&tx)
		c.JSON(http.StatusCreated, gin.H{"message": "Saved online", "transaction": tx})
	})
	r.GET("/api/v1/transactions", func(c *gin.Context) {
		var offline, online []Transaction
		offlineDB.Find(&offline)
		if onlineDB != nil {
			onlineDB.Find(&online)
		}
		c.JSON(http.StatusOK, gin.H{"offline": offline, "online": online})
	})
	r.POST("/api/v1/sync", func(c *gin.Context) {
		if c.GetHeader("X-Admin-Key") != os.Getenv("ADMIN_TOKEN") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		syncTransactions()
		c.JSON(http.StatusOK, gin.H{"message": "Sync done"})
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
