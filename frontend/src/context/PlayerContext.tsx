import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { Player, Hero } from '../types/api';
import { apiService } from '../services/api';

interface PlayerContextType {
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

const PlayerContext = createContext<PlayerContextType | undefined>(undefined);

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
    return cookie ? cookie.split('=')[1] : null;
  };

  const initializePlayer = async () => {
    try {
      setLoading(true);
      setError(null);

      // Check if we have a player_secret cookie
      const playerSecret = getCookieValue('player_secret');
      
      if (!playerSecret) {
        // No player exists, create a new one
        const newPlayer = await apiService.createPlayer(`Player${Date.now()}`);
        setPlayer(newPlayer);
      } else {
        // We have a secret, try to create a hero to get player info
        try {
          const hero = await apiService.createHero({
            title: 'Welcome Hero',
            description: 'Your first hero in the Pixel Protocol battle arena!'
          });
          setPlayer(hero.Player || null);
        } catch (err) {
          console.error('Failed to create initial hero:', err);
          // If hero creation fails, create a new player
          const newPlayer = await apiService.createPlayer(`Player${Date.now()}`);
          setPlayer(newPlayer);
        }
      }
      
      await refreshHeroes();
    } catch (err) {
      console.error('Failed to initialize player:', err);
      setError(err instanceof Error ? err.message : 'Failed to initialize player');
    } finally {
      setLoading(false);
    }
  };

  const refreshHeroes = async () => {
    // If player is not loaded yet, skip heroes refresh
    if (!player) return;

    try {
      const heroesData = await apiService.getPlayerHeroes(player.ID);
      setHeroes(heroesData);
    } catch (err) {
      console.error('Failed to refresh heroes:', err);
      // For now, just log the error and set empty heroes list
      // The user can still create new heroes
      setHeroes([]);
    }
  };

  const createHero = async (title: string, description: string): Promise<Hero> => {
    try {
      const newHero = await apiService.createHero({ title, description });
      
      // Update player info if it came with the hero
      if (newHero.Player) {
        setPlayer(newHero.Player);
      }
      
      await refreshHeroes();
      return newHero;
    } catch (err) {
      console.error('Failed to create hero:', err);
      throw err;
    }
  };

  const selectHero = (hero: Hero) => {
    setCurrentHero(hero);
  };

  useEffect(() => {
    initializePlayer();
  }, []);

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

export const usePlayer = (): PlayerContextType => {
  const context = useContext(PlayerContext);
  if (context === undefined) {
    throw new Error('usePlayer must be used within a PlayerProvider');
  }
  return context;
};