// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/article_provider.dart';
import '../providers/folder_provider.dart';
import '../widgets/article_list.dart';

class FolderScreen extends ConsumerWidget {
  final String folderId;
  const FolderScreen({super.key, required this.folderId});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final folders = ref.watch(folderProvider).valueOrNull ?? [];
    final folder = folders.where((f) => f.id == folderId).firstOrNull;
    final title = folder?.name ?? 'Folder';

    return ArticleList(
      params: ArticleListParams(folderId: folderId, unreadOnly: true),
      title: title,
    );
  }
}
