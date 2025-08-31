import React, { useState } from 'react';
import { PlayerProvider } from './context/PlayerContext';
import { HeroList } from './components/HeroList';
import { HeroView } from './components/HeroView';
import { Hero } from './types/api';
import './App.css';

function App() {
  const [currentView, setCurrentView] = useState<'list' | 'hero'>('list');
  const [selectedHero, setSelectedHero] = useState<Hero | null>(null);

  const handleSelectHero = (hero: Hero) => {
    setSelectedHero(hero);
    setCurrentView('hero');
  };

  const handleBackToList = () => {
    setCurrentView('list');
    setSelectedHero(null);
  };

  return (
    <PlayerProvider>
      <div className="app">
        <header className="app-header">
          <div className="app-title">
            <h1>🎮 Pixel Protocol</h1>
            <p>Battle Arena</p>
          </div>
        </header>
        
        <main className="app-main">
          {currentView === 'list' && (
            <HeroList onSelectHero={handleSelectHero} />
          )}
          
          {currentView === 'hero' && selectedHero && (
            <HeroView 
              hero={selectedHero} 
              onBack={handleBackToList} 
            />
          )}
        </main>
      </div>
    </PlayerProvider>
  );
}

export default App;
