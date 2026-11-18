// Chiptune player for music module files
// This module handles playing chiptune/music files in the browser

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
}

/**
 * Play chiptune music file
 */
async function playChiptune() {
  const button = document.getElementById('chiptune-play-button');
  const downloadLink = document.getElementById('artifact-download-link');
  if (!button || !downloadLink) return;

  if (!chiptune) {
    button.innerHTML = '<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span><span class="visually-hidden">Loading...</span>';
    button.disabled = true;

    try {
      const { ChiptuneJsPlayer } = await import('/js/chiptune3.js');
      const player = new ChiptuneJsPlayer();

      await new Promise((resolve, reject) => {
        const timer = setTimeout(() => {
          reject(new Error('Chiptune AudioWorklet initialization timed out'));
        }, 10000);

        player.onInitialized(() => {
          clearTimeout(timer);
          resolve();
        });
      });

      chiptune = player;
      setupEventHandlers();
    } catch (error) {
      console.error('Failed to load chiptune player:', error);
      chiptune = null;
      button.innerHTML = '🎵 Play music';
      button.disabled = false;
      alert('Failed to load chiptune player: ' + error.message);
      return;
    }
  }

  const fileUrl = downloadLink.href;

  try {
    button.innerHTML = '<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> Loading file...';
    button.disabled = true;

    await chiptune.load(fileUrl);

    button.innerHTML = '⏸️ Stop music';
    button.onclick = stopChiptune;
    button.disabled = false;
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

  chiptune.onMetadata((metadata) => {
    console.log('Metadata received:', metadata);
  });

  chiptune.onEnded(() => {
    stopChiptune();
  });
}

// Initialize the player when this module is loaded
initChiptunePlayer();
