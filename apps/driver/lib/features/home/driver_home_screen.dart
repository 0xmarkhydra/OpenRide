import 'dart:async';

import 'package:flutter/material.dart';

import '../../core/theme/flashx_theme.dart';
import '../auth/auth_controller.dart';
import '../trips/driver_trip_controller.dart';
import 'driver_map_view.dart';

class DriverHomeScreen extends StatefulWidget {
  const DriverHomeScreen({super.key, required this.auth});
  final AuthController auth;

  @override
  State<DriverHomeScreen> createState() => _DriverHomeScreenState();
}

class _DriverHomeScreenState extends State<DriverHomeScreen> {
  late final DriverTripController trips;
  Timer? _offerTicker;
  int index = 0;

  @override
  void initState() {
    super.initState();
    trips = DriverTripController(auth: widget.auth)..addListener(_refresh);
    trips.initialize();
    _offerTicker = Timer.periodic(const Duration(seconds: 1), (_) {
      if (mounted && trips.currentOffer != null) setState(() {});
    });
  }

  void _refresh() {
    if (mounted) setState(() {});
  }

  @override
  void dispose() {
    _offerTicker?.cancel();
    trips.removeListener(_refresh);
    trips.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: IndexedStack(
          index: index,
          children: [_home(), _earnings(), _history(), _account()]),
      bottomNavigationBar: NavigationBar(
        selectedIndex: index,
        onDestinationSelected: (value) => setState(() => index = value),
        destinations: const [
          NavigationDestination(
              icon: Icon(Icons.home_outlined),
              selectedIcon: Icon(Icons.home_rounded),
              label: 'Trang chủ'),
          NavigationDestination(
              icon: Icon(Icons.account_balance_wallet_outlined),
              selectedIcon: Icon(Icons.account_balance_wallet_rounded),
              label: 'Thu nhập'),
          NavigationDestination(
              icon: Icon(Icons.receipt_long_outlined),
              selectedIcon: Icon(Icons.receipt_long_rounded),
              label: 'Chuyến đi'),
          NavigationDestination(
              icon: Icon(Icons.person_outline_rounded),
              selectedIcon: Icon(Icons.person_rounded),
              label: 'Tài khoản'),
        ],
      ),
    );
  }

  Widget _home() {
    return Stack(
      children: [
        Positioned.fill(
          child: DriverMapView(
            online: trips.online,
            position: trips.lastPosition,
            trip: trips.activeTrip ?? trips.offeredTrip,
          ),
        ),
        Positioned(
          top: 0,
          left: 0,
          right: 0,
          child: SafeArea(
            bottom: false,
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
              child: Row(
                children: [
                  const _DriverMiniBrand(),
                  const Spacer(),
                  _OnlineBadge(online: trips.online),
                  const SizedBox(width: 10),
                  _DriverMapAction(
                      icon: Icons.person_rounded,
                      onTap: () => setState(() => index = 3)),
                ],
              ),
            ),
          ),
        ),
        Align(
          alignment: Alignment.bottomCenter,
          child: SafeArea(
            top: false,
            minimum: const EdgeInsets.fromLTRB(12, 12, 12, 10),
            child: Container(
              constraints: const BoxConstraints(maxWidth: 620),
              padding: const EdgeInsets.fromLTRB(20, 10, 20, 20),
              decoration: BoxDecoration(
                color: FlashXTheme.navy,
                borderRadius: BorderRadius.circular(30),
                boxShadow: const [
                  BoxShadow(
                      color: Color(0x400B132B),
                      blurRadius: 38,
                      offset: Offset(0, 18))
                ],
              ),
              child: ConstrainedBox(
                constraints: BoxConstraints(
                  maxHeight: MediaQuery.sizeOf(context).height * .58,
                ),
                child: SingleChildScrollView(
                  physics: const BouncingScrollPhysics(),
                  child: _panel(),
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _panel() {
    if (!trips.approved) return _pendingApproval();
    final active = trips.activeTrip;
    if (active != null) return _activeTrip(active);
    final offer = trips.currentOffer;
    if (offer != null) return _offer(offer);
    return _idlePanel();
  }

  Widget _pendingApproval() {
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Center(
            child: Container(
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                    color: const Color(0xFF4B5877),
                    borderRadius: BorderRadius.circular(99)))),
        const SizedBox(height: 18),
        Row(
          children: [
            Container(
                width: 50,
                height: 50,
                decoration: BoxDecoration(
                    color: FlashXTheme.yellow,
                    borderRadius: BorderRadius.circular(16)),
                child: const Icon(Icons.fact_check_outlined,
                    color: FlashXTheme.navy)),
            const SizedBox(width: 12),
            const Expanded(
                child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                  Text('Hồ sơ đang chờ duyệt',
                      style: TextStyle(
                          color: Colors.white,
                          fontSize: 22,
                          fontWeight: FontWeight.w900)),
                  SizedBox(height: 3),
                  Text('Bạn chưa thể nhận chuyến lúc này.',
                      style: TextStyle(color: Color(0xFFAFB8CF), fontSize: 13))
                ])),
          ],
        ),
        const SizedBox(height: 16),
        Container(
          padding: const EdgeInsets.all(14),
          decoration: BoxDecoration(
              color: FlashXTheme.navySoft,
              borderRadius: BorderRadius.circular(18)),
          child: Row(children: [
            const Icon(Icons.schedule_rounded,
                color: FlashXTheme.yellow, size: 20),
            const SizedBox(width: 10),
            Expanded(
                child: Text('Trạng thái hiện tại: ${trips.approval}',
                    style: const TextStyle(
                        color: Colors.white, fontWeight: FontWeight.w700)))
          ]),
        ),
        const SizedBox(height: 16),
        OutlinedButton.icon(
          onPressed: widget.auth.refreshProfile,
          style: OutlinedButton.styleFrom(
              foregroundColor: Colors.white,
              side: const BorderSide(color: Color(0xFF46516C))),
          icon: const Icon(Icons.refresh_rounded),
          label: const Text('Kiểm tra lại trạng thái'),
        ),
      ],
    );
  }

  Widget _idlePanel() {
    final theme = Theme.of(context);
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Center(
            child: Container(
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                    color: const Color(0xFF4B5877),
                    borderRadius: BorderRadius.circular(99)))),
        const SizedBox(height: 18),
        Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(trips.online ? 'Bạn đang online' : 'Bắt đầu ca làm việc',
                      style: const TextStyle(
                          color: Colors.white,
                          fontSize: 24,
                          fontWeight: FontWeight.w900,
                          letterSpacing: -.5)),
                  const SizedBox(height: 5),
                  Text(
                    trips.online
                        ? 'FlashX đang tìm chuyến phù hợp gần vị trí của bạn.'
                        : 'Bật nhận chuyến khi bạn đã sẵn sàng di chuyển.',
                    style: const TextStyle(
                        color: Color(0xFFAFB8CF), fontSize: 14, height: 1.4),
                  ),
                ],
              ),
            ),
            const SizedBox(width: 12),
            Container(
              width: 58,
              height: 58,
              decoration: BoxDecoration(
                  color:
                      trips.online ? FlashXTheme.lime : const Color(0xFF27324E),
                  borderRadius: BorderRadius.circular(19)),
              child: Icon(
                  trips.online
                      ? Icons.navigation_rounded
                      : Icons.power_settings_new_rounded,
                  color:
                      trips.online ? FlashXTheme.navy : const Color(0xFF8792AA),
                  size: 30),
            ),
          ],
        ),
        if (trips.online) ...[
          const SizedBox(height: 16),
          Row(
            children: [
              Expanded(
                  child: _DriverStat(
                      label: 'GPS',
                      value: trips.lastPosition == null ? 'Đang lấy' : 'Tốt',
                      icon: Icons.gps_fixed_rounded)),
              const SizedBox(width: 10),
              const Expanded(
                  child: _DriverStat(
                      label: 'Trạng thái',
                      value: 'Sẵn sàng',
                      icon: Icons.bolt_rounded)),
            ],
          ),
        ],
        if (trips.error != null) ...[
          const SizedBox(height: 12),
          Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                  color: const Color(0xFF3A2330),
                  borderRadius: BorderRadius.circular(14)),
              child: Text(trips.error!,
                  style: const TextStyle(
                      color: Color(0xFFFFB4B8),
                      fontSize: 13,
                      fontWeight: FontWeight.w600))),
        ],
        const SizedBox(height: 18),
        FilledButton.icon(
          onPressed: trips.busy ? null : () => trips.setOnline(!trips.online),
          style: FilledButton.styleFrom(
            backgroundColor:
                trips.online ? const Color(0xFF2A3551) : FlashXTheme.yellow,
            foregroundColor: trips.online ? Colors.white : FlashXTheme.navy,
          ),
          icon: Icon(trips.online
              ? Icons.power_settings_new_rounded
              : Icons.play_arrow_rounded),
          label: Text(trips.online ? 'Tắt nhận chuyến' : 'Bật nhận chuyến'),
        ),
        if (!trips.online) ...[
          const SizedBox(height: 10),
          Text('Bạn vẫn có thể xem lịch sử và thu nhập khi offline.',
              textAlign: TextAlign.center,
              style: theme.textTheme.bodySmall
                  ?.copyWith(color: const Color(0xFF7F8AA4))),
        ],
      ],
    );
  }

  Widget _offer(Map<String, dynamic> offer) {
    final distance = (offer['distance_to_pickup_m'] as num?)?.toInt() ?? 0;
    final trip = trips.offeredTrip;
    final estimatedFare = trip?['estimated_fare_minor'] ?? 0;
    final countdown = _offerCountdown(offer['expires_at']);
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Center(
            child: Container(
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                    color: const Color(0xFF4B5877),
                    borderRadius: BorderRadius.circular(99)))),
        const SizedBox(height: 15),
        Row(children: [
          const Expanded(
              child: Text('Có chuyến mới',
                  style: TextStyle(
                      color: Colors.white,
                      fontSize: 25,
                      fontWeight: FontWeight.w900,
                      letterSpacing: -.5))),
          Container(
              padding: const EdgeInsets.symmetric(horizontal: 11, vertical: 7),
              decoration: BoxDecoration(
                  color: FlashXTheme.yellow,
                  borderRadius: BorderRadius.circular(999)),
              child: Row(children: [
                const Icon(Icons.timer_outlined,
                    color: FlashXTheme.navy, size: 17),
                const SizedBox(width: 5),
                Text(countdown,
                    style: const TextStyle(
                        color: FlashXTheme.navy,
                        fontSize: 11,
                        fontWeight: FontWeight.w900,
                        fontFeatures: [FontFeature.tabularFigures()]))
              ]))
        ]),
        const SizedBox(height: 6),
        Text('Điểm đón cách bạn ${(distance / 1000).toStringAsFixed(1)} km',
            style: const TextStyle(color: Color(0xFFAFB8CF), fontSize: 14)),
        const SizedBox(height: 16),
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
              color: FlashXTheme.navySoft,
              borderRadius: BorderRadius.circular(20)),
          child: Column(
            children: [
              _DriverRouteRow(
                  icon: Icons.radio_button_checked_rounded,
                  color: FlashXTheme.lime,
                  title: 'Điểm đón',
                  subtitle:
                      'Khoảng ${(distance / 1000).toStringAsFixed(1)} km từ vị trí hiện tại'),
              const Padding(
                  padding: EdgeInsets.only(left: 10),
                  child: Divider(height: 20, color: Color(0xFF34405E))),
              _DriverRouteRow(
                  icon: Icons.location_on_rounded,
                  color: FlashXTheme.yellow,
                  title: 'Điểm trả',
                  subtitle: trip == null
                      ? 'Nhận chuyến để xem chi tiết'
                      : 'Điểm đến đã xác nhận'),
            ],
          ),
        ),
        const SizedBox(height: 14),
        Row(children: [
          Expanded(
              child: _DriverStat(
                  label: 'Khoảng đón',
                  value: '${(distance / 1000).toStringAsFixed(1)} km',
                  icon: Icons.near_me_rounded)),
          const SizedBox(width: 10),
          Expanded(
              child: _DriverStat(
                  label: 'Giá chuyến',
                  value: _money(estimatedFare),
                  icon: Icons.payments_outlined))
        ]),
        const SizedBox(height: 18),
        Row(
          children: [
            Expanded(
                child: OutlinedButton(
                    onPressed: trips.busy ? null : trips.rejectOffer,
                    style: OutlinedButton.styleFrom(
                        foregroundColor: Colors.white,
                        side: const BorderSide(color: Color(0xFF46516C))),
                    child: const Text('Từ chối'))),
            const SizedBox(width: 10),
            Expanded(
                flex: 2,
                child: FilledButton.icon(
                    onPressed: trips.busy ? null : trips.acceptOffer,
                    style: FilledButton.styleFrom(
                        backgroundColor: FlashXTheme.yellow,
                        foregroundColor: FlashXTheme.navy),
                    icon: const Icon(Icons.bolt_rounded),
                    label: const Text('Nhận chuyến'))),
          ],
        ),
      ],
    );
  }

  Widget _activeTrip(Map<String, dynamic> trip) {
    final status = trip['status']?.toString() ?? '';
    final (label, hint, action, callback, icon) = switch (status) {
      'accepted' => (
          'Đã nhận chuyến',
          'Đi tới điểm đón của khách.',
          'Bắt đầu đến đón',
          trips.arriving,
          Icons.navigation_rounded
        ),
      'arriving' => (
          'Đang đến điểm đón',
          'Xác nhận khi bạn đã tới đúng vị trí.',
          'Đã đến điểm đón',
          trips.arrived,
          Icons.pin_drop_rounded
        ),
      'arrived' => (
          'Đã tới điểm đón',
          'Chỉ bắt đầu sau khi khách đã lên xe.',
          'Bắt đầu chuyến',
          trips.startTrip,
          Icons.play_arrow_rounded
        ),
      'in_progress' => (
          'Đang thực hiện chuyến',
          'Lái xe an toàn và tập trung vào hành trình.',
          'Hoàn thành chuyến',
          trips.completeTrip,
          Icons.flag_rounded
        ),
      _ => (
          'Chuyến đang hoạt động',
          'FlashX đang đồng bộ trạng thái chuyến.',
          'Làm mới',
          trips.refreshOffer,
          Icons.refresh_rounded
        ),
    };
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Center(
            child: Container(
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                    color: const Color(0xFF4B5877),
                    borderRadius: BorderRadius.circular(99)))),
        const SizedBox(height: 16),
        Row(children: [
          Container(
              width: 52,
              height: 52,
              decoration: BoxDecoration(
                  color: FlashXTheme.yellow,
                  borderRadius: BorderRadius.circular(17)),
              child: Icon(icon, color: FlashXTheme.navy, size: 28)),
          const SizedBox(width: 12),
          Expanded(
              child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                Text(label,
                    style: const TextStyle(
                        color: Colors.white,
                        fontSize: 23,
                        fontWeight: FontWeight.w900)),
                const SizedBox(height: 3),
                Text(hint,
                    style: const TextStyle(
                        color: Color(0xFFAFB8CF), fontSize: 13, height: 1.35))
              ]))
        ]),
        const SizedBox(height: 16),
        _DriverProgress(status: status),
        const SizedBox(height: 16),
        Container(
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(
                color: FlashXTheme.navySoft,
                borderRadius: BorderRadius.circular(18)),
            child: Row(children: [
              const Icon(Icons.confirmation_number_outlined,
                  color: FlashXTheme.yellow, size: 20),
              const SizedBox(width: 10),
              Expanded(
                  child: Text(
                      'Chuyến ${_shortId(trip['id']?.toString() ?? '')}',
                      style: const TextStyle(
                          color: Colors.white, fontWeight: FontWeight.w700))),
              IconButton(
                  onPressed: () {},
                  tooltip: 'Gọi khách',
                  icon: const Icon(Icons.call_rounded, color: Colors.white))
            ])),
        const SizedBox(height: 16),
        FilledButton.icon(
            onPressed: trips.busy ? null : () => callback(),
            style: FilledButton.styleFrom(
                backgroundColor: FlashXTheme.yellow,
                foregroundColor: FlashXTheme.navy),
            icon: Icon(icon),
            label: Text(action)),
      ],
    );
  }

  Widget _earnings() {
    final theme = Theme.of(context);
    return SafeArea(
      child: RefreshIndicator(
        onRefresh: trips.loadHistory,
        child: ListView(
          padding: const EdgeInsets.fromLTRB(20, 24, 20, 32),
          children: [
            const _DriverPageBrand(),
            const SizedBox(height: 24),
            Text('Thu nhập', style: theme.textTheme.headlineMedium),
            const SizedBox(height: 6),
            Text('Tổng quan hiệu suất từ các chuyến đã hoàn thành.',
                style: theme.textTheme.bodyMedium),
            const SizedBox(height: 20),
            Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                  color: FlashXTheme.navy,
                  borderRadius: BorderRadius.circular(26)),
              child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text('Doanh thu chuyến',
                        style: TextStyle(
                            color: Color(0xFFAEB8CF),
                            fontSize: 13,
                            fontWeight: FontWeight.w700)),
                    const SizedBox(height: 7),
                    Text(_money(trips.totalEarningsMinor),
                        style: const TextStyle(
                            color: Colors.white,
                            fontSize: 34,
                            fontWeight: FontWeight.w900,
                            letterSpacing: -1.2)),
                    const SizedBox(height: 18),
                    Row(children: [
                      Expanded(
                          child: _EarningsMetric(
                              label: 'Chuyến hoàn thành',
                              value: '${trips.completedTrips}')),
                      const SizedBox(width: 10),
                      const Expanded(
                          child: _EarningsMetric(
                              label: 'Đối soát', value: 'Gross fare'))
                    ])
                  ]),
            ),
            const SizedBox(height: 14),
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(20),
                border: Border.all(color: FlashXTheme.border),
              ),
              child: const Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Icon(Icons.info_outline_rounded,
                      color: FlashXTheme.navy, size: 21),
                  SizedBox(width: 10),
                  Expanded(
                    child: Text(
                      'MVP đang hiển thị gross trip fare. Commission và settlement sẽ theo chính sách vận hành khi cấu hình production.',
                      style: TextStyle(
                          color: FlashXTheme.textSecondary,
                          fontSize: 13,
                          height: 1.45),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _history() {
    final theme = Theme.of(context);
    return SafeArea(
      child: RefreshIndicator(
        onRefresh: trips.loadHistory,
        child: ListView(
          padding: const EdgeInsets.fromLTRB(20, 24, 20, 32),
          children: [
            const _DriverPageBrand(),
            const SizedBox(height: 24),
            Text('Chuyến đi', style: theme.textTheme.headlineMedium),
            const SizedBox(height: 6),
            Text('Lịch sử hoạt động gần đây của bạn.',
                style: theme.textTheme.bodyMedium),
            const SizedBox(height: 20),
            if (trips.history.isEmpty) const _DriverEmpty(),
            ...trips.history.map((trip) => Padding(
                  padding: const EdgeInsets.only(bottom: 10),
                  child: Container(
                    padding: const EdgeInsets.all(15),
                    decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(20),
                        border: Border.all(color: FlashXTheme.border)),
                    child: Row(children: [
                      Container(
                          width: 46,
                          height: 46,
                          decoration: BoxDecoration(
                              color: const Color(0xFFFFF7C2),
                              borderRadius: BorderRadius.circular(15)),
                          child: const Icon(Icons.route_rounded,
                              color: FlashXTheme.navy)),
                      const SizedBox(width: 12),
                      Expanded(
                          child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                            Text(_tripStatus(trip['status']?.toString() ?? ''),
                                style: const TextStyle(
                                    fontWeight: FontWeight.w800)),
                            const SizedBox(height: 4),
                            Text(_shortId(trip['id']?.toString() ?? ''),
                                style: theme.textTheme.bodySmall?.copyWith(
                                    color: FlashXTheme.textSecondary))
                          ])),
                      Text(
                          _money(trip['final_fare_minor'] ??
                              trip['estimated_fare_minor'] ??
                              0),
                          style: const TextStyle(
                              color: FlashXTheme.navy,
                              fontWeight: FontWeight.w900))
                    ]),
                  ),
                )),
          ],
        ),
      ),
    );
  }

  Widget _account() {
    final theme = Theme.of(context);
    final profile = widget.auth.profile;
    return SafeArea(
      child: ListView(
        padding: const EdgeInsets.fromLTRB(20, 24, 20, 32),
        children: [
          const _DriverPageBrand(),
          const SizedBox(height: 24),
          Text('Tài khoản tài xế', style: theme.textTheme.headlineMedium),
          const SizedBox(height: 18),
          Container(
            padding: const EdgeInsets.all(18),
            decoration: BoxDecoration(
                color: FlashXTheme.navy,
                borderRadius: BorderRadius.circular(24)),
            child: Row(children: [
              Container(
                  width: 58,
                  height: 58,
                  decoration: BoxDecoration(
                      color: FlashXTheme.yellow,
                      borderRadius: BorderRadius.circular(19)),
                  child: const Icon(Icons.two_wheeler_rounded,
                      color: FlashXTheme.navy, size: 30)),
              const SizedBox(width: 14),
              Expanded(
                  child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                    Text(
                        profile?['full_name']?.toString().isNotEmpty == true
                            ? profile!['full_name'].toString()
                            : 'Tài xế FlashX',
                        style: const TextStyle(
                            color: Colors.white,
                            fontSize: 18,
                            fontWeight: FontWeight.w800)),
                    const SizedBox(height: 4),
                    Text(profile?['phone']?.toString() ?? '',
                        style: const TextStyle(color: Color(0xFFB9C1D5)))
                  ])),
              _ApprovalChip(status: trips.approval)
            ]),
          ),
          const SizedBox(height: 16),
          const _DriverAccountItem(
              icon: Icons.badge_outlined,
              title: 'Hồ sơ & KYC',
              subtitle: 'Giấy tờ tài xế và trạng thái duyệt'),
          const SizedBox(height: 10),
          const _DriverAccountItem(
              icon: Icons.directions_car_outlined,
              title: 'Phương tiện',
              subtitle: 'Thông tin xe đang hoạt động'),
          const SizedBox(height: 10),
          const _DriverAccountItem(
              icon: Icons.shield_outlined,
              title: 'An toàn',
              subtitle: 'Quyền vị trí và hỗ trợ khẩn cấp'),
          const SizedBox(height: 24),
          OutlinedButton.icon(
              onPressed: widget.auth.logout,
              icon: const Icon(Icons.logout_rounded),
              label: const Text('Đăng xuất')),
        ],
      ),
    );
  }

  static String _offerCountdown(dynamic expiresAt) {
    final expiry = DateTime.tryParse(expiresAt?.toString() ?? '');
    if (expiry == null) return '--:--';
    final seconds = expiry.toUtc().difference(DateTime.now().toUtc()).inSeconds;
    final safeSeconds = seconds < 0 ? 0 : seconds;
    final minutes = safeSeconds ~/ 60;
    final remainder = safeSeconds % 60;
    return '${minutes.toString().padLeft(2, '0')}:${remainder.toString().padLeft(2, '0')}';
  }

  static String _money(dynamic value) {
    final amount =
        value is num ? value.toInt() : int.tryParse(value.toString()) ?? 0;
    return '${amount.toString().replaceAllMapped(RegExp(r'(?=(\d{3})+(?!\d))'), (m) => '.')}₫';
  }

  static String _shortId(String value) => value.length <= 8
      ? value
      : value.substring(value.length - 8).toUpperCase();

  static String _tripStatus(String status) => switch (status) {
        'searching' => 'Đang tìm tài xế',
        'accepted' => 'Đã nhận chuyến',
        'arriving' => 'Đang đến đón',
        'arrived' => 'Đã tới điểm đón',
        'in_progress' => 'Đang thực hiện',
        'completed' => 'Hoàn thành',
        'cancelled' => 'Đã hủy',
        _ => 'Chuyến đi',
      };
}

class _DriverMiniBrand extends StatelessWidget {
  const _DriverMiniBrand();
  @override
  Widget build(BuildContext context) => Container(
      padding: const EdgeInsets.fromLTRB(8, 7, 12, 7),
      decoration: BoxDecoration(
          color: FlashXTheme.navy,
          borderRadius: BorderRadius.circular(18),
          boxShadow: const [
            BoxShadow(
                color: Color(0x22000000), blurRadius: 16, offset: Offset(0, 6))
          ]),
      child: const Row(mainAxisSize: MainAxisSize.min, children: [
        Icon(Icons.bolt_rounded, color: FlashXTheme.yellow, size: 25),
        SizedBox(width: 4),
        Text('FlashX Driver',
            style: TextStyle(
                color: Colors.white,
                fontSize: 15,
                fontWeight: FontWeight.w900,
                letterSpacing: -.4))
      ]));
}

class _OnlineBadge extends StatelessWidget {
  const _OnlineBadge({required this.online});
  final bool online;
  @override
  Widget build(BuildContext context) => Container(
      padding: const EdgeInsets.symmetric(horizontal: 11, vertical: 9),
      decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(999),
          border: Border.all(color: FlashXTheme.border),
          boxShadow: const [
            BoxShadow(
                color: Color(0x18000000), blurRadius: 12, offset: Offset(0, 5))
          ]),
      child: Row(mainAxisSize: MainAxisSize.min, children: [
        Container(
            width: 8,
            height: 8,
            decoration: BoxDecoration(
                color: online ? FlashXTheme.success : const Color(0xFF9CA3AF),
                shape: BoxShape.circle)),
        const SizedBox(width: 7),
        Text(online ? 'ONLINE' : 'OFFLINE',
            style: const TextStyle(
                color: FlashXTheme.navy,
                fontSize: 10,
                fontWeight: FontWeight.w900,
                letterSpacing: .7))
      ]));
}

class _DriverMapAction extends StatelessWidget {
  const _DriverMapAction({required this.icon, required this.onTap});
  final IconData icon;
  final VoidCallback onTap;
  @override
  Widget build(BuildContext context) => Material(
      color: Colors.white,
      shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(17),
          side: const BorderSide(color: FlashXTheme.border)),
      elevation: 3,
      shadowColor: const Color(0x22000000),
      child: IconButton(
          onPressed: onTap, icon: Icon(icon, color: FlashXTheme.navy)));
}

class _DriverStat extends StatelessWidget {
  const _DriverStat(
      {required this.label, required this.value, required this.icon});
  final String label;
  final String value;
  final IconData icon;
  @override
  Widget build(BuildContext context) => Container(
      padding: const EdgeInsets.all(13),
      decoration: BoxDecoration(
          color: FlashXTheme.navySoft, borderRadius: BorderRadius.circular(17)),
      child: Row(children: [
        Container(
            width: 34,
            height: 34,
            decoration: BoxDecoration(
                color: const Color(0xFF24304C),
                borderRadius: BorderRadius.circular(11)),
            child: Icon(icon, color: FlashXTheme.yellow, size: 18)),
        const SizedBox(width: 9),
        Expanded(
            child:
                Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text(label,
              style: const TextStyle(
                  color: Color(0xFF8995AF),
                  fontSize: 10.5,
                  fontWeight: FontWeight.w600)),
          const SizedBox(height: 2),
          Text(value,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                  color: Colors.white,
                  fontSize: 13,
                  fontWeight: FontWeight.w800))
        ]))
      ]));
}

class _DriverRouteRow extends StatelessWidget {
  const _DriverRouteRow(
      {required this.icon,
      required this.color,
      required this.title,
      required this.subtitle});
  final IconData icon;
  final Color color;
  final String title;
  final String subtitle;
  @override
  Widget build(BuildContext context) => Row(children: [
        Icon(icon, color: color, size: 20),
        const SizedBox(width: 11),
        Expanded(
            child:
                Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text(title,
              style: const TextStyle(
                  color: Colors.white,
                  fontSize: 13,
                  fontWeight: FontWeight.w800)),
          const SizedBox(height: 2),
          Text(subtitle,
              style: const TextStyle(color: Color(0xFF8995AF), fontSize: 11.5))
        ]))
      ]);
}

class _DriverProgress extends StatelessWidget {
  const _DriverProgress({required this.status});
  final String status;
  @override
  Widget build(BuildContext context) {
    const steps = ['accepted', 'arriving', 'arrived', 'in_progress'];
    final index = steps.indexOf(status);
    return Row(
        children: List.generate(
            steps.length,
            (i) => Expanded(
                    child: Row(children: [
                  Expanded(
                      child: AnimatedContainer(
                          duration: const Duration(milliseconds: 220),
                          height: 5,
                          decoration: BoxDecoration(
                              color: i <= index
                                  ? FlashXTheme.yellow
                                  : const Color(0xFF34405E),
                              borderRadius: BorderRadius.circular(99)))),
                  if (i < steps.length - 1) const SizedBox(width: 5)
                ]))));
  }
}

class _DriverPageBrand extends StatelessWidget {
  const _DriverPageBrand();
  @override
  Widget build(BuildContext context) => const Row(children: [
        Icon(Icons.bolt_rounded, color: FlashXTheme.navy, size: 28),
        SizedBox(width: 4),
        Text('FlashX Driver',
            style: TextStyle(
                color: FlashXTheme.navy,
                fontSize: 20,
                fontWeight: FontWeight.w900,
                letterSpacing: -.7))
      ]);
}

class _EarningsMetric extends StatelessWidget {
  const _EarningsMetric({required this.label, required this.value});
  final String label;
  final String value;
  @override
  Widget build(BuildContext context) => Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
          color: FlashXTheme.navySoft, borderRadius: BorderRadius.circular(17)),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Text(label,
            style: const TextStyle(
                color: Color(0xFF8995AF),
                fontSize: 10.5,
                fontWeight: FontWeight.w600)),
        const SizedBox(height: 5),
        Text(value,
            style: const TextStyle(
                color: Colors.white, fontSize: 15, fontWeight: FontWeight.w800))
      ]));
}

class _DriverEmpty extends StatelessWidget {
  const _DriverEmpty();
  @override
  Widget build(BuildContext context) => Container(
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(22),
          border: Border.all(color: FlashXTheme.border)),
      child: Column(children: [
        Container(
            width: 52,
            height: 52,
            decoration: BoxDecoration(
                color: const Color(0xFFF2F4F7),
                borderRadius: BorderRadius.circular(17)),
            child: const Icon(Icons.route_outlined, color: FlashXTheme.navy)),
        const SizedBox(height: 12),
        const Text('Chưa có chuyến đi',
            style: TextStyle(fontWeight: FontWeight.w800)),
        const SizedBox(height: 4),
        Text('Khi bạn hoàn thành chuyến đầu tiên, lịch sử sẽ xuất hiện ở đây.',
            textAlign: TextAlign.center,
            style: Theme.of(context).textTheme.bodyMedium)
      ]));
}

class _ApprovalChip extends StatelessWidget {
  const _ApprovalChip({required this.status});
  final String status;
  @override
  Widget build(BuildContext context) {
    final approved = status == 'approved';
    return Container(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 7),
        decoration: BoxDecoration(
            color: approved ? const Color(0xFF163D37) : const Color(0xFF3D3420),
            borderRadius: BorderRadius.circular(999)),
        child: Text(approved ? 'ĐÃ DUYỆT' : 'CHỜ DUYỆT',
            style: TextStyle(
                color: approved ? FlashXTheme.lime : FlashXTheme.yellow,
                fontSize: 9.5,
                fontWeight: FontWeight.w900,
                letterSpacing: .6)));
  }
}

class _DriverAccountItem extends StatelessWidget {
  const _DriverAccountItem(
      {required this.icon, required this.title, required this.subtitle});
  final IconData icon;
  final String title;
  final String subtitle;
  @override
  Widget build(BuildContext context) => Container(
      padding: const EdgeInsets.all(15),
      decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(19),
          border: Border.all(color: FlashXTheme.border)),
      child: Row(children: [
        Container(
            width: 42,
            height: 42,
            decoration: BoxDecoration(
                color: const Color(0xFFF2F4F7),
                borderRadius: BorderRadius.circular(14)),
            child: Icon(icon, color: FlashXTheme.navy, size: 21)),
        const SizedBox(width: 12),
        Expanded(
            child:
                Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text(title, style: const TextStyle(fontWeight: FontWeight.w800)),
          const SizedBox(height: 3),
          Text(subtitle,
              style: Theme.of(context)
                  .textTheme
                  .bodySmall
                  ?.copyWith(color: FlashXTheme.textSecondary))
        ])),
      ]));
}
