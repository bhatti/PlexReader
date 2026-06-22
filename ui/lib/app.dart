// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'theme/app_theme.dart';
import 'screens/app_shell.dart';
import 'screens/today_screen.dart';
import 'screens/all_articles_screen.dart';
import 'screens/feed_screen.dart';
import 'screens/folder_screen.dart';
import 'screens/search_screen.dart';
import 'screens/saved_screen.dart';
import 'screens/recently_read_screen.dart';
import 'screens/starred_screen.dart';
import 'screens/preferences_screen.dart';
import 'models/preferences.dart';
import 'providers/navigation_provider.dart';
import 'providers/preferences_provider.dart';

// Exposed so sidebar widgets can show dialogs without needing a BuildContext
// from an already-disposed widget (e.g. after popup menu selection triggers
// a hover-state change that unmounts the _DotsMenu widget).
final rootNavigatorKey = GlobalKey<NavigatorState>();
final _shellNavigatorKey = GlobalKey<NavigatorState>();

final routerProvider = Provider<GoRouter>((ref) {
  final router = GoRouter(
    navigatorKey: rootNavigatorKey,
    initialLocation: '/today',
    observers: [_ArticleSelectionClearer(ref)],
    routes: [
      ShellRoute(
        navigatorKey: _shellNavigatorKey,
        builder: (context, state, child) => AppShell(child: child),
        routes: [
          GoRoute(path: '/today', builder: (c, s) => const TodayScreen()),
          GoRoute(path: '/all', builder: (c, s) => const AllArticlesScreen()),
          GoRoute(path: '/saved', builder: (c, s) => const SavedScreen()),
          GoRoute(path: '/starred', builder: (c, s) => const StarredScreen()),
          GoRoute(path: '/recently-read', builder: (c, s) => const RecentlyReadScreen()),
          GoRoute(path: '/search', builder: (c, s) => const SearchScreen()),
          GoRoute(
            path: '/feed/:feedId',
            builder: (c, s) => FeedScreen(feedId: s.pathParameters['feedId']!),
          ),
          GoRoute(
            path: '/folder/:folderId',
            builder: (c, s) => FolderScreen(folderId: s.pathParameters['folderId']!),
          ),
        ],
      ),
      GoRoute(path: '/preferences', builder: (c, s) => const PreferencesScreen()),
    ],
  );
  return router;
});

// Clears the selected article whenever the user navigates to a different route.
class _ArticleSelectionClearer extends NavigatorObserver {
  final Ref _ref;
  _ArticleSelectionClearer(this._ref);

  @override
  void didPush(Route<dynamic> route, Route<dynamic>? previousRoute) =>
      _ref.read(selectedArticleIdProvider.notifier).state = null;

  @override
  void didReplace({Route<dynamic>? newRoute, Route<dynamic>? oldRoute}) =>
      _ref.read(selectedArticleIdProvider.notifier).state = null;
}

class PlexReaderApp extends ConsumerWidget {
  const PlexReaderApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);
    final prefsAsync = ref.watch(preferencesProvider);
    final themeMode = prefsAsync.whenOrNull(
      data: (p) => p.theme == AppThemeMode.light ? ThemeMode.light : ThemeMode.dark,
    ) ?? ThemeMode.dark;

    return MaterialApp.router(
      title: 'PlexReader',
      theme: AppTheme.lightTheme,
      darkTheme: AppTheme.darkTheme,
      themeMode: themeMode,
      routerConfig: router,
      debugShowCheckedModeBanner: false,
    );
  }
}
