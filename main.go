package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type File struct {
	ID           uint      `gorm:"primaryKey"`
	FileName     string    `gorm:"size:255;not null"`
	OriginalName string    `gorm:"size:255;not null"`
	Extension    string    `gorm:"size:50;not null"`
	Size         int64     `gorm:"not null"`
	UploadDate   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

var db *gorm.DB
var minioClient *minio.Client
var bucketName = "fisrt-bucket"

func initDB() {
	var err error
	dsn := "host=postgres user=postgres password=postgres dbname=filedb port=5432 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("DB error:", err)
	}
	db.AutoMigrate(&File{})
}

func initMinIO() {
	var err error
	minioClient, err = minio.New("minio:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("admin", "admin12345", ""),
		Secure: false,
	})
	if err != nil {
		log.Fatal("MinIO error:", err)
	}
}

func main() {

	initDB()
	initMinIO()

	r := gin.Default()
	r.Run(":8080")
}
