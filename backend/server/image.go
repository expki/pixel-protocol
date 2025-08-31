package server

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/expki/backend/pixel-protocol/database"
	"github.com/expki/backend/pixel-protocol/logger"
	"github.com/google/uuid"
	"gorm.io/plugin/dbresolver"
)

// HandleHeroImage handles /api/hero/:id/image
func (s *Server) HandleHeroImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract hero ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/hero/")
	segments := strings.Split(path, "/")
	
	if len(segments) < 2 || segments[1] != "image" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	heroID, err := uuid.Parse(segments[0])
	if err != nil {
		http.Error(w, "Invalid hero ID", http.StatusBadRequest)
		return
	}

	// Get hero details from database
	var hero database.Hero
	err = s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).
		Where("id = ? AND deleted_at IS NULL", heroID).
		First(&hero).Error
	
	if err != nil {
		http.Error(w, "Hero not found", http.StatusNotFound)
		return
	}

	// Generate image from third-party API
	imageData, err := s.getHeroImage(hero)
	if err != nil {
		logger.Sugar().Errorf("Failed to get hero image: %v", err)
		// Return a default placeholder image on error
		s.serveDefaultImage(w)
		return
	}

	// Set proper headers for PNG image
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	w.Header().Set("ETag", generateETag(hero))
	
	// Check if client has cached version
	if match := r.Header.Get("If-None-Match"); match == generateETag(hero) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// Write image data
	w.Write(imageData)
}

func (s *Server) getHeroImage(hero database.Hero) ([]byte, error) {
	// Using RoboHash as an example third-party API for generating unique avatars
	// You can replace this with any other image generation API
	
	// Create a unique seed based on hero properties
	seed := fmt.Sprintf("%s-%s-%s-%d", hero.ID, hero.Title, hero.Country, hero.Elo)
	
	// Generate hash for consistent image generation
	hasher := md5.New()
	hasher.Write([]byte(seed))
	hash := hex.EncodeToString(hasher.Sum(nil))
	
	// Construct API URL - using RoboHash as example
	// RoboHash generates unique robot avatars based on any text input
	apiURL := fmt.Sprintf("https://robohash.org/%s.png?size=256x256&set=set1", hash)
	
	// For production, you might want to use a different API like:
	// - DiceBear Avatars: https://avatars.dicebear.com/
	// - Adorable Avatars: http://avatars.adorable.io/
	// - UI Avatars: https://ui-avatars.com/
	// - Or a custom AI image generation service
	
	// Make HTTP request to third-party API
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image from API: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	
	// Read the image data
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}
	
	return imageData, nil
}

func (s *Server) serveDefaultImage(w http.ResponseWriter) {
	// Serve a simple 1x1 transparent PNG as fallback
	// In production, you'd want to serve a proper placeholder image
	transparentPNG := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x62, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}
	
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(transparentPNG)
}

// HandleFightImage handles /api/hero/:id/fight/:fightId/image
func (s *Server) HandleFightImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract hero ID and fight ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/hero/")
	segments := strings.Split(path, "/")
	
	if len(segments) < 4 || segments[1] != "fight" || segments[3] != "image" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	heroID, err := uuid.Parse(segments[0])
	if err != nil {
		http.Error(w, "Invalid hero ID", http.StatusBadRequest)
		return
	}

	fightID, err := uuid.Parse(segments[2])
	if err != nil {
		http.Error(w, "Invalid fight ID", http.StatusBadRequest)
		return
	}

	// Parse multipart form to get the uploaded image
	err = r.ParseMultipartForm(32 << 20) // 32 MB limit
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "No image file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	contentType := fileHeader.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		http.Error(w, "File must be an image", http.StatusBadRequest)
		return
	}

	// Read the uploaded image data
	uploadedImageData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read uploaded image", http.StatusBadRequest)
		return
	}

	// Get fight details from database
	var fight database.Fight
	err = s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).
		Where("id = ? AND (attacker_id = ? OR defender_id = ?)", fightID, heroID, heroID).
		Preload("Attacker").
		Preload("Defender").
		First(&fight).Error
	
	if err != nil {
		http.Error(w, "Fight not found", http.StatusNotFound)
		return
	}

	// Generate fight result image from third-party API
	imageData, err := s.getFightImage(fight, uploadedImageData)
	if err != nil {
		logger.Sugar().Errorf("Failed to get fight image: %v", err)
		// Return a default placeholder image on error
		s.serveFightDefaultImage(w)
		return
	}

	// Set proper headers for PNG image
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable") // Cache for 1 year (infinite)
	w.Header().Set("ETag", generateFightETag(fight))
	
	// Write image data
	w.Write(imageData)
}

func (s *Server) getFightImage(fight database.Fight, uploadedImageData []byte) ([]byte, error) {
	// Create a unique seed based on fight properties
	seed := fmt.Sprintf("%s-%s-%s-%d-%s", 
		fight.ID, 
		fight.Attacker.Title, 
		fight.Defender.Title, 
		fight.Outcome,
		fight.Transcript[:min(50, len(fight.Transcript))]) // First 50 chars of transcript
	
	// Generate hash for consistent image generation
	hasher := md5.New()
	hasher.Write([]byte(seed))
	hasher.Write(uploadedImageData) // Include uploaded image data in hash
	hash := hex.EncodeToString(hasher.Sum(nil))
	
	// Using RoboHash as example - you could use different sets based on fight outcome
	var setParam string
	switch fight.Outcome {
	case database.FightOutcome_Victory:
		setParam = "set1" // Robots for victory
	case database.FightOutcome_Defeat:
		setParam = "set2" // Monsters for defeat  
	case database.FightOutcome_Draw:
		setParam = "set3" // Kittens for draw (because why not)
	default:
		setParam = "set1"
	}
	
	// Construct API URL with fight-specific parameters
	apiURL := fmt.Sprintf("https://robohash.org/%s.png?size=512x512&set=%s", hash, setParam)
	
	// Make HTTP request to third-party API
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch fight image from API: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fight image API returned status %d", resp.StatusCode)
	}
	
	// Read the image data
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read fight image data: %w", err)
	}
	
	return imageData, nil
}

func (s *Server) serveFightDefaultImage(w http.ResponseWriter) {
	// Serve a simple 1x1 transparent PNG as fallback for fight images
	transparentPNG := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x62, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}
	
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(transparentPNG)
}

func generateETag(hero database.Hero) string {
	// Generate ETag based on hero properties that affect the image
	data := fmt.Sprintf("%s-%s-%s-%d", 
		hero.ID, hero.Title, hero.Country, hero.Elo)
	hasher := md5.New()
	hasher.Write([]byte(data))
	return fmt.Sprintf(`"%s"`, hex.EncodeToString(hasher.Sum(nil)))
}

func generateFightETag(fight database.Fight) string {
	// Generate ETag based on fight properties that affect the image
	data := fmt.Sprintf("%s-%s-%s-%d-%s", 
		fight.ID, fight.Attacker.Title, fight.Defender.Title, fight.Outcome, fight.Transcript)
	hasher := md5.New()
	hasher.Write([]byte(data))
	return fmt.Sprintf(`"%s"`, hex.EncodeToString(hasher.Sum(nil)))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}