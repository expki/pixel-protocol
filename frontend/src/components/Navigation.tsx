import React from 'react';
import { usePlayer } from '../hooks/usePlayer';

interface NavigationProps {
  currentView: 'list' | 'hero';
  onNavigate: (view: 'list' | 'hero') => void;
}

export const Navigation: React.FC<NavigationProps> = ({ currentView, onNavigate }) => {
  const { player, heroes } = usePlayer();

  return (
    <nav className="nav-bar">
      <div className="nav-container">
        <a href="#" className="nav-brand" onClick={() => onNavigate('list')}>
          🎮 Pixel Protocol
        </a>
        
        <ul className="nav-links">
          <li>
            <a 
              href="#" 
              className={`nav-link ${currentView === 'list' ? 'active' : ''}`}
              onClick={(e) => {
                e.preventDefault();
                onNavigate('list');
              }}
            >
              Battle Arena
            </a>
          </li>
          <li>
            <a href="#" className="nav-link">
              Leaderboard
            </a>
          </li>
          <li>
            <a href="#" className="nav-link">
              Tournament
            </a>
          </li>
          <li>
            <a href="#" className="nav-link">
              Shop
            </a>
          </li>
        </ul>

        <div className="nav-user">
          {player && (
            <div className="user-info">
              <span className="user-name">
                {player.UserName}#{player.UserNameSuffix}
              </span>
              <span className="user-status">
                {heroes.length} {heroes.length === 1 ? 'Hero' : 'Heroes'}
              </span>
            </div>
          )}
        </div>
      </div>
    </nav>
  );
};