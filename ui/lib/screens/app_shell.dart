// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../theme/app_theme.dart';
import '../providers/navigation_provider.dart';
import '../widgets/sidebar.dart';
import '../widgets/keyboard_shortcuts.dart';

class AppShell extends ConsumerStatefulWidget {
  final Widget child;
  const AppShell({super.key, required this.child});

  @override
  ConsumerState<AppShell> createState() => _AppShellState();
}

class _AppShellState extends ConsumerState<AppShell> {
  final _scaffoldKey = GlobalKey<ScaffoldState>();

  @override
  Widget build(BuildContext context) {
    final isWide = MediaQuery.sizeOf(context).width >= 700;
    final sidebarCollapsed = ref.watch(sidebarCollapsedProvider);

    return KeyboardShortcutsHandler(
      child: Scaffold(
        key: _scaffoldKey,
        backgroundColor: AppColors.background,
        drawer: isWide ? null : const Drawer(child: Sidebar()),
        body: Row(
          children: [
            if (isWide)
              SizedBox(
                width: sidebarCollapsed ? 48 : 260,
                child: const Sidebar(),
              ),
            if (isWide)
              Container(width: 1, color: AppColors.divider),
            Expanded(child: widget.child),
          ],
        ),
        floatingActionButton: !isWide
            ? FloatingActionButton(
                mini: true,
                backgroundColor: AppColors.surface,
                onPressed: () => _scaffoldKey.currentState?.openDrawer(),
                child: const Icon(Icons.menu, color: AppColors.textPrimary),
              )
            : null,
      ),
    );
  }
}
