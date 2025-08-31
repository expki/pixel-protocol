import React, { useState } from 'react';
import { PlayerProvider } from './context/PlayerContext';
import { Navigation } from './components/Navigation';
import { HeroList } from './components/HeroList';
import { HeroView } from './components/HeroView';
import type { Hero } from './types/api';
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

  const handleNavigate = (view: 'list' | 'hero') => {
    if (view === 'list') {
      handleBackToList();
    }
  };

  return (
    <PlayerProvider>
      <div className="app">
        <Navigation currentView={currentView} onNavigate={handleNavigate} />
        
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
