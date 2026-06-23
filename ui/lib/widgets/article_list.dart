// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:url_launcher/url_launcher.dart';
import '../models/article.dart';
import '../models/preferences.dart';
import '../providers/article_provider.dart';
import '../providers/feed_provider.dart';
import '../providers/navigation_provider.dart';
import '../theme/app_theme.dart';
import 'article_card_magazine.dart';
import 'article_card_title.dart';
import 'article_card_grid.dart';
import 'article_detail.dart';
import 'empty_state.dart';
import 'error_state.dart';

class ArticleList extends ConsumerStatefulWidget {
  final ArticleListParams params;
  final String title;
  final int? unreadCount;
  final VoidCallback? onMarkAllRead;
  // feedId/folderId for the favorite toggle — null means no favorite button
  final String? favoritesFeedId;
  final bool isFavorite;
  // null = feed has never been fetched (show spinner instead of "All caught up")
  final String? lastFetchedTime;

  const ArticleList({
    super.key,
    required this.params,
    required this.title,
    this.unreadCount,
    this.onMarkAllRead,
    this.favoritesFeedId,
    this.isFavorite = false,
    this.lastFetchedTime,
  });

  @override
  ConsumerState<ArticleList> createState() => _ArticleListState();
}

class _ArticleListState extends ConsumerState<ArticleList> {
  late ArticleListParams _params;
  final ScrollController _scrollController = ScrollController();

  @override
  void initState() {
    super.initState();
    _params = widget.params;
    _scrollController.addListener(_onScroll);
  }

  @override
  void didUpdateWidget(ArticleList oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.params != oldWidget.params) {
      _params = widget.params;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        ref.read(articleListProvider(_params).notifier).updateParams(_params);
      });
    }
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  void _onScroll() {
    if (_scrollController.position.pixels >=
        _scrollController.position.maxScrollExtent - 300) {
      ref.read(articleListProvider(_params).notifier).loadMore();
    }
  }

  @override
  Widget build(BuildContext context) {
    final viewMode = ref.watch(viewModeProvider);
    final articleState = ref.watch(articleListProvider(_params));
    final notifier = ref.read(articleListProvider(_params).notifier);
    final selectedId = ref.watch(selectedArticleIdProvider);

    return Column(
      children: [
        _buildToolbar(context, viewMode, articleState, notifier),
        const Divider(height: 1),
        if (articleState.isMultiSelectMode)
          _buildMultiSelectBar(articleState, notifier),
        Expanded(
          child: selectedId != null
              ? Row(
                  children: [
                    SizedBox(
                      width: 420,
                      child: _buildListContent(
                          context, viewMode, articleState, notifier, selectedId),
                    ),
                    Container(width: 1, color: AppColors.divider),
                    Expanded(
                      child: ArticleDetail(
                        articleId: selectedId,
                        params: _params,
                      ),
                    ),
                  ],
                )
              : RefreshIndicator(
                  onRefresh: () => notifier.refresh(),
                  color: AppColors.primary,
                  backgroundColor: AppColors.surface,
                  child: _buildListContent(
                      context, viewMode, articleState, notifier, null),
                ),
        ),
      ],
    );
  }

  Widget _buildToolbar(BuildContext context, ViewMode viewMode,
      ArticleListState articleState, ArticleListNotifier notifier) {
    final unread = widget.unreadCount;
    return Container(
      color: AppColors.surface,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          Row(
            children: [
              // Title + unread count (like Feedly)
              Expanded(
                child: Row(
                  children: [
                    Flexible(
                      child: Text(
                        widget.title,
                        style: const TextStyle(
                          color: AppColors.textPrimary,
                          fontSize: 20,
                          fontWeight: FontWeight.w700,
                          letterSpacing: -0.3,
                        ),
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    if (unread != null && unread > 0) ...[
                      const SizedBox(width: 8),
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 2),
                        decoration: BoxDecoration(
                          color: AppColors.unreadBadge,
                          borderRadius: BorderRadius.circular(10),
                        ),
                        child: Text(
                          unread > 999 ? '999+' : '$unread',
                          style: const TextStyle(
                            color: AppColors.textSecondary,
                            fontSize: 12,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ),
                    ],
                    // Favorite star — only for feed screens
                    if (widget.favoritesFeedId != null) ...[
                      const SizedBox(width: 6),
                      Tooltip(
                        message: widget.isFavorite ? 'Remove from Favorites' : 'Add to Favorites',
                        child: InkWell(
                          onTap: () {
                            final feed = ref.read(feedProvider).valueOrNull
                                ?.where((f) => f.id == widget.favoritesFeedId)
                                .firstOrNull;
                            if (feed != null) {
                              ref.read(feedProvider.notifier).toggleFavorite(feed);
                            }
                          },
                          borderRadius: BorderRadius.circular(4),
                          child: Padding(
                            padding: const EdgeInsets.all(3),
                            child: Icon(
                              widget.isFavorite ? Icons.star : Icons.star_border,
                              size: 20,
                              color: widget.isFavorite ? AppColors.star : AppColors.textSecondary,
                            ),
                          ),
                        ),
                      ),
                    ],
                  ],
                ),
              ),
              // Sort toggle
              Tooltip(
                message: notifier.sortOrder == SortOrder.newestFirst
                    ? 'Newest first'
                    : 'Oldest first',
                child: IconButton(
                  icon: Icon(
                    notifier.sortOrder == SortOrder.newestFirst
                        ? Icons.arrow_downward
                        : Icons.arrow_upward,
                    size: 18,
                  ),
                  color: AppColors.textSecondary,
                  onPressed: () => notifier.toggleSortOrder(),
                  padding: EdgeInsets.zero,
                  visualDensity: VisualDensity.compact,
                ),
              ),
              // Unread filter toggle
              if (!notifier.hasFixedFilter)
                Tooltip(
                  message: notifier.isUnreadOnly ? 'Showing unread' : 'Showing all',
                  child: IconButton(
                    icon: Icon(
                      notifier.isUnreadOnly
                          ? Icons.radio_button_checked
                          : Icons.radio_button_unchecked,
                      size: 18,
                    ),
                    color: notifier.isUnreadOnly
                        ? AppColors.primary
                        : AppColors.textSecondary,
                    onPressed: () => notifier.toggleUnreadFilter(),
                    padding: EdgeInsets.zero,
                    visualDensity: VisualDensity.compact,
                  ),
                ),
              // Mark all read — Feedly-style checkmark with count
              if (!widget.params.readOnly) ...[
                const SizedBox(width: 2),
                Tooltip(
                  message: 'Mark all as read',
                  child: InkWell(
                    onTap: () => _confirmMarkAllRead(context, notifier),
                    borderRadius: BorderRadius.circular(6),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 4),
                      decoration: BoxDecoration(
                        border: Border.all(color: AppColors.divider),
                        borderRadius: BorderRadius.circular(6),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          const Icon(Icons.check, size: 14, color: AppColors.textSecondary),
                          if (unread != null && unread > 0) ...[
                            const SizedBox(width: 3),
                            Text(
                              unread > 999 ? '999+' : '$unread',
                              style: const TextStyle(
                                  color: AppColors.textSecondary, fontSize: 11),
                            ),
                          ],
                        ],
                      ),
                    ),
                  ),
                ),
              ],
              const SizedBox(width: 2),
              // View mode toggles
              _viewModeButton(Icons.article_outlined, ViewMode.magazine, viewMode),
              _viewModeButton(Icons.format_list_bulleted, ViewMode.titleOnly, viewMode),
              _viewModeButton(Icons.grid_view, ViewMode.cards, viewMode),
              // Refresh
              IconButton(
                icon: const Icon(Icons.refresh, size: 18),
                tooltip: 'Refresh',
                color: AppColors.textSecondary,
                onPressed: () => notifier.refresh(),
                padding: EdgeInsets.zero,
                visualDensity: VisualDensity.compact,
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _viewModeButton(IconData icon, ViewMode mode, ViewMode current) {
    return IconButton(
      icon: Icon(icon, size: 18),
      color: current == mode ? AppColors.primary : AppColors.textSecondary,
      onPressed: () => ref.read(viewModeProvider.notifier).state = mode,
    );
  }

  Widget _buildMultiSelectBar(
      ArticleListState state, ArticleListNotifier notifier) {
    return Container(
      color: AppColors.surface,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        children: [
          Text(
            '${state.selectedIds.length} selected',
            style: const TextStyle(color: AppColors.textPrimary, fontSize: 14),
          ),
          const Spacer(),
          TextButton(
            onPressed: () => notifier.markSelectedAsRead(),
            child: const Text('Mark Read'),
          ),
          TextButton(
            onPressed: () => notifier.clearSelection(),
            child: const Text('Cancel'),
          ),
        ],
      ),
    );
  }

  Widget _buildListContent(
    BuildContext context,
    ViewMode viewMode,
    ArticleListState state,
    ArticleListNotifier notifier,
    String? selectedId,
  ) {
    if (state.articles.isEmpty && state.isLoading) {
      return const Center(
        child: CircularProgressIndicator(color: AppColors.primary),
      );
    }

    if (state.error != null && state.articles.isEmpty) {
      return ErrorState(
        message: state.error!,
        onRetry: () => notifier.refresh(),
      );
    }

    if (state.articles.isEmpty) {
      // Feed has never been fetched — initial background refresh still running.
      final neverFetched = widget.params.feedId != null &&
          (widget.lastFetchedTime == null || widget.lastFetchedTime!.isEmpty);
      if (neverFetched) {
        return const Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              CircularProgressIndicator(color: AppColors.primary),
              SizedBox(height: 16),
              Text(
                'Fetching articles…',
                style: TextStyle(color: AppColors.textSecondary, fontSize: 14),
              ),
            ],
          ),
        );
      }
      return const EmptyState(
        icon: Icons.inbox_outlined,
        title: 'All caught up!',
        message: 'No articles to show.',
      );
    }

    if (viewMode == ViewMode.cards) {
      return _buildGridList(context, state, notifier, selectedId);
    }

    return ListView.builder(
      controller: _scrollController,
      itemCount: state.articles.length + (state.hasMore ? 1 : 0),
      itemBuilder: (context, index) {
        if (index == state.articles.length) {
          return const Padding(
            padding: EdgeInsets.all(16),
            child: Center(
              child: CircularProgressIndicator(
                color: AppColors.primary,
                strokeWidth: 2,
              ),
            ),
          );
        }
        final article = state.articles[index];
        final isSelected = article.id == selectedId;
        final isChecked = state.selectedIds.contains(article.id);

        return viewMode == ViewMode.titleOnly
            ? ArticleCardTitle(
                article: article,
                selected: isSelected,
                checked: isChecked,
                isMultiSelect: state.isMultiSelectMode,
                onTap: () => _openArticle(article, notifier),
                onLongPress: () => notifier.toggleSelect(article.id),
                onToggle: () => notifier.toggleSelect(article.id),
              )
            : ArticleCardMagazine(
                article: article,
                selected: isSelected,
                checked: isChecked,
                isMultiSelect: state.isMultiSelectMode,
                onTap: () => _openArticle(article, notifier),
                onLongPress: () => notifier.toggleSelect(article.id),
                onToggle: () => notifier.toggleSelect(article.id),
                onStar: () => notifier.toggleStar(article.id),
                onSave: () => notifier.toggleSave(article.id),
                onMarkRead: () => notifier.markAsRead(article.id),
              );
      },
    );
  }

  Widget _buildGridList(
    BuildContext context,
    ArticleListState state,
    ArticleListNotifier notifier,
    String? selectedId,
  ) {
    return GridView.builder(
      controller: _scrollController,
      padding: const EdgeInsets.all(12),
      gridDelegate: const SliverGridDelegateWithMaxCrossAxisExtent(
        maxCrossAxisExtent: 320,
        mainAxisExtent: 260,
        mainAxisSpacing: 12,
        crossAxisSpacing: 12,
      ),
      itemCount: state.articles.length + (state.hasMore ? 1 : 0),
      itemBuilder: (context, index) {
        if (index == state.articles.length) {
          return const Center(
              child: CircularProgressIndicator(
                  color: AppColors.primary, strokeWidth: 2));
        }
        final article = state.articles[index];
        return ArticleCardGrid(
          article: article,
          selected: article.id == selectedId,
          onTap: () => _openArticle(article, notifier),
          onStar: () => notifier.toggleStar(article.id),
        );
      },
    );
  }

  void _openArticle(Article article, ArticleListNotifier notifier) {
    // Cmd+click (Mac) or Ctrl+click (Win/Linux) opens the article link directly
    // in a new browser tab, matching Feedly's behaviour.
    final meta = HardwareKeyboard.instance.isMetaPressed;
    final ctrl = HardwareKeyboard.instance.isControlPressed;
    if ((meta || ctrl) && article.link != null && article.link!.isNotEmpty) {
      final uri = Uri.tryParse(article.link!);
      if (uri != null) {
        launchUrl(uri, mode: LaunchMode.externalApplication);
      }
      return;
    }
    // Do NOT auto-mark as read on select — user must explicitly use Mark Read.
    ref.read(selectedArticleIdProvider.notifier).state = article.id;
  }

  void _confirmMarkAllRead(
      BuildContext context, ArticleListNotifier notifier) {
    showDialog(
      context: context,
      useRootNavigator: true,
      builder: (dialogCtx) => AlertDialog(
        title: const Text('Mark All as Read'),
        content: const Text('Mark all visible articles as read?'),
        actions: [
          TextButton(
              onPressed: () => Navigator.of(dialogCtx).pop(),
              child: const Text('Cancel')),
          ElevatedButton(
            onPressed: () {
              Navigator.of(dialogCtx).pop();
              notifier.markAllAsRead(
                feedId: _params.feedId,
                folderId: _params.folderId,
              );
              widget.onMarkAllRead?.call();
            },
            child: const Text('Mark All Read'),
          ),
        ],
      ),
    );
  }
}
