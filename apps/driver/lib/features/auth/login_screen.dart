import 'package:flutter/material.dart';

import 'auth_controller.dart';

class DriverLoginScreen extends StatefulWidget {
  const DriverLoginScreen({super.key, required this.auth});
  final AuthController auth;

  @override
  State<DriverLoginScreen> createState() => _DriverLoginScreenState();
}

class _DriverLoginScreenState extends State<DriverLoginScreen> {
  final phone = TextEditingController(text: '+84987654321');
  final otp = TextEditingController();

  @override
  void dispose() {
    phone.dispose();
    otp.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final waiting = widget.auth.challengeId != null;
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 420),
              child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    const CircleAvatar(
                        radius: 36,
                        child: Icon(Icons.two_wheeler_rounded, size: 34)),
                    const SizedBox(height: 22),
                    Text('FlashX Driver',
                        textAlign: TextAlign.center,
                        style: Theme.of(context)
                            .textTheme
                            .headlineMedium
                            ?.copyWith(fontWeight: FontWeight.w900)),
                    const SizedBox(height: 8),
                    const Text('Đăng nhập để quản lý hồ sơ và nhận chuyến.',
                        textAlign: TextAlign.center),
                    const SizedBox(height: 28),
                    TextField(
                        controller: phone,
                        enabled: !waiting && !widget.auth.busy,
                        keyboardType: TextInputType.phone,
                        decoration: const InputDecoration(
                            labelText: 'Số điện thoại',
                            border: OutlineInputBorder())),
                    if (waiting) ...[
                      const SizedBox(height: 14),
                      TextField(
                          controller: otp,
                          enabled: !widget.auth.busy,
                          keyboardType: TextInputType.number,
                          maxLength: 6,
                          decoration: InputDecoration(
                              labelText: 'Mã OTP',
                              border: const OutlineInputBorder(),
                              helperText: widget.auth.developmentCode == null
                                  ? null
                                  : 'DEV OTP: ${widget.auth.developmentCode}')),
                    ],
                    if (widget.auth.error != null) ...[
                      const SizedBox(height: 8),
                      Text(widget.auth.error!,
                          style: TextStyle(
                              color: Theme.of(context).colorScheme.error))
                    ],
                    const SizedBox(height: 18),
                    FilledButton(
                      onPressed: widget.auth.busy
                          ? null
                          : () async {
                              if (waiting) {
                                await widget.auth.verifyOTP(otp.text.trim());
                              } else {
                                await widget.auth.requestOTP(phone.text.trim());
                              }
                            },
                      child: Padding(
                          padding: const EdgeInsets.symmetric(vertical: 14),
                          child: widget.auth.busy
                              ? const SizedBox.square(
                                  dimension: 22,
                                  child:
                                      CircularProgressIndicator(strokeWidth: 2))
                              : Text(waiting ? 'Xác nhận OTP' : 'Tiếp tục')),
                    ),
                  ]),
            ),
          ),
        ),
      ),
    );
  }
}
