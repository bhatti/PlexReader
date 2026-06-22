// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter/material.dart';
import 'package:timeago/timeago.dart' as timeago;
import '../models/article.dart';
import '../theme/app_theme.dart';

/// Compact title-only row card for the title-only view mode.
/// Shows unread dot, title, feed name, and time. Hover reveals star/save icons.
class ArticleCardTitle extends StatefulWidget {
  final Article article;
  final bool selected;
  final bool checked;
  final bool isMultiSelect;
  final VoidCallback onTap;
  final VoidCallback onLongPress;
  final VoidCallback onToggle;

  const ArticleCardTitle({
    super.key,
    required this.article,
    required this.selected,
    required this.checked,
    required this.isMultiSelect,
    required this.onTap,
    required this.onLongPress,
    required this.onToggle,
  });

  @override
  State<ArticleCardTitle> createState() => _ArticleCardTitleState();
}

class _ArticleCardTitleState extends State<ArticleCardTitle> {
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
            height: 40,
            decoration: BoxDecoration(
              color: widget.selected
                  ? AppColors.selectedItem
                  : _isHovered
                      ? AppColors.surface
                      : Colors.transparent,
              border: Border(
                left: BorderSide(
                  color: widget.selected
                      ? AppColors.primary
                      : Colors.transparent,
                  width: 2,
                ),
                bottom: const BorderSide(
                    color: AppColors.divider, width: 0.5),
              ),
            ),
            padding: const EdgeInsets.symmetric(horizontal: 12),
            child: Row(
              children: [
                // Checkbox or unread dot
                if (widget.isMultiSelect)
                  Padding(
                    padding: const EdgeInsets.only(right: 8),
                    child: Checkbox(
                      value: widget.checked,
                      onChanged: (_) => widget.onToggle(),
                    ),
                  )
                else ...[
                  Container(
                    width: 6,
                    height: 6,
                    margin: const EdgeInsets.only(right: 8),
                    decoration: BoxDecoration(
                      color: article.isRead
                          ? Colors.transparent
                          : AppColors.primary,
                      shape: BoxShape.circle,
                    ),
                  ),
                ],
                // Title
                Expanded(
                  child: Text(
                    article.title,
                    style: TextStyle(
                      color: article.isRead
                          ? AppColors.textSecondary
                          : AppColors.textPrimary,
                      fontSize: 13,
                      fontWeight:
                          article.isRead ? FontWeight.normal : FontWeight.w500,
                    ),
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                // Feed name
                if (article.feedTitle != null && !_isHovered)
                  Padding(
                    padding: const EdgeInsets.only(left: 8),
                    child: Text(
                      article.feedTitle!,
                      style: const TextStyle(
                          color: AppColors.textSecondary, fontSize: 11),
                    ),
                  ),
                // Time
                if (timeStr != null && !_isHovered)
                  Padding(
                    padding: const EdgeInsets.only(left: 8),
                    child: Text(
                      timeStr,
                      style: const TextStyle(
                          color: AppColors.textSecondary, fontSize: 11),
                    ),
                  ),
                // Hover actions (replace time/feed labels)
                if (_isHovered) ...[
                  if (article.isStarred)
                    _icon(Icons.star, AppColors.star, () {})
                  else
                    _icon(Icons.star_outline, AppColors.textSecondary, () {}),
                  if (article.isSavedForLater)
                    _icon(Icons.bookmark, AppColors.primary, () {})
                  else
                    _icon(
                        Icons.bookmark_outline, AppColors.textSecondary, () {}),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _icon(IconData icon, Color color, VoidCallback onTap) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(4),
      child: Padding(
        padding: const EdgeInsets.all(4),
        child: Icon(icon, size: 15, color: color),
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
