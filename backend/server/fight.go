package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/expki/backend/pixel-protocol/database"
	"github.com/expki/backend/pixel-protocol/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
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

	// Handle different methods and paths
	if r.Method == http.MethodPost && segments[1] == "fight" {
		// POST /api/hero/:id/fight - Create a new fight
		s.createHeroFight(w, r, heroID)
	} else if r.Method == http.MethodGet {
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
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
		Order("timestamp DESC").
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
		Order("timestamp DESC").
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

type FightResult struct {
	Fight   database.Fight `json:"fight"`
	Victory bool           `json:"victory"`
	EloGain int32          `json:"elo_gain"`
}

func (s *Server) createHeroFight(w http.ResponseWriter, r *http.Request, attackerID uuid.UUID) {
	// Parse request body for secret
	var req PlayerSecret
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Secret == "" {
		http.Error(w, "Secret is required", http.StatusBadRequest)
		return
	}

	// Parse secret as UUID
	secret, err := uuid.Parse(req.Secret)
	if err != nil {
		http.Error(w, "Invalid secret", http.StatusBadRequest)
		return
	}

	// Get the attacker hero and verify ownership
	var attacker database.Hero
	err = s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).
		Where("id = ? AND deleted_at IS NULL", attackerID).
		Preload("Player").
		First(&attacker).Error
	if err != nil {
		http.Error(w, "Hero not found", http.StatusNotFound)
		return
	}

	// Verify the player owns this hero via secret
	if attacker.Player == nil || attacker.Player.Secret != secret {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Find a suitable opponent with similar ELO
	defender, err := s.findOpponent(r.Context(), attacker)
	if err != nil {
		logger.Sugar().Errorf("Failed to find opponent: %v", err)
		http.Error(w, "No suitable opponent found", http.StatusNotFound)
		return
	}

	// Simple 50/50 random fight outcome
	// TODO: proper fights
	outcome := database.FightOutcome_Defeat
	if randomBool() {
		outcome = database.FightOutcome_Victory
	}

	// Calculate ELO changes using standard ELO formula
	eloGain := calculateEloChange(attacker.Elo, defender.Elo, outcome == database.FightOutcome_Victory)

	// Create fight record
	fight := database.Fight{
		ID:         uuid.New(),
		AttackerID: attackerID,
		DefenderID: defender.ID,
		Timestamp:  time.Now(),
		Outcome:    outcome,
		Transcript: "", // TODO: create transcript
	}

	// Start transaction to update ELOs and create fight
	tx := s.db.DB.Clauses(dbresolver.Write).WithContext(r.Context()).Begin()

	// Create fight record
	if err := tx.Create(&fight).Error; err != nil {
		tx.Rollback()
		logger.Sugar().Errorf("Failed to create fight: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Update attacker ELO
	newAttackerElo := attacker.Elo
	if outcome == database.FightOutcome_Victory {
		newAttackerElo = uint32(int32(attacker.Elo) + eloGain)
	} else {
		// Ensure ELO doesn't go below 0
		if int32(attacker.Elo) > eloGain {
			newAttackerElo = uint32(int32(attacker.Elo) - eloGain)
		} else {
			newAttackerElo = 0
		}
	}

	if err := tx.Model(&database.Hero{}).Where("id = ?", attackerID).
		Update("elo", newAttackerElo).Error; err != nil {
		tx.Rollback()
		logger.Sugar().Errorf("Failed to update attacker ELO: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Update defender ELO (opposite of attacker)
	newDefenderElo := defender.Elo
	if outcome == database.FightOutcome_Victory {
		// Attacker won, defender loses ELO
		if int32(defender.Elo) > eloGain {
			newDefenderElo = uint32(int32(defender.Elo) - eloGain)
		} else {
			newDefenderElo = 0
		}
	} else {
		// Attacker lost, defender gains ELO
		newDefenderElo = uint32(int32(defender.Elo) + eloGain)
	}

	if err := tx.Model(&database.Hero{}).Where("id = ?", defender.ID).
		Update("elo", newDefenderElo).Error; err != nil {
		tx.Rollback()
		logger.Sugar().Errorf("Failed to update defender ELO: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		logger.Sugar().Errorf("Failed to commit transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Load the complete fight with relationships
	s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).
		Where("id = ?", fight.ID).
		Preload("Attacker").
		Preload("Defender").
		First(&fight)

	// Prepare response
	result := FightResult{
		Fight:   fight,
		Victory: outcome == database.FightOutcome_Victory,
		EloGain: eloGain,
	}
	if outcome == database.FightOutcome_Defeat {
		result.EloGain = -eloGain
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func (s *Server) findOpponent(ctx context.Context, attacker database.Hero) (*database.Hero, error) {
	// Define ELO range for matchmaking (±200 ELO points)
	eloRange := uint32(200)
	minElo := attacker.Elo
	maxElo := attacker.Elo + eloRange

	if attacker.Elo > eloRange {
		minElo = attacker.Elo - eloRange
	} else {
		minElo = 0
	}

	// Find local potential opponent within ELO range
	if randomBool() {
		var opponentLocal database.Hero
		errLocal := s.db.DB.Clauses(dbresolver.Read).WithContext(ctx).
			Where("id != ? AND deleted_at IS NULL AND country = ? AND elo BETWEEN ? AND ?", attacker.ID, attacker.Country, minElo, maxElo).
			Order("RANDOM()"). // Random selection for variety
			Take(&opponentLocal).Error
		if errLocal != nil && !errors.Is(errLocal, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("database error: %w", errLocal)
		} else if errLocal == nil {
			return &opponentLocal, nil
		}
	}

	// Find global potential opponents within ELO range
	var opponentGlobal database.Hero
	errGlobal := s.db.DB.Clauses(dbresolver.Read).WithContext(ctx).
		Where("id != ? AND deleted_at IS NULL AND elo BETWEEN ? AND ?", attacker.ID, minElo, maxElo).
		Order("RANDOM()"). // Random selection for variety
		Take(&opponentGlobal).Error
	if errGlobal != nil && !errors.Is(errGlobal, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("database error: %w", errGlobal)
	} else if errors.Is(errGlobal, gorm.ErrRecordNotFound) {
		errGlobal = s.db.DB.Clauses(dbresolver.Read).WithContext(ctx).
			Where("id != ? AND deleted_at IS NULL", attacker.ID).
			Order("RANDOM()"). // Just pick a random opponent if none in range
			First(&opponentGlobal).Error
	}
	return &opponentGlobal, errGlobal
}

func calculateEloChange(attackerElo, defenderElo uint32, won bool) int32 {
	// K-factor (determines how much ratings can change)
	K := float64(32)

	// Expected score based on ELO difference
	expectedScore := 1 / (1 + math.Pow(10, float64(defenderElo-attackerElo)/400))

	// Actual score (1 for win, 0 for loss)
	actualScore := 0.0
	if won {
		actualScore = 1.0
	}

	// Calculate ELO change
	eloChange := K * (actualScore - expectedScore)

	return int32(math.Round(eloChange))
}

func randomBool() bool {
	nano := time.Now().UnixNano()
	return ((nano>>1)^nano)%2 == 0
}
