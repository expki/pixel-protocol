package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/expki/backend/pixel-protocol/database"
	"github.com/expki/backend/pixel-protocol/logger"
	"github.com/google/uuid"
	"gorm.io/plugin/dbresolver"
)

type PlayerRequest struct {
	UserName string `json:"username"`
}

type PlayerUpdateRequest struct {
	UserName *string `json:"username,omitempty"`
}

func (s *Server) HandlePlayer(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/player")
	segments := strings.Split(strings.Trim(path, "/"), "/")

	switch r.Method {
	case http.MethodGet:
		if len(segments) == 1 && segments[0] != "" {
			s.getPlayer(w, r, segments[0])
		} else {
			http.Error(w, "Player ID required", http.StatusBadRequest)
		}
	case http.MethodPost:
		s.createPlayer(w, r)
	case http.MethodPut:
		if len(segments) == 1 && segments[0] != "" {
			s.updatePlayer(w, r, segments[0])
		} else {
			http.Error(w, "Player ID required", http.StatusBadRequest)
		}
	case http.MethodPatch:
		if len(segments) == 1 && segments[0] != "" {
			s.patchPlayer(w, r, segments[0])
		} else {
			http.Error(w, "Player ID required", http.StatusBadRequest)
		}
	case http.MethodDelete:
		if len(segments) == 1 && segments[0] != "" {
			s.deletePlayer(w, r, segments[0])
		} else {
			http.Error(w, "Player ID required", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getPlayer(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}

	var player database.Player
	result := s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).Where("id = ? AND deleted_at IS NULL", id).First(&player)
	if result.Error != nil {
		if result.RowsAffected == 0 {
			http.Error(w, "Player not found", http.StatusNotFound)
		} else {
			logger.Sugar().Errorf("Failed to get player: %v", result.Error)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}

func (s *Server) createPlayer(w http.ResponseWriter, r *http.Request) {
	var req PlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserName == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// Generate a unique suffix for the username
	suffix := s.generateUserNameSuffix(r.Context(), req.UserName)

	player := database.Player{
		ID:             uuid.New(),
		UserName:       req.UserName,
		UserNameSuffix: suffix,
	}

	result := s.db.DB.Clauses(dbresolver.Write).WithContext(r.Context()).Create(&player)
	if result.Error != nil {
		logger.Sugar().Errorf("Failed to create player: %v", result.Error)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(player)
}

func (s *Server) generateUserNameSuffix(ctx context.Context, username string) uint32 {
	// Get all existing suffixes for this username, sorted ascending
	var existingSuffixes []uint32
	s.db.DB.Clauses(dbresolver.Read).WithContext(ctx).Model(&database.Player{}).
		Where("user_name = ?", username).
		Order("user_name_suffix ASC").
		Pluck("user_name_suffix", &existingSuffixes)

	// If no existing suffixes, start with 1
	if len(existingSuffixes) == 0 {
		return 1
	}

	// Since the list is sorted, we can efficiently find the first gap
	// by checking if each position matches its expected value
	for i, suffix := range existingSuffixes {
		expectedSuffix := uint32(i + 1)
		if suffix != expectedSuffix {
			// Found a gap, return the expected suffix
			return expectedSuffix
		}
	}

	// If no gaps found (all positions match 1,2,3,...n),
	// return the next number after the last suffix
	return uint32(len(existingSuffixes) + 1)
}

func (s *Server) updatePlayer(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}

	var req PlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserName == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	var player database.Player
	result := s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).Where("id = ? AND deleted_at IS NULL", id).First(&player)
	if result.Error != nil {
		if result.RowsAffected == 0 {
			http.Error(w, "Player not found", http.StatusNotFound)
		} else {
			logger.Sugar().Errorf("Failed to find player: %v", result.Error)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// If username is changing, generate new suffix
	if player.UserName != req.UserName {
		player.UserName = req.UserName
		player.UserNameSuffix = s.generateUserNameSuffix(r.Context(), req.UserName)
	}

	result = s.db.DB.Clauses(dbresolver.Write).WithContext(r.Context()).Save(&player)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate") || strings.Contains(result.Error.Error(), "unique") {
			// Try with a different suffix
			for attempts := 0; attempts < 10; attempts++ {
				player.UserNameSuffix = s.generateUserNameSuffix(r.Context(), req.UserName)
				result = s.db.DB.Clauses(dbresolver.Write).WithContext(r.Context()).Save(&player)
				if result.Error == nil {
					break
				}
			}
			if result.Error != nil {
				http.Error(w, "Unable to update with unique username", http.StatusConflict)
				return
			}
		} else {
			logger.Sugar().Errorf("Failed to update player: %v", result.Error)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}

func (s *Server) patchPlayer(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}

	var req PlayerUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var player database.Player
	result := s.db.DB.Clauses(dbresolver.Read).WithContext(r.Context()).Where("id = ? AND deleted_at IS NULL", id).First(&player)
	if result.Error != nil {
		if result.RowsAffected == 0 {
			http.Error(w, "Player not found", http.StatusNotFound)
		} else {
			logger.Sugar().Errorf("Failed to find player: %v", result.Error)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if req.UserName != nil && *req.UserName != player.UserName {
		player.UserName = *req.UserName
		// Generate new suffix when username changes
		player.UserNameSuffix = s.generateUserNameSuffix(r.Context(), *req.UserName)
	}

	result = s.db.DB.Clauses(dbresolver.Write).WithContext(r.Context()).Save(&player)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate") || strings.Contains(result.Error.Error(), "unique") {
			// Try with a different suffix
			for attempts := 0; attempts < 10; attempts++ {
				player.UserNameSuffix = s.generateUserNameSuffix(r.Context(), player.UserName)
				result = s.db.DB.Clauses(dbresolver.Write).WithContext(r.Context()).Save(&player)
				if result.Error == nil {
					break
				}
			}
			if result.Error != nil {
				http.Error(w, "Unable to update with unique username", http.StatusConflict)
				return
			}
		} else {
			logger.Sugar().Errorf("Failed to patch player: %v", result.Error)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}

func (s *Server) deletePlayer(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}

	// Soft delete by setting DeletedAt timestamp
	now := time.Now()
	result := s.db.DB.Clauses(dbresolver.Write).WithContext(r.Context()).Model(&database.Player{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now)

	if result.Error != nil {
		logger.Sugar().Errorf("Failed to delete player: %v", result.Error)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected == 0 {
		http.Error(w, "Player not found or already deleted", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
