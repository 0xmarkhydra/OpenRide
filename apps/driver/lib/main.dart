import 'package:flutter/material.dart';

import 'core/theme/flashx_theme.dart';
import 'features/auth/auth_controller.dart';
import 'features/auth/login_screen.dart';
import 'features/home/driver_home_screen.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const FlashXDriverApp());
}

class FlashXDriverApp extends StatefulWidget {
  const FlashXDriverApp({super.key});

  @override
  State<FlashXDriverApp> createState() => _FlashXDriverAppState();
}

class _FlashXDriverAppState extends State<FlashXDriverApp> {
  late final AuthController auth;

  @override
  void initState() {
    super.initState();
    auth = AuthController()..addListener(_refresh);
    auth.initialize();
  }

  void _refresh() {
    if (mounted) setState(() {});
  }

  @override
  void dispose() {
    auth.removeListener(_refresh);
    auth.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'OpenRide Driver',
      debugShowCheckedModeBanner: false,
      theme: FlashXTheme.light(),
      home: !auth.initialized
          ? const Scaffold(body: Center(child: CircularProgressIndicator()))
          : auth.authenticated
              ? DriverHomeScreen(auth: auth)
              : DriverLoginScreen(auth: auth),
    );
  }
}
