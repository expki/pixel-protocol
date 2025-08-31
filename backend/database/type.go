package database

type FightOutcome uint8

const (
	FightOutcome_Draw FightOutcome = iota
	FightOutcome_Victory
	FightOutcome_Defeat
)

func (value FightOutcome) Invert() FightOutcome {
	switch value {
	case FightOutcome_Defeat:
		return FightOutcome_Victory
	case FightOutcome_Victory:
		return FightOutcome_Defeat
	default:
		return value
	}
}
