import { useContext } from 'react';
import { PlayerContext } from '../context/PlayerContext';
import type { PlayerContextType } from '../context/PlayerContext';

export const usePlayer = (): PlayerContextType => {
  const context = useContext(PlayerContext);
  if (context === undefined) {
    throw new Error('usePlayer must be used within a PlayerProvider');
  }
  return context;
};