import 'package:flutter/material.dart';

import '../../core/theme/flashx_theme.dart';
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
  final _destinationSearch = TextEditingController();
  var _index = 0;
  var _serviceType = 'bike';

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
    _destinationSearch.dispose();
    trips.removeListener(_refresh);
    trips.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: IndexedStack(
          index: _index, children: [_home(), _history(), _account()]),
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
              driverLocation: trips.driverLocation),
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
                  const _FlashXMiniBrand(),
                  const Spacer(),
                  _MapAction(
                    icon: Icons.my_location_rounded,
                    semanticLabel: 'Lấy lại vị trí hiện tại',
                    onTap: trips.busy ? null : trips.refreshPickupLocation,
                  ),
                  const SizedBox(width: 10),
                  _MapAction(
                    icon: Icons.person_rounded,
                    semanticLabel: 'Mở tài khoản',
                    onTap: () => setState(() => _index = 2),
                  ),
                ],
              ),
            ),
          ),
        ),
        if (active != null && trips.hasActiveTrip)
          Positioned(
            top: 82,
            left: 16,
            right: 16,
            child: SafeArea(
              bottom: false,
              child:
                  _TripStatusPill(status: active['status']?.toString() ?? ''),
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
                color: Colors.white,
                borderRadius: BorderRadius.circular(30),
                border: Border.all(color: FlashXTheme.border),
                boxShadow: const [
                  BoxShadow(
                      color: Color(0x260B132B),
                      blurRadius: 34,
                      offset: Offset(0, 16)),
                ],
              ),
              child: ConstrainedBox(
                constraints: BoxConstraints(
                  maxHeight: MediaQuery.sizeOf(context).height * .62,
                ),
                child: SingleChildScrollView(
                  physics: const BouncingScrollPhysics(),
                  child: active == null || !trips.hasActiveTrip
                      ? _bookingPanel()
                      : _activeTripPanel(active),
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _bookingPanel() {
    final theme = Theme.of(context);
    final query = _destinationSearch.text.trim().toLowerCase();
    final suggestions = RiderTripController.destinations.entries
        .where(
            (entry) => query.isEmpty || entry.key.toLowerCase().contains(query))
        .toList();
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Center(
          child: Container(
              width: 40,
              height: 4,
              decoration: BoxDecoration(
                  color: const Color(0xFFD7DAE0),
                  borderRadius: BorderRadius.circular(99))),
        ),
        const SizedBox(height: 16),
        Text('Bạn muốn đi đâu?', style: theme.textTheme.headlineSmall),
        const SizedBox(height: 6),
        Text('Chọn điểm đến để xem giá và thời gian dự kiến.',
            style: theme.textTheme.bodyMedium),
        const SizedBox(height: 16),
        Container(
          padding: const EdgeInsets.all(14),
          decoration: BoxDecoration(
              color: const Color(0xFFF7F8FA),
              borderRadius: BorderRadius.circular(20),
              border: Border.all(color: FlashXTheme.border)),
          child: const _LocationRow(
            icon: Icons.radio_button_checked_rounded,
            iconColor: FlashXTheme.lime,
            title: 'Vị trí hiện tại',
            subtitle: 'Điểm đón của bạn',
          ),
        ),
        const SizedBox(height: 10),
        TextField(
          controller: _destinationSearch,
          onChanged: (_) => setState(() {}),
          textInputAction: TextInputAction.search,
          decoration: InputDecoration(
            hintText: 'Tìm điểm đến...',
            prefixIcon: const Icon(Icons.search_rounded),
            suffixIcon: query.isEmpty
                ? null
                : IconButton(
                    tooltip: 'Xóa tìm kiếm',
                    onPressed: () {
                      _destinationSearch.clear();
                      setState(() {});
                    },
                    icon: const Icon(Icons.close_rounded),
                  ),
          ),
        ),
        const SizedBox(height: 16),
        Row(
          children: [
            Text('Dịch vụ', style: theme.textTheme.titleMedium),
            const Spacer(),
            Text('Giá hiển thị trước',
                style: theme.textTheme.bodySmall
                    ?.copyWith(color: FlashXTheme.textSecondary)),
          ],
        ),
        const SizedBox(height: 10),
        Row(
          children: [
            Expanded(
                child: _ServiceCard(
                    selected: _serviceType == 'bike',
                    icon: Icons.two_wheeler_rounded,
                    title: 'FlashX Bike',
                    subtitle: 'Nhanh & tiết kiệm',
                    onTap: () => setState(() => _serviceType = 'bike'))),
            const SizedBox(width: 10),
            Expanded(
                child: _ServiceCard(
                    selected: _serviceType == 'car',
                    icon: Icons.directions_car_filled_rounded,
                    title: 'FlashX Car',
                    subtitle: 'Thoải mái hơn',
                    onTap: () => setState(() => _serviceType = 'car'))),
          ],
        ),
        const SizedBox(height: 16),
        Row(
          children: [
            Text(query.isEmpty ? 'Điểm đến gần đây' : 'Kết quả tìm kiếm',
                style: theme.textTheme.titleMedium),
            const Spacer(),
            Text('${suggestions.length} địa điểm',
                style: theme.textTheme.bodySmall
                    ?.copyWith(color: FlashXTheme.textSecondary)),
          ],
        ),
        const SizedBox(height: 8),
        if (suggestions.isEmpty)
          const _EmptySearchResult()
        else
          ...suggestions.map((entry) => Padding(
                padding: const EdgeInsets.only(bottom: 8),
                child: _DestinationTile(
                  name: entry.key,
                  onTap: trips.busy
                      ? null
                      : () => _selectDestination(entry.key, entry.value),
                ),
              )),
        if (trips.error != null) ...[
          const SizedBox(height: 4),
          _ErrorBanner(message: trips.error!),
        ],
      ],
    );
  }

  Future<void> _selectDestination(
      String name, Map<String, double> destination) async {
    final ok = await trips.estimateTo(destination, serviceType: _serviceType);
    if (!mounted || !ok || trips.estimate == null) return;
    final fare =
        (trips.estimate!['fare'] as Map<String, dynamic>?)?['total_minor'] ?? 0;
    final distance = trips.estimate!['distance_m'] ?? 0;
    final duration = trips.estimate!['duration_s'] ?? 0;
    final accepted = await showModalBottomSheet<bool>(
      context: context,
      isScrollControlled: true,
      showDragHandle: false,
      builder: (context) => SafeArea(
        top: false,
        child: Padding(
          padding: const EdgeInsets.fromLTRB(20, 12, 20, 24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Center(
                  child: Container(
                      width: 42,
                      height: 4,
                      decoration: BoxDecoration(
                          color: const Color(0xFFD7DAE0),
                          borderRadius: BorderRadius.circular(99)))),
              const SizedBox(height: 20),
              Row(
                children: [
                  Container(
                      width: 48,
                      height: 48,
                      decoration: BoxDecoration(
                          color: FlashXTheme.yellow,
                          borderRadius: BorderRadius.circular(16)),
                      child: Icon(
                          _serviceType == 'car'
                              ? Icons.directions_car_filled_rounded
                              : Icons.two_wheeler_rounded,
                          color: FlashXTheme.navy)),
                  const SizedBox(width: 12),
                  Expanded(
                      child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                        Text(
                            _serviceType == 'car'
                                ? 'FlashX Car'
                                : 'FlashX Bike',
                            style: Theme.of(context).textTheme.titleLarge),
                        const SizedBox(height: 2),
                        Text(name,
                            style: Theme.of(context).textTheme.bodyMedium)
                      ])),
                  Text(_money(fare),
                      style: const TextStyle(
                          fontSize: 22,
                          fontWeight: FontWeight.w900,
                          color: FlashXTheme.navy)),
                ],
              ),
              const SizedBox(height: 18),
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                    color: const Color(0xFFF7F8FA),
                    borderRadius: BorderRadius.circular(18)),
                child: Row(
                  children: [
                    Expanded(
                        child: _EstimateStat(
                            label: 'Quãng đường',
                            value:
                                '${((distance as num).toDouble() / 1000).toStringAsFixed(1)} km')),
                    Container(width: 1, height: 34, color: FlashXTheme.border),
                    Expanded(
                        child: _EstimateStat(
                            label: 'Thời gian',
                            value:
                                '${((duration as num).toDouble() / 60).ceil()} phút')),
                    Container(width: 1, height: 34, color: FlashXTheme.border),
                    const Expanded(
                        child: _EstimateStat(
                            label: 'Thanh toán', value: 'Tiền mặt')),
                  ],
                ),
              ),
              const SizedBox(height: 18),
              FilledButton.icon(
                onPressed: () => Navigator.pop(context, true),
                icon: const Icon(Icons.bolt_rounded),
                label: const Text('Đặt chuyến ngay'),
              ),
            ],
          ),
        ),
      ),
    );
    if (accepted == true) {
      await trips.createTrip(destination, serviceType: _serviceType);
    }
  }

  Widget _activeTripPanel(Map<String, dynamic> trip) {
    final status = trip['status']?.toString() ?? '';
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
                    color: const Color(0xFFD7DAE0),
                    borderRadius: BorderRadius.circular(99)))),
        const SizedBox(height: 16),
        Row(
          children: [
            Container(
                width: 48,
                height: 48,
                decoration: BoxDecoration(
                    color: FlashXTheme.yellow,
                    borderRadius: BorderRadius.circular(16)),
                child: const Icon(Icons.bolt_rounded,
                    color: FlashXTheme.navy, size: 28)),
            const SizedBox(width: 12),
            Expanded(
                child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                  Text(_statusLabel(status),
                      style: theme.textTheme.headlineSmall),
                  const SizedBox(height: 3),
                  Text(_statusHint(status), style: theme.textTheme.bodyMedium)
                ])),
          ],
        ),
        const SizedBox(height: 18),
        _TripTimeline(status: status),
        const SizedBox(height: 16),
        if (trip['driver_id']?.toString().isNotEmpty == true)
          Container(
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(
                color: const Color(0xFFF7F8FA),
                borderRadius: BorderRadius.circular(18)),
            child: Row(
              children: [
                const CircleAvatar(
                    backgroundColor: FlashXTheme.navy,
                    foregroundColor: Colors.white,
                    child: Icon(Icons.person_rounded)),
                const SizedBox(width: 12),
                Expanded(
                    child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                      const Text('Tài xế FlashX',
                          style: TextStyle(fontWeight: FontWeight.w800)),
                      Text(
                          'Mã ${_shortId(trip['driver_id']?.toString() ?? '')}',
                          style: theme.textTheme.bodySmall)
                    ])),
                IconButton(
                    onPressed: () {},
                    tooltip: 'Gọi tài xế',
                    icon: const Icon(Icons.call_rounded)),
                IconButton(
                    onPressed: () {},
                    tooltip: 'Nhắn tin',
                    icon: const Icon(Icons.chat_bubble_outline_rounded)),
              ],
            ),
          ),
        if (trips.driverLocation != null) ...[
          const SizedBox(height: 10),
          const Row(children: [
            Icon(Icons.radar_rounded, size: 18, color: FlashXTheme.success),
            SizedBox(width: 8),
            Text('Vị trí tài xế đang cập nhật trực tiếp',
                style: TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.w700,
                    color: FlashXTheme.success))
          ]),
        ],
        const SizedBox(height: 14),
        if (!['in_progress', 'completed', 'cancelled'].contains(status))
          OutlinedButton.icon(
              onPressed: trips.busy ? null : trips.cancel,
              icon: const Icon(Icons.close_rounded),
              label: const Text('Hủy chuyến')),
        if (status == 'completed')
          FilledButton(
              onPressed: trips.loadHistory,
              child: const Text('Xem lịch sử chuyến')),
      ],
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
            const _PageBrand(),
            const SizedBox(height: 24),
            Text('Chuyến đi của bạn', style: theme.textTheme.headlineMedium),
            const SizedBox(height: 6),
            Text('Theo dõi các chuyến gần đây và đánh giá tài xế.',
                style: theme.textTheme.bodyMedium),
            const SizedBox(height: 20),
            if (trips.history.isEmpty)
              const _EmptyCard(
                  icon: Icons.route_outlined,
                  title: 'Chưa có chuyến đi',
                  subtitle: 'Chuyến đầu tiên của bạn sẽ xuất hiện ở đây.'),
            ...trips.history.map((trip) => Padding(
                  padding: const EdgeInsets.only(bottom: 10),
                  child: _TripHistoryCard(
                    trip: trip,
                    statusLabel: _statusLabel(trip['status']?.toString() ?? ''),
                    money: _money(trip['final_fare_minor'] ??
                        trip['estimated_fare_minor'] ??
                        0),
                    canRate: trip['status'] == 'completed' &&
                        !trips.ratedTripIds
                            .contains(trip['id']?.toString() ?? ''),
                    onRate: () => _rateTrip(trip),
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
          shape:
              RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
          title: const Text('Chuyến đi thế nào?'),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text('Đánh giá trải nghiệm với tài xế FlashX.',
                  style: Theme.of(context).textTheme.bodyMedium),
              const SizedBox(height: 12),
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: List.generate(
                    5,
                    (index) => IconButton(
                        onPressed: () =>
                            setDialogState(() => stars = index + 1),
                        icon: Icon(
                            index < stars
                                ? Icons.star_rounded
                                : Icons.star_border_rounded,
                            color: FlashXTheme.warning,
                            size: 30))),
              ),
              const SizedBox(height: 8),
              TextField(
                  controller: comment,
                  maxLength: 500,
                  maxLines: 3,
                  decoration: const InputDecoration(
                      labelText: 'Nhận xét (không bắt buộc)',
                      hintText: 'Tài xế thân thiện, chuyến đi an toàn...')),
            ],
          ),
          actions: [
            TextButton(
                onPressed: () => Navigator.pop(dialogContext, false),
                child: const Text('Để sau')),
            FilledButton(
                onPressed: () => Navigator.pop(dialogContext, true),
                child: const Text('Gửi đánh giá')),
          ],
        ),
      ),
    );
    if (submit == true) {
      final ok =
          await trips.rateTrip(trip['id'].toString(), stars, comment.text);
      if (mounted && !ok && trips.error != null) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text(trips.error!)));
      }
    }
    comment.dispose();
  }

  Widget _account() {
    final theme = Theme.of(context);
    final phone = widget.auth.phone ?? '';
    return SafeArea(
      child: ListView(
        padding: const EdgeInsets.fromLTRB(20, 24, 20, 32),
        children: [
          const _PageBrand(),
          const SizedBox(height: 24),
          Text('Tài khoản', style: theme.textTheme.headlineMedium),
          const SizedBox(height: 18),
          Container(
            padding: const EdgeInsets.all(18),
            decoration: BoxDecoration(
                color: FlashXTheme.navy,
                borderRadius: BorderRadius.circular(24)),
            child: Row(
              children: [
                Container(
                    width: 56,
                    height: 56,
                    decoration: BoxDecoration(
                        color: FlashXTheme.yellow,
                        borderRadius: BorderRadius.circular(18)),
                    child: const Icon(Icons.person_rounded,
                        color: FlashXTheme.navy, size: 30)),
                const SizedBox(width: 14),
                Expanded(
                    child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                      const Text('Rider FlashX',
                          style: TextStyle(
                              color: Colors.white,
                              fontSize: 18,
                              fontWeight: FontWeight.w800)),
                      const SizedBox(height: 4),
                      Text(phone,
                          style: const TextStyle(color: Color(0xFFB9C1D5)))
                    ])),
              ],
            ),
          ),
          const SizedBox(height: 16),
          const _AccountItem(
              icon: Icons.credit_card_rounded,
              title: 'Thanh toán',
              subtitle: 'Tiền mặt · phương thức mặc định'),
          const SizedBox(height: 10),
          const _AccountItem(
              icon: Icons.shield_outlined,
              title: 'An toàn & quyền riêng tư',
              subtitle: 'Quản lý dữ liệu và quyền truy cập'),
          const SizedBox(height: 10),
          const _AccountItem(
              icon: Icons.help_outline_rounded,
              title: 'Trợ giúp',
              subtitle: 'Hỗ trợ chuyến đi và tài khoản'),
          const SizedBox(height: 24),
          OutlinedButton.icon(
              onPressed: widget.auth.logout,
              icon: const Icon(Icons.logout_rounded),
              label: const Text('Đăng xuất')),
        ],
      ),
    );
  }

  static String _money(dynamic value) {
    final amount =
        value is num ? value.toInt() : int.tryParse(value.toString()) ?? 0;
    return '${amount.toString().replaceAllMapped(RegExp(r'(?=(\d{3})+(?!\d))'), (m) => '.')}₫';
  }

  static String _shortId(String value) => value.length <= 8
      ? value
      : value.substring(value.length - 8).toUpperCase();

  static String _statusLabel(String status) => switch (status) {
        'searching' => 'Đang tìm tài xế gần bạn',
        'accepted' => 'Tài xế đã nhận chuyến',
        'arriving' => 'Tài xế đang đến đón bạn',
        'arrived' => 'Tài xế đã đến điểm đón',
        'in_progress' => 'Bạn đang trên chuyến đi',
        'completed' => 'Chuyến đi đã hoàn thành',
        'cancelled' => 'Chuyến đi đã hủy',
        _ => 'Chuyến đi',
      };

  static String _statusHint(String status) => switch (status) {
        'searching' => 'FlashX đang kết nối với tài xế phù hợp.',
        'accepted' => 'Thông tin tài xế đã được xác nhận.',
        'arriving' => 'Theo dõi vị trí tài xế trực tiếp trên bản đồ.',
        'arrived' => 'Hãy kiểm tra biển số trước khi lên xe.',
        'in_progress' => 'Chúc bạn có một chuyến đi an toàn.',
        'completed' => 'Cảm ơn bạn đã sử dụng FlashX.',
        'cancelled' => 'Bạn có thể đặt một chuyến mới bất cứ lúc nào.',
        _ => 'Trạng thái chuyến đang được cập nhật.',
      };
}

class _FlashXMiniBrand extends StatelessWidget {
  const _FlashXMiniBrand();
  @override
  Widget build(BuildContext context) => Container(
        padding: const EdgeInsets.fromLTRB(8, 7, 12, 7),
        decoration: BoxDecoration(
            color: FlashXTheme.navy,
            borderRadius: BorderRadius.circular(18),
            boxShadow: const [
              BoxShadow(
                  color: Color(0x22000000),
                  blurRadius: 16,
                  offset: Offset(0, 6))
            ]),
        child: const Row(mainAxisSize: MainAxisSize.min, children: [
          Icon(Icons.bolt_rounded, color: FlashXTheme.yellow, size: 25),
          SizedBox(width: 4),
          Text('FlashX',
              style: TextStyle(
                  color: Colors.white,
                  fontSize: 16,
                  fontWeight: FontWeight.w900,
                  letterSpacing: -.5))
        ]),
      );
}

class _MapAction extends StatelessWidget {
  const _MapAction(
      {required this.icon, required this.semanticLabel, required this.onTap});
  final IconData icon;
  final String semanticLabel;
  final VoidCallback? onTap;
  @override
  Widget build(BuildContext context) => Material(
        color: Colors.white,
        shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(17),
            side: const BorderSide(color: FlashXTheme.border)),
        elevation: 3,
        shadowColor: const Color(0x22000000),
        child: IconButton(
            onPressed: onTap,
            tooltip: semanticLabel,
            icon: Icon(icon, color: FlashXTheme.navy)),
      );
}

class _TripStatusPill extends StatelessWidget {
  const _TripStatusPill({required this.status});
  final String status;
  @override
  Widget build(BuildContext context) => Center(
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
          decoration: BoxDecoration(
              color: FlashXTheme.navy,
              borderRadius: BorderRadius.circular(999),
              boxShadow: const [
                BoxShadow(
                    color: Color(0x260B132B),
                    blurRadius: 20,
                    offset: Offset(0, 8))
              ]),
          child: Row(mainAxisSize: MainAxisSize.min, children: [
            const SizedBox(
                width: 8,
                height: 8,
                child: CircularProgressIndicator(
                    strokeWidth: 2, color: FlashXTheme.yellow)),
            const SizedBox(width: 9),
            Text(_RiderHomeScreenState._statusLabel(status),
                style: const TextStyle(
                    color: Colors.white,
                    fontSize: 12,
                    fontWeight: FontWeight.w800))
          ]),
        ),
      );
}

class _LocationRow extends StatelessWidget {
  const _LocationRow(
      {required this.icon,
      required this.iconColor,
      required this.title,
      required this.subtitle});
  final IconData icon;
  final Color iconColor;
  final String title;
  final String subtitle;
  @override
  Widget build(BuildContext context) => Row(children: [
        Icon(icon, color: iconColor, size: 20),
        const SizedBox(width: 12),
        Expanded(
            child:
                Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text(title,
              style:
                  const TextStyle(fontSize: 14, fontWeight: FontWeight.w800)),
          const SizedBox(height: 2),
          Text(subtitle,
              style: Theme.of(context)
                  .textTheme
                  .bodySmall
                  ?.copyWith(color: FlashXTheme.textSecondary))
        ]))
      ]);
}

class _ServiceCard extends StatelessWidget {
  const _ServiceCard(
      {required this.selected,
      required this.icon,
      required this.title,
      required this.subtitle,
      required this.onTap});
  final bool selected;
  final IconData icon;
  final String title;
  final String subtitle;
  final VoidCallback onTap;
  @override
  Widget build(BuildContext context) => Material(
        color: selected ? const Color(0xFFFFF9D6) : Colors.white,
        shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(18),
            side: BorderSide(
                color: selected ? FlashXTheme.yellow : FlashXTheme.border,
                width: selected ? 1.5 : 1)),
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(18),
          child: Padding(
            padding: const EdgeInsets.all(13),
            child: Row(children: [
              Container(
                  width: 38,
                  height: 38,
                  decoration: BoxDecoration(
                      color: selected
                          ? FlashXTheme.yellow
                          : const Color(0xFFF2F4F7),
                      borderRadius: BorderRadius.circular(13)),
                  child: Icon(icon, color: FlashXTheme.navy, size: 22)),
              const SizedBox(width: 10),
              Expanded(
                  child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                    Text(title,
                        style: const TextStyle(
                            fontSize: 13, fontWeight: FontWeight.w800)),
                    const SizedBox(height: 2),
                    Text(subtitle,
                        style: const TextStyle(
                            fontSize: 10.5, color: FlashXTheme.textSecondary))
                  ]))
            ]),
          ),
        ),
      );
}

class _EmptySearchResult extends StatelessWidget {
  const _EmptySearchResult();

  @override
  Widget build(BuildContext context) => Container(
        padding: const EdgeInsets.all(18),
        decoration: BoxDecoration(
          color: const Color(0xFFF7F8FA),
          borderRadius: BorderRadius.circular(18),
          border: Border.all(color: FlashXTheme.border),
        ),
        child: Column(
          children: [
            const Icon(Icons.search_off_rounded,
                color: FlashXTheme.textSecondary, size: 28),
            const SizedBox(height: 8),
            const Text('Chưa thấy địa điểm phù hợp',
                style: TextStyle(fontWeight: FontWeight.w800)),
            const SizedBox(height: 4),
            Text(
              'Thử một từ khóa khác. Khi Places API được cấu hình, FlashX sẽ tìm địa điểm trực tiếp trên bản đồ.',
              textAlign: TextAlign.center,
              style: Theme.of(context)
                  .textTheme
                  .bodySmall
                  ?.copyWith(color: FlashXTheme.textSecondary),
            ),
          ],
        ),
      );
}

class _DestinationTile extends StatelessWidget {
  const _DestinationTile({required this.name, required this.onTap});
  final String name;
  final VoidCallback? onTap;
  @override
  Widget build(BuildContext context) => Material(
        color: Colors.white,
        shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(17),
            side: const BorderSide(color: FlashXTheme.border)),
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(17),
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
            child: Row(children: [
              Container(
                  width: 36,
                  height: 36,
                  decoration: BoxDecoration(
                      color: const Color(0xFFF2F4F7),
                      borderRadius: BorderRadius.circular(12)),
                  child: const Icon(Icons.history_rounded,
                      color: FlashXTheme.navy, size: 19)),
              const SizedBox(width: 11),
              Expanded(
                  child: Text(name,
                      style: const TextStyle(fontWeight: FontWeight.w700))),
              const Icon(Icons.chevron_right_rounded,
                  color: FlashXTheme.textSecondary)
            ]),
          ),
        ),
      );
}

class _EstimateStat extends StatelessWidget {
  const _EstimateStat({required this.label, required this.value});
  final String label;
  final String value;
  @override
  Widget build(BuildContext context) => Column(children: [
        Text(label,
            style: const TextStyle(
                color: FlashXTheme.textSecondary,
                fontSize: 10.5,
                fontWeight: FontWeight.w600)),
        const SizedBox(height: 5),
        Text(value,
            textAlign: TextAlign.center,
            style: const TextStyle(
                color: FlashXTheme.navy,
                fontSize: 13,
                fontWeight: FontWeight.w800))
      ]);
}

class _TripTimeline extends StatelessWidget {
  const _TripTimeline({required this.status});
  final String status;
  @override
  Widget build(BuildContext context) {
    const steps = [
      'searching',
      'accepted',
      'arrived',
      'in_progress',
      'completed'
    ];
    final normalized = status == 'arriving' ? 'accepted' : status;
    final index = steps.indexOf(normalized);
    return Row(
      children: List.generate(
          steps.length,
          (i) => Expanded(
                  child: Row(children: [
                Expanded(
                    child: AnimatedContainer(
                        duration: const Duration(milliseconds: 250),
                        height: 5,
                        decoration: BoxDecoration(
                            color: i <= index
                                ? FlashXTheme.yellow
                                : const Color(0xFFE9EBF0),
                            borderRadius: BorderRadius.circular(99)))),
                if (i < steps.length - 1) const SizedBox(width: 5)
              ]))),
    );
  }
}

class _ErrorBanner extends StatelessWidget {
  const _ErrorBanner({required this.message});
  final String message;
  @override
  Widget build(BuildContext context) => Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
          color: const Color(0xFFFFEEEE),
          borderRadius: BorderRadius.circular(14)),
      child: Row(children: [
        const Icon(Icons.error_outline_rounded,
            color: FlashXTheme.danger, size: 18),
        const SizedBox(width: 8),
        Expanded(
            child: Text(message,
                style: const TextStyle(
                    color: FlashXTheme.danger,
                    fontSize: 12.5,
                    fontWeight: FontWeight.w600)))
      ]));
}

class _PageBrand extends StatelessWidget {
  const _PageBrand();
  @override
  Widget build(BuildContext context) => const Row(children: [
        Icon(Icons.bolt_rounded, color: FlashXTheme.navy, size: 28),
        SizedBox(width: 4),
        Text('FlashX',
            style: TextStyle(
                color: FlashXTheme.navy,
                fontSize: 20,
                fontWeight: FontWeight.w900,
                letterSpacing: -.7))
      ]);
}

class _EmptyCard extends StatelessWidget {
  const _EmptyCard(
      {required this.icon, required this.title, required this.subtitle});
  final IconData icon;
  final String title;
  final String subtitle;
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
            child: Icon(icon, color: FlashXTheme.navy)),
        const SizedBox(height: 12),
        Text(title, style: const TextStyle(fontWeight: FontWeight.w800)),
        const SizedBox(height: 4),
        Text(subtitle,
            textAlign: TextAlign.center,
            style: Theme.of(context).textTheme.bodyMedium)
      ]));
}

class _TripHistoryCard extends StatelessWidget {
  const _TripHistoryCard(
      {required this.trip,
      required this.statusLabel,
      required this.money,
      required this.canRate,
      required this.onRate});
  final Map<String, dynamic> trip;
  final String statusLabel;
  final String money;
  final bool canRate;
  final VoidCallback onRate;
  @override
  Widget build(BuildContext context) => Container(
        padding: const EdgeInsets.all(16),
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
              child: const Icon(Icons.route_rounded, color: FlashXTheme.navy)),
          const SizedBox(width: 12),
          Expanded(
              child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                Text(statusLabel,
                    style: const TextStyle(fontWeight: FontWeight.w800)),
                const SizedBox(height: 4),
                Text(
                    _RiderHomeScreenState._shortId(
                        trip['id']?.toString() ?? ''),
                    style: Theme.of(context)
                        .textTheme
                        .bodySmall
                        ?.copyWith(color: FlashXTheme.textSecondary)),
                if (canRate)
                  Padding(
                      padding: const EdgeInsets.only(top: 7),
                      child: GestureDetector(
                          onTap: onRate,
                          child: const Text('★ Đánh giá tài xế',
                              style: TextStyle(
                                  color: FlashXTheme.warning,
                                  fontSize: 12,
                                  fontWeight: FontWeight.w800))))
              ])),
          Text(money,
              style: const TextStyle(
                  color: FlashXTheme.navy,
                  fontSize: 15,
                  fontWeight: FontWeight.w900))
        ]),
      );
}

class _AccountItem extends StatelessWidget {
  const _AccountItem(
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
