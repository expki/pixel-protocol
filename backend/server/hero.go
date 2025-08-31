package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/expki/backend/pixel-protocol/database"
	"github.com/expki/backend/pixel-protocol/geolookup"
	"github.com/expki/backend/pixel-protocol/logger"
	"github.com/google/uuid"
	"gorm.io/plugin/dbresolver"
)

type HeroRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type HeroUpdateRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
}

func (s *Server) HandleHero(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/hero")
	segments := strings.Split(strings.Trim(path, "/"), "/")

	var secretStruct PlayerSecret
	err := json.NewDecoder(r.Body).Decode(&secretStruct)
	if err != nil {
		http.Error(w, "Missing player _secret", http.StatusBadRequest)
		return
	}
	if secretStruct.Secret == "" {
		http.Error(w, "Secret is required", http.StatusBadRequest)
		return
	}
	secret, err := uuid.Parse(secretStruct.Secret)
	if err != nil {
		http.Error(w, "Invalid player _secret", http.StatusBadRequest)
		return
	}
	var player database.Player
	result := s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).Where("secret = ? AND deleted_at IS NULL", secret).First(&player)
	if result.Error != nil {
		if result.RowsAffected == 0 {
			http.Error(w, "Player not found", http.StatusNotFound)
		} else {
			logger.Sugar().Errorf("Failed to get player: %v", result.Error)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if r.Method == http.MethodPost {
		s.createHero(w, r, player)
		return
	}
	id, err := uuid.Parse(segments[0])
	if err != nil {
		http.Error(w, "Invalid Hero ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getHero(w, r, player, id)
	case http.MethodPut:
		s.updateHero(w, r, player, id)
	case http.MethodPatch:
		s.patchHero(w, r, player, id)
	case http.MethodDelete:
		s.deleteHero(w, r, player, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getHero(w http.ResponseWriter, r *http.Request, player database.Player, id uuid.UUID) {

	var hero database.Hero
	result := s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).
		Where("player_id = ? AND id = ? AND deleted_at IS NULL", player.ID, id).
		First(&hero)
	if result.Error != nil {
		if result.RowsAffected == 0 {
			http.Error(w, "Hero not found", http.StatusNotFound)
		} else {
			logger.Sugar().Errorf("Failed to get hero: %v", result.Error)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hero)
}

func (s *Server) createHero(w http.ResponseWriter, r *http.Request, player database.Player) {
	var req HeroRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Title == "" || req.Description == "" {
		http.Error(w, "Title and Description are required", http.StatusBadRequest)
		return
	}

	hero := database.Hero{
		ID:          uuid.New(),
		Country:     geolookup.GetClientCountry(r),
		Elo:         1000, // Starting Elo rating
		Title:       req.Title,
		Description: req.Description,
		PlayerID:    player.ID,
		Player:      &player,
	}

	result := s.db.DB.Clauses(dbresolver.Write).WithContext(r.Context()).Create(&hero)
	if result.Error != nil {
		logger.Sugar().Errorf("Failed to create hero: %v", result.Error)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(hero)
}

func (s *Server) updateHero(w http.ResponseWriter, r *http.Request, player database.Player, id uuid.UUID) {

	var req HeroRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Title == "" || req.Description == "" {
		http.Error(w, "Country, title, and description are required", http.StatusBadRequest)
		return
	}

	// Get the hero first to authenticate
	var hero database.Hero
	result := s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).
		Where("player_id = ? AND id = ? AND deleted_at IS NULL", player.ID, id).First(&hero)
	if result.Error != nil {
		if result.RowsAffected == 0 {
			http.Error(w, "Hero not found", http.StatusNotFound)
		} else {
			logger.Sugar().Errorf("Failed to find hero: %v", result.Error)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Update fields
	hero.Country = geolookup.GetClientCountry(r)
	hero.Title = req.Title
	hero.Description = req.Description

	result = s.db.DB.Clauses(dbresolver.Write).WithContext(r.Context()).Save(&hero)
	if result.Error != nil {
		logger.Sugar().Errorf("Failed to update hero: %v", result.Error)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hero)
}

func (s *Server) patchHero(w http.ResponseWriter, r *http.Request, player database.Player, id uuid.UUID) {

	var req HeroUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get the hero first to authenticate
	var hero database.Hero
	result := s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).
		Where("player_id =? AND id = ? AND deleted_at IS NULL", player.ID, id).First(&hero)
	if result.Error != nil {
		if result.RowsAffected == 0 {
			http.Error(w, "Hero not found", http.StatusNotFound)
		} else {
			logger.Sugar().Errorf("Failed to find hero: %v", result.Error)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Update only provided fields
	hero.Country = geolookup.GetClientCountry(r)
	if req.Title != nil {
		hero.Title = *req.Title
	}
	if req.Description != nil {
		hero.Description = *req.Description
	}

	result = s.db.DB.Clauses(dbresolver.Write).WithContext(r.Context()).Save(&hero)
	if result.Error != nil {
		logger.Sugar().Errorf("Failed to patch hero: %v", result.Error)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hero)
}

func (s *Server) deleteHero(w http.ResponseWriter, r *http.Request, player database.Player, id uuid.UUID) {

	// Soft delete by setting DeletedAt timestamp
	now := time.Now()
	result := s.db.DB.Clauses(dbresolver.Write).WithContext(r.Context()).
		Model(&database.Hero{}).
		Where("player_id = ? AND id = ? AND deleted_at IS NULL", player.ID, id).
		Update("deleted_at", now)

	if result.Error != nil {
		logger.Sugar().Errorf("Failed to delete hero: %v", result.Error)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected == 0 {
		http.Error(w, "Hero not found or already deleted", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
