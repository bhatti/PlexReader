// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/preferences.dart';

// Currently selected article id (for the detail panel)
final selectedArticleIdProvider = StateProvider<String?>((ref) => null);

// Current view mode (can be overridden per screen)
final viewModeProvider = StateProvider<ViewMode>((ref) => ViewMode.magazine);

// Sidebar collapsed state
final sidebarCollapsedProvider = StateProvider<bool>((ref) => false);

// Currently expanded folder ids in sidebar
final expandedFolderIdsProvider = StateProvider<Set<String>>((ref) => {});
