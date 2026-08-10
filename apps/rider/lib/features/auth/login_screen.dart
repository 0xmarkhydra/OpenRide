import 'package:flutter/material.dart';

import 'auth_controller.dart';

class RiderLoginScreen extends StatefulWidget {
  const RiderLoginScreen({super.key, required this.auth});
  final AuthController auth;

  @override
  State<RiderLoginScreen> createState() => _RiderLoginScreenState();
}

class _RiderLoginScreenState extends State<RiderLoginScreen> {
  final _phone = TextEditingController(text: '+84912345678');
  final _otp = TextEditingController();

  @override
  void dispose() {
    _phone.dispose();
    _otp.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final awaitingCode = widget.auth.challengeId != null;
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
                      radius: 34, child: Icon(Icons.route_rounded, size: 32)),
                  const SizedBox(height: 24),
                  Text('Chào mừng đến FlashX',
                      textAlign: TextAlign.center,
                      style: Theme.of(context)
                          .textTheme
                          .headlineMedium
                          ?.copyWith(fontWeight: FontWeight.w900)),
                  const SizedBox(height: 8),
                  Text('Đăng nhập bằng số điện thoại để đặt chuyến.',
                      textAlign: TextAlign.center,
                      style: TextStyle(
                          color:
                              Theme.of(context).colorScheme.onSurfaceVariant)),
                  const SizedBox(height: 28),
                  TextField(
                    controller: _phone,
                    enabled: !awaitingCode && !widget.auth.busy,
                    keyboardType: TextInputType.phone,
                    decoration: const InputDecoration(
                        labelText: 'Số điện thoại',
                        prefixIcon: Icon(Icons.phone_outlined),
                        border: OutlineInputBorder()),
                  ),
                  if (awaitingCode) ...[
                    const SizedBox(height: 14),
                    TextField(
                      controller: _otp,
                      enabled: !widget.auth.busy,
                      keyboardType: TextInputType.number,
                      maxLength: 6,
                      decoration: InputDecoration(
                        labelText: 'Mã OTP',
                        prefixIcon: const Icon(Icons.lock_outline_rounded),
                        border: const OutlineInputBorder(),
                        helperText: widget.auth.developmentCode == null
                            ? null
                            : 'DEV OTP: ${widget.auth.developmentCode}',
                      ),
                    ),
                  ],
                  if (widget.auth.error != null) ...[
                    const SizedBox(height: 8),
                    Text(widget.auth.error!,
                        style: TextStyle(
                            color: Theme.of(context).colorScheme.error)),
                  ],
                  const SizedBox(height: 18),
                  FilledButton(
                    onPressed: widget.auth.busy
                        ? null
                        : () async {
                            if (!awaitingCode) {
                              await widget.auth.requestOTP(_phone.text.trim());
                            } else {
                              await widget.auth.verifyOTP(_otp.text.trim());
                            }
                          },
                    child: Padding(
                      padding: const EdgeInsets.symmetric(vertical: 14),
                      child: widget.auth.busy
                          ? const SizedBox(
                              width: 22,
                              height: 22,
                              child: CircularProgressIndicator(strokeWidth: 2))
                          : Text(awaitingCode ? 'Xác nhận OTP' : 'Tiếp tục'),
                    ),
                  ),
                  if (awaitingCode)
                    TextButton(
                        onPressed: widget.auth.busy
                            ? null
                            : () => setState(() {
                                  widget.auth.challengeId = null;
                                }),
                        child: const Text('Đổi số điện thoại')),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
