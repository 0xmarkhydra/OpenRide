import 'package:flutter/material.dart';

import '../auth/auth_controller.dart';
import '../trips/rider_trip_controller.dart';
import 'rider_map_view.dart';

class RiderHomeScreen extends StatefulWidget {
  const RiderHomeScreen({super.key, required this.auth});
  final AuthController auth;

  @override
  State<RiderHomeScreen> createState() => _RiderHomeScreenState();
}

class _RiderHomeScreenState extends State<RiderHomeScreen> {
  late final RiderTripController trips;
  var _index = 0;

  @override
  void initState() {
    super.initState();
    trips = RiderTripController(
        api: widget.auth.api, realtime: widget.auth.realtime)
      ..addListener(_refresh);
    trips.loadHistory();
    trips.refreshPickupLocation();
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
        index: _index,
        children: [
          _home(),
          _history(),
          _account(),
        ],
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _index,
        onDestinationSelected: (value) => setState(() => _index = value),
        destinations: const [
          NavigationDestination(
              icon: Icon(Icons.home_outlined),
              selectedIcon: Icon(Icons.home_rounded),
              label: 'Trang chủ'),
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
    final active = trips.activeTrip;
    return Stack(
      children: [
        Positioned.fill(
          child: RiderMapView(
            pickup: trips.pickup,
            activeTrip: active,
            driverLocation: trips.driverLocation,
          ),
        ),
        SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              children: [
                CircleAvatar(
                    child: IconButton(
                        onPressed: () {},
                        icon: const Icon(Icons.menu_rounded))),
                const SizedBox(width: 12),
                Expanded(
                  child: Card(
                    child: Padding(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 14, vertical: 12),
                      child: Text(active == null
                          ? 'Sẵn sàng cho chuyến mới'
                          : _statusLabel(active['status']?.toString() ?? '')),
                    ),
                  ),
                ),
              ],
            ),
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
              child: active == null || !trips.hasActiveTrip
                  ? _bookingPanel()
                  : _activeTripPanel(active),
            ),
          ),
        ),
      ],
    );
  }

  Widget _bookingPanel() {
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const Text('Bạn muốn đi đâu?',
            style: TextStyle(fontSize: 23, fontWeight: FontWeight.w900)),
        const SizedBox(height: 6),
        const Text(
            'MVP đang dùng các điểm mẫu cho tới khi tài khoản Maps của khách được cấu hình.'),
        const SizedBox(height: 14),
        ...RiderTripController.destinations.entries.map((entry) => Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: OutlinedButton.icon(
                onPressed: trips.busy
                    ? null
                    : () => _selectDestination(entry.key, entry.value),
                icon: const Icon(Icons.location_on_outlined),
                label: Align(
                    alignment: Alignment.centerLeft, child: Text(entry.key)),
              ),
            )),
        if (trips.error != null)
          Text(trips.error!,
              style: TextStyle(color: Theme.of(context).colorScheme.error)),
      ],
    );
  }

  Future<void> _selectDestination(
      String name, Map<String, double> destination) async {
    final ok = await trips.estimateTo(destination);
    if (!mounted || !ok || trips.estimate == null) return;
    final fare =
        (trips.estimate!['fare'] as Map<String, dynamic>?)?['total_minor'] ?? 0;
    final distance = trips.estimate!['distance_m'] ?? 0;
    final accepted = await showModalBottomSheet<bool>(
      context: context,
      showDragHandle: true,
      builder: (context) => Padding(
        padding: const EdgeInsets.fromLTRB(20, 4, 20, 28),
        child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(name,
                  style: const TextStyle(
                      fontSize: 22, fontWeight: FontWeight.w900)),
              const SizedBox(height: 12),
              Text(
                  '${((distance as num).toDouble() / 1000).toStringAsFixed(1)} km · ${_money(fare)}'),
              const SizedBox(height: 18),
              FilledButton(
                  onPressed: () => Navigator.pop(context, true),
                  child: const Text('Đặt xe máy')),
            ]),
      ),
    );
    if (accepted == true) await trips.createTrip(destination);
  }

  Widget _activeTripPanel(Map<String, dynamic> trip) {
    final status = trip['status']?.toString() ?? '';
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text(_statusLabel(status),
            style: const TextStyle(fontSize: 22, fontWeight: FontWeight.w900)),
        const SizedBox(height: 8),
        Text('Mã chuyến: ${trip['id']}'),
        if (trip['driver_id']?.toString().isNotEmpty == true)
          Text('Tài xế: ${trip['driver_id']}'),
        if (trips.driverLocation != null)
          const Text('Vị trí tài xế đang được cập nhật realtime.'),
        const SizedBox(height: 14),
        LinearProgressIndicator(value: _progress(status)),
        const SizedBox(height: 14),
        if (!['in_progress', 'completed', 'cancelled'].contains(status))
          OutlinedButton(
              onPressed: trips.busy ? null : trips.cancel,
              child: const Text('Hủy chuyến')),
        if (status == 'completed')
          FilledButton(
              onPressed: trips.loadHistory, child: const Text('Hoàn tất')),
      ],
    );
  }

  Widget _history() {
    return SafeArea(
      child: RefreshIndicator(
        onRefresh: trips.loadHistory,
        child: ListView(
          padding: const EdgeInsets.all(20),
          children: [
            const Text('Chuyến đi',
                style: TextStyle(fontSize: 28, fontWeight: FontWeight.w900)),
            const SizedBox(height: 16),
            if (trips.history.isEmpty)
              const Card(
                  child: Padding(
                      padding: EdgeInsets.all(20),
                      child: Text('Chưa có chuyến đi.'))),
            ...trips.history.map((trip) => Card(
                  child: ListTile(
                    leading:
                        const CircleAvatar(child: Icon(Icons.route_rounded)),
                    title: Text(_statusLabel(trip['status']?.toString() ?? '')),
                    subtitle: Text(trip['id']?.toString() ?? ''),
                    trailing: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Text(_money(trip['final_fare_minor'] ??
                            trip['estimated_fare_minor'] ??
                            0)),
                        if (trip['status'] == 'completed' &&
                            !trips.ratedTripIds
                                .contains(trip['id']?.toString() ?? ''))
                          InkWell(
                            onTap: () => _rateTrip(trip),
                            child: const Text('Đánh giá',
                                style: TextStyle(
                                    fontSize: 12,
                                    fontWeight: FontWeight.w800)),
                          ),
                      ],
                    ),
                  ),
                )),
          ],
        ),
      ),
    );
  }

  Future<void> _rateTrip(Map<String, dynamic> trip) async {
    var stars = 5;
    final comment = TextEditingController();
    final submit = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => StatefulBuilder(
        builder: (context, setDialogState) => AlertDialog(
          title: const Text('Đánh giá tài xế'),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: List.generate(
                  5,
                  (index) => IconButton(
                    onPressed: () => setDialogState(() => stars = index + 1),
                    icon: Icon(index < stars ? Icons.star_rounded : Icons.star_border_rounded),
                  ),
                ),
              ),
              TextField(
                controller: comment,
                maxLength: 500,
                decoration: const InputDecoration(
                  labelText: 'Nhận xét (không bắt buộc)',
                  border: OutlineInputBorder(),
                ),
              ),
            ],
          ),
          actions: [
            TextButton(onPressed: () => Navigator.pop(dialogContext, false), child: const Text('Để sau')),
            FilledButton(onPressed: () => Navigator.pop(dialogContext, true), child: const Text('Gửi đánh giá')),
          ],
        ),
      ),
    );
    if (submit == true) {
      final ok = await trips.rateTrip(trip['id'].toString(), stars, comment.text);
      if (mounted && !ok && trips.error != null) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(trips.error!)));
      }
    }
    comment.dispose();
  }

  Widget _account() {
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child:
            Column(crossAxisAlignment: CrossAxisAlignment.stretch, children: [
          const Icon(Icons.person_rounded, size: 54),
          const SizedBox(height: 16),
          const Text('Tài khoản',
              style: TextStyle(fontSize: 28, fontWeight: FontWeight.w900)),
          const SizedBox(height: 12),
          Text(widget.auth.phone ?? ''),
          Text(widget.auth.actorId ?? '',
              style: TextStyle(
                  color: Theme.of(context).colorScheme.onSurfaceVariant)),
          const Spacer(),
          OutlinedButton.icon(
              onPressed: widget.auth.logout,
              icon: const Icon(Icons.logout_rounded),
              label: const Text('Đăng xuất')),
        ]),
      ),
    );
  }

  static String _money(dynamic value) {
    final amount =
        value is num ? value.toInt() : int.tryParse(value.toString()) ?? 0;
    return '${amount.toString().replaceAllMapped(RegExp(r'(?=(\d{3})+(?!\d))'), (m) => '.')}₫';
  }

  static String _statusLabel(String status) => switch (status) {
        'searching' => 'Đang tìm tài xế',
        'accepted' => 'Tài xế đã nhận chuyến',
        'arriving' => 'Tài xế đang đến đón',
        'arrived' => 'Tài xế đã đến',
        'in_progress' => 'Đang di chuyển',
        'completed' => 'Chuyến đã hoàn thành',
        'cancelled' => 'Chuyến đã hủy',
        _ => 'Chuyến đi',
      };

  static double _progress(String status) => switch (status) {
        'searching' => .12,
        'accepted' => .3,
        'arriving' => .45,
        'arrived' => .6,
        'in_progress' => .8,
        'completed' => 1,
        _ => .05,
      };
}
