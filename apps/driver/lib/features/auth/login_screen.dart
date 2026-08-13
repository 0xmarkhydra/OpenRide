import 'package:flutter/material.dart';

import '../../core/theme/flashx_theme.dart';
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
    final theme = Theme.of(context);
    return Scaffold(
      backgroundColor: FlashXTheme.navy,
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(20),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 440),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  const Row(
                    children: [
                      _DriverMark(),
                      SizedBox(width: 12),
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text('FlashX',
                              style: TextStyle(
                                  color: Colors.white,
                                  fontSize: 24,
                                  fontWeight: FontWeight.w900,
                                  letterSpacing: -1)),
                          Text('DRIVER',
                              style: TextStyle(
                                  color: Color(0xFF98A3BC),
                                  fontSize: 11,
                                  fontWeight: FontWeight.w700,
                                  letterSpacing: 1.4)),
                        ],
                      ),
                    ],
                  ),
                  const SizedBox(height: 40),
                  Text(
                    waiting
                        ? 'Xác thực tài xế'
                        : 'Sẵn sàng nhận việc cùng FlashX.',
                    style: theme.textTheme.displaySmall
                        ?.copyWith(color: Colors.white),
                  ),
                  const SizedBox(height: 10),
                  Text(
                    waiting
                        ? 'Nhập mã OTP để tiếp tục.'
                        : 'Nhận việc phù hợp gần bạn, xem rõ xe của khách và thực hiện từng bước an toàn.',
                    style: theme.textTheme.bodyLarge
                        ?.copyWith(color: const Color(0xFFB9C1D5)),
                  ),
                  const SizedBox(height: 28),
                  Container(
                    padding: const EdgeInsets.all(20),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(28),
                      boxShadow: const [
                        BoxShadow(
                            color: Color(0x33000000),
                            blurRadius: 36,
                            offset: Offset(0, 16))
                      ],
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        Text(waiting ? 'Mã OTP' : 'Đăng nhập tài xế',
                            style: theme.textTheme.titleLarge),
                        const SizedBox(height: 6),
                        Text(
                            waiting
                                ? 'Mã gồm 6 chữ số và chỉ dùng một lần.'
                                : 'Dùng số điện thoại đã đăng ký hồ sơ tài xế.',
                            style: theme.textTheme.bodyMedium),
                        const SizedBox(height: 20),
                        TextField(
                          controller: phone,
                          enabled: !waiting && !widget.auth.busy,
                          keyboardType: TextInputType.phone,
                          decoration: const InputDecoration(
                              labelText: 'Số điện thoại',
                              hintText: '+84...',
                              prefixIcon: Icon(Icons.phone_iphone_rounded)),
                        ),
                        if (waiting) ...[
                          const SizedBox(height: 14),
                          TextField(
                            controller: otp,
                            autofocus: true,
                            enabled: !widget.auth.busy,
                            keyboardType: TextInputType.number,
                            maxLength: 6,
                            textAlign: TextAlign.center,
                            style: const TextStyle(
                                fontSize: 24,
                                fontWeight: FontWeight.w800,
                                letterSpacing: 8),
                            decoration: InputDecoration(
                              labelText: 'Mã OTP',
                              counterText: '',
                              prefixIcon:
                                  const Icon(Icons.lock_outline_rounded),
                              helperText: widget.auth.developmentCode == null
                                  ? null
                                  : 'DEV OTP: ${widget.auth.developmentCode}',
                            ),
                          ),
                        ],
                        if (widget.auth.error != null) ...[
                          const SizedBox(height: 12),
                          Container(
                            padding: const EdgeInsets.all(12),
                            decoration: BoxDecoration(
                                color: const Color(0xFFFFEEEE),
                                borderRadius: BorderRadius.circular(14)),
                            child: Text(widget.auth.error!,
                                style: const TextStyle(
                                    color: FlashXTheme.danger,
                                    fontWeight: FontWeight.w600)),
                          ),
                        ],
                        const SizedBox(height: 20),
                        FilledButton(
                          onPressed: widget.auth.busy
                              ? null
                              : () async {
                                  if (waiting) {
                                    await widget.auth
                                        .verifyOTP(otp.text.trim());
                                  } else {
                                    await widget.auth
                                        .requestOTP(phone.text.trim());
                                  }
                                },
                          child: widget.auth.busy
                              ? const SizedBox.square(
                                  dimension: 22,
                                  child: CircularProgressIndicator(
                                      strokeWidth: 2.2, color: Colors.white))
                              : Text(waiting
                                  ? 'Xác nhận OTP'
                                  : 'Vào FlashX Driver'),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 18),
                  const Text(
                      'GPS chỉ được dùng cho vận hành chuyến khi cần thiết.',
                      textAlign: TextAlign.center,
                      style: TextStyle(color: Color(0xFF8590A9), fontSize: 12)),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _DriverMark extends StatelessWidget {
  const _DriverMark();

  @override
  Widget build(BuildContext context) => Container(
        width: 50,
        height: 50,
        decoration: BoxDecoration(
            color: FlashXTheme.yellow, borderRadius: BorderRadius.circular(16)),
        child:
            const Icon(Icons.bolt_rounded, color: FlashXTheme.navy, size: 31),
      );
}
