import 'package:flutter/material.dart';

abstract final class FlashXTheme {
  static const navy = Color(0xFF0B132B);
  static const navySoft = Color(0xFF17213D);
  static const yellow = Color(0xFFFFD600);
  static const lime = Color(0xFFC7F36B);
  static const canvas = Color(0xFFF4F5F7);
  static const surface = Color(0xFFFFFFFF);
  static const textPrimary = Color(0xFF111827);
  static const textSecondary = Color(0xFF6B7280);
  static const border = Color(0xFFE5E7EB);
  static const success = Color(0xFF10B981);
  static const warning = Color(0xFFF59E0B);
  static const danger = Color(0xFFE5484D);

  static ThemeData light() {
    const scheme = ColorScheme.light(
      primary: navy,
      onPrimary: Colors.white,
      secondary: yellow,
      onSecondary: navy,
      tertiary: lime,
      surface: surface,
      onSurface: textPrimary,
      error: danger,
      outline: border,
      surfaceContainer: Color(0xFFF4F5F7),
    );
    final base = ThemeData(
      useMaterial3: true,
      colorScheme: scheme,
      scaffoldBackgroundColor: canvas,
      visualDensity: VisualDensity.standard,
    );

    return base.copyWith(
      textTheme: base.textTheme.copyWith(
        displaySmall: const TextStyle(
          color: textPrimary,
          fontSize: 34,
          height: 1.05,
          fontWeight: FontWeight.w900,
          letterSpacing: -1.2,
        ),
        headlineMedium: const TextStyle(
          color: textPrimary,
          fontSize: 28,
          height: 1.1,
          fontWeight: FontWeight.w900,
          letterSpacing: -.8,
        ),
        headlineSmall: const TextStyle(
          color: textPrimary,
          fontSize: 23,
          height: 1.15,
          fontWeight: FontWeight.w800,
          letterSpacing: -.45,
        ),
        titleLarge: const TextStyle(
          color: textPrimary,
          fontSize: 19,
          height: 1.2,
          fontWeight: FontWeight.w800,
        ),
        bodyLarge: const TextStyle(
          color: textPrimary,
          fontSize: 16,
          height: 1.45,
          fontWeight: FontWeight.w500,
        ),
        bodyMedium: const TextStyle(
          color: textSecondary,
          fontSize: 14,
          height: 1.4,
          fontWeight: FontWeight.w500,
        ),
        labelLarge: const TextStyle(
          fontSize: 16,
          fontWeight: FontWeight.w800,
        ),
      ),
      cardTheme: CardThemeData(
        color: surface,
        surfaceTintColor: Colors.transparent,
        elevation: 0,
        margin: EdgeInsets.zero,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(22),
          side: const BorderSide(color: border),
        ),
      ),
      navigationBarTheme: NavigationBarThemeData(
        height: 72,
        elevation: 0,
        backgroundColor: surface,
        indicatorColor: yellow,
        surfaceTintColor: Colors.transparent,
        labelTextStyle: WidgetStateProperty.resolveWith((states) => TextStyle(
              color:
                  states.contains(WidgetState.selected) ? navy : textSecondary,
              fontSize: 11,
              fontWeight: states.contains(WidgetState.selected)
                  ? FontWeight.w800
                  : FontWeight.w600,
            )),
        iconTheme: WidgetStateProperty.resolveWith((states) => IconThemeData(
              color:
                  states.contains(WidgetState.selected) ? navy : textSecondary,
              size: 24,
            )),
      ),
      filledButtonTheme: FilledButtonThemeData(
        style: FilledButton.styleFrom(
          backgroundColor: navy,
          foregroundColor: Colors.white,
          minimumSize: const Size.fromHeight(60),
          padding: const EdgeInsets.symmetric(horizontal: 22, vertical: 18),
          shape:
              RoundedRectangleBorder(borderRadius: BorderRadius.circular(19)),
          textStyle: const TextStyle(fontSize: 17, fontWeight: FontWeight.w900),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: navy,
          minimumSize: const Size.fromHeight(56),
          padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 16),
          side: const BorderSide(color: border),
          shape:
              RoundedRectangleBorder(borderRadius: BorderRadius.circular(18)),
          textStyle: const TextStyle(fontSize: 16, fontWeight: FontWeight.w800),
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: surface,
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 16, vertical: 17),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(18),
          borderSide: const BorderSide(color: border),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(18),
          borderSide: const BorderSide(color: navy, width: 1.5),
        ),
      ),
      dividerTheme: const DividerThemeData(color: border, thickness: 1),
      progressIndicatorTheme: const ProgressIndicatorThemeData(
        color: yellow,
        linearTrackColor: Color(0xFF293552),
      ),
    );
  }
}
