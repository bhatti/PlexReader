// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/feed.dart';
import '../providers/article_provider.dart';
import '../providers/feed_provider.dart';
import '../providers/api_client_provider.dart';
import '../theme/app_theme.dart';
import '../widgets/article_list.dart';

class FeedScreen extends ConsumerStatefulWidget {
  final String feedId;
  const FeedScreen({super.key, required this.feedId});

  @override
  ConsumerState<FeedScreen> createState() => _FeedScreenState();
}

class _FeedScreenState extends ConsumerState<FeedScreen> {
  @override
  void initState() {
    super.initState();
    // Trigger a background refresh when navigating to a feed.
    // This ensures feeds that have never been fetched (or are stale) load
    // their articles immediately rather than waiting for the scheduler cycle.
    WidgetsBinding.instance.addPostFrameCallback((_) => _maybeRefresh());
  }

  Future<void> _maybeRefresh() async {
    final feeds = ref.read(feedProvider).valueOrNull ?? [];
    final feed = feeds.where((f) => f.id == widget.feedId).firstOrNull;
    // Refresh if never fetched or if there was a prior error.
    if (feed == null) return;
    if (feed.lastFetchedTime != null && feed.lastFetchedTime!.isNotEmpty && !feed.hasError) {
      return; // already fetched successfully, scheduler handles future refreshes
    }
    try {
      await ref.read(feedServiceProvider).refreshFeed(widget.feedId);
      ref.read(feedProvider.notifier).loadFeeds();
    } catch (_) {
      // Best-effort — the user can right-click → Refresh manually
    }
  }

  @override
  Widget build(BuildContext context) {
    final feeds = ref.watch(feedProvider).valueOrNull ?? [];
    final feed = feeds.where((f) => f.id == widget.feedId).firstOrNull;
    final title = feed?.title ?? 'Feed';

    return Column(
      children: [
        if (feed != null && feed.hasError)
          _FeedErrorBanner(feed: feed),
        Expanded(
          child: ArticleList(
            params: ArticleListParams(feedId: widget.feedId, unreadOnly: true),
            title: title,
          ),
        ),
      ],
    );
  }
}

class _FeedErrorBanner extends StatelessWidget {
  final Feed feed;
  const _FeedErrorBanner({required this.feed});

  @override
  Widget build(BuildContext context) {
    return Container(
      color: AppColors.error.withValues(alpha: 0.12),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        children: [
          const Icon(Icons.error_outline, color: AppColors.error, size: 16),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              feed.lastError ?? 'Feed fetch failed',
              style: const TextStyle(color: AppColors.error, fontSize: 13),
              overflow: TextOverflow.ellipsis,
            ),
          ),
        ],
      ),
    );
  }
}
