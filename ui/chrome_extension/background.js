// Background service worker — handles feed detection messages from content_script.js
// and routes subscribe requests to the popup/server.
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.type === 'FEED_DETECTED') {
    chrome.storage.session.set({ detectedFeeds: message.feeds });
    sendResponse({ ok: true });
  } else if (message.type === 'GET_DETECTED_FEEDS') {
    chrome.storage.session.get(['detectedFeeds'], (result) => {
      sendResponse({ feeds: result.detectedFeeds || [] });
    });
    return true; // keep channel open for async response
  }
});
