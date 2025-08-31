import React, { createContext, useState, useEffect, useCallback } from 'react';
import type { ReactNode } from 'react';
import type { Player, Hero } from '../types/api';
import { apiService } from '../services/api';

export interface PlayerContextType {
  player: Player | null;
  heroes: Hero[];
  currentHero: Hero | null;
  loading: boolean;
  error: string | null;
  initializePlayer: () => Promise<void>;
  createHero: (title: string, description: string) => Promise<Hero>;
  selectHero: (hero: Hero) => void;
  refreshHeroes: () => Promise<void>;
}

// eslint-disable-next-line react-refresh/only-export-components
export const PlayerContext = createContext<PlayerContextType | undefined>(undefined);

interface PlayerProviderProps {
  children: ReactNode;
}

export const PlayerProvider: React.FC<PlayerProviderProps> = ({ children }) => {
  const [player, setPlayer] = useState<Player | null>(null);
  const [heroes, setHeroes] = useState<Hero[]>([]);
  const [currentHero, setCurrentHero] = useState<Hero | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const getCookieValue = (name: string): string | null => {
    const cookies = document.cookie.split(';');
    const cookie = cookies.find(c => c.trim().startsWith(`${name}=`));
    return cookie ? (cookie.split('=')[1] ?? null) : null;
  };

  const refreshHeroes = useCallback(async (playerToUse?: Player) => {
    // Use passed player or current player state
    const currentPlayer = playerToUse || player;
    
    // If no player is available, skip heroes refresh
    if (!currentPlayer) {
      return;
    }

    try {
      const heroesData = await apiService.getPlayerHeroes(currentPlayer.ID);
      setHeroes(heroesData);
    } catch (_err) {
      // For now, just set empty heroes list
      // The user can still create new heroes
      setHeroes([]);
    }
  }, [player]);

  const initializePlayer = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      // Check if we have existing player ID cookie (player_secret is HttpOnly)
      const playerId = getCookieValue('player_id');
      let currentPlayer: Player;
      
      if (!playerId) {
        // No player ID cookie, create a new player
        const newPlayer = await apiService.createPlayer(`Player${String(Date.now())}`);
        setPlayer(newPlayer);
        currentPlayer = newPlayer;
      } else {
        // We have a player ID, try to fetch the existing player info
        try {
          const existingPlayer = await apiService.getPlayer(playerId);
          setPlayer(existingPlayer);
          currentPlayer = existingPlayer;
        } catch (_err) {
          // Session invalid or expired, create a new player
          const newPlayer = await apiService.createPlayer(`Player${String(Date.now())}`);
          setPlayer(newPlayer);
          currentPlayer = newPlayer;
        }
      }
      
      await refreshHeroes(currentPlayer);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to initialize player');
    } finally {
      setLoading(false);
    }
  }, [refreshHeroes]);

  const createHero = async (title: string, description: string): Promise<Hero> => {
    const newHero = await apiService.createHero({ title, description });
    
    // Update player info if it came with the hero
    if (newHero.Player) {
      setPlayer(newHero.Player);
    }
    
    await refreshHeroes();
    return newHero;
  };

  const selectHero = (hero: Hero) => {
    setCurrentHero(hero);
  };

  useEffect(() => {
    void initializePlayer();
  }, []); // Only run once on mount

  const value: PlayerContextType = {
    player,
    heroes,
    currentHero,
    loading,
    error,
    initializePlayer,
    createHero,
    selectHero,
    refreshHeroes,
  };

  return (
    <PlayerContext.Provider value={value}>
      {children}
    </PlayerContext.Provider>
  );
};

