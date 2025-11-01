package main

import (
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"uniqueIndex" json:"email"`
	Name      string    `json:"name"`
	Picture   string    `json:"picture"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Credits
	FreeGenerationsLeft int       `gorm:"default:0" json:"free_generations_left"`
	LastFreeGeneration  time.Time `json:"last_free_generation"`
	PaidGenerations     int       `gorm:"default:0" json:"paid_generations"`
}

type Generation struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index" json:"user_id"`
	ImageURL  string    `json:"image_url"`
	Provider  string    `json:"provider"`
	IsFree    bool      `json:"is_free"`
	CreatedAt time.Time `json:"created_at"`
}

type Transaction struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	UserID      string    `gorm:"index" json:"user_id"`
	PackageType string    `json:"package_type"` // "pack1", "pack2", "pack3"
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"` // "USD" or "RUB"
	Status      string    `json:"status"`    // "pending", "completed", "failed"
	LavaOrderID string    `gorm:"uniqueIndex" json:"lava_order_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Package struct {
	Type     string  `json:"type"`     // "pack1", "pack2", "pack3"
	Name     string  `json:"name"`
	Count    int     `json:"count"`    // количество генераций
	PriceUSD float64 `json:"price_usd"`
	PriceRUB float64 `json:"price_rub"`
	Popular  bool    `json:"popular"` // флаг "Популярный"
}

var Packages = []Package{
	{Type: "pack1", Name: "Стартовый", Count: 10, PriceUSD: 2.99, PriceRUB: 249, Popular: false},
	{Type: "pack2", Name: "Базовый", Count: 30, PriceUSD: 7.99, PriceRUB: 599, Popular: true},
	{Type: "pack3", Name: "Профессиональный", Count: 100, PriceUSD: 19.99, PriceRUB: 1499, Popular: false},
}

func InitDB() (*gorm.DB, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "coverflow.db"
	}
	
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto migrate
	err = db.AutoMigrate(&User{}, &Generation{}, &Transaction{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func GetOrCreateUser(db *gorm.DB, userID string, email string, name string, picture string) (*User, error) {
	var user User
	err := db.Where("id = ?", userID).First(&user).Error
	
	if err == gorm.ErrRecordNotFound {
		// Create new user
		user = User{
			ID:                  userID,
			Email:               email,
			Name:                name,
			Picture:             picture,
			FreeGenerationsLeft: 1, // Start with 1 free generation
			LastFreeGeneration:  time.Time{},
			PaidGenerations:     0,
		}
		err = db.Create(&user).Error
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		// Update user info if changed
		user.Email = email
		user.Name = name
		user.Picture = picture
		db.Save(&user)
	}

	return &user, nil
}

func CheckGenerationLimit(db *gorm.DB, userID string) (bool, int, error) {
	var user User
	err := db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return false, 0, err
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	
	// Reset free generation if new day
	if user.LastFreeGeneration.Before(today) {
		user.FreeGenerationsLeft = 1
		user.LastFreeGeneration = time.Time{}
		db.Save(&user)
	}

	// Check if user can generate
	canGenerate := user.FreeGenerationsLeft > 0 || user.PaidGenerations > 0
	remaining := user.FreeGenerationsLeft + user.PaidGenerations

	return canGenerate, remaining, nil
}

func UseGeneration(db *gorm.DB, userID string, isFree bool) error {
	var user User
	err := db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return err
	}

	if isFree {
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		
		// Reset if new day
		if user.LastFreeGeneration.Before(today) {
			user.FreeGenerationsLeft = 1
		}
		
		if user.FreeGenerationsLeft > 0 {
			user.FreeGenerationsLeft--
			user.LastFreeGeneration = time.Now()
		} else {
			return gorm.ErrRecordNotFound // No free generations left
		}
	} else {
		if user.PaidGenerations > 0 {
			user.PaidGenerations--
		} else {
			return gorm.ErrRecordNotFound // No paid generations left
		}
	}

	return db.Save(&user).Error
}

func AddPaidGenerations(db *gorm.DB, userID string, count int) error {
	var user User
	err := db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return err
	}

	user.PaidGenerations += count
	return db.Save(&user).Error
}

