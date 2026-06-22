// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter/material.dart';
import '../providers/article_provider.dart';
import '../widgets/article_list.dart';

class TodayScreen extends StatelessWidget {
  const TodayScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return ArticleList(
      params: const ArticleListParams(todayOnly: true, unreadOnly: false),
      title: 'Today',
    );
  }
}
