// SPDX-License-Identifier: LGPL-2.1-or-later
import 'dart:async';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/feed.dart';
import '../services/feed_service.dart';
import 'api_client_provider.dart';

class FeedNotifier extends StateNotifier<AsyncValue<List<Feed>>> {
  final FeedService _service;
  Timer? _pollTimer;

  FeedNotifier(this._service) : super(const AsyncValue.loading()) {
    loadFeeds();
    // Poll every 60 s so sidebar unread counts stay in sync with the backend
    // scheduler without requiring a manual page refresh.
    _pollTimer = Timer.periodic(const Duration(seconds: 60), (_) => _silentReload());
  }

  @override
  void dispose() {
    _pollTimer?.cancel();
    super.dispose();
  }

  // Reload without flashing the loading state — keeps existing data visible.
  Future<void> _silentReload() async {
    if (!mounted) return;
    try {
      final feeds = await _service.listFeeds();
      if (mounted) state = AsyncValue.data(feeds);
    } catch (_) {
      // Best-effort background poll — ignore errors, don't wipe existing state.
    }
  }

  Future<void> loadFeeds() async {
    state = const AsyncValue.loading();
    try {
      final feeds = await _service.listFeeds();
      state = AsyncValue.data(feeds);
    } catch (e, st) {
      state = AsyncValue.error(e, st);
    }
  }

  Future<Feed?> createFeed({
    required String title,
    required String xmlUrl,
    String? htmlUrl,
    String? folderId,
    String? iconUrl,
  }) async {
    try {
      final feed = await _service.createFeed(
        title: title,
        xmlUrl: xmlUrl,
        htmlUrl: htmlUrl,
        folderId: folderId,
        iconUrl: iconUrl,
      );
      final current = state.valueOrNull ?? [];
      state = AsyncValue.data([...current, feed]);
      return feed;
    } catch (_) {
      return null;
    }
  }

  Future<bool> deleteFeed(String id) async {
    try {
      await _service.deleteFeed(id);
      final current = state.valueOrNull ?? [];
      state = AsyncValue.data(current.where((f) => f.id != id).toList());
      return true;
    } catch (_) {
      return false;
    }
  }

  Future<bool> toggleFavorite(Feed feed) async {
    final updated = feed.copyWith(isFavorite: !feed.isFavorite);
    return updateFeed(updated);
  }

  Future<bool> updateFeed(Feed feed) async {
    try {
      final updated = await _service.updateFeed(feed);
      final current = state.valueOrNull ?? [];
      state = AsyncValue.data(
        current.map((f) => f.id == updated.id ? updated : f).toList(),
      );
      return true;
    } catch (_) {
      return false;
    }
  }

  Future<bool> importOPML(List<int> content) async {
    try {
      await _service.importOPML(content);
      await loadFeeds();
      return true;
    } catch (_) {
      return false;
    }
  }

  void decrementUnread(String feedId) {
    final current = state.valueOrNull;
    if (current == null) return;
    state = AsyncValue.data(
      current.map((f) {
        if (f.id == feedId) {
          return f.copyWith(unreadCount: (f.unreadCount - 1).clamp(0, 999999));
        }
        return f;
      }).toList(),
    );
  }

  void clearUnread(String feedId) {
    final current = state.valueOrNull;
    if (current == null) return;
    state = AsyncValue.data(
      current.map((f) => f.id == feedId ? f.copyWith(unreadCount: 0) : f).toList(),
    );
  }
}

final feedProvider =
    StateNotifierProvider<FeedNotifier, AsyncValue<List<Feed>>>((ref) {
  return FeedNotifier(ref.watch(feedServiceProvider));
});

// Convenience: feeds grouped by folderId (null = no folder)
final feedsByFolderProvider = Provider<Map<String?, List<Feed>>>((ref) {
  final feeds = ref.watch(feedProvider).valueOrNull ?? [];
  final map = <String?, List<Feed>>{};
  for (final feed in feeds) {
    map.putIfAbsent(feed.folderId, () => []).add(feed);
  }
  return map;
});
