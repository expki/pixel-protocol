export interface Player {
  ID: string;
  UserName: string;
  UserNameSuffix: number;
  Secret: string;
  DeletedAt?: string | null;
}

export interface Hero {
  ID: string;
  Country: string;
  Elo: number;
  Title: string;
  Description: string;
  PlayerID: string;
  Player?: Player;
  DeletedAt?: string | null;
}

export interface Fight {
  ID: string;
  AttackerID: string;
  Attacker: Hero;
  DefenderID: string;
  Defender: Hero;
  Timestamp: string;
  Outcome: 0 | 1 | 2; // 0=Draw, 1=Victory, 2=Defeat
  Transcript: string;
}

export interface FightResult {
  fight: Fight;
  victory: boolean;
  elo_gain: number;
}

export interface FightsResponse {
  fights: Fight[];
  has_more: boolean;
  next_cursor?: string;
}

export interface SecretRequest {
  _secret: string;
}