// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../theme/app_theme.dart';

class KeyboardShortcutsHandler extends ConsumerWidget {
  final Widget child;
  const KeyboardShortcutsHandler({super.key, required this.child});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Focus(
      autofocus: true,
      onKeyEvent: (node, event) {
        if (event is! KeyDownEvent) return KeyEventResult.ignored;
        final key = event.logicalKey;
        final shift = HardwareKeyboard.instance.isShiftPressed;

        if (key == LogicalKeyboardKey.slash && shift) {
          // ? key — show shortcuts
          _showShortcutsDialog(context);
          return KeyEventResult.handled;
        }
        if (key == LogicalKeyboardKey.keyA && shift) {
          // Shift+A — mark all read (handled per screen via FocusScope)
          return KeyEventResult.ignored;
        }
        return KeyEventResult.ignored;
      },
      child: child,
    );
  }

  void _showShortcutsDialog(BuildContext context) {
    showDialog(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('Keyboard Shortcuts'),
        content: SizedBox(
          width: 360,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: const [
              _ShortcutRow('j', 'Next article'),
              _ShortcutRow('k', 'Previous article'),
              _ShortcutRow('m', 'Toggle read'),
              _ShortcutRow('s', 'Star/unstar'),
              _ShortcutRow('l', 'Save for later'),
              _ShortcutRow('Shift + A', 'Mark all as read'),
              _ShortcutRow('?', 'Show this help'),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Close'),
          ),
        ],
      ),
    );
  }
}

class _ShortcutRow extends StatelessWidget {
  final String label;
  final String description;
  const _ShortcutRow(this.label, this.description);

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
            decoration: BoxDecoration(
              color: AppColors.surfaceVariant,
              borderRadius: BorderRadius.circular(4),
              border: Border.all(color: AppColors.divider),
            ),
            child: Text(label, style: const TextStyle(
              color: AppColors.textPrimary,
              fontSize: 12,
              fontFamily: 'monospace',
            )),
          ),
          const SizedBox(width: 12),
          Text(description, style: const TextStyle(color: AppColors.textSecondary, fontSize: 14)),
        ],
      ),
    );
  }
}
