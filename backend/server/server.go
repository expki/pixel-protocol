package server

import (
	"github.com/expki/backend/pixel-protocol/claude"
	"github.com/expki/backend/pixel-protocol/database"
)

type Server struct {
	db     *database.Database
	claude *claude.Client
}

func New(db *database.Database, claudeClient *claude.Client) *Server {
	return &Server{
		db:     db,
		claude: claudeClient,
	}
}
