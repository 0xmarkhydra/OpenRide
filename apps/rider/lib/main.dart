import 'package:flutter/material.dart';

import 'core/theme/flashx_theme.dart';
import 'features/auth/auth_controller.dart';
import 'features/auth/login_screen.dart';
import 'features/home/rider_home_screen.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const FlashXRiderApp());
}

class FlashXRiderApp extends StatefulWidget {
  const FlashXRiderApp({super.key});

  @override
  State<FlashXRiderApp> createState() => _FlashXRiderAppState();
}

class _FlashXRiderAppState extends State<FlashXRiderApp> {
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
      title: 'FlashX',
      debugShowCheckedModeBanner: false,
      theme: FlashXTheme.light(),
      home: !auth.initialized
          ? const Scaffold(body: Center(child: CircularProgressIndicator()))
          : auth.authenticated
              ? RiderHomeScreen(auth: auth)
              : RiderLoginScreen(auth: auth),
    );
  }
}
