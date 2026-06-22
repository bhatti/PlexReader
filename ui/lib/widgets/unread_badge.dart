// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter/material.dart';
import '../theme/app_theme.dart';

class UnreadBadge extends StatelessWidget {
  final int count;

  const UnreadBadge({super.key, required this.count});

  @override
  Widget build(BuildContext context) {
    if (count <= 0) return const SizedBox.shrink();
    final label = count > 999 ? '999+' : count.toString();
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: AppColors.unreadBadge,
        borderRadius: BorderRadius.circular(10),
      ),
      child: Text(
        label,
        style: const TextStyle(
          color: AppColors.textSecondary,
          fontSize: 11,
          fontWeight: FontWeight.w500,
        ),
      ),
    );
  }
}
