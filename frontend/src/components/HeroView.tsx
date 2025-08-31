import React, { useState, useEffect } from 'react';
import { Hero, FightResult } from '../types/api';
import { apiService } from '../services/api';

interface HeroViewProps {
  hero: Hero;
  onBack: () => void;
}

export const HeroView: React.FC<HeroViewProps> = ({ hero, onBack }) => {
  const [heroImage, setHeroImage] = useState<string | null>(null);
  const [fighting, setFighting] = useState(false);
  const [lastFight, setLastFight] = useState<FightResult | null>(null);
  const [showFightResult, setShowFightResult] = useState(false);

  useEffect(() => {
    const loadHeroImage = async () => {
      try {
        const imageUrl = await apiService.getHeroImage(hero.ID);
        setHeroImage(imageUrl);
      } catch (err) {
        console.error('Failed to load hero image:', err);
        // Image is optional, so we don't show an error
      }
    };

    loadHeroImage();

    return () => {
      if (heroImage) {
        URL.revokeObjectURL(heroImage);
      }
    };
  }, [hero.ID]);

  const handleFight = async () => {
    setFighting(true);
    try {
      const fightResult = await apiService.startFight(hero.ID);
      setLastFight(fightResult);
      setShowFightResult(true);
    } catch (err) {
      console.error('Fight failed:', err);
      alert(`Fight failed: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setFighting(false);
    }
  };

  const getFightOutcomeText = (outcome: 0 | 1 | 2): string => {
    switch (outcome) {
      case 0: return 'Draw';
      case 1: return 'Victory';
      case 2: return 'Defeat';
      default: return 'Unknown';
    }
  };

  const getFightOutcomeClass = (outcome: 0 | 1 | 2): string => {
    switch (outcome) {
      case 0: return 'draw';
      case 1: return 'victory';
      case 2: return 'defeat';
      default: return '';
    }
  };

  return (
    <div className="hero-view-container">
      <div className="hero-view-header">
        <button className="btn btn-secondary back-button" onClick={onBack}>
          ← Back to Heroes
        </button>
      </div>

      <div className="hero-main-info">
        <div className="hero-image-container">
          {heroImage ? (
            <img src={heroImage} alt={hero.Title} className="hero-image" />
          ) : (
            <div className="hero-image-placeholder">
              <span>🦸</span>
            </div>
          )}
        </div>
        
        <div className="hero-details">
          <h1>{hero.Title}</h1>
          <div className="hero-stats-detailed">
            <div className="stat">
              <label>ELO Rating:</label>
              <span className="elo-value">{hero.Elo}</span>
            </div>
            <div className="stat">
              <label>Country:</label>
              <span>{hero.Country}</span>
            </div>
            <div className="stat">
              <label>Player:</label>
              <span>{hero.Player?.UserName}#{hero.Player?.UserNameSuffix}</span>
            </div>
          </div>
        </div>
      </div>

      <div className="hero-description-section">
        <h3>Hero Description</h3>
        <div className="description-box">
          <p>{hero.Description}</p>
        </div>
      </div>

      <div className="hero-actions-section">
        <button 
          className="btn btn-primary fight-button"
          onClick={handleFight}
          disabled={fighting}
        >
          {fighting ? 'Fighting...' : '⚔️ Start Fight'}
        </button>
        <p className="fight-hint">
          Fights are determined by creativity! More unique and entertaining hero descriptions have better chances of winning.
        </p>
      </div>

      {showFightResult && lastFight && (
        <div className="fight-result-modal">
          <div className="fight-result-content">
            <h3 className={`fight-outcome ${getFightOutcomeClass(lastFight.fight.Outcome)}`}>
              {getFightOutcomeText(lastFight.fight.Outcome)}!
            </h3>
            
            <div className="fight-summary">
              <div className="fighters">
                <div className="fighter attacker">
                  <h4>{lastFight.fight.Attacker.Title}</h4>
                  <p>ELO: {lastFight.fight.Attacker.Elo}</p>
                </div>
                <div className="vs">VS</div>
                <div className="fighter defender">
                  <h4>{lastFight.fight.Defender.Title}</h4>
                  <p>ELO: {lastFight.fight.Defender.Elo}</p>
                </div>
              </div>
              
              <div className="elo-change">
                <p>ELO Change: {lastFight.elo_gain > 0 ? '+' : ''}{lastFight.elo_gain}</p>
              </div>
            </div>

            <div className="fight-transcript">
              <h4>Battle Story</h4>
              <div className="transcript-text">
                {lastFight.fight.Transcript.split('\n').map((paragraph, index) => (
                  <p key={index}>{paragraph}</p>
                ))}
              </div>
            </div>

            <div className="fight-result-actions">
              <button 
                className="btn btn-primary"
                onClick={() => setShowFightResult(false)}
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};