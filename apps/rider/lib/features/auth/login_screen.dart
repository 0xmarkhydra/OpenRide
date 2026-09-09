import 'package:flutter/material.dart';

import '../../core/theme/flashx_theme.dart';
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
    final theme = Theme.of(context);
    return Scaffold(
      backgroundColor: FlashXTheme.navy,
      body: Stack(
        children: [
          const Positioned(
            right: -90,
            top: -70,
            child: _GlowOrb(size: 260, color: FlashXTheme.yellow),
          ),
          const Positioned(
            left: -110,
            bottom: 100,
            child: _GlowOrb(size: 240, color: FlashXTheme.lime),
          ),
          SafeArea(
            child: Center(
              child: SingleChildScrollView(
                padding: const EdgeInsets.fromLTRB(20, 32, 20, 24),
                child: ConstrainedBox(
                  constraints: const BoxConstraints(maxWidth: 440),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      const _BrandLockup(),
                      const SizedBox(height: 42),
                      Text(
                        awaitingCode
                            ? 'Xác nhận số điện thoại'
                            : 'Bạn chọn chuyến. Tài xế chọn giá.',
                        style: theme.textTheme.displaySmall?.copyWith(
                          color: Colors.white,
                        ),
                      ),
                      const SizedBox(height: 12),
                      Text(
                        awaitingCode
                            ? 'Nhập mã OTP 6 số vừa được gửi đến ${_phone.text.trim()}.'
                            : 'OpenRide kết nối bạn với các tài xế phù hợp để so sánh giá, thời gian đón và độ tin cậy một cách minh bạch.',
                        style: theme.textTheme.bodyLarge?.copyWith(
                          color: const Color(0xFFB9C1D5),
                        ),
                      ),
                      const SizedBox(height: 30),
                      Container(
                        padding: const EdgeInsets.all(20),
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: BorderRadius.circular(28),
                          boxShadow: const [
                            BoxShadow(
                              color: Color(0x33000000),
                              blurRadius: 40,
                              offset: Offset(0, 18),
                            ),
                          ],
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.stretch,
                          children: [
                            Text(
                              awaitingCode ? 'Nhập mã OTP' : 'Đăng nhập OpenRide',
                              style: theme.textTheme.titleLarge,
                            ),
                            const SizedBox(height: 6),
                            Text(
                              awaitingCode
                                  ? 'Mã xác thực có hiệu lực trong thời gian ngắn.'
                                  : 'Dùng số điện thoại của bạn để tiếp tục.',
                              style: theme.textTheme.bodyMedium,
                            ),
                            const SizedBox(height: 20),
                            TextField(
                              controller: _phone,
                              enabled: !awaitingCode && !widget.auth.busy,
                              keyboardType: TextInputType.phone,
                              textInputAction: awaitingCode
                                  ? TextInputAction.next
                                  : TextInputAction.done,
                              decoration: const InputDecoration(
                                labelText: 'Số điện thoại',
                                hintText: '+84...',
                                prefixIcon: Icon(Icons.phone_iphone_rounded),
                              ),
                            ),
                            if (awaitingCode) ...[
                              const SizedBox(height: 14),
                              TextField(
                                controller: _otp,
                                autofocus: true,
                                enabled: !widget.auth.busy,
                                keyboardType: TextInputType.number,
                                maxLength: 6,
                                textAlign: TextAlign.center,
                                style: const TextStyle(
                                  fontSize: 24,
                                  fontWeight: FontWeight.w800,
                                  letterSpacing: 8,
                                ),
                                decoration: InputDecoration(
                                  labelText: 'Mã OTP',
                                  counterText: '',
                                  prefixIcon:
                                      const Icon(Icons.lock_outline_rounded),
                                  helperText: widget.auth.developmentCode ==
                                          null
                                      ? null
                                      : 'DEV OTP: ${widget.auth.developmentCode}',
                                ),
                              ),
                            ],
                            if (widget.auth.error != null) ...[
                              const SizedBox(height: 12),
                              _InlineError(message: widget.auth.error!),
                            ],
                            const SizedBox(height: 20),
                            FilledButton(
                              onPressed: widget.auth.busy
                                  ? null
                                  : () async {
                                      if (!awaitingCode) {
                                        await widget.auth
                                            .requestOTP(_phone.text.trim());
                                      } else {
                                        await widget.auth
                                            .verifyOTP(_otp.text.trim());
                                      }
                                    },
                              child: widget.auth.busy
                                  ? const SizedBox.square(
                                      dimension: 22,
                                      child: CircularProgressIndicator(
                                        strokeWidth: 2.2,
                                        color: Colors.white,
                                      ),
                                    )
                                  : Row(
                                      mainAxisAlignment:
                                          MainAxisAlignment.center,
                                      children: [
                                        Text(awaitingCode
                                            ? 'Xác nhận OTP'
                                            : 'Tiếp tục'),
                                        const SizedBox(width: 8),
                                        const Icon(Icons.arrow_forward_rounded,
                                            size: 20),
                                      ],
                                    ),
                            ),
                            if (awaitingCode) ...[
                              const SizedBox(height: 6),
                              TextButton(
                                onPressed: widget.auth.busy
                                    ? null
                                    : () => setState(() {
                                          widget.auth.challengeId = null;
                                          _otp.clear();
                                        }),
                                child: const Text('Đổi số điện thoại'),
                              ),
                            ],
                          ],
                        ),
                      ),
                      const SizedBox(height: 20),
                      Text(
                        'Bằng việc tiếp tục, bạn đồng ý với điều khoản sử dụng và chính sách quyền riêng tư của OpenRide.',
                        textAlign: TextAlign.center,
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: const Color(0xFF8590A9),
                          height: 1.45,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _BrandLockup extends StatelessWidget {
  const _BrandLockup();

  @override
  Widget build(BuildContext context) => Row(
        children: [
          Container(
            width: 50,
            height: 50,
            decoration: BoxDecoration(
              color: FlashXTheme.yellow,
              borderRadius: BorderRadius.circular(16),
            ),
            child: const Icon(Icons.swap_horiz_rounded,
                color: FlashXTheme.navy, size: 30),
          ),
          const SizedBox(width: 12),
          const Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'OpenRide',
                style: TextStyle(
                  color: Colors.white,
                  fontSize: 24,
                  fontWeight: FontWeight.w900,
                  letterSpacing: -1,
                ),
              ),
              Text(
                'RIDER',
                style: TextStyle(
                  color: Color(0xFF98A3BC),
                  fontSize: 12,
                  fontWeight: FontWeight.w700,
                  letterSpacing: 1.1,
                ),
              ),
            ],
          ),
        ],
      );
}

class _GlowOrb extends StatelessWidget {
  const _GlowOrb({required this.size, required this.color});
  final double size;
  final Color color;

  @override
  Widget build(BuildContext context) => Container(
        width: size,
        height: size,
        decoration: BoxDecoration(
          shape: BoxShape.circle,
          color: color.withValues(alpha: .08),
          boxShadow: [
            BoxShadow(
              color: color.withValues(alpha: .08),
              blurRadius: 80,
              spreadRadius: 20,
            ),
          ],
        ),
      );
}

class _InlineError extends StatelessWidget {
  const _InlineError({required this.message});
  final String message;

  @override
  Widget build(BuildContext context) => Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: const Color(0xFFFFEEEE),
          borderRadius: BorderRadius.circular(14),
          border: Border.all(color: const Color(0xFFFFD2D2)),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Icon(Icons.error_outline_rounded,
                color: FlashXTheme.danger, size: 19),
            const SizedBox(width: 9),
            Expanded(
              child: Text(
                message,
                style: const TextStyle(
                  color: FlashXTheme.danger,
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ),
          ],
        ),
      );
}
