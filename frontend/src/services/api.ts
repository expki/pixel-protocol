import type { Hero, Player, FightResult, FightsResponse, Fight } from '../types/api';

const API_BASE = '/api';

interface CreateHeroRequest {
  title: string;
  description: string;
}

interface CreatePlayerRequest {
  username: string;
}

class ApiService {
  private getHeaders(includeSecret = true): HeadersInit {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    };
    
    return headers;
  }

  private async getCookieSecret(): Promise<string | null> {
    const cookies = document.cookie.split(';');
    const playerSecretCookie = cookies.find(cookie => 
      cookie.trim().startsWith('player_secret=')
    );
    
    if (playerSecretCookie) {
      return playerSecretCookie.split('=')[1];
    }
    
    return null;
  }

  private async makeAuthenticatedRequest(
    url: string, 
    options: RequestInit = {},
    requireAuth = true
  ): Promise<Response> {
    const secret = await this.getCookieSecret();
    
    if (requireAuth && !secret) {
      throw new Error('Authentication required - no player_secret cookie found');
    }

    const body = options.body ? JSON.parse(options.body as string) : {};
    
    if (secret) {
      body._secret = secret;
    }

    return fetch(url, {
      ...options,
      headers: {
        ...this.getHeaders(),
        ...options.headers,
      },
      body: Object.keys(body).length > 0 ? JSON.stringify(body) : options.body,
      credentials: 'include',
    });
  }

  // Player endpoints
  async createPlayer(username: string): Promise<Player> {
    const response = await fetch(`${API_BASE}/player`, {
      method: 'POST',
      headers: this.getHeaders(false),
      body: JSON.stringify({ username }),
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to create player: ${response.statusText}`);
    }

    return response.json();
  }

  async getPlayer(id: string): Promise<Player> {
    const response = await this.makeAuthenticatedRequest(`${API_BASE}/player/${id}`, {
      method: 'GET',
    });

    if (!response.ok) {
      throw new Error(`Failed to get player: ${response.statusText}`);
    }

    return response.json();
  }

  // Hero endpoints
  async createHero(heroData: CreateHeroRequest): Promise<Hero> {
    const response = await fetch(`${API_BASE}/hero`, {
      method: 'POST',
      headers: this.getHeaders(false),
      body: JSON.stringify(heroData),
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to create hero: ${response.statusText}`);
    }

    return response.json();
  }

  async getHero(id: string): Promise<Hero> {
    const response = await this.makeAuthenticatedRequest(`${API_BASE}/hero/${id}`, {
      method: 'GET',
    });

    if (!response.ok) {
      throw new Error(`Failed to get hero: ${response.statusText}`);
    }

    return response.json();
  }

  async getHeroImage(id: string): Promise<string> {
    const response = await fetch(`${API_BASE}/hero/${id}/image`, {
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to get hero image: ${response.statusText}`);
    }

    const blob = await response.blob();
    return URL.createObjectURL(blob);
  }

  // Fight endpoints
  async startFight(heroId: string): Promise<FightResult> {
    const response = await this.makeAuthenticatedRequest(`${API_BASE}/hero/${heroId}/fight`, {
      method: 'POST',
    });

    if (!response.ok) {
      throw new Error(`Failed to start fight: ${response.statusText}`);
    }

    return response.json();
  }

  async getHeroFights(heroId: string, lastId?: string, limit = 20): Promise<FightsResponse> {
    const params = new URLSearchParams();
    if (lastId) params.append('last_id', lastId);
    params.append('limit', limit.toString());

    const response = await fetch(`${API_BASE}/hero/${heroId}/fights?${params}`, {
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to get hero fights: ${response.statusText}`);
    }

    return response.json();
  }

  async getFight(heroId: string, fightId: string): Promise<Fight> {
    const response = await fetch(`${API_BASE}/hero/${heroId}/fight/${fightId}`, {
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to get fight: ${response.statusText}`);
    }

    return response.json();
  }

  // Player heroes endpoint (we'll need to create this in the backend)
  async getPlayerHeroes(playerId: string): Promise<Hero[]> {
    const response = await this.makeAuthenticatedRequest(`${API_BASE}/player/${playerId}/heroes`, {
      method: 'GET',
    });

    if (!response.ok) {
      throw new Error(`Failed to get player heroes: ${response.statusText}`);
    }

    return response.json();
  }
}

export const apiService = new ApiService();