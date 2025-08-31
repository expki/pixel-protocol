package server

import "github.com/expki/backend/pixel-protocol/database"

type Server struct {
	db *database.Database
}

func New(db *database.Database) *Server {
	return &Server{
		db: db,
	}
}
