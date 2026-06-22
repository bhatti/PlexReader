// SPDX-License-Identifier: LGPL-2.1-or-later
// ignore: avoid_web_libraries_in_flutter
import 'dart:html' as html;
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../theme/app_theme.dart';
import '../models/feed.dart';
import '../models/folder.dart';
import '../providers/folder_provider.dart';
import '../providers/feed_provider.dart';
import '../providers/navigation_provider.dart';
import '../providers/api_client_provider.dart';
import '../providers/article_provider.dart' show articleRefreshSignalProvider;
import '../app.dart' show rootNavigatorKey, routerProvider;
import 'add_feed_dialog.dart';
import 'create_folder_dialog.dart';
import 'import_opml_dialog.dart' show ImportOpmlDialog;

class Sidebar extends ConsumerWidget {
  const Sidebar({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final collapsed = ref.watch(sidebarCollapsedProvider);
    return Container(
      color: Theme.of(context).brightness == Brightness.dark
          ? AppColors.sidebar
          : AppColorsLight.sidebar,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _SidebarHeader(collapsed: collapsed),
          const Divider(height: 1),
          Expanded(
            child: collapsed
                ? _CollapsedNav()
                : _ExpandedNav(),
          ),
          const Divider(height: 1),
          _SidebarFooter(collapsed: collapsed),
        ],
      ),
    );
  }
}

class _SidebarHeader extends ConsumerWidget {
  final bool collapsed;
  const _SidebarHeader({required this.collapsed});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return InkWell(
      onTap: () => ref.read(sidebarCollapsedProvider.notifier).state = !collapsed,
      child: Container(
        height: 56,
        padding: const EdgeInsets.symmetric(horizontal: 12),
        child: Row(
          children: [
            Container(
              width: 32,
              height: 32,
              decoration: BoxDecoration(
                color: AppColors.primary,
                borderRadius: BorderRadius.circular(8),
              ),
              child: const Icon(Icons.rss_feed, color: Colors.white, size: 18),
            ),
            if (!collapsed) ...[
              const SizedBox(width: 10),
              const Expanded(
                child: Text(
                  'PlexReader',
                  style: TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.w700,
                    letterSpacing: -0.3,
                  ),
                ),
              ),
              Icon(
                collapsed ? Icons.chevron_right : Icons.chevron_left,
                size: 18,
                color: AppColors.textSecondary,
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _CollapsedNav extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final location = GoRouterState.of(context).matchedLocation;
    return Column(
      children: [
        _CollapseNavItem(Icons.wb_sunny_outlined, '/today', 'Today', location),
        _CollapseNavItem(Icons.menu_outlined, '/all', 'All', location),
        _CollapseNavItem(Icons.bookmark_border, '/saved', 'Saved', location),
        _CollapseNavItem(Icons.star_border_rounded, '/starred', 'Starred', location),
        _CollapseNavItem(Icons.history, '/recently-read', 'Recent', location),
        _CollapseNavItem(Icons.search, '/search', 'Search', location),
      ],
    );
  }
}

class _CollapseNavItem extends StatelessWidget {
  final IconData icon;
  final String route;
  final String label;
  final String currentLocation;
  const _CollapseNavItem(this.icon, this.route, this.label, this.currentLocation);

  @override
  Widget build(BuildContext context) {
    final active = currentLocation == route;
    return Tooltip(
      message: label,
      child: InkWell(
        onTap: () => context.go(route),
        child: Container(
          height: 44,
          width: double.infinity,
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: active ? AppColors.selectedItem : Colors.transparent,
            border: active
                ? const Border(left: BorderSide(color: AppColors.primary, width: 2))
                : null,
          ),
          child: Icon(icon,
              color: active ? AppColors.primary : AppColors.textSecondary,
              size: 20),
        ),
      ),
    );
  }
}

class _ExpandedNav extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final location = GoRouterState.of(context).matchedLocation;
    final foldersAsync = ref.watch(folderProvider);
    ref.watch(feedProvider); // rebuild on feed changes
    final feedsByFolder = ref.watch(feedsByFolderProvider);
    final expandedIds = ref.watch(expandedFolderIdsProvider);

    // Total unread across all feeds.
    final allFeeds = ref.watch(feedProvider).valueOrNull ?? [];
    final totalUnread = allFeeds.fold<int>(0, (sum, f) => sum + f.unreadCount);

    return ListView(
      padding: const EdgeInsets.symmetric(vertical: 4),
      children: [
        _NavItem(
          icon: Icons.wb_sunny_outlined,
          label: 'Today',
          route: '/today',
          currentLocation: location,
        ),
        _NavItem(
          icon: Icons.bookmark_border,
          label: 'Saved for Later',
          route: '/saved',
          currentLocation: location,
        ),
        _NavItem(
          icon: Icons.star_border_rounded,
          label: 'Starred',
          route: '/starred',
          currentLocation: location,
        ),
        _NavItem(
          icon: Icons.history,
          label: 'Recently Read',
          route: '/recently-read',
          currentLocation: location,
        ),
        _NavItem(
          icon: Icons.search,
          label: 'Search',
          route: '/search',
          currentLocation: location,
        ),

        // ── Feeds section header ──────────────────────────────────────────
        const Padding(
          padding: EdgeInsets.fromLTRB(12, 16, 12, 4),
          child: Text(
            'FEEDS',
            style: TextStyle(
              color: AppColors.textSecondary,
              fontSize: 11,
              fontWeight: FontWeight.w600,
              letterSpacing: 0.8,
            ),
          ),
        ),

        // "All" — shows everything across all feeds (like Feedly)
        _AllFeedsRow(unreadCount: totalUnread, currentLocation: location),

        // Uncategorized feeds (no folder)
        ...feedsByFolder[null]?.map((feed) => _FeedTile(
              feed: feed,
              currentLocation: location,
            )) ??
            [],

        // Folders with their feeds
        ...foldersAsync.valueOrNull?.map((folder) {
              final feeds = feedsByFolder[folder.id] ?? [];
              final expanded = expandedIds.contains(folder.id);
              return _FolderSection(
                folder: folder,
                feeds: feeds,
                expanded: expanded,
                currentLocation: location,
                onToggle: () {
                  final current = Set<String>.from(expandedIds);
                  if (expanded) {
                    current.remove(folder.id);
                  } else {
                    current.add(folder.id);
                  }
                  ref.read(expandedFolderIdsProvider.notifier).state = current;
                },
              );
            }) ??
            [],
        const SizedBox(height: 8),
      ],
    );
  }
}

// ── "All" feeds row ─────────────────────────────────────────────────────────

class _AllFeedsRow extends ConsumerWidget {
  final int unreadCount;
  final String currentLocation;
  const _AllFeedsRow({required this.unreadCount, required this.currentLocation});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final active = currentLocation == '/all';
    return _HoverRow(
      active: active,
      onTap: () => context.go('/all'),
      builder: (ctx, hovering) => Row(
        children: [
          Container(
            width: 18,
            height: 18,
            decoration: BoxDecoration(
              color: AppColors.primary.withValues(alpha: 0.15),
              borderRadius: BorderRadius.circular(4),
            ),
            child: const Icon(Icons.menu_outlined, size: 12, color: AppColors.primary),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              'All',
              style: TextStyle(
                color: active ? AppColors.textPrimary : AppColors.textSecondary,
                fontSize: 13,
                fontWeight: active ? FontWeight.w600 : FontWeight.normal,
              ),
            ),
          ),
          if (!hovering && unreadCount > 0) _UnreadBadge(unreadCount),
          Offstage(
            offstage: !hovering,
            child: _DotsMenu(items: [
              _MenuItem(
                icon: Icons.done_all,
                label: 'Mark All as Read',
                onTap: () async {
                  try {
                    await ref.read(articleServiceProvider).markAllAsRead(all: true);
                    ref.read(feedProvider.notifier).loadFeeds();
                    ref.read(folderProvider.notifier).loadFolders();
                    ref.read(articleRefreshSignalProvider.notifier).state++;
                  } catch (_) {}
                },
              ),
            ]),
          ),
        ],
      ),
    );
  }
}

// ── Nav item (Today, Saved, Starred…) ───────────────────────────────────────

class _NavItem extends StatelessWidget {
  final IconData icon;
  final String label;
  final String route;
  final String currentLocation;

  const _NavItem({
    required this.icon,
    required this.label,
    required this.route,
    required this.currentLocation,
  });

  @override
  Widget build(BuildContext context) {
    final active = currentLocation == route;
    return _HoverRow(
      active: active,
      onTap: () => context.go(route),
      builder: (ctx, hovering) => Row(
        children: [
          Icon(icon,
              color: active ? AppColors.primary : AppColors.textSecondary,
              size: 18),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              label,
              style: TextStyle(
                color: active ? AppColors.textPrimary : AppColors.textSecondary,
                fontSize: 13,
                fontWeight: active ? FontWeight.w500 : FontWeight.normal,
              ),
              overflow: TextOverflow.ellipsis,
            ),
          ),
        ],
      ),
    );
  }
}

// ── Folder row ───────────────────────────────────────────────────────────────

class _FolderSection extends ConsumerWidget {
  final Folder folder;
  final List<Feed> feeds;
  final bool expanded;
  final String currentLocation;
  final VoidCallback onToggle;

  const _FolderSection({
    required this.folder,
    required this.feeds,
    required this.expanded,
    required this.currentLocation,
    required this.onToggle,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final active = currentLocation == '/folder/${folder.id}';
    // Sum unread from all feeds in this folder.
    final folderUnread = feeds.fold<int>(0, (sum, f) => sum + f.unreadCount);

    return Column(
      children: [
        _HoverRow(
          active: active,
          onTap: () => context.go('/folder/${folder.id}'),
          builder: (ctx, hovering) => Row(
            children: [
              // Chevron — expands/collapses, stops propagation to row nav tap.
              GestureDetector(
                behavior: HitTestBehavior.opaque,
                onTap: () {
                  onToggle();
                },
                child: Padding(
                  padding: const EdgeInsets.all(4),
                  child: Icon(
                    expanded ? Icons.keyboard_arrow_down : Icons.keyboard_arrow_right,
                    color: AppColors.textSecondary,
                    size: 16,
                  ),
                ),
              ),
              // Folder icon — small colored square
              Container(
                width: 16,
                height: 16,
                decoration: BoxDecoration(
                  color: _folderColor(folder.name),
                  borderRadius: BorderRadius.circular(3),
                ),
                child: const Icon(Icons.folder, size: 10, color: Colors.white),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  folder.name,
                  style: TextStyle(
                    color: active ? AppColors.textPrimary : AppColors.textSecondary,
                    fontSize: 13,
                    fontWeight: active ? FontWeight.w500 : FontWeight.normal,
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              if (!hovering && folderUnread > 0) _UnreadBadge(folderUnread),
              Offstage(
                offstage: !hovering,
                child: _DotsMenu(items: [
                  _MenuItem(
                    icon: Icons.done_all,
                    label: 'Mark All as Read',
                    onTap: () async {
                      try {
                        await ref.read(articleServiceProvider).markAllAsRead(folderId: folder.id);
                        ref.read(feedProvider.notifier).loadFeeds();
                        ref.read(folderProvider.notifier).loadFolders();
                        ref.read(articleRefreshSignalProvider.notifier).state++;
                      } catch (_) {}
                    },
                  ),
                  _MenuItem(
                    icon: Icons.edit_outlined,
                    label: 'Rename',
                    onTap: () => _renameFolder(ref),
                  ),
                  _MenuItem(
                    icon: Icons.delete_outline,
                    label: 'Delete',
                    onTap: () => _deleteFolder(ref),
                    destructive: true,
                  ),
                ]),
              ),
            ],
          ),
        ),
        if (expanded)
          ...feeds.map((feed) => Padding(
                padding: const EdgeInsets.only(left: 20),
                child: _FeedTile(feed: feed, currentLocation: currentLocation),
              )),
      ],
    );
  }

  // Deterministic color from folder name.
  static Color _folderColor(String name) {
    const colors = [
      Color(0xFF1976D2), Color(0xFF388E3C), Color(0xFF7B1FA2),
      Color(0xFFF57C00), Color(0xFF0097A7), Color(0xFFD32F2F),
      Color(0xFF5D4037), Color(0xFF455A64),
    ];
    final idx = name.isEmpty ? 0 : name.codeUnits.fold(0, (a, b) => a + b) % colors.length;
    return colors[idx];
  }

  void _renameFolder(WidgetRef ref) {
    final ctx = rootNavigatorKey.currentContext;
    if (ctx == null) return;
    final controller = TextEditingController(text: folder.name);
    showDialog(
      context: ctx,
      builder: (dialogCtx) => AlertDialog(
        title: const Text('Rename Folder'),
        content: TextField(
          controller: controller,
          autofocus: true,
          decoration: const InputDecoration(labelText: 'Name'),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.of(dialogCtx).pop(), child: const Text('Cancel')),
          ElevatedButton(
            onPressed: () {
              Navigator.of(dialogCtx).pop();
              ref.read(folderProvider.notifier).updateFolder(
                    folder.copyWith(name: controller.text.trim()),
                  );
            },
            child: const Text('Save'),
          ),
        ],
      ),
    );
  }

  void _deleteFolder(WidgetRef ref) {
    final ctx = rootNavigatorKey.currentContext;
    if (ctx == null) return;
    showDialog(
      context: ctx,
      builder: (dialogCtx) => AlertDialog(
        title: const Text('Delete Folder'),
        content: Text('Delete "${folder.name}"? Feeds will become uncategorized.'),
        actions: [
          TextButton(onPressed: () => Navigator.of(dialogCtx).pop(), child: const Text('Cancel')),
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: AppColors.error),
            onPressed: () async {
              Navigator.of(dialogCtx).pop();
              await ref.read(folderProvider.notifier).deleteFolder(folder.id);
              ref.read(feedProvider.notifier).loadFeeds();
              ref.read(routerProvider).go('/today');
            },
            child: const Text('Delete'),
          ),
        ],
      ),
    );
  }
}

// ── Feed tile ────────────────────────────────────────────────────────────────

class _FeedTile extends ConsumerWidget {
  final Feed feed;
  final String currentLocation;

  const _FeedTile({required this.feed, required this.currentLocation});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final active = currentLocation == '/feed/${feed.id}';
    return _HoverRow(
      active: active,
      height: 32,
      onTap: () => context.go('/feed/${feed.id}'),
      builder: (ctx, hovering) => Row(
        children: [
          _FeedIcon(feed: feed, size: 16),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              feed.title,
              style: TextStyle(
                color: active ? AppColors.textPrimary : AppColors.textSecondary,
                fontSize: 12.5,
                fontWeight: active ? FontWeight.w500 : FontWeight.normal,
              ),
              overflow: TextOverflow.ellipsis,
            ),
          ),
          if (feed.hasError && !hovering)
            Tooltip(
              message: feed.lastError ?? 'Feed error',
              child: const Padding(
                padding: EdgeInsets.only(right: 2),
                child: Icon(Icons.error_outline, color: AppColors.error, size: 13),
              ),
            ),
          if (!hovering && feed.unreadCount > 0) _UnreadBadge(feed.unreadCount),
          Offstage(
            offstage: !hovering,
            child: _DotsMenu(items: [
              _MenuItem(
                icon: Icons.done_all,
                label: 'Mark All as Read',
                onTap: () async {
                  try {
                    await ref.read(articleServiceProvider).markAllAsRead(feedId: feed.id);
                    ref.read(feedProvider.notifier).loadFeeds();
                    ref.read(articleRefreshSignalProvider.notifier).state++;
                  } catch (_) {}
                },
              ),
              _MenuItem(
                icon: Icons.refresh,
                label: 'Refresh',
                onTap: () async {
                  try {
                    await ref.read(feedServiceProvider).refreshFeed(feed.id);
                    ref.read(feedProvider.notifier).loadFeeds();
                  } catch (_) {}
                },
              ),
              _MenuItem(
                icon: Icons.edit_outlined,
                label: 'Edit',
                onTap: () => _editFeed(ref),
              ),
              _MenuItem(
                icon: Icons.delete_outline,
                label: 'Delete',
                onTap: () => _deleteFeed(ref),
                destructive: true,
              ),
            ]),
          ),
        ],
      ),
    );
  }

  void _editFeed(WidgetRef ref) {
    final ctx = rootNavigatorKey.currentContext;
    if (ctx == null) return;
    showDialog(
      context: ctx,
      builder: (_) => _EditFeedDialog(feed: feed, ref: ref),
    );
  }

  void _deleteFeed(WidgetRef ref) {
    final ctx = rootNavigatorKey.currentContext;
    if (ctx == null) return;
    showDialog(
      context: ctx,
      builder: (dialogCtx) => AlertDialog(
        title: const Text('Delete Feed'),
        content: Text('Delete "${feed.title}"? All articles will be removed.'),
        actions: [
          TextButton(onPressed: () => Navigator.of(dialogCtx).pop(), child: const Text('Cancel')),
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: AppColors.error),
            onPressed: () async {
              Navigator.of(dialogCtx).pop();
              await ref.read(feedProvider.notifier).deleteFeed(feed.id);
              ref.read(routerProvider).go('/today');
            },
            child: const Text('Delete'),
          ),
        ],
      ),
    );
  }
}

// ── Feed favicon icon with color-hash fallback ───────────────────────────────

class _FeedIcon extends StatelessWidget {
  final Feed feed;
  final double size;
  const _FeedIcon({required this.feed, required this.size});

  @override
  Widget build(BuildContext context) {
    // Try explicit iconUrl first.
    if (feed.iconUrl != null && feed.iconUrl!.isNotEmpty) {
      return _NetworkIcon(url: feed.iconUrl!, size: size, fallbackColor: _colorFor(feed));
    }
    // Derive Google favicon from htmlUrl or xmlUrl.
    final pageUrl = (feed.htmlUrl != null && feed.htmlUrl!.isNotEmpty)
        ? feed.htmlUrl!
        : feed.xmlUrl;
    final faviconUrl = _googleFavicon(pageUrl);
    if (faviconUrl != null) {
      return _NetworkIcon(url: faviconUrl, size: size, fallbackColor: _colorFor(feed));
    }
    return _InitialIcon(title: feed.title, size: size, color: _colorFor(feed));
  }

  /// Google's public favicon service — works for most sites, no API key needed.
  static String? _googleFavicon(String url) {
    try {
      final uri = Uri.parse(url);
      if (!uri.hasScheme || uri.host.isEmpty) return null;
      final origin = '${uri.scheme}://${uri.host}';
      return 'https://www.google.com/s2/favicons?domain=$origin&sz=32';
    } catch (_) {
      return null;
    }
  }

  static Color _colorFor(Feed feed) {
    const palette = [
      Color(0xFFE53935), Color(0xFFE91E63), Color(0xFF9C27B0),
      Color(0xFF3F51B5), Color(0xFF2196F3), Color(0xFF009688),
      Color(0xFF4CAF50), Color(0xFFFF9800), Color(0xFF795548),
      Color(0xFF607D8B), Color(0xFF00BCD4), Color(0xFFFF5722),
    ];
    final key = feed.xmlUrl.isEmpty ? feed.title : feed.xmlUrl;
    final idx = key.codeUnits.fold(0, (a, b) => a + b) % palette.length;
    return palette[idx];
  }
}

class _NetworkIcon extends StatelessWidget {
  final String url;
  final double size;
  final Color fallbackColor;
  const _NetworkIcon({required this.url, required this.size, required this.fallbackColor});

  @override
  Widget build(BuildContext context) {
    return ClipRRect(
      borderRadius: BorderRadius.circular(3),
      child: Image.network(
        url,
        width: size,
        height: size,
        fit: BoxFit.cover,
        errorBuilder: (_, __, ___) => _InitialIcon(
          title: url,
          size: size,
          color: fallbackColor,
        ),
      ),
    );
  }
}

class _InitialIcon extends StatelessWidget {
  final String title;
  final double size;
  final Color color;
  const _InitialIcon({required this.title, required this.size, required this.color});

  @override
  Widget build(BuildContext context) {
    final letter = title.isNotEmpty ? title[0].toUpperCase() : '?';
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        color: color,
        borderRadius: BorderRadius.circular(3),
      ),
      alignment: Alignment.center,
      child: Text(
        letter,
        style: TextStyle(
          color: Colors.white,
          fontSize: size * 0.6,
          fontWeight: FontWeight.w700,
        ),
      ),
    );
  }
}

// ── Unread badge ─────────────────────────────────────────────────────────────

class _UnreadBadge extends StatelessWidget {
  final int count;
  const _UnreadBadge(this.count);

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(left: 4),
      padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 1),
      decoration: BoxDecoration(
        color: AppColors.unreadBadge,
        borderRadius: BorderRadius.circular(10),
      ),
      child: Text(
        count > 999 ? '999+' : '$count',
        style: const TextStyle(
          color: AppColors.textSecondary,
          fontSize: 11,
          fontWeight: FontWeight.w500,
        ),
      ),
    );
  }
}

// ── Hover row — shared base for all clickable sidebar rows ───────────────────
// Uses a builder so children can react to hover state (e.g. show dots menu
// only on hover, show unread count when not hovered — like Feedly).

class _HoverRow extends StatefulWidget {
  final Widget Function(BuildContext context, bool hovering) builder;
  final bool active;
  final VoidCallback? onTap;
  final double height;

  const _HoverRow({
    required this.builder,
    required this.active,
    this.onTap,
    this.height = 34,
  });

  @override
  State<_HoverRow> createState() => _HoverRowState();
}

class _HoverRowState extends State<_HoverRow> {
  bool _hovering = false;

  @override
  Widget build(BuildContext context) {
    final bg = widget.active
        ? AppColors.selectedItem
        : _hovering
            ? AppColors.hoverItem
            : Colors.transparent;

    return MouseRegion(
      onEnter: (_) => setState(() => _hovering = true),
      onExit: (_) => setState(() => _hovering = false),
      child: GestureDetector(
        onTap: widget.onTap,
        child: Container(
          height: widget.height,
          decoration: BoxDecoration(
            color: bg,
            border: Border(
              left: BorderSide(
                color: widget.active ? AppColors.primary : Colors.transparent,
                width: 2,
              ),
            ),
          ),
          padding: const EdgeInsets.only(left: 10, right: 4),
          child: widget.builder(context, _hovering),
        ),
      ),
    );
  }
}

// ── … dots menu ──────────────────────────────────────────────────────────────

class _MenuItem {
  final IconData icon;
  final String label;
  final VoidCallback onTap;
  final bool destructive;
  const _MenuItem({
    required this.icon,
    required this.label,
    required this.onTap,
    this.destructive = false,
  });
}

class _DotsMenu extends StatefulWidget {
  final List<_MenuItem> items;
  const _DotsMenu({required this.items});

  @override
  State<_DotsMenu> createState() => _DotsMenuState();
}

class _DotsMenuState extends State<_DotsMenu> {
  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 22,
      height: 22,
      child: PopupMenuButton<int>(
        padding: EdgeInsets.zero,
        iconSize: 14,
        tooltip: '',
        icon: const Icon(Icons.more_horiz, size: 14, color: AppColors.textSecondary),
        onSelected: (i) => widget.items[i].onTap(),
        itemBuilder: (_) => widget.items.asMap().entries.map((e) {
          final item = e.value;
          return PopupMenuItem<int>(
            value: e.key,
            height: 36,
            child: Row(
              children: [
                Icon(
                  item.icon,
                  size: 15,
                  color: item.destructive ? AppColors.error : null,
                ),
                const SizedBox(width: 10),
                Text(
                  item.label,
                  style: TextStyle(
                    fontSize: 13,
                    color: item.destructive ? AppColors.error : null,
                  ),
                ),
              ],
            ),
          );
        }).toList(),
      ),
    );
  }
}

// ── Edit feed dialog ─────────────────────────────────────────────────────────

class _EditFeedDialog extends StatefulWidget {
  final Feed feed;
  final WidgetRef ref;
  const _EditFeedDialog({required this.feed, required this.ref});

  @override
  State<_EditFeedDialog> createState() => _EditFeedDialogState();
}

class _EditFeedDialogState extends State<_EditFeedDialog> {
  late final TextEditingController _title;
  String? _folderId;

  @override
  void initState() {
    super.initState();
    _title = TextEditingController(text: widget.feed.title);
    _folderId = widget.feed.folderId;
  }

  @override
  void dispose() {
    _title.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final folders = widget.ref.read(folderProvider).valueOrNull ?? [];
    return AlertDialog(
      title: const Text('Edit Feed'),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          TextField(
            controller: _title,
            decoration: const InputDecoration(labelText: 'Title'),
          ),
          const SizedBox(height: 12),
          DropdownButtonFormField<String?>(
            value: _folderId,
            decoration: const InputDecoration(labelText: 'Folder'),
            items: [
              const DropdownMenuItem(value: null, child: Text('None')),
              ...folders.map((f) => DropdownMenuItem(value: f.id, child: Text(f.name))),
            ],
            onChanged: (v) => setState(() => _folderId = v),
          ),
        ],
      ),
      actions: [
        TextButton(onPressed: () => Navigator.of(context).pop(), child: const Text('Cancel')),
        ElevatedButton(
          onPressed: () {
            Navigator.of(context).pop();
            widget.ref.read(feedProvider.notifier).updateFeed(
                  widget.feed.copyWith(title: _title.text.trim(), folderId: _folderId),
                );
          },
          child: const Text('Save'),
        ),
      ],
    );
  }
}

// ── Sidebar footer ────────────────────────────────────────────────────────────

class _SidebarFooter extends ConsumerWidget {
  final bool collapsed;
  const _SidebarFooter({required this.collapsed});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    if (collapsed) {
      return Padding(
        padding: const EdgeInsets.symmetric(vertical: 8),
        child: Column(
          children: [
            IconButton(
              icon: const Icon(Icons.add, size: 20),
              onPressed: () => _showAddMenu(context, ref),
              tooltip: 'Add',
            ),
            IconButton(
              icon: const Icon(Icons.download_outlined, size: 20),
              onPressed: () => _exportOpml(context, ref),
              tooltip: 'Export OPML',
            ),
            IconButton(
              icon: const Icon(Icons.settings_outlined, size: 20),
              onPressed: () => context.push('/preferences'),
              tooltip: 'Preferences',
            ),
          ],
        ),
      );
    }
    final versionAsync = ref.watch(versionProvider);
    final versionText = versionAsync.whenOrNull(data: (v) => v) ?? '';

    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Padding(
          padding: const EdgeInsets.all(8),
          child: Row(
            children: [
              Expanded(
                child: OutlinedButton.icon(
                  onPressed: () => _showAddMenu(context, ref),
                  icon: const Icon(Icons.add, size: 16),
                  label: const Text('Add Feed', style: TextStyle(fontSize: 13)),
                  style: OutlinedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
                  ),
                ),
              ),
              const SizedBox(width: 4),
              IconButton(
                icon: const Icon(Icons.download_outlined, size: 18),
                onPressed: () => _exportOpml(context, ref),
                tooltip: 'Export OPML',
              ),
              IconButton(
                icon: const Icon(Icons.settings_outlined, size: 18),
                onPressed: () => context.push('/preferences'),
                tooltip: 'Preferences',
              ),
            ],
          ),
        ),
        if (versionText.isNotEmpty)
          Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: Text(
              'v$versionText',
              style: const TextStyle(
                color: AppColors.textSecondary,
                fontSize: 10,
              ),
            ),
          ),
      ],
    );
  }

  Future<void> _exportOpml(BuildContext context, WidgetRef ref) async {
    try {
      final bytes = await ref.read(feedServiceProvider).exportOPML();
      final blob = html.Blob([bytes], 'text/xml');
      final url = html.Url.createObjectUrlFromBlob(blob);
      html.AnchorElement(href: url)
        ..setAttribute('download', 'plexreader-feeds.opml')
        ..click();
      html.Url.revokeObjectUrl(url);
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Export failed: $e')),
        );
      }
    }
  }

  void _showAddMenu(BuildContext context, WidgetRef ref) {
    // Capture a root-level context before entering the modal so dialogs opened
    // after closing the sheet have a valid navigator to attach to.
    final rootContext = context;
    showModalBottomSheet(
      context: context,
      // useRootNavigator keeps the sheet on the root navigator so
      // Navigator.pop(sheetCtx) only dismisses the sheet, never a GoRouter page.
      useRootNavigator: true,
      builder: (sheetCtx) => Material(
        child: SafeArea(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              ListTile(
                leading: const Icon(Icons.rss_feed),
                title: const Text('Follow a Source'),
                onTap: () {
                  Navigator.of(sheetCtx).pop();
                  showDialog(
                    context: rootContext,
                    builder: (_) => const AddFeedDialog(),
                  );
                },
              ),
              ListTile(
                leading: const Icon(Icons.folder_outlined),
                title: const Text('New Folder'),
                onTap: () {
                  Navigator.of(sheetCtx).pop();
                  showDialog(
                    context: rootContext,
                    builder: (_) => const CreateFolderDialog(),
                  );
                },
              ),
              ListTile(
                leading: const Icon(Icons.upload_file),
                title: const Text('Import OPML'),
                onTap: () {
                  Navigator.of(sheetCtx).pop();
                  showDialog(
                    context: rootContext,
                    builder: (_) => const ImportOpmlDialog(),
                  );
                },
              ),
            ],
          ),
        ),
      ),
    );
  }
}
