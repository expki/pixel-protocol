import React, { useState } from 'react';
import type { Hero } from '../types/api';
import { usePlayer } from '../hooks/usePlayer';

interface HeroListProps {
  onSelectHero: (hero: Hero) => void;
}

export const HeroList: React.FC<HeroListProps> = ({ onSelectHero }) => {
  const { heroes, createHero, loading, error, refreshHeroes } = usePlayer();
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [creating, setCreating] = useState(false);

  const handleCreateHero = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim() || !description.trim()) {
      return;
    }

    setCreating(true);
    try {
      await createHero(title.trim(), description.trim());
      setTitle('');
      setDescription('');
      setShowCreateForm(false);
      await refreshHeroes();
    } catch (_err) {
      // Handle error silently or use proper error reporting
    } finally {
      setCreating(false);
    }
  };

  if (loading) {
    return (
      <div className="loading-container">
        <div className="loading-spinner"></div>
        <p>Loading your heroes...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="error-container">
        <h2>Error</h2>
        <p>{error}</p>
      </div>
    );
  }

  return (
    <div className="hero-list-container">
      <div className="hero-list-header">
        <h1>Your Heroes</h1>
        <button 
          className="btn btn-primary"
          onClick={() => {
            setShowCreateForm(!showCreateForm);
          }}
          disabled={creating}
        >
          Create New Hero
        </button>
      </div>

      {showCreateForm && (
        <div className="create-hero-form">
          <h3>Create a New Hero</h3>
          <form onSubmit={(e) => {
            void handleCreateHero(e);
          }}>
            <div className="form-group">
              <label htmlFor="title">Hero Title:</label>
              <input
                id="title"
                type="text"
                value={title}
                onChange={(e) => {
                  setTitle(e.target.value);
                }}
                placeholder="e.g., Dragon Slayer, Shadow Ninja..."
                required
                disabled={creating}
              />
            </div>
            <div className="form-group">
              <label htmlFor="description">Hero Description:</label>
              <textarea
                id="description"
                value={description}
                onChange={(e) => {
                  setDescription(e.target.value);
                }}
                placeholder="Describe your hero's abilities, background, and unique traits. Be creative - more creative heroes win more fights!"
                rows={4}
                required
                disabled={creating}
              />
            </div>
            <div className="form-actions">
              <button 
                type="submit" 
                className="btn btn-primary"
                disabled={creating || !title.trim() || !description.trim()}
              >
                {creating ? 'Creating...' : 'Create Hero'}
              </button>
              <button 
                type="button" 
                className="btn btn-secondary"
                onClick={() => {
                  setShowCreateForm(false);
                }}
                disabled={creating}
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      {heroes.length === 0 ? (
        <div className="empty-state">
          <h3>No Heroes Yet!</h3>
          <p>Create your first hero to start battling!</p>
        </div>
      ) : (
        <div className="heroes-grid">
          {heroes.map((hero) => (
            <div key={hero.ID} className="hero-card">
              <div className="hero-header">
                <h3>{hero.Title}</h3>
                <div className="hero-stats">
                  <span className="elo">ELO: {hero.Elo}</span>
                  <span className="country">{hero.Country}</span>
                </div>
              </div>
              <div className="hero-description">
                <p>{hero.Description}</p>
              </div>
              <div className="hero-actions">
                <button 
                  className="btn btn-primary"
                  onClick={() => {
                    onSelectHero(hero);
                  }}
                >
                  View Hero
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};