// Content script: detect RSS/Atom feed links on the current page and report them
// to the background service worker so the popup can offer a "subscribe" button.
(function detectFeeds() {
  const links = document.querySelectorAll(
    'link[type="application/rss+xml"], link[type="application/atom+xml"], link[type="application/feed+json"]'
  );

  if (links.length === 0) return;

  const feeds = Array.from(links).map((link) => ({
    url: link.href,
    title: link.title || document.title || link.href,
    type: link.type,
  }));

  chrome.runtime.sendMessage({ type: 'FEED_DETECTED', feeds });
})();
