package database

import (
	"time"

	"github.com/google/uuid"
)

type Player struct {
	ID             uuid.UUID `gorm:"primarykey"`
	UserName       string    `gorm:"index:uq_player_username,unique;not null"`
	UserNameSuffix uint32    `gorm:"index:uq_player_username,unique;not null"`
	Heros          []*Hero
	DeletedAt      *time.Time
}

type Hero struct {
	ID          uuid.UUID `gorm:"primarykey"`
	Country     string    `gorm:"index:idx_hero_country;not null"`
	Elo         uint32    `gorm:"index:idx_hero_elo;not null"`
	Title       string    `gorm:"not null"`
	Description string    `gorm:"not null"`
	PlayerID    uuid.UUID `gorm:"not null"`
	Player      *Player   `gorm:"foreignKey:PlayerID"`
	DeletedAt   *time.Time
}

type Fight struct {
	ID         uuid.UUID    `gorm:"primarykey"`
	AttackerID uuid.UUID    `gorm:"not null"`
	Attacker   *Hero        `gorm:"foreignKey:AttackerID"`
	DefenderID uuid.UUID    `gorm:"not null"`
	Defender   *Hero        `gorm:"foreignKey:DefenderID"`
	TimeStamp  time.Time    `gorm:"not null"`
	Outcome    FightOutcome `gorm:"not null"`
}

func (f Fight) OutcomeAttacker() FightOutcome {
	return f.Outcome
}

func (f Fight) OutcomeDefender() FightOutcome {
	return f.Outcome.Invert()
}
