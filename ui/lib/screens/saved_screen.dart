// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter/material.dart';
import '../providers/article_provider.dart';
import '../widgets/article_list.dart';

class SavedScreen extends StatelessWidget {
  const SavedScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return ArticleList(
      params: const ArticleListParams(savedForLaterOnly: true, unreadOnly: false),
      title: 'Saved for Later',
    );
  }
}
