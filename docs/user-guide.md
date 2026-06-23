# PlexReader — User Guide

## Table of Contents

- [First Launch](#first-launch)
- [Importing Feeds from Feedly (OPML)](#importing-feeds-from-feedly-opml)
- [Adding Individual Feeds](#adding-individual-feeds)
- [Organising Feeds into Folders](#organising-feeds-into-folders)
- [Favorites](#favorites)
- [Reading Articles](#reading-articles)
- [View Modes](#view-modes)
- [Keyboard Shortcuts](#keyboard-shortcuts)
- [Starred Articles](#starred-articles)
- [Saved for Later](#saved-for-later)
- [Search](#search)
- [Today Feed](#today-feed)
- [Exporting Your Feeds (OPML)](#exporting-your-feeds-opml)
- [Preferences](#preferences)
- [Chrome Extension](#chrome-extension)

---

## First Launch

After starting PlexReader (see [Installation](installation.md)), open **http://localhost:3000** in your browser. You will land on the **Today** page, which is empty until you subscribe to some feeds.

The left sidebar shows:
- **Today** — articles from the last 24 hours across all feeds
- **All Articles** — every article, newest first
- **Starred** — articles you have explicitly starred
- **Saved for Later** — your reading list
- **Recently Read** — articles you have marked as read in the current session
- Your **folders** and individual **feeds** below a divider

---

## Importing Feeds from Feedly (OPML)

If you are migrating from Feedly, Inoreader, or any other reader that supports OPML export, you can import all your feeds at once.

### Export from Feedly

1. In Feedly, go to **Organise** → **Import/Export** → **Export as OPML**.
2. Save the downloaded `.opml` file to your computer.

### Import into PlexReader

1. Click the **import icon** (↑ arrow) in the top-right corner of the sidebar, or use the three-dot menu next to "Feeds".
2. Select **Import OPML** from the menu.
3. In the dialog that appears, click **Choose File** and select your `.opml` file.
4. Click **Import**.

PlexReader will:
- Create a folder for each `<outline>` group in the OPML.
- Subscribe to every `xmlUrl` feed found.
- Skip feeds that are already subscribed (no duplicates).
- Report a summary: *"12 feeds added, 3 folders created, 2 feeds skipped"*.

The import runs in the background. A background refresh begins immediately so your articles start populating within seconds.

---

## Adding Individual Feeds

To subscribe to a single feed:

1. Click **+ Add Feed** at the bottom of the sidebar, or click the **+** button next to a folder name.
2. Paste the RSS/Atom feed URL (e.g. `https://example.com/feed.rss`) or the website URL — PlexReader will auto-discover the feed from the page's `<link>` tag.
3. Optionally assign the feed to a **folder**.
4. Click **Subscribe**.

PlexReader fetches the feed immediately and displays the first batch of articles. The feed title and icon are auto-populated from the feed metadata; you can rename either at any time by right-clicking the feed in the sidebar and selecting **Edit**.

---

## Organising Feeds into Folders

Folders are optional but useful when you follow many feeds.

- **Create folder**: Click **+ New Folder** at the bottom of the sidebar. Type a name and press Enter.
- **Move feed to folder**: Drag a feed onto a folder in the sidebar, or right-click → **Move to Folder**.
- **Rename folder**: Right-click the folder → **Rename**.
- **Delete folder**: Right-click the folder → **Delete Folder**. Feeds inside the folder are **not** deleted; they move to the uncategorised section.

---

## Favorites

Favorites let you pin important feeds at the top of the sidebar in a dedicated **FAVORITES** section, separate from the regular folder/feed tree — the same pattern Feedly uses.

### Add a feed to Favorites

Two ways:
1. **From the toolbar**: Open the feed, then click the ☆ star icon next to the feed title in the top toolbar. The star fills amber (★) when active.
2. **From the sidebar**: Right-click (or hover → ⋯ menu) on any feed → **Add to Favorites**.

### Remove from Favorites

Click the filled ★ star in the toolbar, or hover the feed in the sidebar → ⋯ menu → **Remove from Favorites**.

The feed remains in its original folder — Favorites is a view, not a second location.

---

## Reading Articles

Click any article in the article list to open it in the reading pane on the right. The full article body is displayed using the stored HTML content fetched from the feed.

- **Open in browser**: Click the article title or the external-link icon (↗) to open the original URL in a new tab.
- **Mark as read**: Selecting an article in the list opens it in the reading pane but does **not** automatically mark it as read. Mark articles read explicitly using one of these methods:
  - Press `m` while the article is focused
  - Click the circle/dot icon on the article card
  - Click the **✓ N** (Mark all) button in the feed toolbar to mark everything in the current view as read
  - Press `Shift+A` to mark all as read
- **Mark all as read**: Click **Mark all read** in the feed toolbar, or press `Shift+A`.
- **Scroll through articles**: Use `j` (next) and `k` (previous) to move through the article list without touching the mouse.

---

## View Modes

Switch between three view modes using the toolbar icons at the top of any feed or folder view:

### Magazine

The default view. Each article shows a **thumbnail image**, the feed name, the article title in large text, a two-line summary, and the relative timestamp. This is the most visually rich mode and works best for image-heavy feeds like news sites.

### Title

A compact single-line list of article titles with unread indicators and timestamps. Use this mode when you follow high-volume feeds and want to triage quickly.

### Cards

A grid layout that displays articles as tiles — thumbnail on top, title below. Good for content-heavy feeds like photography or design blogs.

Your preferred view mode per feed/folder is remembered across sessions.

---

## Keyboard Shortcuts

PlexReader is fully keyboard-navigable. Press `?` at any time to display the shortcut reference.

| Key | Action |
|-----|--------|
| `j` | Move to the **next** article in the list |
| `k` | Move to the **previous** article in the list |
| `o` | **Open** the selected article in your browser (original URL) |
| `m` | **Toggle read/unread** on the selected article |
| `s` | **Star / unstar** the selected article |
| `l` | **Save for later** / remove from saved list |
| `Shift+A` | **Mark all as read** in the current feed or folder |
| `?` | Show the keyboard shortcut reference dialog |

Keyboard focus is maintained as you navigate. After pressing `j` or `k`, the focused article is highlighted and the reading pane updates automatically.

---

## Starred Articles

Starring an article permanently bookmarks it regardless of its read state or age. Starred articles are never deleted by the retention policy.

- **Star**: Press `s` on a focused article, or click the ★ icon on an article card.
- **View all starred**: Click **Starred** in the left sidebar.
- **Unstar**: Press `s` again, or click the filled ★ icon.

Use starred articles for content you want to reference again later — tutorials, reference posts, or recipes.

---

## Saved for Later

"Saved for Later" is a separate reading list from Starred. It is intended for articles you intend to read soon but have not read yet.

- **Save**: Press `l` on a focused article, or click the bookmark icon on an article card.
- **View**: Click **Saved for Later** in the left sidebar.
- **Remove**: Press `l` again, or click the filled bookmark icon.

Unlike starred, saved-for-later articles can be read and removed from the list once done.

---

## Search

Click the **search icon** (🔍) in the sidebar header, or navigate to the Search screen.

Type your query in the search box. PlexReader uses **SQLite FTS5** with Porter stemming to search article titles and full content across all your subscriptions. Results are ranked by relevance (BM25).

Search tips:
- Partial words are matched: `rustac` matches *rustacean*.
- Phrases must be quoted: `"async rust"` matches the exact phrase.
- Prefix search: `async*` matches *async*, *asynchronous*, *asyncio*.
- Searches are case-insensitive.

Results display in Magazine view. Click any result to open the article.

---

## Today Feed

The **Today** feed (first item in the sidebar) aggregates all articles published in the **last 24 hours** across every feed you subscribe to, sorted by publish time descending.

It provides a single consolidated view of what is new today without having to visit each feed individually — similar to Feedly's Today view.

---

## Exporting Your Feeds (OPML)

To export your current subscriptions as an OPML file (for backup or migration):

1. Click the three-dot menu (⋮) at the top of the sidebar.
2. Select **Export OPML**.
3. The browser downloads an `opml.xml` file containing all your feeds and folders.

The exported OPML is compatible with Feedly, Inoreader, NetNewsWire, and other standard RSS readers.

---

## Preferences

Open **Preferences** from the bottom of the left sidebar (gear icon ⚙). The following settings are available:

| Setting | Options | Description |
|---------|---------|-------------|
| Start page | Today, All Articles, Starred | Which screen opens when you launch PlexReader |
| Default view mode | Magazine, Title, Cards | Default article list layout |
| Default sort | Newest first, Oldest first | Article sort order |
| Hide read articles | On / Off | Filter out read articles from list views |
| Refresh interval | 5m – 24h | Global default feed refresh frequency |
| Retention | 7 – 365 days | How long to keep articles before auto-deletion. Starred articles are never deleted. |
| Theme | Dark | Currently only dark theme is available |

Preferences are stored server-side (in the SQLite database) and apply across all devices and browsers.

### Per-feed refresh interval

To override the global refresh interval for a specific feed:

1. Right-click the feed in the sidebar → **Edit Feed**.
2. Change the **Refresh interval** field (minimum 60 seconds).
3. Click **Save**.

The scheduler respects per-feed overrides on the next cycle.

---

## Chrome Extension

The Chrome extension embeds the full PlexReader UI in the browser toolbar popup, and can detect RSS feeds on any page.

> The extension only works in Chrome and Chromium-based browsers. For Firefox, Safari, or Edge, use the web app instead.

### Building and loading (developer mode)

```bash
# From the repo root — builds Flutter web and copies it into chrome_extension/
make chrome-ext
```

Then:
1. Go to `chrome://extensions/`
2. Enable **Developer mode** (top-right toggle)
3. Click **Load unpacked** → select `ui/chrome_extension/`

The PlexReader icon appears in the Chrome toolbar. Click it to open the full reader UI in a 400×600 px popup.

### Using any browser (not just Chrome)

The web app at **http://localhost:3000** (Docker) or **http://localhost:3001** (static build) works in any modern browser — Firefox, Safari, Edge, Brave. The Chrome extension is optional and only adds the one-click subscribe feature.

### Subscribing to a feed from any page

When you visit a website that publishes an RSS or Atom feed, the extension's content script detects the `<link rel="alternate" type="application/rss+xml">` tag and surfaces a **Subscribe** button in the popup. Click it to add the feed directly to your PlexReader instance.
