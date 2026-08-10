import 'package:flutter/material.dart';

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
  int index = 0;

  @override
  void initState() {
    super.initState();
    trips = DriverTripController(auth: widget.auth)..addListener(_refresh);
    trips.initialize();
  }

  void _refresh() {
    if (mounted) setState(() {});
  }

  @override
  void dispose() {
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
    return Stack(children: [
      Positioned.fill(
        child: DriverMapView(
          online: trips.online,
          position: trips.lastPosition,
          trip: trips.activeTrip ?? trips.offeredTrip,
        ),
      ),
      SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(children: [
            const CircleAvatar(child: Icon(Icons.two_wheeler_rounded)),
            const Spacer(),
            Chip(
              avatar: CircleAvatar(
                  radius: 5,
                  backgroundColor: trips.online ? Colors.green : Colors.grey),
              label: Text(trips.online ? 'Đang online' : 'Đang offline'),
            ),
          ]),
        ),
      ),
      Align(
        alignment: Alignment.bottomCenter,
        child: SafeArea(
          top: false,
          minimum: const EdgeInsets.all(12),
          child: Container(
            padding: const EdgeInsets.all(18),
            decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(28),
                boxShadow: const [
                  BoxShadow(blurRadius: 28, color: Color(0x22000000))
                ]),
            child: _panel(),
          ),
        ),
      ),
    ]);
  }

  Widget _panel() {
    if (!trips.approved) {
      return Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Icon(Icons.fact_check_outlined, size: 40),
            const SizedBox(height: 10),
            const Text('Hồ sơ đang chờ duyệt',
                style: TextStyle(fontSize: 22, fontWeight: FontWeight.w900)),
            const SizedBox(height: 6),
            Text(
                'Trạng thái: ${trips.approval}. Admin cần duyệt tài xế trước khi nhận chuyến.'),
            const SizedBox(height: 14),
            OutlinedButton(
                onPressed: widget.auth.refreshProfile,
                child: const Text('Kiểm tra lại trạng thái')),
          ]);
    }
    final active = trips.activeTrip;
    if (active != null) return _activeTrip(active);
    final offer = trips.currentOffer;
    if (offer != null) return _offer(offer);
    return Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(trips.online ? 'Sẵn sàng nhận chuyến' : 'Bắt đầu làm việc',
              style:
                  const TextStyle(fontSize: 23, fontWeight: FontWeight.w900)),
          const SizedBox(height: 6),
          Text(trips.online
              ? 'GPS đang gửi vị trí để hệ thống tìm khách gần bạn.'
              : 'Bật online khi bạn đã sẵn sàng.'),
          if (trips.error != null) ...[
            const SizedBox(height: 8),
            Text(trips.error!,
                style: TextStyle(color: Theme.of(context).colorScheme.error))
          ],
          const SizedBox(height: 16),
          FilledButton.icon(
            onPressed: trips.busy ? null : () => trips.setOnline(!trips.online),
            style: FilledButton.styleFrom(
                backgroundColor: trips.online
                    ? Theme.of(context).colorScheme.error
                    : Theme.of(context).colorScheme.primary),
            icon: Icon(trips.online
                ? Icons.power_settings_new_rounded
                : Icons.play_arrow_rounded),
            label: Padding(
                padding: const EdgeInsets.symmetric(vertical: 12),
                child:
                    Text(trips.online ? 'Tắt nhận chuyến' : 'Bật nhận chuyến')),
          ),
        ]);
  }

  Widget _offer(Map<String, dynamic> offer) {
    final distance = (offer['distance_to_pickup_m'] as num?)?.toInt() ?? 0;
    return Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const Text('Có chuyến mới',
              style: TextStyle(fontSize: 24, fontWeight: FontWeight.w900)),
          const SizedBox(height: 6),
          Text(
              'Cách điểm đón khoảng ${(distance / 1000).toStringAsFixed(1)} km'),
          Text('Offer: ${offer['id']}', style: const TextStyle(fontSize: 12)),
          const SizedBox(height: 16),
          Row(children: [
            Expanded(
                child: OutlinedButton(
                    onPressed: trips.busy ? null : trips.rejectOffer,
                    child: const Text('Từ chối'))),
            const SizedBox(width: 10),
            Expanded(
                child: FilledButton(
                    onPressed: trips.busy ? null : trips.acceptOffer,
                    child: const Text('Nhận chuyến'))),
          ]),
        ]);
  }

  Widget _activeTrip(Map<String, dynamic> trip) {
    final status = trip['status']?.toString() ?? '';
    final (label, action, callback) = switch (status) {
      'accepted' => ('Đã nhận chuyến', 'Bắt đầu đến đón', trips.arriving),
      'arriving' => ('Đang đến điểm đón', 'Đã đến điểm đón', trips.arrived),
      'arrived' => ('Đã tới điểm đón', 'Bắt đầu chuyến', trips.startTrip),
      'in_progress' => (
          'Đang thực hiện chuyến',
          'Hoàn thành chuyến',
          trips.completeTrip
        ),
      _ => ('Chuyến đang hoạt động', 'Làm mới', trips.refreshOffer),
    };
    return Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(label,
              style:
                  const TextStyle(fontSize: 23, fontWeight: FontWeight.w900)),
          const SizedBox(height: 6),
          Text('Mã chuyến: ${trip['id']}'),
          const SizedBox(height: 14),
          FilledButton(
              onPressed: trips.busy ? null : () => callback(),
              child: Padding(
                  padding: const EdgeInsets.symmetric(vertical: 12),
                  child: Text(action))),
        ]);
  }

  Widget _earnings() => SafeArea(
        child: RefreshIndicator(
          onRefresh: trips.loadHistory,
          child: ListView(
            padding: const EdgeInsets.all(24),
            children: [
              const Text('Thu nhập',
                  style: TextStyle(fontSize: 28, fontWeight: FontWeight.w900)),
              const SizedBox(height: 16),
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(20),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text('Tổng chuyến đã hoàn thành'),
                      const SizedBox(height: 6),
                      Text('${trips.completedTrips}',
                          style: const TextStyle(
                              fontSize: 30, fontWeight: FontWeight.w900)),
                      const SizedBox(height: 18),
                      const Text('Doanh thu chuyến'),
                      const SizedBox(height: 6),
                      Text(_money(trips.totalEarningsMinor),
                          style: const TextStyle(
                              fontSize: 30, fontWeight: FontWeight.w900)),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 8),
              const Text(
                  'MVP hiển thị gross trip fare. Commission/settlement sẽ được áp theo chính sách tài xế của khách hàng.'),
            ],
          ),
        ),
      );

  Widget _history() => SafeArea(
        child: RefreshIndicator(
          onRefresh: trips.loadHistory,
          child: ListView(
            padding: const EdgeInsets.all(24),
            children: [
              const Text('Chuyến đi',
                  style: TextStyle(fontSize: 28, fontWeight: FontWeight.w900)),
              const SizedBox(height: 16),
              if (trips.history.isEmpty)
                const Card(
                    child: Padding(
                        padding: EdgeInsets.all(20),
                        child: Text('Chưa có chuyến đi.'))),
              ...trips.history.map(
                (trip) => Card(
                  child: ListTile(
                    leading: const CircleAvatar(child: Icon(Icons.route_rounded)),
                    title: Text(_tripStatus(trip['status']?.toString() ?? '')),
                    subtitle: Text(trip['id']?.toString() ?? ''),
                    trailing: Text(_money(trip['final_fare_minor'] ??
                        trip['estimated_fare_minor'] ??
                        0)),
                  ),
                ),
              ),
            ],
          ),
        ),
      );

  static String _money(dynamic value) {
    final amount =
        value is num ? value.toInt() : int.tryParse(value.toString()) ?? 0;
    return '${amount.toString().replaceAllMapped(RegExp(r'(?=(\\d{3})+(?!\\d))'), (m) => '.')}₫';
  }

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

  Widget _account() => SafeArea(
      child: Padding(
          padding: const EdgeInsets.all(24),
          child:
              Column(crossAxisAlignment: CrossAxisAlignment.stretch, children: [
            const Icon(Icons.person_rounded, size: 54),
            const SizedBox(height: 16),
            const Text('Tài khoản tài xế',
                style: TextStyle(fontSize: 28, fontWeight: FontWeight.w900)),
            const SizedBox(height: 10),
            Text(widget.auth.profile?['phone']?.toString() ?? ''),
            Text('Duyệt: ${trips.approval}'),
            const Spacer(),
            OutlinedButton.icon(
                onPressed: widget.auth.logout,
                icon: const Icon(Icons.logout_rounded),
                label: const Text('Đăng xuất')),
          ])));
}
