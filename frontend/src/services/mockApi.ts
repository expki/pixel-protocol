import type { Hero, Player, FightResult, FightsResponse, Fight } from '../types/api';

// Mock data for development
const mockPlayers: Player[] = [
  {
    ID: '1',
    UserName: 'Player',
    UserNameSuffix: 1234,
    Secret: 'mock-secret-123',
    DeletedAt: null
  }
];

const mockHeroes: Hero[] = [
  {
    ID: '1',
    Country: 'USA',
    Elo: 1000,
    Title: 'Dragon Slayer',
    Description: 'A fierce warrior who has conquered the ancient dragons of the northern mountains. Armed with enchanted blade and unbreakable will.',
    PlayerID: '1',
    Player: mockPlayers[0]!,
    DeletedAt: null
  },
  {
    ID: '2',
    Country: 'Japan',
    Elo: 1150,
    Title: 'Shadow Ninja',
    Description: 'Master of stealth and ancient martial arts. Can move through darkness like a whisper and strike with lightning precision.',
    PlayerID: '1',
    Player: mockPlayers[0]!,
    DeletedAt: null
  }
];

const generateMockHero = (title: string, description: string): Hero => {
  const id = String(mockHeroes.length + 1);
  const countries = ['USA', 'Japan', 'Germany', 'France', 'Canada', 'Australia', 'Brazil', 'India'];
  const country = countries[Math.floor(Math.random() * countries.length)];
  
  const hero: Hero = {
    ID: id,
    Country: country,
    Elo: 1000 + Math.floor(Math.random() * 200),
    Title: title,
    Description: description,
    PlayerID: '1',
    Player: mockPlayers[0]!,
    DeletedAt: null
  };
  
  mockHeroes.push(hero);
  return hero;
};

const generateMockFight = (attackerHero: Hero): FightResult => {
  // Create a random opponent
  const opponents = mockHeroes.filter(h => h.ID !== attackerHero.ID);
  const defender = opponents[Math.floor(Math.random() * opponents.length)] || mockHeroes[0];
  
  // Determine outcome (slightly favor attacker for engagement)
  const outcomes: (0 | 1 | 2)[] = [1, 1, 1, 2, 2, 0]; // 50% win, 33% loss, 17% draw
  const outcome = outcomes[Math.floor(Math.random() * outcomes.length)];
  
  let eloGain = 0;
  switch (outcome) {
    case 1: // Victory
      eloGain = 15 + Math.floor(Math.random() * 20);
      break;
    case 2: // Defeat
      eloGain = -(10 + Math.floor(Math.random() * 15));
      break;
    case 0: // Draw
      eloGain = Math.floor(Math.random() * 10) - 5;
      break;
  }
  
  // Update hero ELO
  const heroIndex = mockHeroes.findIndex(h => h.ID === attackerHero.ID);
  if (heroIndex !== -1) {
    mockHeroes[heroIndex].Elo += eloGain;
  }
  
  const battleStories = [
    `The battle began at dawn in the mystic arena. ${attackerHero.Title} faced off against ${defender.Title} in an epic confrontation.\n\nSwords clashed and magic sparked as both warriors displayed incredible skill. The crowd watched in awe as the fighters demonstrated their legendary abilities.\n\nAfter an intense struggle, the battle reached its dramatic conclusion!`,
    
    `In the ancient colosseum, ${attackerHero.Title} challenged ${defender.Title} to single combat.\n\nThe fight was fierce - ${attackerHero.Title}'s ${attackerHero.Description.split('.')[0].toLowerCase()} met ${defender.Title}'s prowess in a display of incredible martial skill.\n\nSpells and steel rang through the air as the warriors pushed their limits. The outcome was truly spectacular!`,
    
    `The arena trembled as ${attackerHero.Title} and ${defender.Title} began their legendary duel.\n\nEach fighter brought their unique abilities to bear - this was a clash of titans that would be remembered for ages. The battle raged with incredible intensity.\n\nWhen the dust settled, the crowd erupted in cheers for this amazing display of combat!`
  ];
  
  const fight: Fight = {
    ID: String(Date.now()),
    AttackerID: attackerHero.ID,
    Attacker: attackerHero,
    DefenderID: defender.ID,
    Defender: defender,
    Timestamp: new Date().toISOString(),
    Outcome: outcome,
    Transcript: battleStories[Math.floor(Math.random() * battleStories.length)]
  };
  
  return {
    fight,
    victory: outcome === 1,
    elo_gain: eloGain
  };
};

export class MockApiService {
  private async delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
  
  private getHeaders(_includeSecret = true): HeadersInit {
    return {
      'Content-Type': 'application/json',
    };
  }
  
  private getCookieSecret(): string | null {
    const cookies = document.cookie.split(';');
    const playerSecretCookie = cookies.find(cookie => 
      cookie.trim().startsWith('player_secret=')
    );
    
    if (playerSecretCookie) {
      return playerSecretCookie.split('=')[1] ?? null;
    }
    
    return null;
  }
  
  private setCookie(name: string, value: string): void {
    document.cookie = `${name}=${value}; path=/; max-age=31536000`; // 1 year
  }

  async createPlayer(username: string): Promise<Player> {
    await this.delay(500); // Simulate network delay
    
    const player: Player = {
      ID: '1',
      UserName: username,
      UserNameSuffix: Math.floor(Math.random() * 9999) + 1,
      Secret: `mock-secret-${Date.now()}`,
      DeletedAt: null
    };
    
    // Set cookie to simulate authentication
    this.setCookie('player_secret', player.Secret);
    
    // Update mock data
    mockPlayers[0] = player;
    
    return player;
  }

  async getPlayer(id: string): Promise<Player> {
    await this.delay(300);
    const player = mockPlayers.find(p => p.ID === id);
    if (!player) {
      throw new Error('Player not found');
    }
    return player;
  }

  async createHero(heroData: { title: string; description: string }): Promise<Hero> {
    await this.delay(800); // Simulate AI processing time
    return generateMockHero(heroData.title, heroData.description);
  }

  async getHero(id: string): Promise<Hero> {
    await this.delay(300);
    const hero = mockHeroes.find(h => h.ID === id);
    if (!hero) {
      throw new Error('Hero not found');
    }
    return hero;
  }

  async getHeroImage(id: string): Promise<string> {
    await this.delay(600);
    
    // Generate a colorful placeholder image using a data URL
    const canvas = document.createElement('canvas');
    canvas.width = 300;
    canvas.height = 300;
    const ctx = canvas.getContext('2d');
    
    if (!ctx) {
      throw new Error('Could not create canvas context');
    }
    
    // Create a gradient background
    const gradient = ctx.createLinearGradient(0, 0, 300, 300);
    const colors = ['#FF6B6B', '#4ECDC4', '#45B7D1', '#FFA07A', '#98D8C8'];
    const heroIndex = mockHeroes.findIndex(h => h.ID === id);
    const baseColor = colors[heroIndex % colors.length];
    
    gradient.addColorStop(0, baseColor);
    gradient.addColorStop(1, `${baseColor}88`);
    
    ctx.fillStyle = gradient;
    ctx.fillRect(0, 0, 300, 300);
    
    // Add some geometric shapes for character representation
    ctx.fillStyle = '#FFFFFF';
    ctx.fillRect(120, 100, 60, 80); // Body
    ctx.fillRect(135, 50, 30, 50);  // Head
    ctx.fillRect(110, 120, 20, 60); // Left arm
    ctx.fillRect(170, 120, 20, 60); // Right arm
    ctx.fillRect(130, 180, 15, 80); // Left leg
    ctx.fillRect(155, 180, 15, 80); // Right leg
    
    // Add title text
    ctx.fillStyle = '#FFFFFF';
    ctx.font = '16px Arial';
    ctx.textAlign = 'center';
    const hero = mockHeroes.find(h => h.ID === id);
    if (hero) {
      ctx.fillText(hero.Title, 150, 280);
    }
    
    return canvas.toDataURL('image/png');
  }

  async startFight(heroId: string): Promise<FightResult> {
    await this.delay(1500); // Simulate AI battle processing
    
    const hero = mockHeroes.find(h => h.ID === heroId);
    if (!hero) {
      throw new Error('Hero not found');
    }
    
    return generateMockFight(hero);
  }

  async getHeroFights(heroId: string, lastId?: string, limit = 20): Promise<FightsResponse> {
    await this.delay(400);
    
    // For mock, just return empty for now
    return {
      fights: [],
      has_more: false,
      next_cursor: undefined
    };
  }

  async getFight(heroId: string, fightId: string): Promise<Fight> {
    await this.delay(300);
    
    // For mock, create a sample fight
    const hero = mockHeroes.find(h => h.ID === heroId);
    if (!hero) {
      throw new Error('Hero not found');
    }
    
    const result = generateMockFight(hero);
    return result.fight;
  }

  async getPlayerHeroes(playerId: string): Promise<Hero[]> {
    await this.delay(400);
    return mockHeroes.filter(h => h.PlayerID === playerId);
  }
}

export const mockApiService = new MockApiService();