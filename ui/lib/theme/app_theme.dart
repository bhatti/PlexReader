// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

// Dark-mode palette.
class AppColors {
  static const background = Color(0xFF1a1a1a);
  static const sidebar = Color(0xFF1e1e1e);
  static const surface = Color(0xFF242424);
  static const surfaceVariant = Color(0xFF2a2a2a);
  static const primary = Color(0xFFf57c00);
  static const primaryDark = Color(0xFFe65100);
  static const textPrimary = Color(0xFFe8e8e8);
  static const textSecondary = Color(0xFF8a8a8a);
  static const error = Color(0xFFe53935);
  static const divider = Color(0xFF2e2e2e);
  static const selectedItem = Color(0xFF2d2d2d);
  static const hoverItem = Color(0xFF252525);
  static const unreadBadge = Color(0xFF3a3a3a);
  static const success = Color(0xFF4caf50);
  static const star = Color(0xFFffd54f);
}

// Light-mode palette.
class AppColorsLight {
  static const background = Color(0xFFf5f5f5);
  static const sidebar = Color(0xFFfafafa);
  static const surface = Color(0xFFffffff);
  static const surfaceVariant = Color(0xFFf0f0f0);
  static const primary = Color(0xFFe65100);
  static const textPrimary = Color(0xFF1a1a1a);
  static const textSecondary = Color(0xFF757575);
  static const error = Color(0xFFe53935);
  static const divider = Color(0xFFe0e0e0);
  static const selectedItem = Color(0xFFfff3e0);
  static const unreadBadge = Color(0xFFeeeeee);
  static const success = Color(0xFF388e3c);
  static const star = Color(0xFFf9a825);
}

class AppTheme {
  static ThemeData get darkTheme {
    final base = ThemeData.dark();
    final textTheme = GoogleFonts.interTextTheme(base.textTheme).copyWith(
      displayLarge: GoogleFonts.inter(color: AppColors.textPrimary, fontSize: 32, fontWeight: FontWeight.w600),
      displayMedium: GoogleFonts.inter(color: AppColors.textPrimary, fontSize: 28, fontWeight: FontWeight.w600),
      displaySmall: GoogleFonts.inter(color: AppColors.textPrimary, fontSize: 24, fontWeight: FontWeight.w600),
      headlineLarge: GoogleFonts.inter(color: AppColors.textPrimary, fontSize: 22, fontWeight: FontWeight.w600),
      headlineMedium: GoogleFonts.inter(color: AppColors.textPrimary, fontSize: 18, fontWeight: FontWeight.w600),
      headlineSmall: GoogleFonts.inter(color: AppColors.textPrimary, fontSize: 16, fontWeight: FontWeight.w600),
      titleLarge: GoogleFonts.inter(color: AppColors.textPrimary, fontSize: 16, fontWeight: FontWeight.w600),
      titleMedium: GoogleFonts.inter(color: AppColors.textPrimary, fontSize: 14, fontWeight: FontWeight.w500),
      titleSmall: GoogleFonts.inter(color: AppColors.textSecondary, fontSize: 12, fontWeight: FontWeight.w500),
      bodyLarge: GoogleFonts.inter(color: AppColors.textPrimary, fontSize: 15),
      bodyMedium: GoogleFonts.inter(color: AppColors.textPrimary, fontSize: 14),
      bodySmall: GoogleFonts.inter(color: AppColors.textSecondary, fontSize: 12),
      labelLarge: GoogleFonts.inter(color: AppColors.textPrimary, fontSize: 14, fontWeight: FontWeight.w500),
      labelMedium: GoogleFonts.inter(color: AppColors.textSecondary, fontSize: 12),
      labelSmall: GoogleFonts.inter(color: AppColors.textSecondary, fontSize: 11),
    );

    return base.copyWith(
      colorScheme: const ColorScheme.dark(
        primary: AppColors.primary,
        onPrimary: Colors.white,
        secondary: AppColors.primary,
        onSecondary: Colors.white,
        surface: AppColors.surface,
        onSurface: AppColors.textPrimary,
        error: AppColors.error,
        onError: Colors.white,
        outline: AppColors.divider,
        surfaceContainerHighest: AppColors.surfaceVariant,
      ),
      scaffoldBackgroundColor: AppColors.background,
      textTheme: textTheme,
      primaryTextTheme: textTheme,
      appBarTheme: AppBarTheme(
        backgroundColor: AppColors.sidebar,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
        titleTextStyle: GoogleFonts.inter(
          color: AppColors.textPrimary,
          fontSize: 18,
          fontWeight: FontWeight.w600,
        ),
      ),
      dividerTheme: const DividerThemeData(
        color: AppColors.divider,
        thickness: 1,
        space: 1,
      ),
      cardTheme: const CardThemeData(
        color: AppColors.surface,
        elevation: 0,
        margin: EdgeInsets.zero,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.all(Radius.circular(8)),
        ),
      ),
      listTileTheme: const ListTileThemeData(
        tileColor: Colors.transparent,
        selectedTileColor: AppColors.selectedItem,
        iconColor: AppColors.textSecondary,
        textColor: AppColors.textPrimary,
        dense: true,
        contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 2),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: AppColors.surfaceVariant,
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: const BorderSide(color: AppColors.divider),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: const BorderSide(color: AppColors.divider),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: const BorderSide(color: AppColors.primary, width: 1.5),
        ),
        hintStyle: GoogleFonts.inter(color: AppColors.textSecondary, fontSize: 14),
        contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: AppColors.primary,
          foregroundColor: Colors.white,
          elevation: 0,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
          textStyle: GoogleFonts.inter(fontSize: 14, fontWeight: FontWeight.w500),
        ),
      ),
      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          foregroundColor: AppColors.primary,
          textStyle: GoogleFonts.inter(fontSize: 14, fontWeight: FontWeight.w500),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: AppColors.textPrimary,
          side: const BorderSide(color: AppColors.divider),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
        ),
      ),
      switchTheme: SwitchThemeData(
        thumbColor: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) return AppColors.primary;
          return AppColors.textSecondary;
        }),
        trackColor: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) return AppColors.primary.withValues(alpha: 0.4);
          return AppColors.surfaceVariant;
        }),
      ),
      checkboxTheme: CheckboxThemeData(
        fillColor: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) return AppColors.primary;
          return Colors.transparent;
        }),
        checkColor: WidgetStateProperty.all(Colors.white),
        side: const BorderSide(color: AppColors.textSecondary, width: 1.5),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(4)),
      ),
      radioTheme: RadioThemeData(
        fillColor: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) return AppColors.primary;
          return AppColors.textSecondary;
        }),
      ),
      tooltipTheme: TooltipThemeData(
        decoration: BoxDecoration(
          color: AppColors.surface,
          borderRadius: BorderRadius.circular(6),
          border: Border.all(color: AppColors.divider),
        ),
        textStyle: GoogleFonts.inter(color: AppColors.textPrimary, fontSize: 12),
      ),
      popupMenuTheme: PopupMenuThemeData(
        color: AppColors.surface,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
          side: const BorderSide(color: AppColors.divider),
        ),
        elevation: 4,
        textStyle: GoogleFonts.inter(color: AppColors.textPrimary, fontSize: 14),
      ),
      dialogTheme: DialogThemeData(
        backgroundColor: AppColors.surface,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        titleTextStyle: GoogleFonts.inter(
          color: AppColors.textPrimary,
          fontSize: 18,
          fontWeight: FontWeight.w600,
        ),
        contentTextStyle: GoogleFonts.inter(color: AppColors.textPrimary, fontSize: 14),
      ),
      scrollbarTheme: ScrollbarThemeData(
        thumbColor: WidgetStateProperty.all(AppColors.divider),
        trackColor: WidgetStateProperty.all(Colors.transparent),
        radius: const Radius.circular(4),
        thickness: WidgetStateProperty.all(4),
      ),
      iconTheme: const IconThemeData(color: AppColors.textSecondary, size: 20),
    );
  }

  static ThemeData get lightTheme {
    final base = ThemeData.light();
    final textTheme = GoogleFonts.interTextTheme(base.textTheme).copyWith(
      displayLarge: GoogleFonts.inter(color: AppColorsLight.textPrimary, fontSize: 32, fontWeight: FontWeight.w600),
      displayMedium: GoogleFonts.inter(color: AppColorsLight.textPrimary, fontSize: 28, fontWeight: FontWeight.w600),
      displaySmall: GoogleFonts.inter(color: AppColorsLight.textPrimary, fontSize: 24, fontWeight: FontWeight.w600),
      headlineLarge: GoogleFonts.inter(color: AppColorsLight.textPrimary, fontSize: 22, fontWeight: FontWeight.w600),
      headlineMedium: GoogleFonts.inter(color: AppColorsLight.textPrimary, fontSize: 18, fontWeight: FontWeight.w600),
      headlineSmall: GoogleFonts.inter(color: AppColorsLight.textPrimary, fontSize: 16, fontWeight: FontWeight.w600),
      titleLarge: GoogleFonts.inter(color: AppColorsLight.textPrimary, fontSize: 16, fontWeight: FontWeight.w600),
      titleMedium: GoogleFonts.inter(color: AppColorsLight.textPrimary, fontSize: 14, fontWeight: FontWeight.w500),
      titleSmall: GoogleFonts.inter(color: AppColorsLight.textSecondary, fontSize: 12, fontWeight: FontWeight.w500),
      bodyLarge: GoogleFonts.inter(color: AppColorsLight.textPrimary, fontSize: 15),
      bodyMedium: GoogleFonts.inter(color: AppColorsLight.textPrimary, fontSize: 14),
      bodySmall: GoogleFonts.inter(color: AppColorsLight.textSecondary, fontSize: 12),
      labelLarge: GoogleFonts.inter(color: AppColorsLight.textPrimary, fontSize: 14, fontWeight: FontWeight.w500),
      labelMedium: GoogleFonts.inter(color: AppColorsLight.textSecondary, fontSize: 12),
      labelSmall: GoogleFonts.inter(color: AppColorsLight.textSecondary, fontSize: 11),
    );

    return base.copyWith(
      colorScheme: const ColorScheme.light(
        primary: AppColorsLight.primary,
        onPrimary: Colors.white,
        secondary: AppColorsLight.primary,
        onSecondary: Colors.white,
        surface: AppColorsLight.surface,
        onSurface: AppColorsLight.textPrimary,
        error: AppColorsLight.error,
        onError: Colors.white,
        outline: AppColorsLight.divider,
        surfaceContainerHighest: AppColorsLight.surfaceVariant,
      ),
      scaffoldBackgroundColor: AppColorsLight.background,
      textTheme: textTheme,
      primaryTextTheme: textTheme,
      appBarTheme: AppBarTheme(
        backgroundColor: AppColorsLight.sidebar,
        foregroundColor: AppColorsLight.textPrimary,
        elevation: 0,
        titleTextStyle: GoogleFonts.inter(
          color: AppColorsLight.textPrimary,
          fontSize: 18,
          fontWeight: FontWeight.w600,
        ),
      ),
      dividerTheme: const DividerThemeData(
        color: AppColorsLight.divider,
        thickness: 1,
        space: 1,
      ),
      cardTheme: const CardThemeData(
        color: AppColorsLight.surface,
        elevation: 0,
        margin: EdgeInsets.zero,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.all(Radius.circular(8)),
        ),
      ),
      listTileTheme: const ListTileThemeData(
        tileColor: Colors.transparent,
        selectedTileColor: AppColorsLight.selectedItem,
        iconColor: AppColorsLight.textSecondary,
        textColor: AppColorsLight.textPrimary,
        dense: true,
        contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 2),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: AppColorsLight.surfaceVariant,
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: const BorderSide(color: AppColorsLight.divider),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: const BorderSide(color: AppColorsLight.divider),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: const BorderSide(color: AppColorsLight.primary, width: 1.5),
        ),
        hintStyle: GoogleFonts.inter(color: AppColorsLight.textSecondary, fontSize: 14),
        contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: AppColorsLight.primary,
          foregroundColor: Colors.white,
          elevation: 0,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
          textStyle: GoogleFonts.inter(fontSize: 14, fontWeight: FontWeight.w500),
        ),
      ),
      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          foregroundColor: AppColorsLight.primary,
          textStyle: GoogleFonts.inter(fontSize: 14, fontWeight: FontWeight.w500),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: AppColorsLight.textPrimary,
          side: const BorderSide(color: AppColorsLight.divider),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
        ),
      ),
      switchTheme: SwitchThemeData(
        thumbColor: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) return AppColorsLight.primary;
          return AppColorsLight.textSecondary;
        }),
        trackColor: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) return AppColorsLight.primary.withValues(alpha: 0.4);
          return AppColorsLight.surfaceVariant;
        }),
      ),
      checkboxTheme: CheckboxThemeData(
        fillColor: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) return AppColorsLight.primary;
          return Colors.transparent;
        }),
        checkColor: WidgetStateProperty.all(Colors.white),
        side: const BorderSide(color: AppColorsLight.textSecondary, width: 1.5),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(4)),
      ),
      radioTheme: RadioThemeData(
        fillColor: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) return AppColorsLight.primary;
          return AppColorsLight.textSecondary;
        }),
      ),
      tooltipTheme: TooltipThemeData(
        decoration: BoxDecoration(
          color: AppColorsLight.surface,
          borderRadius: BorderRadius.circular(6),
          border: Border.all(color: AppColorsLight.divider),
        ),
        textStyle: GoogleFonts.inter(color: AppColorsLight.textPrimary, fontSize: 12),
      ),
      popupMenuTheme: PopupMenuThemeData(
        color: AppColorsLight.surface,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
          side: const BorderSide(color: AppColorsLight.divider),
        ),
        elevation: 4,
        textStyle: GoogleFonts.inter(color: AppColorsLight.textPrimary, fontSize: 14),
      ),
      dialogTheme: DialogThemeData(
        backgroundColor: AppColorsLight.surface,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        titleTextStyle: GoogleFonts.inter(
          color: AppColorsLight.textPrimary,
          fontSize: 18,
          fontWeight: FontWeight.w600,
        ),
        contentTextStyle: GoogleFonts.inter(color: AppColorsLight.textPrimary, fontSize: 14),
      ),
      scrollbarTheme: ScrollbarThemeData(
        thumbColor: WidgetStateProperty.all(AppColorsLight.divider),
        trackColor: WidgetStateProperty.all(Colors.transparent),
        radius: const Radius.circular(4),
        thickness: WidgetStateProperty.all(4),
      ),
      iconTheme: const IconThemeData(color: AppColorsLight.textSecondary, size: 20),
    );
  }
}
