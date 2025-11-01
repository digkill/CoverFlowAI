package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googleOAuth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

type GenerateCoverRequest struct {
	Image    string `json:"image" binding:"required"`
	Provider string `json:"provider,omitempty"` // "openai" or "nanobanana", defaults to "nanobanana"
	Prompt   string `json:"prompt,omitempty"`   // optional custom prompt for generation
}

type GenerateCoverResponse struct {
	ID       string `json:"id"`
	ImageURL string `json:"image_url"`
}

// Nano Banana API structures
type NanoBananaCreateTaskRequest struct {
	Model       string          `json:"model"`
	Input       NanoBananaInput `json:"input"`
	CallBackUrl string          `json:"callBackUrl,omitempty"`
}

type NanoBananaInput struct {
	Prompt       string   `json:"prompt"`
	ImageUrls    []string `json:"image_urls"`
	OutputFormat string   `json:"output_format,omitempty"`
	ImageSize    string   `json:"image_size,omitempty"`
}

type NanoBananaCreateTaskResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		TaskID string `json:"taskId"`
	} `json:"data"`
}

type NanoBananaTaskResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		TaskID       string `json:"taskId"`
		Model        string `json:"model"`
		State        string `json:"state"` // "waiting", "success", "fail"
		Param        string `json:"param"`
		ResultJSON   string `json:"resultJson"`
		FailCode     string `json:"failCode,omitempty"`
		FailMsg      string `json:"failMsg,omitempty"`
		CostTime     int    `json:"costTime,omitempty"`
		CompleteTime int64  `json:"completeTime,omitempty"`
		CreateTime   int64  `json:"createTime"`
	} `json:"data"`
}

type NanoBananaResult struct {
	ResultUrls []string `json:"resultUrls"`
}

type OpenAIRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	Image          string `json:"image"`
	N              int    `json:"n"`
	Size           string `json:"size"`
	ResponseFormat struct {
		Type string `json:"type"`
	} `json:"response_format"`
}

type OpenAIResponse struct {
	Data []struct {
		URL string `json:"url"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found")
	}

	// Get API keys
	openAIKey := os.Getenv("OPENAI_API_KEY")
	nanoBananaKey := os.Getenv("NANO_BANANA_API_KEY")

	if openAIKey == "" {
		fmt.Println("Warning: OPENAI_API_KEY not set in environment variables")
	}
	if nanoBananaKey == "" {
		fmt.Println("Warning: NANO_BANANA_API_KEY not set in environment variables")
	}

	// Create temp directory for images
	tempDir := filepath.Join(".", "temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		fmt.Printf("Warning: Failed to create temp directory: %v\n", err)
	}

	// Create storage directory for user images
	storageDir := filepath.Join(".", "storage")
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		fmt.Printf("Warning: Failed to create storage directory: %v\n", err)
	}

	// Initialize Redis client
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: os.Getenv("REDIS_PASSWORD"), // no password set
		DB:       0,                           // use default DB
	})

	// Test Redis connection
	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("Warning: Failed to connect to Redis: %v\n", err)
		fmt.Println("Redis is required for image caching. Please ensure Redis is running.")
	} else {
		fmt.Println("Connected to Redis successfully")
	}

	// Initialize database
	db, err := InitDB()
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Database initialized successfully")

	r := gin.Default()

	// Initialize session store
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		sessionSecret = "coverflow-ai-secret-key-change-in-production"
		fmt.Println("Warning: SESSION_SECRET not set, using default. Change in production!")
	}
	store := cookie.NewStore([]byte(sessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions("coverflow_session", store))

	// Initialize Google OAuth2 config
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	googleRedirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	if googleRedirectURL == "" {
		googleRedirectURL = "http://localhost:8080/api/auth/callback"
	}

	var oauthConfig *oauth2.Config
	if googleClientID != "" && googleClientSecret != "" {
		oauthConfig = &oauth2.Config{
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
			RedirectURL:  googleRedirectURL,
			Scopes:       []string{"openid", "profile", "email"},
			Endpoint:     google.Endpoint,
		}
		fmt.Println("Google OAuth configured")
	} else {
		fmt.Println("Warning: GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET not set. OAuth will not work.")
	}

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Serve static files from temp directory
	r.Static("/temp", tempDir)

	// Serve user storage files
	r.Static("/storage", storageDir)

	// Endpoint to serve images from Redis
	r.GET("/api/image/:imageId", func(c *gin.Context) {
		imageID := c.Param("imageId")
		ctx := context.Background()

		// Get image from Redis
		imageData, err := redisClient.Get(ctx, fmt.Sprintf("image:%s", imageID)).Bytes()
		if err == redis.Nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get image from cache"})
			return
		}

		// Determine content type
		contentType := "image/png"
		if strings.HasSuffix(imageID, ".jpg") || strings.HasSuffix(imageID, ".jpeg") {
			contentType = "image/jpeg"
		} else if strings.HasSuffix(imageID, ".webp") {
			contentType = "image/webp"
		}

		c.Data(http.StatusOK, contentType, imageData)
	})

	// Auth endpoints
	if oauthConfig != nil {
		// Start OAuth flow
		r.GET("/api/auth/google", func(c *gin.Context) {
			session := sessions.Default(c)

			// Generate state token for CSRF protection
			state := uuid.New().String()
			session.Set("oauth_state", state)
			if err := session.Save(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
				return
			}

			// Redirect to Google OAuth
			url := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
			c.Redirect(http.StatusTemporaryRedirect, url)
		})

		// OAuth callback
		r.GET("/api/auth/callback", func(c *gin.Context) {
			session := sessions.Default(c)

			// Verify state
			storedState := session.Get("oauth_state")
			state := c.Query("state")
			code := c.Query("code")

			if storedState == nil || storedState != state {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
				return
			}

			// Exchange code for token
			token, err := oauthConfig.Exchange(ctx, code)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to exchange token", "details": err.Error()})
				return
			}

			// Get user info
			client := oauthConfig.Client(ctx, token)
			oauth2Service, err := googleOAuth2.NewService(ctx, option.WithHTTPClient(client))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create OAuth2 service"})
				return
			}

			userInfo, err := oauth2Service.Userinfo.Get().Do()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
				return
			}

			// Get or create user in database
			user, err := GetOrCreateUser(db, userInfo.Id, userInfo.Email, userInfo.Name, userInfo.Picture)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
				return
			}

			// Save user info in session
			session.Set("user_id", user.ID)
			session.Set("user_email", user.Email)
			session.Set("user_name", user.Name)
			session.Set("user_picture", user.Picture)
			session.Delete("oauth_state")
			if err := session.Save(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
				return
			}

			// Redirect to frontend
			frontendURL := os.Getenv("FRONTEND_URL")
			if frontendURL == "" {
				frontendURL = "http://localhost:3000"
			}
			c.Redirect(http.StatusTemporaryRedirect, frontendURL)
		})
	}

	// Get current user
	r.GET("/api/auth/me", func(c *gin.Context) {
		session := sessions.Default(c)
		userIDValue := session.Get("user_id")

		if userIDValue == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		userID, _ := userIDValue.(string)

		// Get user from database
		var user User
		if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{
				"id":      session.Get("user_id"),
				"email":   session.Get("user_email"),
				"name":    session.Get("user_name"),
				"picture": session.Get("user_picture"),
			})
			return
		}

		// Check limits
		canGenerate, remaining, _ := CheckGenerationLimit(db, userID)

		c.JSON(http.StatusOK, gin.H{
			"id":                    user.ID,
			"email":                 user.Email,
			"name":                  user.Name,
			"picture":               user.Picture,
			"can_generate":          canGenerate,
			"generations_remaining": remaining,
			"free_generations_left": user.FreeGenerationsLeft,
			"paid_generations":      user.PaidGenerations,
		})
	})

	// Logout
	r.POST("/api/auth/logout", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		if err := session.Save(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear session"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
	})

	// User limits endpoint
	r.GET("/api/user/limits", func(c *gin.Context) {
		session := sessions.Default(c)
		userIDValue := session.Get("user_id")

		if userIDValue == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		userID, _ := userIDValue.(string)
		canGenerate, remaining, err := CheckGenerationLimit(db, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check limits"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"can_generate": canGenerate,
			"remaining":    remaining,
		})
	})

	// Get available packages
	r.GET("/api/packages", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"packages": Packages})
	})

	// Create payment order (Lava Top)
	r.POST("/api/payment/create", func(c *gin.Context) {
		session := sessions.Default(c)
		userIDValue := session.Get("user_id")

		if userIDValue == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		userID, _ := userIDValue.(string)

		var req struct {
			PackageType string `json:"package_type" binding:"required"` // "pack1", "pack2", "pack3"
			Currency    string `json:"currency" binding:"required"`     // "USD" or "RUB"
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Find package
		var selectedPackage *Package
		for _, pkg := range Packages {
			if pkg.Type == req.PackageType {
				selectedPackage = &pkg
				break
			}
		}

		if selectedPackage == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid package type"})
			return
		}

		// Get price based on currency
		var amount float64
		if req.Currency == "RUB" {
			amount = selectedPackage.PriceRUB
		} else {
			amount = selectedPackage.PriceUSD
		}

		// Create transaction
		transactionID := uuid.New().String()
		transaction := Transaction{
			ID:          transactionID,
			UserID:      userID,
			PackageType: req.PackageType,
			Amount:      amount,
			Currency:    req.Currency,
			Status:      "pending",
		}

		if err := db.Create(&transaction).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
			return
		}

		// Create Lava Top order
		orderID, paymentURL, err := createLavaTopOrder(transactionID, amount, req.Currency, selectedPackage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment order", "details": err.Error()})
			return
		}

		// Update transaction with Lava order ID
		transaction.LavaOrderID = orderID
		db.Save(&transaction)

		c.JSON(http.StatusOK, gin.H{
			"transaction_id": transactionID,
			"payment_url":    paymentURL,
			"order_id":       orderID,
		})
	})

	// Lava Top webhook
	r.POST("/api/payment/webhook", func(c *gin.Context) {
		// Verify webhook signature from Lava Top
		// Process payment confirmation
		var webhookData map[string]interface{}
		if err := c.ShouldBindJSON(&webhookData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook data"})
			return
		}

		// Extract order ID and status from webhook
		orderID, _ := webhookData["order_id"].(string)
		status, _ := webhookData["status"].(string)

		// Find transaction
		var transaction Transaction
		if err := db.Where("lava_order_id = ?", orderID).First(&transaction).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}

		if status == "success" || status == "completed" {
			// Update transaction status
			transaction.Status = "completed"
			db.Save(&transaction)

			// Find package and add generations
			for _, pkg := range Packages {
				if pkg.Type == transaction.PackageType {
					if err := AddPaidGenerations(db, transaction.UserID, pkg.Count); err != nil {
						fmt.Printf("Failed to add generations: %v\n", err)
					}
					break
				}
			}
		} else {
			transaction.Status = "failed"
			db.Save(&transaction)
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Health check endpoint
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Generate cover endpoint
	r.POST("/api/generate-cover", func(c *gin.Context) {
		var req GenerateCoverRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Get user ID from session
		session := sessions.Default(c)
		userIDValue := session.Get("user_id")
		userIDStr := "anonymous"
		if userIDValue != nil {
			if id, ok := userIDValue.(string); ok {
				userIDStr = id
			}
		}

		// Default to nanobanana if not specified
		provider := req.Provider
		if provider == "" {
			provider = "nanobanana"
		}

		// Remove data:image prefix if present
		imageData := req.Image
		var imageFormat string
		if strings.HasPrefix(imageData, "data:image") {
			parts := strings.Split(imageData, ";base64,")
			if len(parts) == 2 {
				// Extract format
				formatPart := strings.Split(parts[0], "/")
				if len(formatPart) == 2 {
					imageFormat = formatPart[1]
				}
				imageData = parts[1]
			} else {
				// Try old format
				parts := strings.Split(imageData, ",")
				if len(parts) == 2 {
					imageData = parts[1]
				}
			}
		}

		if imageFormat == "" {
			imageFormat = "png"
		}

		var coverURL string
		var err error

		// Check generation limit
		canGenerate, remaining, err := CheckGenerationLimit(db, userIDStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check generation limit"})
			return
		}

		if !canGenerate {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":     "No generations left",
				"remaining": remaining,
				"message":   "You have reached your generation limit. Please purchase a package to continue.",
			})
			return
		}

		// Determine if using free or paid generation
		var user User
		db.Where("id = ?", userIDStr).First(&user)
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		useFree := user.FreeGenerationsLeft > 0 && (user.LastFreeGeneration.Before(today) || user.LastFreeGeneration.IsZero())

		if provider == "nanobanana" {
			if nanoBananaKey == "" {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Nano Banana API key not configured"})
				return
			}
			ctx := context.Background()
			coverURL, err = generateCoverWithNanoBanana(imageData, imageFormat, nanoBananaKey, redisClient, ctx, userIDStr, storageDir, req.Prompt)
		} else if provider == "openai" {
			if openAIKey == "" {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "OpenAI API key not configured"})
				return
			}
			coverURL, err = generateCoverWithOpenAI(imageData, openAIKey, userIDStr, storageDir, req.Prompt)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider. Use 'openai' or 'nanobanana'"})
			return
		}

		// If generation successful, use up generation credit
		if err == nil {
			if err := UseGeneration(db, userIDStr, useFree); err != nil {
				fmt.Printf("Warning: Failed to update generation count: %v\n", err)
			}

			// Record generation
			generationID := uuid.New().String()
			generation := Generation{
				ID:       generationID,
				UserID:   userIDStr,
				ImageURL: coverURL,
				Provider: provider,
				IsFree:   useFree,
			}
			db.Create(&generation)
		}

		if err != nil {
			fmt.Printf("Error generating cover: %v\n", err)
			// Determine appropriate status code
			statusCode := http.StatusInternalServerError
			errMsg := err.Error()

			// Check for specific error types
			if strings.Contains(errMsg, "API key not configured") || strings.Contains(errMsg, "IMGBB_API_KEY not set") {
				statusCode = http.StatusBadRequest
			} else if strings.Contains(errMsg, "authentication failed") || strings.Contains(errMsg, "insufficient account balance") {
				statusCode = http.StatusUnauthorized
			}

			c.JSON(statusCode, gin.H{
				"error":   "Failed to generate cover",
				"details": errMsg,
			})
			return
		}

		response := GenerateCoverResponse{
			ID:       uuid.New().String(),
			ImageURL: coverURL,
		}

		c.JSON(http.StatusOK, response)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s\n", port)
	if err := r.Run(":" + port); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}

func generateCoverWithOpenAI(imageData string, apiKey string, userID string, storageDir string, customPrompt string) (string, error) {
	prompt := customPrompt
	if prompt == "" {
		prompt = "Create a professional YouTube thumbnail cover based on this collage. Make it visually appealing, modern, and optimized for video thumbnails. Ensure high quality and attention-grabbing design."
	}

	openAIReq := map[string]interface{}{
		"model":  "dall-e-3",
		"prompt": prompt,
		"n":      1,
		"size":   "1024x1024",
	}

	reqBody, err := json.Marshal(openAIReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/images/generations", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if openAIResp.Error != nil {
		return "", fmt.Errorf("OpenAI API error: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Data) == 0 {
		return "", fmt.Errorf("no image URL in response")
	}

	// Download and save generated image
	resultURL := openAIResp.Data[0].URL
	savedPath, err := downloadAndSaveImage(resultURL, userID, storageDir)
	if err != nil {
		fmt.Printf("Warning: Failed to save image locally: %v\n", err)
		// Return original URL if save fails
		return resultURL, nil
	}

	// Return local URL
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return fmt.Sprintf("%s/storage/%s", baseURL, savedPath), nil
}

func generateCoverWithNanoBanana(imageData string, imageFormat string, apiKey string, redisClient *redis.Client, ctx context.Context, userID string, storageDir string, customPrompt string) (string, error) {

	// Decode base64 image
	decodedData, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 image: %w", err)
	}

	// Validate image size (max 10MB for Nano Banana API)
	if len(decodedData) > 10*1024*1024 {
		return "", fmt.Errorf("image size exceeds 10MB limit")
	}

	// Save image to Redis with expiration
	imageID := fmt.Sprintf("%s.%s", uuid.New().String(), imageFormat)
	redisKey := fmt.Sprintf("image:%s", imageID)

	err = redisClient.Set(ctx, redisKey, decodedData, 30*time.Minute).Err()
	if err != nil {
		return "", fmt.Errorf("failed to save image to Redis: %w", err)
	}
	fmt.Printf("Image saved to Redis: %s (size: %d bytes)\n", imageID, len(decodedData))

	// Create public URL for the image
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	imageURL := fmt.Sprintf("%s/api/image/%s", baseURL, imageID)
	fmt.Printf("Image accessible at: %s\n", imageURL)

	// Create task
	taskID, err := createNanoBananaTask(imageURL, apiKey, customPrompt)
	if err != nil {
		// Clean up Redis on error
		redisClient.Del(ctx, redisKey)
		return "", fmt.Errorf("failed to create Nano Banana task: %w", err)
	}

	fmt.Printf("Nano Banana task created: %s\n", taskID)

	// Poll for result
	maxAttempts := 120 // 10 minutes max (5 second intervals)
	interval := 5 * time.Second

	resultURL, err := pollNanoBananaTask(taskID, apiKey, maxAttempts, interval)
	if err != nil {
		// Clean up Redis on error
		redisClient.Del(ctx, redisKey)
		return "", fmt.Errorf("failed to get Nano Banana result: %w", err)
	}

	fmt.Printf("Nano Banana task completed successfully. Result URL: %s\n", resultURL)

	// Clean up Redis cache after successful generation
	redisClient.Del(ctx, redisKey)
	fmt.Printf("Cleaned up Redis cache for image: %s\n", imageID)

	// Download and save generated image
	savedPath, err := downloadAndSaveImage(resultURL, userID, storageDir)
	if err != nil {
		fmt.Printf("Warning: Failed to save image locally: %v\n", err)
		// Return original URL if save fails
		return resultURL, nil
	}

	// Return local URL
	return fmt.Sprintf("%s/storage/%s", baseURL, savedPath), nil
}

func createNanoBananaTask(imageURL string, apiKey string, customPrompt string) (string, error) {
	// Use custom prompt if provided, otherwise use default
	prompt := customPrompt
	if prompt == "" {
		prompt = "Transform this collage into a professional YouTube thumbnail cover. " +
			"Make it visually striking, modern, and optimized for video thumbnails. " +
			"Ensure high quality, attention-grabbing design with good contrast and readable text. " +
			"Maintain the key elements from the collage but enhance them professionally. " +
			"Use 16:9 aspect ratio suitable for YouTube thumbnails."
	}

	reqBody := NanoBananaCreateTaskRequest{
		Model: "google/nano-banana-edit",
		Input: NanoBananaInput{
			Prompt:       prompt,
			ImageUrls:    []string{imageURL},
			OutputFormat: "png",
			ImageSize:    "16:9", // YouTube thumbnail standard
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.kie.ai/api/v1/jobs/createTask", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode == 401 {
		return "", fmt.Errorf("authentication failed: check your NANO_BANANA_API_KEY")
	}
	if resp.StatusCode == 402 {
		return "", fmt.Errorf("insufficient account balance")
	}
	if resp.StatusCode == 429 {
		return "", fmt.Errorf("rate limit exceeded, please try again later")
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("nano banana API error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var taskResp NanoBananaCreateTaskResponse
	if err := json.Unmarshal(body, &taskResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if taskResp.Code != 200 {
		errorMsg := taskResp.Msg
		switch taskResp.Code {
		case 400:
			errorMsg = "invalid request parameters: " + errorMsg
		case 401:
			errorMsg = "authentication failed: " + errorMsg
		case 402:
			errorMsg = "insufficient account balance: " + errorMsg
		case 422:
			errorMsg = "parameter validation failed: " + errorMsg
		case 429:
			errorMsg = "rate limit exceeded: " + errorMsg
		case 500:
			errorMsg = "internal server error: " + errorMsg
		}
		return "", fmt.Errorf("nano banana API error: %s (code: %d)", errorMsg, taskResp.Code)
	}

	if taskResp.Data.TaskID == "" {
		return "", fmt.Errorf("no task ID in response")
	}

	return taskResp.Data.TaskID, nil
}

func pollNanoBananaTask(taskID string, apiKey string, maxAttempts int, interval time.Duration) (string, error) {
	url := fmt.Sprintf("https://api.kie.ai/api/v1/jobs/recordInfo?taskId=%s", taskID)

	for i := 0; i < maxAttempts; i++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+apiKey)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Poll attempt %d/%d failed, retrying...\n", i+1, maxAttempts)
			time.Sleep(interval)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Printf("Poll attempt %d/%d failed to read response, retrying...\n", i+1, maxAttempts)
			time.Sleep(interval)
			continue
		}

		var taskResp NanoBananaTaskResponse
		if err := json.Unmarshal(body, &taskResp); err != nil {
			fmt.Printf("Poll attempt %d/%d failed to parse response, retrying...\n", i+1, maxAttempts)
			time.Sleep(interval)
			continue
		}

		if taskResp.Code != 200 {
			return "", fmt.Errorf("nano banana API error: %s (code: %d)", taskResp.Msg, taskResp.Code)
		}

		state := taskResp.Data.State
		if state == "success" {
			// Parse result JSON
			if taskResp.Data.ResultJSON == "" {
				return "", fmt.Errorf("empty result JSON in response")
			}

			var result NanoBananaResult
			if err := json.Unmarshal([]byte(taskResp.Data.ResultJSON), &result); err != nil {
				return "", fmt.Errorf("failed to parse result JSON: %w", err)
			}

			if len(result.ResultUrls) == 0 {
				return "", fmt.Errorf("no result URLs in response")
			}

			if taskResp.Data.CostTime > 0 {
				fmt.Printf("Task completed in %d ms\n", taskResp.Data.CostTime)
			}

			return result.ResultUrls[0], nil
		} else if state == "fail" {
			failMsg := taskResp.Data.FailMsg
			if failMsg == "" {
				failMsg = "unknown error"
			}
			return "", fmt.Errorf("task failed: %s (failCode: %s)", failMsg, taskResp.Data.FailCode)
		}

		// Task is still processing (waiting)
		if (i+1)%6 == 0 { // Log every 30 seconds
			fmt.Printf("Task status: %s (waiting for completion, attempt %d/%d)...\n", state, i+1, maxAttempts)
		}
		time.Sleep(interval)
	}

	return "", fmt.Errorf("task timeout after %d attempts (approximately %.1f minutes)", maxAttempts, float64(maxAttempts)*interval.Seconds()/60)
}

// downloadAndSaveImage downloads an image from URL and saves it to storage/userid/
func downloadAndSaveImage(imageURL string, userID string, storageDir string) (string, error) {
	// Download image
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	// Read image data
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}

	// Create user directory
	userDir := filepath.Join(storageDir, userID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create user directory: %w", err)
	}

	// Generate filename with timestamp
	filename := fmt.Sprintf("%s_%d.png", uuid.New().String(), time.Now().Unix())
	filePath := filepath.Join(userDir, filename)

	// Save image
	if err := os.WriteFile(filePath, imageData, 0644); err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	fmt.Printf("Image saved to: %s (size: %d bytes)\n", filePath, len(imageData))

	// Return relative path from storage directory
	return filepath.Join(userID, filename), nil
}
