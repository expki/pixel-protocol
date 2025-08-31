package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/expki/backend/pixel-protocol/database"
	"github.com/expki/backend/pixel-protocol/logger"
	"github.com/google/uuid"
	"gorm.io/plugin/dbresolver"
)

type FightsResponse struct {
	Fights     []database.Fight `json:"fights"`
	HasMore    bool             `json:"has_more"`
	NextCursor string           `json:"next_cursor,omitempty"`
}

// HandlePlayerFights handles /api/player/:id/fight and /api/player/:id/fight/:fightId
func (s *Server) HandlePlayerFights(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/player/")
	segments := strings.Split(path, "/")

	if len(segments) < 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	playerID, err := uuid.Parse(segments[0])
	if err != nil {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}

	// Check if it's for fights list or specific fight
	if segments[1] == "fights" {
		s.getPlayerFights(w, r, playerID)
	} else if segments[1] == "fight" && len(segments) > 2 {
		fightID, err := uuid.Parse(segments[2])
		if err != nil {
			http.Error(w, "Invalid fight ID", http.StatusBadRequest)
			return
		}
		s.getPlayerFight(w, r, playerID, fightID)
	} else {
		http.Error(w, "Invalid path", http.StatusBadRequest)
	}
}

// HandleHeroFights handles /api/hero/:id/fights and /api/hero/:id/fight/:fightId
func (s *Server) HandleHeroFights(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/hero/")
	segments := strings.Split(path, "/")

	if len(segments) < 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	heroID, err := uuid.Parse(segments[0])
	if err != nil {
		http.Error(w, "Invalid hero ID", http.StatusBadRequest)
		return
	}

	// Check if it's for fights list or specific fight
	if segments[1] == "fights" {
		s.getHeroFights(w, r, heroID)
	} else if segments[1] == "fight" && len(segments) > 2 {
		fightID, err := uuid.Parse(segments[2])
		if err != nil {
			http.Error(w, "Invalid fight ID", http.StatusBadRequest)
			return
		}
		s.getHeroFight(w, r, heroID, fightID)
	} else {
		http.Error(w, "Invalid path", http.StatusBadRequest)
	}
}

func (s *Server) getPlayerFights(w http.ResponseWriter, r *http.Request, playerID uuid.UUID) {
	// Parse query parameters
	lastIDStr := r.URL.Query().Get("last_id")

	// First verify the player exists
	var player database.Player
	if err := s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).
		Where("id = ? AND deleted_at IS NULL", playerID).
		First(&player).Error; err != nil {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}

	// Get all hero IDs for this player
	var heroIDs []uuid.UUID
	s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).
		Model(&database.Hero{}).
		Where("player_id = ? AND deleted_at IS NULL", playerID).
		Pluck("id", &heroIDs)

	if len(heroIDs) == 0 {
		// No heroes, return empty fights
		response := FightsResponse{
			Fights:  []database.Fight{},
			HasMore: false,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Build the query for fights
	query := s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).
		Model(&database.Fight{}).
		Where("(attacker_id IN ? OR defender_id IN ?)", heroIDs, heroIDs).
		Order("timestamp DESC, id DESC").
		Preload("Attacker").
		Preload("Defender").
		Limit(20 + 1) // Get one extra to check if there are more

	// If last_id is provided, use it for cursor-based pagination
	if lastIDStr != "" {
		lastID, err := uuid.Parse(lastIDStr)
		if err != nil {
			http.Error(w, "Invalid last_id", http.StatusBadRequest)
			return
		}

		// Get the timestamp of the last fight to properly continue pagination
		var lastFight database.Fight
		if err := s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).
			Where("id = ?", lastID).
			First(&lastFight).Error; err == nil {
			// Continue from where we left off
			query = query.Where("(timestamp < ? OR (timestamp = ? AND id < ?))",
				lastFight.Timestamp, lastFight.Timestamp, lastID)
		}
	}

	var fights []database.Fight
	if err := query.Find(&fights).Error; err != nil {
		logger.Sugar().Errorf("Failed to get fights: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if there are more results
	hasMore := len(fights) > 20
	if hasMore {
		fights = fights[:20] // Remove the extra one
	}

	response := FightsResponse{
		Fights:  fights,
		HasMore: hasMore,
	}

	// Add next cursor if there are more results
	if hasMore && len(fights) > 0 {
		response.NextCursor = fights[len(fights)-1].ID.String()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) getHeroFights(w http.ResponseWriter, r *http.Request, heroID uuid.UUID) {
	// Parse query parameters
	lastIDStr := r.URL.Query().Get("last_id")

	// First verify the hero exists
	var hero database.Hero
	if err := s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).
		Where("id = ? AND deleted_at IS NULL", heroID).
		First(&hero).Error; err != nil {
		http.Error(w, "Hero not found", http.StatusNotFound)
		return
	}

	// Build the query for fights
	query := s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).
		Model(&database.Fight{}).
		Where("attacker_id = ? OR defender_id = ?", heroID, heroID).
		Order("timestamp DESC, id DESC").
		Preload("Attacker").
		Preload("Defender").
		Limit(20 + 1) // Get one extra to check if there are more

	// If last_id is provided, use it for cursor-based pagination
	if lastIDStr != "" {
		lastID, err := uuid.Parse(lastIDStr)
		if err != nil {
			http.Error(w, "Invalid last_id", http.StatusBadRequest)
			return
		}

		// Get the timestamp of the last fight to properly continue pagination
		var lastFight database.Fight
		if err := s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).
			Where("id = ?", lastID).
			First(&lastFight).Error; err == nil {
			// Continue from where we left off
			query = query.Where("(timestamp < ? OR (timestamp = ? AND id < ?))",
				lastFight.Timestamp, lastFight.Timestamp, lastID)
		}
	}

	var fights []database.Fight
	if err := query.Find(&fights).Error; err != nil {
		logger.Sugar().Errorf("Failed to get fights: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if there are more results
	hasMore := len(fights) > 20
	if hasMore {
		fights = fights[:20] // Remove the extra one
	}

	response := FightsResponse{
		Fights:  fights,
		HasMore: hasMore,
	}

	// Add next cursor if there are more results
	if hasMore && len(fights) > 0 {
		response.NextCursor = fights[len(fights)-1].ID.String()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) getPlayerFight(w http.ResponseWriter, r *http.Request, playerID, fightID uuid.UUID) {
	// First verify the player exists
	var player database.Player
	if err := s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).
		Where("id = ? AND deleted_at IS NULL", playerID).
		First(&player).Error; err != nil {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}

	// Get all hero IDs for this player
	var heroIDs []uuid.UUID
	s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).
		Model(&database.Hero{}).
		Where("player_id = ? AND deleted_at IS NULL", playerID).
		Pluck("id", &heroIDs)

	if len(heroIDs) == 0 {
		http.Error(w, "Fight not found", http.StatusNotFound)
		return
	}

	// Get the specific fight
	var fight database.Fight
	result := s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).
		Where("id = ? AND (attacker_id IN ? OR defender_id IN ?)", fightID, heroIDs, heroIDs).
		Preload("Attacker").
		Preload("Defender").
		First(&fight)

	if result.Error != nil {
		if result.RowsAffected == 0 {
			http.Error(w, "Fight not found", http.StatusNotFound)
		} else {
			logger.Sugar().Errorf("Failed to get fight: %v", result.Error)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fight)
}

func (s *Server) getHeroFight(w http.ResponseWriter, r *http.Request, heroID, fightID uuid.UUID) {
	// First verify the hero exists
	var hero database.Hero
	if err := s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).
		Where("id = ? AND deleted_at IS NULL", heroID).
		First(&hero).Error; err != nil {
		http.Error(w, "Hero not found", http.StatusNotFound)
		return
	}

	// Get the specific fight
	var fight database.Fight
	result := s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).
		Where("id = ? AND (attacker_id = ? OR defender_id = ?)", fightID, heroID, heroID).
		Preload("Attacker").
		Preload("Defender").
		First(&fight)

	if result.Error != nil {
		if result.RowsAffected == 0 {
			http.Error(w, "Fight not found", http.StatusNotFound)
		} else {
			logger.Sugar().Errorf("Failed to get fight: %v", result.Error)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fight)
}
