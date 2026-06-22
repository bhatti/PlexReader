// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../models/preferences.dart';
import '../providers/preferences_provider.dart';
import '../theme/app_theme.dart';

class PreferencesScreen extends ConsumerWidget {
  const PreferencesScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final prefsAsync = ref.watch(preferencesProvider);

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        title: const Text('Preferences'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => context.canPop() ? context.pop() : context.go('/today'),
        ),
      ),
      body: prefsAsync.when(
        loading: () => const Center(child: CircularProgressIndicator(color: AppColors.primary)),
        error: (e, _) => Center(child: Text('Error: $e', style: const TextStyle(color: AppColors.error))),
        data: (prefs) => _PreferencesForm(prefs: prefs),
      ),
    );
  }
}

class _PreferencesForm extends ConsumerWidget {
  final Preferences prefs;
  const _PreferencesForm({required this.prefs});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final notifier = ref.read(preferencesProvider.notifier);

    return Row(
      children: [
        // Left nav (desktop style)
        if (MediaQuery.sizeOf(context).width > 600)
          SizedBox(
            width: 200,
            child: Material(
              color: AppColors.sidebar,
              child: ListView(
                children: const [
                  _NavItem(label: 'General', icon: Icons.settings_outlined, isSelected: true),
                ],
              ),
            ),
          ),
        Expanded(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(32),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 560),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('General', style: TextStyle(
                    color: AppColors.textPrimary,
                    fontSize: 22,
                    fontWeight: FontWeight.w700,
                  )),
                  const SizedBox(height: 24),

                  // Start Page
                  _sectionHeader('Start Page'),
                  ...StartPage.values.map((v) => RadioListTile<StartPage>(
                    value: v,
                    groupValue: prefs.startPage,
                    title: Text(_startPageLabel(v)),
                    onChanged: (val) {
                      if (val != null) notifier.update(prefs.copyWith(startPage: val));
                    },
                    activeColor: AppColors.primary,
                    contentPadding: EdgeInsets.zero,
                  )),

                  const SizedBox(height: 24),

                  // Default View
                  _sectionHeader('Default Presentation'),
                  ...ViewMode.values.map((v) => RadioListTile<ViewMode>(
                    value: v,
                    groupValue: prefs.defaultView,
                    title: Text(_viewModeLabel(v)),
                    subtitle: Text(_viewModeDesc(v), style: const TextStyle(color: AppColors.textSecondary, fontSize: 12)),
                    onChanged: (val) {
                      if (val != null) notifier.update(prefs.copyWith(defaultView: val));
                    },
                    activeColor: AppColors.primary,
                    contentPadding: EdgeInsets.zero,
                  )),

                  const SizedBox(height: 24),

                  // Default Sort
                  _sectionHeader('Default Sort'),
                  ...SortOrder.values.map((v) => RadioListTile<SortOrder>(
                    value: v,
                    groupValue: prefs.defaultSort,
                    title: Text(_sortLabel(v)),
                    onChanged: (val) {
                      if (val != null) notifier.update(prefs.copyWith(defaultSort: val));
                    },
                    activeColor: AppColors.primary,
                    contentPadding: EdgeInsets.zero,
                  )),

                  const SizedBox(height: 24),

                  // Appearance
                  _sectionHeader('Appearance'),
                  ...AppThemeMode.values.map((v) => RadioListTile<AppThemeMode>(
                    value: v,
                    groupValue: prefs.theme,
                    title: Text(_themeModeLabel(v)),
                    secondary: Icon(
                      v == AppThemeMode.dark ? Icons.dark_mode_outlined : Icons.light_mode_outlined,
                      size: 20,
                    ),
                    onChanged: (val) {
                      if (val != null) notifier.update(prefs.copyWith(theme: val));
                    },
                    activeColor: AppColors.primary,
                    contentPadding: EdgeInsets.zero,
                  )),

                  const SizedBox(height: 24),

                  // Hide read articles
                  _sectionHeader('Reading'),
                  SwitchListTile(
                    value: prefs.hideReadArticles,
                    onChanged: (val) => notifier.update(prefs.copyWith(hideReadArticles: val)),
                    title: const Text('Hide read articles'),
                    subtitle: const Text(
                      'Only show unread articles by default',
                      style: TextStyle(color: AppColors.textSecondary, fontSize: 12),
                    ),
                    activeColor: AppColors.primary,
                    contentPadding: EdgeInsets.zero,
                  ),
                ],
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _sectionHeader(String label) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Text(
        label,
        style: const TextStyle(
          color: AppColors.textPrimary,
          fontSize: 16,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }

  String _themeModeLabel(AppThemeMode t) {
    switch (t) {
      case AppThemeMode.dark: return 'Dark';
      case AppThemeMode.light: return 'Light';
    }
  }

  String _startPageLabel(StartPage p) {
    switch (p) {
      case StartPage.today: return 'Today';
      case StartPage.firstFolder: return 'First Folder';
      case StartPage.all: return 'All Articles';
    }
  }

  String _viewModeLabel(ViewMode v) {
    switch (v) {
      case ViewMode.magazine: return 'Magazine';
      case ViewMode.titleOnly: return 'Title Only';
      case ViewMode.cards: return 'Cards';
      case ViewMode.article: return 'Article View';
    }
  }

  String _viewModeDesc(ViewMode v) {
    switch (v) {
      case ViewMode.magazine: return 'Thumbnail + title + summary';
      case ViewMode.titleOnly: return 'Compact list of titles';
      case ViewMode.cards: return 'Image-forward grid layout';
      case ViewMode.article: return 'Full article inline';
    }
  }

  String _sortLabel(SortOrder s) {
    switch (s) {
      case SortOrder.newestFirst: return 'Newest first';
      case SortOrder.oldestFirst: return 'Oldest first';
    }
  }
}

class _NavItem extends StatelessWidget {
  final String label;
  final IconData icon;
  final bool isSelected;
  const _NavItem({required this.label, required this.icon, this.isSelected = false});

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: Icon(icon, size: 18, color: isSelected ? AppColors.primary : AppColors.textSecondary),
      title: Text(label, style: TextStyle(
        color: isSelected ? AppColors.textPrimary : AppColors.textSecondary,
        fontSize: 14,
        fontWeight: isSelected ? FontWeight.w500 : FontWeight.normal,
      )),
      selected: isSelected,
      selectedTileColor: AppColors.selectedItem,
    );
  }
}
