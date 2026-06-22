// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter/material.dart';
import '../providers/article_provider.dart';
import '../widgets/article_list.dart';

class StarredScreen extends StatelessWidget {
  const StarredScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return const ArticleList(
      params: ArticleListParams(starredOnly: true, unreadOnly: false),
      title: 'Starred',
    );
  }
}
