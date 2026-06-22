// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../services/api_client.dart';
import '../services/folder_service.dart';
import '../services/feed_service.dart';
import '../services/article_service.dart';
import '../services/preferences_service.dart';
import '../services/version_service.dart';

final apiClientProvider = Provider<ApiClient>((ref) {
  final client = ApiClient();
  ref.onDispose(client.dispose);
  return client;
});

final folderServiceProvider = Provider<FolderService>((ref) {
  return FolderService(ref.watch(apiClientProvider));
});

final feedServiceProvider = Provider<FeedService>((ref) {
  return FeedService(ref.watch(apiClientProvider));
});

final articleServiceProvider = Provider<ArticleService>((ref) {
  return ArticleService(ref.watch(apiClientProvider));
});

final preferencesServiceProvider = Provider<PreferencesService>((ref) {
  return PreferencesService(ref.watch(apiClientProvider));
});

final versionServiceProvider = Provider<VersionService>((ref) {
  final client = ref.watch(apiClientProvider);
  return VersionService(client.baseUrl);
});

final versionProvider = FutureProvider<String>((ref) async {
  return ref.watch(versionServiceProvider).getVersion();
});
