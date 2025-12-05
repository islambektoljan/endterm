package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
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
	r.POST("/upload-txt", uplaodTxtFile)
	r.POST("/upload-json", uplaodJSONHandler)
	r.POST("/upload-txt", uplaodTXTHandler)
	r.Run(":8080")
}

func checkExtension(filename, allowedExt string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == allowedExt
}

func uploadFileWithExtension(c *gin.Context, allowedExt string) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "No file received"})
		return
	}
	defer file.Close()

	if !checkExtension(header.Filename, allowedExt) {
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("Invalid file type. Expected %s, got %s",
				allowedExt, filepath.Ext(header.Filename)),
		})
		return
	}

	objectName := fmt.Sprintf("%s_%d%s",
		strings.TrimPrefix(allowedExt, "."),
		time.Now().UnixNano(),
		allowedExt)

	_, err = minioClient.PutObject(c, bucketName, objectName, file, header.Size, minio.PutObjectOptions{})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to upload to MinIO: " + err.Error()})
		return
	}

	fileRecord := File{
		FileName:     objectName,
		OriginalName: header.Filename,
		Extension:    allowedExt,
		Size:         header.Size,
	}

	db.Create(&fileRecord)

	c.JSON(200, gin.H{
		"id":       fileRecord.ID,
		"filename": header.Filename,
		"type":     strings.TrimPrefix(allowedExt, "."),
		"message":  "File uploaded successfully",
	})
}

func uplaodTxtFile(c *gin.Context) {
	uploadFileWithExtension(c, ".txt")
}

func uplaodJSONHandler(c *gin.Context) {
	uploadFileWithExtension(c, ".json")
}

func uplaodTXTHandler(c *gin.Context) {
	uploadFileWithExtension(c, ".txt")
}
