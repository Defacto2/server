// Chiptune player for music module files
// This module handles playing chiptune/music files in the browser

import {ChiptuneJsPlayer} from 'https://DrSnuggles.github.io/chiptune/chiptune3.js';

let chiptune = null;

/**
 * Initialize the chiptune player when the page loads
 */
export function initChiptunePlayer() {
  // Check if we're on a page that needs the chiptune player
  const playButton = document.getElementById('chiptune-play-button');
  if (!playButton) return;

  // Set up the play button click handler
  playButton.onclick = playChiptune;

  // Set up cleanup for when leaving the page
  window.addEventListener('beforeunload', () => {
    if (chiptune) chiptune.stop();
  });

  console.log('Chiptune player initialized');
}

/**
 * Play chiptune music file
 */
async function playChiptune() {
  const button = document.getElementById('chiptune-play-button');
  if (!button) return;

  if (!chiptune) {
    button.innerHTML = '<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span><span class="visually-hidden">Loading...</span>';
    button.disabled = true;

    try {
      // Create new player instance
      chiptune = new ChiptuneJsPlayer();

      // Wait for initialization
      await new Promise((resolve) => {
        chiptune.onInitialized = () => {
          setupEventHandlers();
          resolve();
        };

        // Fallback timeout
        setTimeout(resolve, 10000);
      });
    } catch (error) {
      console.error('Failed to load chiptune player:', error);
      button.innerHTML = '🎵 Play music';
      button.disabled = false;
      alert('Failed to load chiptune player: ' + error.message);
      return;
    }
  }

  // Get the download URL
  const downloadLink = document.getElementById('artifact-download-link');
  if (!downloadLink) return;

  const fileUrl = downloadLink.href;

  try {
    button.innerHTML = '<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> Loading file...';
    button.disabled = true;

    // Load and play the music file
    await chiptune.load(fileUrl);

    button.innerHTML = '⏸️ Stop music';
    button.onclick = stopChiptune;
    button.disabled = false;

    // Setup handlers if not already done
    if (!chiptune.onMetadata) {
      setupEventHandlers();
    }

  } catch (error) {
    console.error('Failed to play music:', error);
    button.innerHTML = '🎵 Play music';
    button.disabled = false;
    alert('Failed to play music: ' + error.message);
  }
}

/**
 * Stop chiptune playback
 */
function stopChiptune() {
  const button = document.getElementById('chiptune-play-button');
  if (!button || !chiptune) return;

  chiptune.stop();

  button.innerHTML = '🎵 Play music';
  button.onclick = playChiptune;
}

/**
 * Set up event handlers for the chiptune player
 */
function setupEventHandlers() {
  if (!chiptune) return;

  // Try to get metadata if available
  chiptune.onMetadata = (metadata) => {
    // Metadata handling can be added back if needed
    console.log('Metadata received:', metadata);
  };
}

// Initialize the player when this module is loaded
initChiptunePlayer();