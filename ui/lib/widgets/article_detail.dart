// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter/material.dart';
import 'package:flutter_html/flutter_html.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:intl/intl.dart';
import '../models/article.dart';
import '../providers/article_provider.dart';
import '../providers/navigation_provider.dart';
import '../theme/app_theme.dart';

/// Full-pane article detail widget with toolbar (star, save, mark read, open in browser).
/// Used in the split-pane layout from [ArticleList].
class ArticleDetail extends ConsumerWidget {
  final String articleId;
  final ArticleListParams params;

  const ArticleDetail({
    super.key,
    required this.articleId,
    required this.params,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(articleListProvider(params));
    final article = state.articles.cast<Article?>().firstWhere(
          (a) => a?.id == articleId,
          orElse: () => null,
        );

    if (article == null) {
      return const Center(
        child: Text(
          'Select an article',
          style: TextStyle(color: AppColors.textSecondary),
        ),
      );
    }

    return _ArticleDetailPane(article: article, params: params);
  }
}

class _ArticleDetailPane extends ConsumerWidget {
  final Article article;
  final ArticleListParams params;

  const _ArticleDetailPane({required this.article, required this.params});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final notifier = ref.read(articleListProvider(params).notifier);

    return Column(
      children: [
        // Toolbar
        Container(
          height: 48,
          padding: const EdgeInsets.symmetric(horizontal: 8),
          decoration: const BoxDecoration(
            color: AppColors.surface,
            border: Border(bottom: BorderSide(color: AppColors.divider)),
          ),
          child: Row(
            children: [
              IconButton(
                icon: const Icon(Icons.close, size: 18),
                onPressed: () =>
                    ref.read(selectedArticleIdProvider.notifier).state = null,
                tooltip: 'Close',
              ),
              const Spacer(),
              IconButton(
                icon: Icon(
                  article.isStarred ? Icons.star : Icons.star_outline,
                  size: 18,
                  color: article.isStarred ? AppColors.star : null,
                ),
                onPressed: () => notifier.toggleStar(article.id),
                tooltip: article.isStarred ? 'Unstar' : 'Star',
              ),
              IconButton(
                icon: Icon(
                  article.isSavedForLater
                      ? Icons.bookmark
                      : Icons.bookmark_outline,
                  size: 18,
                  color: article.isSavedForLater ? AppColors.primary : null,
                ),
                onPressed: () => notifier.toggleSave(article.id),
                tooltip: article.isSavedForLater ? 'Unsave' : 'Save for later',
              ),
              if (article.isRead)
                IconButton(
                  icon: const Icon(Icons.mark_email_unread_outlined, size: 18),
                  onPressed: () => notifier.markAsUnread(article.id),
                  tooltip: 'Mark unread',
                )
              else
                IconButton(
                  icon: const Icon(Icons.done, size: 18),
                  onPressed: () => notifier.markAsRead(article.id),
                  tooltip: 'Mark read',
                ),
              if (article.link != null && article.link!.isNotEmpty)
                IconButton(
                  icon: const Icon(Icons.open_in_new, size: 18),
                  onPressed: () => _openInBrowser(article.link!),
                  tooltip: 'Open in browser',
                ),
            ],
          ),
        ),
        // Scrollable content
        Expanded(
          child: ArticleDetailContent(article: article),
        ),
      ],
    );
  }

  void _openInBrowser(String url) {
    launchUrl(Uri.parse(url), mode: LaunchMode.externalApplication);
  }
}

/// Standalone scrollable article content — title, meta, HTML body.
/// Can be embedded in a detail pane or dialog.
class ArticleDetailContent extends StatelessWidget {
  final Article article;

  const ArticleDetailContent({super.key, required this.article});

  @override
  Widget build(BuildContext context) {
    final content = article.content ?? article.summary ?? '';
    final published = _formatDate(article.publishedTime);

    return SelectionArea(
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Feed name
            if (article.feedTitle != null)
              Text(
                article.feedTitle!,
                style: const TextStyle(
                  color: AppColors.primary,
                  fontSize: 12,
                  fontWeight: FontWeight.w500,
                ),
              ),
            const SizedBox(height: 8),
            // Title
            Text(
              article.title,
              style: const TextStyle(
                color: AppColors.textPrimary,
                fontSize: 22,
                fontWeight: FontWeight.w700,
                height: 1.3,
              ),
            ),
            const SizedBox(height: 8),
            // Meta row
            Wrap(
              children: [
                if (article.author != null && article.author!.isNotEmpty) ...[
                  Text(
                    article.author!,
                    style: const TextStyle(
                        color: AppColors.textSecondary, fontSize: 13),
                  ),
                  const Text(' · ',
                      style: TextStyle(color: AppColors.textSecondary)),
                ],
                if (published != null)
                  Text(
                    published,
                    style: const TextStyle(
                        color: AppColors.textSecondary, fontSize: 13),
                  ),
              ],
            ),
            // Link
            if (article.link != null && article.link!.isNotEmpty) ...[
              const SizedBox(height: 8),
              InkWell(
                onTap: () => _openLink(article.link!),
                child: Text(
                  article.link!,
                  style: const TextStyle(
                    color: AppColors.primary,
                    fontSize: 12,
                    decoration: TextDecoration.underline,
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
            // Hero image
            if (article.thumbnailUrl != null &&
                article.thumbnailUrl!.isNotEmpty) ...[
              const SizedBox(height: 16),
              ClipRRect(
                borderRadius: BorderRadius.circular(8),
                child: Image.network(
                  article.thumbnailUrl!,
                  width: double.infinity,
                  fit: BoxFit.contain,
                  errorBuilder: (_, __, ___) => const SizedBox.shrink(),
                ),
              ),
            ],
            const SizedBox(height: 20),
            const Divider(),
            const SizedBox(height: 16),
            // HTML content
            if (content.isNotEmpty)
              Html(
                data: content,
                style: {
                  'body': Style(
                    color: AppColors.textPrimary,
                    fontSize: FontSize(15),
                    lineHeight: LineHeight.em(1.7),
                    margin: Margins.zero,
                    padding: HtmlPaddings.zero,
                  ),
                  'h1': Style(
                      color: AppColors.textPrimary,
                      fontSize: FontSize(20),
                      fontWeight: FontWeight.w700),
                  'h2': Style(
                      color: AppColors.textPrimary,
                      fontSize: FontSize(18),
                      fontWeight: FontWeight.w600),
                  'h3': Style(
                      color: AppColors.textPrimary,
                      fontSize: FontSize(16),
                      fontWeight: FontWeight.w600),
                  'a': Style(color: AppColors.primary),
                  'p': Style(
                      color: AppColors.textPrimary,
                      margin: Margins.only(bottom: 12)),
                  'img': Style(width: Width(100, Unit.percent)),
                  'blockquote': Style(
                    color: AppColors.textSecondary,
                    border: const Border(
                      left: BorderSide(color: AppColors.primary, width: 3),
                    ),
                    padding: HtmlPaddings.only(left: 12),
                  ),
                  'code': Style(
                    backgroundColor: AppColors.surfaceVariant,
                    color: AppColors.textPrimary,
                    fontFamily: 'monospace',
                  ),
                  'pre': Style(
                    backgroundColor: AppColors.surfaceVariant,
                    padding: HtmlPaddings.all(12),
                  ),
                },
                onLinkTap: (url, _, __) {
                  if (url != null) _openLink(url);
                },
              )
            else
              const Text(
                'No content available.',
                style: TextStyle(color: AppColors.textSecondary, fontSize: 14),
              ),
            const SizedBox(height: 24),
            if (article.link != null && article.link!.isNotEmpty)
              OutlinedButton.icon(
                onPressed: () => _openLink(article.link!),
                icon: const Icon(Icons.open_in_new, size: 16),
                label: const Text('Read full article'),
              ),
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }

  String? _formatDate(String? iso) {
    if (iso == null) return null;
    try {
      final dt = DateTime.parse(iso).toLocal();
      return DateFormat('MMM d, yyyy · h:mm a').format(dt);
    } catch (_) {
      return null;
    }
  }

  Future<void> _openLink(String url) async {
    final uri = Uri.tryParse(url);
    if (uri != null && await canLaunchUrl(uri)) {
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    }
  }
}
