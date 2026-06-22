// Load the PlexReader server URL from chrome storage and embed the UI in the popup iframe.
// The user configures their server URL via the extension options or the first-run prompt.
const DEFAULT_URL = 'http://localhost:8080';

async function init() {
  const result = await chrome.storage.sync.get(['serverUrl']);
  const serverUrl = result.serverUrl || DEFAULT_URL;
  const frame = document.getElementById('plexreader-frame');
  frame.src = serverUrl;
}

init();
