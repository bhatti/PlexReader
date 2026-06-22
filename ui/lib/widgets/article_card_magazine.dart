// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter/material.dart';
import 'package:timeago/timeago.dart' as timeago;
import '../models/article.dart';
import '../theme/app_theme.dart';

/// Magazine-style article card matching Feedly's layout:
/// square thumbnail on the LEFT, title + summary + meta on the right.
class ArticleCardMagazine extends StatefulWidget {
  final Article article;
  final bool selected;
  final bool checked;
  final bool isMultiSelect;
  final VoidCallback onTap;
  final VoidCallback onLongPress;
  final VoidCallback onToggle;
  final VoidCallback onStar;
  final VoidCallback onSave;
  final VoidCallback onMarkRead;

  const ArticleCardMagazine({
    super.key,
    required this.article,
    required this.selected,
    required this.checked,
    required this.isMultiSelect,
    required this.onTap,
    required this.onLongPress,
    required this.onToggle,
    required this.onStar,
    required this.onSave,
    required this.onMarkRead,
  });

  @override
  State<ArticleCardMagazine> createState() => _ArticleCardMagazineState();
}

class _ArticleCardMagazineState extends State<ArticleCardMagazine> {
  bool _isHovered = false;

  @override
  Widget build(BuildContext context) {
    final article = widget.article;
    final timeStr = _formatTime(article.publishedTime);

    return MouseRegion(
      onEnter: (_) => setState(() => _isHovered = true),
      onExit: (_) => setState(() => _isHovered = false),
      child: GestureDetector(
        onLongPress: widget.onLongPress,
        child: InkWell(
          onTap: widget.isMultiSelect ? widget.onToggle : widget.onTap,
          child: Container(
            decoration: BoxDecoration(
              color: widget.selected
                  ? AppColors.selectedItem
                  : _isHovered
                      ? AppColors.surface
                      : Colors.transparent,
              border: Border(
                left: BorderSide(
                  color: widget.selected ? AppColors.primary : Colors.transparent,
                  width: 2,
                ),
                bottom: const BorderSide(color: AppColors.divider, width: 0.5),
              ),
            ),
            padding: const EdgeInsets.fromLTRB(14, 10, 14, 6),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Multi-select checkbox
                if (widget.isMultiSelect)
                  Padding(
                    padding: const EdgeInsets.only(right: 10, top: 2),
                    child: Checkbox(
                      value: widget.checked,
                      onChanged: (_) => widget.onToggle(),
                    ),
                  ),

                // Thumbnail — always 72×72 on the LEFT, like Feedly.
                // Placeholder keeps columns aligned on articles with no image.
                if (!widget.isMultiSelect) ...[
                  _ArticleThumbnail(url: article.thumbnailUrl, title: article.title),
                  const SizedBox(width: 12),
                ],

                // Text column
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // Feed name · time · unread dot
                      Row(
                        children: [
                          if (article.feedTitle != null)
                            Flexible(
                              child: Text(
                                article.feedTitle!,
                                style: const TextStyle(
                                  color: AppColors.primary,
                                  fontSize: 11.5,
                                  fontWeight: FontWeight.w500,
                                ),
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
                          if (article.feedTitle != null && timeStr != null)
                            const Text(' · ',
                                style: TextStyle(
                                    color: AppColors.textSecondary, fontSize: 11.5)),
                          if (timeStr != null)
                            Text(timeStr,
                                style: const TextStyle(
                                    color: AppColors.textSecondary, fontSize: 11)),
                          if (!article.isRead) ...[
                            const Spacer(),
                            Container(
                              width: 6,
                              height: 6,
                              decoration: const BoxDecoration(
                                color: AppColors.primary,
                                shape: BoxShape.circle,
                              ),
                            ),
                          ],
                        ],
                      ),
                      const SizedBox(height: 3),
                      // Title
                      Text(
                        article.title,
                        style: TextStyle(
                          color: article.isRead
                              ? AppColors.textSecondary
                              : AppColors.textPrimary,
                          fontSize: 14,
                          fontWeight:
                              article.isRead ? FontWeight.normal : FontWeight.w600,
                          height: 1.3,
                        ),
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                      ),
                      // Summary
                      if (article.summary != null &&
                          article.summary!.isNotEmpty) ...[
                        const SizedBox(height: 3),
                        Text(
                          _stripHtml(article.summary!),
                          style: const TextStyle(
                            color: AppColors.textSecondary,
                            fontSize: 12.5,
                            height: 1.4,
                          ),
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ],
                      // Action icons row — always rendered so height is stable.
                      // Star and bookmark are always visible (colored = active).
                      // Mark-read only appears on hover (it removes the article).
                      SizedBox(
                        height: 28,
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            // Star — always visible, gold when starred
                            _actionIcon(
                              article.isStarred ? Icons.star : Icons.star_outline,
                              article.isStarred
                                  ? AppColors.star
                                  : (_isHovered ? AppColors.textSecondary : AppColors.textSecondary.withValues(alpha: 0.35)),
                              widget.onStar,
                              tooltip: article.isStarred ? 'Unstar' : 'Star',
                            ),
                            // Bookmark — always visible, primary color when saved
                            _actionIcon(
                              article.isSavedForLater
                                  ? Icons.bookmark
                                  : Icons.bookmark_outline,
                              article.isSavedForLater
                                  ? AppColors.primary
                                  : (_isHovered ? AppColors.textSecondary : AppColors.textSecondary.withValues(alpha: 0.35)),
                              widget.onSave,
                              tooltip: article.isSavedForLater ? 'Unsave' : 'Save for later',
                            ),
                            // Mark read — only show on hover (removes from unread list)
                            if (_isHovered || widget.isMultiSelect || _isMobile(context))
                              _actionIcon(
                                article.isRead
                                    ? Icons.radio_button_unchecked
                                    : Icons.check_circle_outline,
                                AppColors.textSecondary,
                                widget.onMarkRead,
                                tooltip: article.isRead ? 'Mark unread' : 'Mark read',
                              ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _actionIcon(IconData icon, Color color, VoidCallback onTap,
      {String? tooltip}) {
    return Tooltip(
      message: tooltip ?? '',
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(4),
        child: Padding(
          padding: const EdgeInsets.all(4),
          child: Icon(icon, size: 16, color: color),
        ),
      ),
    );
  }

  bool _isMobile(BuildContext context) =>
      MediaQuery.sizeOf(context).width < 600;

  String? _formatTime(String? iso) {
    if (iso == null) return null;
    try {
      return timeago.format(DateTime.parse(iso).toLocal());
    } catch (_) {
      return null;
    }
  }

  String _stripHtml(String html) =>
      html.replaceAll(RegExp(r'<[^>]*>'), '').trim();
}

// ── Thumbnail ────────────────────────────────────────────────────────────────
// 72×72 square. Shows the article image; falls back to a colored initial-letter
// placeholder so every row stays aligned regardless of image availability.

class _ArticleThumbnail extends StatelessWidget {
  final String? url;
  final String title;

  const _ArticleThumbnail({required this.url, required this.title});

  @override
  Widget build(BuildContext context) {
    if (url != null && url!.isNotEmpty) {
      return ClipRRect(
        borderRadius: BorderRadius.circular(6),
        child: Image.network(
          url!,
          width: 72,
          height: 72,
          fit: BoxFit.cover,
          errorBuilder: (_, __, ___) => _placeholder(),
        ),
      );
    }
    return _placeholder();
  }

  Widget _placeholder() {
    const palette = [
      Color(0xFF1565C0), Color(0xFF2E7D32), Color(0xFF6A1B9A),
      Color(0xFFE65100), Color(0xFF00695C), Color(0xFFAD1457),
      Color(0xFF4527A0), Color(0xFF37474F),
    ];
    final letter = title.isNotEmpty ? title[0].toUpperCase() : '?';
    final idx = title.codeUnits.fold(0, (a, b) => a + b) % palette.length;
    return Container(
      width: 72,
      height: 72,
      decoration: BoxDecoration(
        color: palette[idx],
        borderRadius: BorderRadius.circular(6),
      ),
      alignment: Alignment.center,
      child: Text(
        letter,
        style: const TextStyle(
          color: Colors.white,
          fontSize: 28,
          fontWeight: FontWeight.w700,
        ),
      ),
    );
  }
}
