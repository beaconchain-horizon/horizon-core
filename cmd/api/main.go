package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
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
	onlineDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("⚠️ خطا در اتصال به PostgreSQL: %v", err)
		return
	}
	onlineDB.AutoMigrate(&Transaction{})
	log.Println("✅ اتصال به PostgreSQL برقرار شد")
}

func connectOfflineDB() {
	var err error
	offlineDB, err = gorm.Open(sqlite.Open("horizon_offline.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("خطا در اتصال به SQLite:", err)
	}
	offlineDB.AutoMigrate(&Transaction{})
	log.Println("✅ اتصال به SQLite برقرار شد")
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
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "online"})
	})
	r.POST("/api/v1/transactions", func(c *gin.Context) {
		var tx Transaction
		c.ShouldBindJSON(&tx)
		if onlineDB == nil {
			tx.Status = "offline"
			offlineDB.Create(&tx)
			c.JSON(201, gin.H{"message": "ذخیره آفلاین", "transaction": tx})
			return
		}
		tx.Status = "confirmed"
		onlineDB.Create(&tx)
		c.JSON(201, gin.H{"message": "ذخیره آنلاین", "transaction": tx})
	})
	r.GET("/api/v1/transactions", func(c *gin.Context) {
		var offline, online []Transaction
		offlineDB.Find(&offline)
		if onlineDB != nil {
			onlineDB.Find(&online)
		}
		c.JSON(200, gin.H{"offline": offline, "online": online})
	})
	r.POST("/api/v1/sync", func(c *gin.Context) {
		if c.GetHeader("X-Admin-Key") != os.Getenv("ADMIN_TOKEN") {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		syncTransactions()
		c.JSON(200, gin.H{"message": "Sync done"})
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
