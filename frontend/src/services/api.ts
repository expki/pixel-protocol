import type { Hero, Player, FightResult, FightsResponse, Fight } from '../types/api';
import { mockApiService } from './mockApi';

const API_BASE = 'http://localhost:5080/api';

interface CreateHeroRequest {
  title: string;
  description: string;
}

interface _CreatePlayerRequest {
  username: string;
}

class ApiService {
  private useMockApi = false; // Set to false when backend is available
  private getHeaders(_includeSecret = true): HeadersInit {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    };
    
    return headers;
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

  private async makeAuthenticatedRequest(
    url: string, 
    options: RequestInit = {},
    requireAuth = true
  ): Promise<Response> {
    const secret = this.getCookieSecret();
    
    if (requireAuth && !secret) {
      throw new Error('Authentication required - no player_secret cookie found');
    }

    const body: Record<string, unknown> = options.body ? JSON.parse(options.body as string) as Record<string, unknown> : {};
    
    if (secret) {
      body['_secret'] = secret;
    }

    return fetch(url, {
      ...options,
      headers: Object.assign(
        {},
        this.getHeaders(),
        options.headers ? (options.headers as Record<string, string>) : {}
      ),
      body: Object.keys(body).length > 0 ? JSON.stringify(body) : (options.body || null),
      credentials: 'include',
    });
  }

  // Player endpoints
  async createPlayer(username: string): Promise<Player> {
    if (this.useMockApi) {
      return mockApiService.createPlayer(username);
    }
    const response = await fetch(`${API_BASE}/player`, {
      method: 'POST',
      headers: this.getHeaders(false),
      body: JSON.stringify({ username }),
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to create player: ${response.statusText}`);
    }

    return response.json() as Promise<Player>;
  }

  async getPlayer(id: string): Promise<Player> {
    if (this.useMockApi) {
      return mockApiService.getPlayer(id);
    }
    const response = await fetch(`${API_BASE}/player/${id}`, {
      method: 'GET',
      headers: this.getHeaders(),
      credentials: 'include', // This will send HttpOnly cookies automatically
    });

    if (!response.ok) {
      throw new Error(`Failed to get player: ${response.statusText}`);
    }

    return response.json() as Promise<Player>;
  }

  // Hero endpoints
  async createHero(heroData: CreateHeroRequest): Promise<Hero> {
    if (this.useMockApi) {
      return mockApiService.createHero(heroData);
    }
    const response = await fetch(`${API_BASE}/hero`, {
      method: 'POST',
      headers: this.getHeaders(false),
      body: JSON.stringify(heroData),
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to create hero: ${response.statusText}`);
    }

    return response.json() as Promise<Hero>;
  }

  async getHero(id: string): Promise<Hero> {
    if (this.useMockApi) {
      return mockApiService.getHero(id);
    }
    const response = await this.makeAuthenticatedRequest(`${API_BASE}/hero/${id}`, {
      method: 'GET',
    });

    if (!response.ok) {
      throw new Error(`Failed to get hero: ${response.statusText}`);
    }

    return response.json() as Promise<Hero>;
  }

  async getHeroImage(id: string): Promise<string> {
    if (this.useMockApi) {
      return mockApiService.getHeroImage(id);
    }
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
    if (this.useMockApi) {
      return mockApiService.startFight(heroId);
    }
    const response = await fetch(`${API_BASE}/hero/${heroId}/fight`, {
      method: 'POST',
      headers: this.getHeaders(),
      credentials: 'include', // This will send HttpOnly cookies automatically
    });

    if (!response.ok) {
      throw new Error(`Failed to start fight: ${response.statusText}`);
    }

    return response.json() as Promise<FightResult>;
  }

  async getHeroFights(heroId: string, lastId?: string, limit = 20): Promise<FightsResponse> {
    const params = new URLSearchParams();
    if (lastId) {
      params.append('last_id', lastId);
    }
    params.append('limit', limit.toString());

    const response = await fetch(`${API_BASE}/hero/${heroId}/fights?${params}`, {
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to get hero fights: ${response.statusText}`);
    }

    return response.json() as Promise<FightsResponse>;
  }

  async getFight(heroId: string, fightId: string): Promise<Fight> {
    const response = await fetch(`${API_BASE}/hero/${heroId}/fight/${fightId}`, {
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to get fight: ${response.statusText}`);
    }

    return response.json() as Promise<Fight>;
  }

  // Player heroes endpoint (we'll need to create this in the backend)
  async getPlayerHeroes(playerId: string): Promise<Hero[]> {
    if (this.useMockApi) {
      return mockApiService.getPlayerHeroes(playerId);
    }
    const response = await fetch(`${API_BASE}/player/${playerId}/heroes`, {
      method: 'GET',
      headers: this.getHeaders(),
      credentials: 'include', // This will send HttpOnly cookies automatically
    });

    if (!response.ok) {
      throw new Error(`Failed to get player heroes: ${response.statusText}`);
    }

    return response.json() as Promise<Hero[]>;
  }
}

export const apiService = new ApiService();