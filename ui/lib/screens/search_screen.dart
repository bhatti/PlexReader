// SPDX-License-Identifier: LGPL-2.1-or-later
import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/preferences.dart';
import '../providers/article_provider.dart';
import '../theme/app_theme.dart';
import '../widgets/article_list.dart';
import '../widgets/empty_state.dart';

class SearchScreen extends ConsumerStatefulWidget {
  const SearchScreen({super.key});

  @override
  ConsumerState<SearchScreen> createState() => _SearchScreenState();
}

class _SearchScreenState extends ConsumerState<SearchScreen> {
  final _controller = TextEditingController();
  Timer? _debounce;
  String _query = '';

  @override
  void dispose() {
    _controller.dispose();
    _debounce?.cancel();
    super.dispose();
  }

  void _onChanged(String value) {
    _debounce?.cancel();
    _debounce = Timer(const Duration(milliseconds: 300), () {
      if (mounted) setState(() => _query = value.trim());
    });
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Container(
          color: AppColors.background,
          padding: const EdgeInsets.all(16),
          child: TextField(
            controller: _controller,
            autofocus: true,
            onChanged: _onChanged,
            decoration: InputDecoration(
              hintText: 'Search articles...',
              prefixIcon: const Icon(Icons.search, color: AppColors.textSecondary),
              suffixIcon: _controller.text.isNotEmpty
                  ? IconButton(
                      icon: const Icon(Icons.clear),
                      onPressed: () {
                        _controller.clear();
                        setState(() => _query = '');
                      },
                    )
                  : null,
            ),
          ),
        ),
        const Divider(height: 1),
        Expanded(
          child: _query.isEmpty
              ? const EmptyState(
                  icon: Icons.search,
                  title: 'Search articles',
                  message: 'Type to search across all your feeds',
                )
              : ArticleList(
                  params: ArticleListParams(
                    query: _query,
                    unreadOnly: false,
                    sortOrder: SortOrder.newestFirst,
                  ),
                  title: 'Results for "$_query"',
                ),
        ),
      ],
    );
  }
}
