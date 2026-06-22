// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/article.dart';
import '../models/preferences.dart';
import '../services/article_service.dart';
import 'api_client_provider.dart';
import 'feed_provider.dart';
import 'folder_provider.dart';

class ArticleListState {
  final List<Article> articles;
  final bool isLoading;
  final bool hasMore;
  final String? nextPageToken;
  final String? error;
  final Set<String> selectedIds;
  final bool isMultiSelectMode;

  const ArticleListState({
    this.articles = const [],
    this.isLoading = false,
    this.hasMore = true,
    this.nextPageToken,
    this.error,
    this.selectedIds = const {},
    this.isMultiSelectMode = false,
  });

  ArticleListState copyWith({
    List<Article>? articles,
    bool? isLoading,
    bool? hasMore,
    String? nextPageToken,
    String? error,
    Set<String>? selectedIds,
    bool? isMultiSelectMode,
  }) {
    return ArticleListState(
      articles: articles ?? this.articles,
      isLoading: isLoading ?? this.isLoading,
      hasMore: hasMore ?? this.hasMore,
      nextPageToken: nextPageToken,
      error: error,
      selectedIds: selectedIds ?? this.selectedIds,
      isMultiSelectMode: isMultiSelectMode ?? this.isMultiSelectMode,
    );
  }
}

class ArticleListParams {
  final String? feedId;
  final String? folderId;
  final bool unreadOnly;
  final bool readOnly;
  final bool? starredOnly;
  final bool? savedForLaterOnly;
  final bool? todayOnly;
  final String? query;
  final SortOrder sortOrder;

  const ArticleListParams({
    this.feedId,
    this.folderId,
    this.unreadOnly = true,
    this.readOnly = false,
    this.starredOnly,
    this.savedForLaterOnly,
    this.todayOnly,
    this.query,
    this.sortOrder = SortOrder.newestFirst,
  });

  @override
  bool operator ==(Object other) =>
      other is ArticleListParams &&
      other.feedId == feedId &&
      other.folderId == folderId &&
      other.unreadOnly == unreadOnly &&
      other.readOnly == readOnly &&
      other.starredOnly == starredOnly &&
      other.savedForLaterOnly == savedForLaterOnly &&
      other.todayOnly == todayOnly &&
      other.query == query &&
      other.sortOrder == sortOrder;

  @override
  int get hashCode => Object.hash(
        feedId, folderId, unreadOnly, readOnly, starredOnly, savedForLaterOnly, todayOnly, query, sortOrder);
}

class ArticleListNotifier extends StateNotifier<ArticleListState> {
  final ArticleService _service;
  final Ref _ref;
  ArticleListParams _params;
  bool _loadingMore = false;

  ArticleListNotifier(this._service, this._ref, this._params)
      : super(const ArticleListState()) {
    _load(refresh: true);
  }

  void updateParams(ArticleListParams params) {
    if (params == _params) return;
    _params = params;
    _load(refresh: true);
  }

  Future<void> refresh() => _load(refresh: true);

  Future<void> loadMore() async {
    if (_loadingMore || !state.hasMore || state.isLoading) return;
    _loadingMore = true;
    try {
      await _load(refresh: false);
    } finally {
      _loadingMore = false;
    }
  }

  Future<void> _load({required bool refresh}) async {
    if (!mounted) return;
    if (state.isLoading) return;
    state = state.copyWith(isLoading: true, error: null);
    try {
      final result = await _service.listArticles(
        feedId: _params.feedId,
        folderId: _params.folderId,
        unreadOnly: _params.unreadOnly ? true : null,
        readOnly: _params.readOnly ? true : null,
        starredOnly: _params.starredOnly,
        savedForLaterOnly: _params.savedForLaterOnly,
        todayOnly: _params.todayOnly,
        query: _params.query,
        sortOrder: _params.sortOrder,
        pageToken: refresh ? null : state.nextPageToken,
      );
      if (!mounted) return;
      final newArticles = refresh
          ? result.articles
          : [...state.articles, ...result.articles];
      state = state.copyWith(
        articles: newArticles,
        isLoading: false,
        hasMore: result.nextPageToken != null,
        nextPageToken: result.nextPageToken,
      );
    } catch (e) {
      if (!mounted) return;
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  void _updateArticle(Article updated) {
    state = state.copyWith(
      articles: state.articles
          .map((a) => a.id == updated.id ? updated : a)
          .toList(),
    );
  }

  Future<void> markAsRead(String id) async {
    if (!mounted) return;
    // Optimistically remove from unread list (or update in-place for read/all views).
    final articleIdx = state.articles.indexWhere((a) => a.id == id);
    if (articleIdx < 0) return;
    final article = state.articles[articleIdx];
    if (_params.unreadOnly) {
      state = state.copyWith(
        articles: state.articles.where((a) => a.id != id).toList(),
      );
      _ref.read(feedProvider.notifier).decrementUnread(article.feedId);
    }
    try {
      final updated = await _service.markAsRead(id);
      if (!mounted) return;
      if (!_params.unreadOnly) {
        _updateArticle(updated);
        if (!article.isRead) {
          _ref.read(feedProvider.notifier).decrementUnread(updated.feedId);
        }
      }
    } catch (_) {
      if (!mounted) return;
      _load(refresh: true);
    }
  }

  Future<void> markAsUnread(String id) async {
    try {
      final updated = await _service.markAsUnread(id);
      _updateArticle(updated);
    } catch (_) {}
  }

  Future<void> markAllAsRead({String? feedId, String? folderId}) async {
    try {
      await _service.markAllAsRead(feedId: feedId, folderId: folderId);
      // Reload from backend so unread-only filter and counts are accurate.
      // Patching local state leaves stale articles visible when unreadOnly=true.
      await _load(refresh: true);
      _ref.read(feedProvider.notifier).loadFeeds();
      _ref.read(folderProvider.notifier).loadFolders();
    } catch (_) {}
  }

  Future<void> markSelectedAsRead() async {
    final ids = state.selectedIds.toList();
    if (ids.isEmpty) return;
    try {
      await _service.markAllAsRead(articleIds: ids);
      state = state.copyWith(
        articles: state.articles
            .map((a) => ids.contains(a.id) ? a.copyWith(isRead: true) : a)
            .toList(),
        selectedIds: {},
        isMultiSelectMode: false,
      );
    } catch (_) {}
  }

  Future<void> toggleStar(String id) async {
    if (!mounted) return;
    final articleIdx = state.articles.indexWhere((a) => a.id == id);
    if (articleIdx < 0) return;
    final article = state.articles[articleIdx];
    // Optimistic update — flip star immediately so UI responds without waiting.
    _updateArticle(article.copyWith(isStarred: !article.isStarred));
    // For starred-only view: remove article if un-starring.
    if (_params.starredOnly == true && article.isStarred) {
      state = state.copyWith(articles: state.articles.where((a) => a.id != id).toList());
    }
    try {
      final updated = await _service.starArticle(id, starred: !article.isStarred);
      if (!mounted) return;
      _updateArticle(updated);
      // Invalidate the starred-screen provider so it reloads next time it's opened.
      _ref.invalidate(articleListProvider(
          const ArticleListParams(starredOnly: true, unreadOnly: false)));
    } catch (_) {
      if (!mounted) return;
      _updateArticle(article);
    }
  }

  Future<void> toggleSave(String id) async {
    if (!mounted) return;
    final articleIdx = state.articles.indexWhere((a) => a.id == id);
    if (articleIdx < 0) return;
    final article = state.articles[articleIdx];
    // Optimistic update — flip saved immediately so UI responds without waiting.
    _updateArticle(article.copyWith(isSavedForLater: !article.isSavedForLater));
    // For saved-only view: remove article if un-saving.
    if (_params.savedForLaterOnly == true && article.isSavedForLater) {
      state = state.copyWith(articles: state.articles.where((a) => a.id != id).toList());
    }
    try {
      final updated = await _service.saveForLater(id, saved: !article.isSavedForLater);
      if (!mounted) return;
      _updateArticle(updated);
      // Invalidate the saved-screen provider so it reloads next time it's opened.
      _ref.invalidate(articleListProvider(
          const ArticleListParams(savedForLaterOnly: true, unreadOnly: false)));
    } catch (_) {
      if (!mounted) return;
      _updateArticle(article);
    }
  }

  void toggleSelect(String id) {
    final newSelected = Set<String>.from(state.selectedIds);
    if (newSelected.contains(id)) {
      newSelected.remove(id);
    } else {
      newSelected.add(id);
    }
    state = state.copyWith(
      selectedIds: newSelected,
      isMultiSelectMode: newSelected.isNotEmpty,
    );
  }

  void clearSelection() {
    state = state.copyWith(selectedIds: {}, isMultiSelectMode: false);
  }

  void toggleUnreadFilter() {
    _params = ArticleListParams(
      feedId: _params.feedId,
      folderId: _params.folderId,
      unreadOnly: !_params.unreadOnly,
      readOnly: _params.readOnly,
      starredOnly: _params.starredOnly,
      savedForLaterOnly: _params.savedForLaterOnly,
      todayOnly: _params.todayOnly,
      query: _params.query,
      sortOrder: _params.sortOrder,
    );
    _load(refresh: true);
  }

  void toggleSortOrder() {
    _params = ArticleListParams(
      feedId: _params.feedId,
      folderId: _params.folderId,
      unreadOnly: _params.unreadOnly,
      readOnly: _params.readOnly,
      starredOnly: _params.starredOnly,
      savedForLaterOnly: _params.savedForLaterOnly,
      todayOnly: _params.todayOnly,
      query: _params.query,
      sortOrder: _params.sortOrder == SortOrder.newestFirst
          ? SortOrder.oldestFirst
          : SortOrder.newestFirst,
    );
    _load(refresh: true);
  }

  bool get isUnreadOnly => _params.unreadOnly;
  SortOrder get sortOrder => _params.sortOrder;
  // True for views where the unread toggle makes no sense (Recently Read, Starred, Saved).
  bool get hasFixedFilter =>
      _params.readOnly ||
      _params.starredOnly == true ||
      _params.savedForLaterOnly == true;
}

// Bump this to force all active article list instances to reload.
// Used by sidebar mark-all-as-read which doesn't know which params are active.
final articleRefreshSignalProvider = StateProvider<int>((ref) => 0);

// Family provider keyed by params
final articleListProvider = StateNotifierProvider.family<ArticleListNotifier,
    ArticleListState, ArticleListParams>((ref, params) {
  ref.watch(articleRefreshSignalProvider); // reload when signal bumps
  return ArticleListNotifier(ref.watch(articleServiceProvider), ref, params);
});
