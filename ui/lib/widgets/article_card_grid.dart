// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:timeago/timeago.dart' as timeago;
import '../models/article.dart';
import '../theme/app_theme.dart';

/// Grid card with large thumbnail on top, title, feed name, and time below.
/// Used in the cards view mode.
class ArticleCardGrid extends StatelessWidget {
  final Article article;
  final bool selected;
  final VoidCallback onTap;
  final VoidCallback onStar;

  const ArticleCardGrid({
    super.key,
    required this.article,
    required this.selected,
    required this.onTap,
    required this.onStar,
  });

  @override
  Widget build(BuildContext context) {
    final timeStr = _formatTime(article.publishedTime);

    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(8),
      child: Container(
        decoration: BoxDecoration(
          color: AppColors.surface,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(
            color: selected ? AppColors.primary : AppColors.divider,
            width: selected ? 1.5 : 1,
          ),
        ),
        clipBehavior: Clip.hardEdge,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Thumbnail
            article.thumbnailUrl != null && article.thumbnailUrl!.isNotEmpty
                ? CachedNetworkImage(
                    imageUrl: article.thumbnailUrl!,
                    height: 130,
                    width: double.infinity,
                    fit: BoxFit.cover,
                    errorWidget: (_, __, ___) => _placeholder(),
                  )
                : _placeholder(),
            // Body
            Expanded(
              child: Padding(
                padding: const EdgeInsets.all(8),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Feed name
                    if (article.feedTitle != null)
                      Text(
                        article.feedTitle!,
                        style: const TextStyle(
                          color: AppColors.primary,
                          fontSize: 11,
                          fontWeight: FontWeight.w500,
                        ),
                        overflow: TextOverflow.ellipsis,
                      ),
                    const SizedBox(height: 3),
                    // Title
                    Expanded(
                      child: Text(
                        article.title,
                        style: TextStyle(
                          color: article.isRead
                              ? AppColors.textSecondary
                              : AppColors.textPrimary,
                          fontSize: 13,
                          fontWeight: article.isRead
                              ? FontWeight.normal
                              : FontWeight.w600,
                          height: 1.3,
                        ),
                        maxLines: 3,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    // Bottom row: time + star
                    Row(
                      children: [
                        if (timeStr != null)
                          Expanded(
                            child: Text(
                              timeStr,
                              style: const TextStyle(
                                  color: AppColors.textSecondary, fontSize: 11),
                              overflow: TextOverflow.ellipsis,
                            ),
                          ),
                        IconButton(
                          icon: Icon(
                            article.isStarred
                                ? Icons.star
                                : Icons.star_outline,
                            size: 15,
                            color: article.isStarred
                                ? AppColors.star
                                : AppColors.textSecondary,
                          ),
                          onPressed: onStar,
                          padding: EdgeInsets.zero,
                          constraints: const BoxConstraints(
                              minWidth: 24, minHeight: 24),
                          tooltip: article.isStarred ? 'Unstar' : 'Star',
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _placeholder() {
    return Container(
      height: 130,
      color: AppColors.surfaceVariant,
      child: const Center(
        child: Icon(Icons.article_outlined,
            size: 40, color: AppColors.textSecondary),
      ),
    );
  }

  String? _formatTime(String? iso) {
    if (iso == null) return null;
    try {
      return timeago.format(DateTime.parse(iso).toLocal());
    } catch (_) {
      return null;
    }
  }
}
