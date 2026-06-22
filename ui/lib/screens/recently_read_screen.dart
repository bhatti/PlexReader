// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter/material.dart';
import '../models/preferences.dart';
import '../providers/article_provider.dart';
import '../widgets/article_list.dart';

class RecentlyReadScreen extends StatelessWidget {
  const RecentlyReadScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return const ArticleList(
      params: ArticleListParams(
        unreadOnly: false,
        readOnly: true,
        sortOrder: SortOrder.newestFirst,
      ),
      title: 'Recently Read',
    );
  }
}
