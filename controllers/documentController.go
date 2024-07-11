package controllers

import (
	"context"
	"fmt"
	"image/png"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"paper-management-backend/database"
	"paper-management-backend/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/unidoc/unipdf/v3/model"
	"github.com/unidoc/unipdf/v3/render"
	"gorm.io/gorm"
)

func renderPageAsPNG(page *model.PdfPage, outputPath string) error {
	// Initialize the renderer.
	r := render.NewImageDevice()

	// Render the page to an image.
	img, err := r.Render(page)
	if err != nil {
		return err
	}

	// Create the output file.
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Encode the image to PNG.
	err = png.Encode(file, img)
	if err != nil {
		return err
	}

	return nil
}

func GetAllJenisDocuments(c *fiber.Ctx) error {
	var jenisDocuments []models.JenisDocument
	if err := database.DB.Find(&jenisDocuments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve documents"})
	}
	return c.JSON(jenisDocuments)
}

func GetAllDocuments(c *fiber.Ctx) error {
	userId := c.Locals("userId").(uuid.UUID)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("perPage", "10"))
	sort := c.Query("sort", "created_at")
	order := strings.ToUpper(c.Query("order", "DESC"))
	search := c.Query("search", "")

	offset := (page - 1) * perPage

	var documents []models.Document
	query := database.DB.Preload("JenisDocument").Where("user_id = ?", userId)

	if search != "" {
		query = query.Where("file_name LIKE ? OR location LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	if err := query.Model(&models.Document{}).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to count documents",
			"details": err.Error(),
		})
	}

	query = query.Order(gorm.Expr(sort + " " + order))

	if err := query.Limit(perPage).Offset(offset).Find(&documents).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve documents",
			"details": err.Error(),
		})
	}

	meta := fiber.Map{
		"page":     page,
		"perPage":  perPage,
		"sort":     sort,
		"order":    order,
		"search":   search,
		"total":    total,
		"lastPage": int(math.Ceil(float64(total) / float64(perPage))),
	}

	response := fiber.Map{
		"data": documents,
		"meta": meta,
	}

	return c.JSON(response)
}

func UploadDocument(c *fiber.Ctx) error {
	// Load environment variables
	godotenv.Load()

	userId := c.Locals("userId").(uuid.UUID)

	// Parse form data
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse form data"})
	}

	files := form.File["file"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No file uploaded"})
	}

	// Get the first file
	file := files[0]
	fileName := file.Filename // Use file.Filename instead of f.Name

	// Load the AWS S3 configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"")),
		config.WithRegion("us-east-1"), // Default region
		config.WithEndpointResolver(aws.EndpointResolverFunc(
			func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL: os.Getenv("BASE_URL_R2"),
				}, nil
			})),
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to load AWS config",
			"details": err.Error(),
		})
	}

	client := s3.NewFromConfig(cfg)

	// Generate a unique key for the file
	jenisDocumentStr := c.FormValue("jenis_document")
	jenisDocumentUint64, err := strconv.ParseUint(jenisDocumentStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid jenis_document value"})
	}
	jenisDocument := uint(jenisDocumentUint64)

	location := c.FormValue("location")
	now := time.Now().Format("02-01-2006")
	fileKey := fmt.Sprintf("documents/%d/%s/%s_%s", jenisDocument, now, uuid.New().String(), fileName)

	// Upload the file to AWS S3
	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot open file"})
	}
	defer func() {
		f.Close()
		// Hapus file sementara setelah digunakan
		os.Remove(file.Filename)
	}()

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(os.Getenv("AWS_BUCKET_NAME")),
		Key:    aws.String(fileKey),
		Body:   f,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to upload file",
			"details": err.Error(),
		})
	}

	// Save the file locally to process OCR
	tempFilePath := fmt.Sprintf("%s/%s", os.Getenv("OCR_TEMP_DIR"), fileKey)
	tempDirPath := fmt.Sprintf("%s/documents/%d/%s", os.Getenv("OCR_TEMP_DIR"), jenisDocument, now)
	os.MkdirAll(tempDirPath, os.ModePerm)
	localFile, err := os.Create(tempFilePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot create temporary file"})
	}
	defer func() {
		localFile.Close()
		// Hapus file sementara setelah digunakan
		os.Remove(tempFilePath)
	}()

	_, err = f.Seek(0, 0)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot seek file"})
	}

	_, err = io.Copy(localFile, f)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot save file locally"})
	}

	// Process OCR for PDF file
	pdfFile, err := os.Open(tempFilePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot open PDF file for OCR"})
	}
	defer pdfFile.Close()

	pdfReader, err := model.NewPdfReader(pdfFile)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read PDF file"})
	}

	// Get the first page for thumbnail
	pageNum := 1
	page, err := pdfReader.GetPage(pageNum)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get PDF page"})
	}

	// Generate a unique filename for the thumbnail
	thumbnailFileName := fmt.Sprintf("%s_%s.png", uuid.New().String(), fileName)

	// Save the thumbnail locally
	thumbnailFilePath := fmt.Sprintf("%s/%s", os.Getenv("OCR_TEMP_DIR"), thumbnailFileName)
	if err := renderPageAsPNG(page, thumbnailFilePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to render PDF page as thumbnail",
			"details": err.Error(),
		})
	}
	defer func() {
		// Hapus file sementara setelah digunakan
		os.Remove(thumbnailFilePath)
	}()

	// Upload the thumbnail to AWS S3
	thumbnailFile, err := os.Open(thumbnailFilePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot open thumbnail file"})
	}
	defer func() {
		thumbnailFile.Close()
		// Hapus file sementara setelah digunakan
		os.Remove(thumbnailFilePath)
	}()

	thumbnailKey := fmt.Sprintf("thumbnails/%d/%s/%s", jenisDocument, now, thumbnailFileName)

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(os.Getenv("AWS_BUCKET_NAME")),
		Key:    aws.String(thumbnailKey),
		Body:   thumbnailFile,
		ACL:    "public-read",
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to upload thumbnail",
			"details": err.Error(),
		})
	}

	// Generate thumbnail URL
	thumbnailURL := fmt.Sprintf("%s/%s", os.Getenv("PUBLIC_URL_R2"), thumbnailKey)

	// Save the document information in the database
	document := models.Document{
		UserID:          userId,
		FileName:        fileName,
		JenisDocumentID: jenisDocument,
		URLFile:         fmt.Sprintf("%s/%s", os.Getenv("BASE_URL_R2"), fileKey),
		Key:             fileKey,
		Location:        location,
		ThumbnailURL:    thumbnailURL,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := database.DB.Create(&document).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to save document in database",
			"details": err.Error(),
		})
	}

	response := fiber.Map{
		"file_name":     document.FileName,
		"url_file":      document.URLFile,
		"key":           document.Key,
		"location":      document.Location,
		"thumbnail_url": document.ThumbnailURL,
		"created_at":    document.CreatedAt,
		"updated_at":    document.UpdatedAt,
	}

	return c.JSON(response)
}

func ShareDocument(c *fiber.Ctx) error {
	// Load environment variables
	godotenv.Load()

	documentID := c.Params("id")
	var document models.Document
	if err := database.DB.First(&document, "id = ?", documentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Document not found"})
	}

	// Parse form data for time limit
	var request struct {
		TimeLimit string `json:"time_limit"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse request"})
	}

	// Determine expiration time
	var expires time.Duration
	if request.TimeLimit == "forever" {
		expires = 0
	} else {
		parsedDuration, err := time.ParseDuration(request.TimeLimit)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid time limit format"})
		}
		expires = parsedDuration
	}

	// Load the Cloudflare R2 configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load AWS configuration"})
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.EndpointResolver = s3.EndpointResolverFromURL(os.Getenv("BASE_URL_R2"))
	})

	// Generate presigned URL
	presignClient := s3.NewPresignClient(client)
	presignParams := &s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("AWS_BUCKET_NAME")),
		Key:    aws.String(document.Key),
	}

	presignOpts := func(po *s3.PresignOptions) {
		if expires > 0 {
			po.Expires = expires
		}
	}

	presignedURL, err := presignClient.PresignGetObject(context.TODO(), presignParams, presignOpts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate presigned URL"})
	}

	return c.JSON(fiber.Map{"url": presignedURL.URL})
}
