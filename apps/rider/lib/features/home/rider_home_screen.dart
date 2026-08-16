import 'dart:async';

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
  var _index = 0;
  var _serviceType = 'designated_driver_car';
  var _bookingMode = 'immediate';
  DateTime? _scheduledAt;
  String _destinationName = 'Vincom Plaza Thanh Hóa';
  final TextEditingController _destinationSearch = TextEditingController();
  Timer? _placeSearchDebounce;
  bool _customDestinationSelected = false;
  bool _pinDestinationMode = false;

  @override
  void initState() {
    super.initState();
    trips = RiderTripController(
        api: widget.auth.api, realtime: widget.auth.realtime)
      ..addListener(_refresh);
    trips.selectDestination(RiderTripController.destinations[_destinationName]);
    trips.initialize();
  }

  void _refresh() {
    if (mounted) setState(() {});
  }

  @override
  void dispose() {
    _placeSearchDebounce?.cancel();
    _destinationSearch.dispose();
    trips.removeListener(_refresh);
    trips.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: IndexedStack(
          index: _index, children: [_home(), _activity(), _account()]),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _index,
        onDestinationSelected: (value) => setState(() => _index = value),
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.home_outlined),
            selectedIcon: Icon(Icons.home_rounded),
            label: 'Trang chủ',
          ),
          NavigationDestination(
            icon: Icon(Icons.receipt_long_outlined),
            selectedIcon: Icon(Icons.receipt_long_rounded),
            label: 'Hoạt động',
          ),
          NavigationDestination(
            icon: Icon(Icons.person_outline_rounded),
            selectedIcon: Icon(Icons.person_rounded),
            label: 'Tài khoản',
          ),
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
            selectedDestination: trips.selectedDestination,
            estimatePolyline: trips.estimate?['route_polyline']?.toString(),
            liveRoutes: trips.liveRoutes,
            onDestinationSelected: _pinDestinationMode
                ? (point) {
                    trips.selectDestination(point);
                    _destinationSearch.clear();
                    trips.clearPlaceSearch();
                    setState(() {
                      _destinationName = 'Điểm đã ghim';
                      _customDestinationSelected = true;
                      _pinDestinationMode = false;
                    });
                    ScaffoldMessenger.of(context)
                      ..hideCurrentSnackBar()
                      ..showSnackBar(const SnackBar(
                          content: Text('Đã chọn điểm đến trên bản đồ.')));
                  }
                : null,
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
                  const _BrandPill(),
                  const Spacer(),
                  _MapButton(
                    icon: Icons.my_location_rounded,
                    tooltip: 'Vị trí hiện tại',
                    onPressed: trips.refreshPickupLocation,
                  ),
                ],
              ),
            ),
          ),
        ),
        if (trips.hasActiveTrip)
          Positioned(
            top: 72,
            left: 16,
            right: 16,
            child: SafeArea(
              bottom: false,
              child: _ActivePill(
                label: _statusLabel(trips.status),
                service:
                    _serviceLabel(active?['service_type']?.toString() ?? ''),
              ),
            ),
          ),
        Align(
          alignment: Alignment.bottomCenter,
          child: SafeArea(
            top: false,
            minimum: const EdgeInsets.fromLTRB(12, 12, 12, 10),
            child: Container(
              constraints: BoxConstraints(
                maxWidth: 650,
                maxHeight: MediaQuery.sizeOf(context).height * .68,
              ),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(30),
                border: Border.all(color: FlashXTheme.border),
                boxShadow: const [
                  BoxShadow(
                    color: Color(0x260B132B),
                    blurRadius: 34,
                    offset: Offset(0, 16),
                  ),
                ],
              ),
              child: SingleChildScrollView(
                padding: const EdgeInsets.fromLTRB(20, 10, 20, 20),
                child: trips.hasActiveTrip
                    ? _activePanel(trips.activeTrip!)
                    : _bookingPanel(),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _bookingPanel() {
    final theme = Theme.of(context);
    final compatible = trips.vehiclesForService(_serviceType);
    final selectedID = compatible.any((v) => v['id'] == trips.selectedVehicleID)
        ? trips.selectedVehicleID
        : null;
    if (selectedID == null && compatible.isNotEmpty && !trips.busy) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted && trips.selectedVehicleID != compatible.first['id']) {
          trips.selectVehicle(compatible.first['id']?.toString());
        }
      });
    }

    final destinationEntries = RiderTripController.destinations.entries
        .where((entry) => _serviceType == 'vehicle_inspection_assist'
            ? entry.key.contains('đăng kiểm')
            : !entry.key.contains('đăng kiểm'))
        .toList();
    if (!_customDestinationSelected &&
        !destinationEntries.any((e) => e.key == _destinationName) &&
        destinationEntries.isNotEmpty) {
      _destinationName = destinationEntries.first.key;
      trips.selectDestination(destinationEntries.first.value);
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Center(child: _handle()),
        const SizedBox(height: 14),
        Text('Bạn cần tài xế làm gì?', style: theme.textTheme.headlineSmall),
        const SizedBox(height: 5),
        Text('Chọn dịch vụ, xe của bạn và thời gian thực hiện.',
            style: theme.textTheme.bodyMedium),
        const SizedBox(height: 16),
        Row(
          children: [
            Expanded(
              child: _ServiceCard(
                selected: _serviceType == 'designated_driver_car',
                icon: Icons.directions_car_filled_rounded,
                title: 'Lái hộ ô tô',
                onTap: () => _selectService('designated_driver_car'),
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: _ServiceCard(
                selected: _serviceType == 'designated_driver_bike',
                icon: Icons.two_wheeler_rounded,
                title: 'Lái hộ xe máy',
                onTap: () => _selectService('designated_driver_bike'),
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: _ServiceCard(
                selected: _serviceType == 'vehicle_inspection_assist',
                icon: Icons.fact_check_rounded,
                title: 'Đăng kiểm hộ',
                onTap: () => _selectService('vehicle_inspection_assist'),
              ),
            ),
          ],
        ),
        const SizedBox(height: 16),
        Row(
          children: [
            Expanded(
                child: Text('Xe của tôi', style: theme.textTheme.titleMedium)),
            TextButton.icon(
              onPressed: trips.busy ? null : _showAddVehicle,
              icon: const Icon(Icons.add_rounded),
              label: const Text('Thêm xe'),
            ),
          ],
        ),
        if (compatible.isEmpty)
          _InfoBox(
            icon: Icons.directions_car_outlined,
            text: _serviceType == 'designated_driver_bike'
                ? 'Bạn chưa lưu xe máy. Thêm xe để tiếp tục đặt dịch vụ.'
                : 'Bạn chưa lưu ô tô. Thêm xe để tiếp tục đặt dịch vụ.',
          )
        else
          DropdownButtonFormField<String>(
            initialValue: selectedID,
            decoration:
                const InputDecoration(prefixIcon: Icon(Icons.key_rounded)),
            hint: const Text('Chọn phương tiện'),
            items: compatible
                .map((vehicle) => DropdownMenuItem<String>(
                      value: vehicle['id']?.toString(),
                      child: Text(_vehicleLabel(vehicle)),
                    ))
                .toList(),
            onChanged: trips.busy ? null : trips.selectVehicle,
          ),
        const SizedBox(height: 14),
        TextField(
          controller: _destinationSearch,
          enabled: !trips.busy,
          textInputAction: TextInputAction.search,
          onChanged: _queuePlaceSearch,
          decoration: InputDecoration(
            labelText: _serviceType == 'vehicle_inspection_assist'
                ? 'Tìm nơi đăng kiểm / địa chỉ'
                : 'Tìm điểm đến',
            hintText: 'Ví dụ: Vincom Thanh Hóa',
            prefixIcon: const Icon(Icons.search_rounded),
            suffixIcon: trips.placesLoading
                ? const Padding(
                    padding: EdgeInsets.all(14),
                    child: SizedBox.square(
                      dimension: 18,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    ),
                  )
                : _destinationSearch.text.isEmpty
                    ? null
                    : IconButton(
                        onPressed: () {
                          _placeSearchDebounce?.cancel();
                          _destinationSearch.clear();
                          trips.clearPlaceSearch();
                          setState(() {});
                        },
                        icon: const Icon(Icons.close_rounded),
                      ),
          ),
        ),
        if (trips.placeResults.isNotEmpty) ...[
          const SizedBox(height: 8),
          Container(
            decoration: BoxDecoration(
              color: FlashXTheme.surface,
              borderRadius: BorderRadius.circular(18),
              border: Border.all(color: FlashXTheme.border),
            ),
            child: Column(
              children: trips.placeResults
                  .map((place) => ListTile(
                        leading: const Icon(Icons.location_on_outlined),
                        title: Text(place['name']?.toString().isNotEmpty == true
                            ? place['name'].toString()
                            : place['address']?.toString() ?? 'Địa điểm'),
                        subtitle:
                            place['address']?.toString().isNotEmpty == true
                                ? Text(place['address'].toString(),
                                    maxLines: 2,
                                    overflow: TextOverflow.ellipsis)
                                : null,
                        onTap: () => _selectPlaceResult(place),
                      ))
                  .toList(),
            ),
          ),
        ] else if (trips.placeSearchUnavailable &&
            _destinationSearch.text.trim().runes.length >= 3) ...[
          const SizedBox(height: 8),
          const _InfoBox(
            icon: Icons.map_outlined,
            text:
                'Chưa tìm được địa chỉ tự động. Bạn vẫn có thể ghim chính xác điểm cần đến trên bản đồ.',
          ),
        ],
        const SizedBox(height: 10),
        DropdownButtonFormField<String>(
          initialValue:
              destinationEntries.any((entry) => entry.key == _destinationName)
                  ? _destinationName
                  : null,
          decoration: InputDecoration(
            labelText: _serviceType == 'vehicle_inspection_assist'
                ? 'Nơi đăng kiểm'
                : 'Điểm đến nhanh',
            hintText: _customDestinationSelected
                ? (_destinationName == 'Điểm đã ghim'
                    ? _pinnedDestinationLabel(trips.selectedDestination)
                    : _destinationName)
                : null,
            prefixIcon: Icon(_serviceType == 'vehicle_inspection_assist'
                ? Icons.fact_check_outlined
                : Icons.location_on_outlined),
          ),
          items: destinationEntries
              .map((entry) =>
                  DropdownMenuItem(value: entry.key, child: Text(entry.key)))
              .toList(),
          onChanged: trips.busy
              ? null
              : (value) {
                  if (value == null) return;
                  final destination = RiderTripController.destinations[value];
                  trips.selectDestination(destination);
                  _destinationSearch.clear();
                  trips.clearPlaceSearch();
                  setState(() {
                    _destinationName = value;
                    _customDestinationSelected = false;
                    _pinDestinationMode = false;
                  });
                },
        ),
        const SizedBox(height: 8),
        OutlinedButton.icon(
          onPressed: trips.busy
              ? null
              : () => setState(() => _pinDestinationMode = true),
          icon: const Icon(Icons.add_location_alt_outlined),
          label: Text(_pinDestinationMode
              ? 'Chạm vị trí cần đến trên phần bản đồ phía trên'
              : _destinationName == 'Điểm đã ghim'
                  ? 'Đổi điểm đã ghim trên bản đồ'
                  : 'Ghim điểm khác trên bản đồ'),
        ),
        if (_customDestinationSelected &&
            trips.selectedDestination != null) ...[
          const SizedBox(height: 8),
          _InfoBox(
            icon: Icons.location_on_rounded,
            text: _destinationName == 'Điểm đã ghim'
                ? _pinnedDestinationLabel(trips.selectedDestination)
                : 'Đã chọn: $_destinationName',
          ),
        ],
        const SizedBox(height: 14),
        Row(
          children: [
            Expanded(
              child: ChoiceChip(
                label: const SizedBox(
                    width: double.infinity,
                    child: Text('Ngay bây giờ', textAlign: TextAlign.center)),
                selected: _bookingMode == 'immediate',
                onSelected: (_) => setState(() {
                  _bookingMode = 'immediate';
                  _scheduledAt = null;
                }),
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: ChoiceChip(
                label: const SizedBox(
                    width: double.infinity,
                    child: Text('Hẹn giờ', textAlign: TextAlign.center)),
                selected: _bookingMode == 'scheduled',
                onSelected: (_) async {
                  setState(() => _bookingMode = 'scheduled');
                  await _pickSchedule();
                },
              ),
            ),
          ],
        ),
        if (_bookingMode == 'scheduled') ...[
          const SizedBox(height: 8),
          OutlinedButton.icon(
            onPressed: _pickSchedule,
            icon: const Icon(Icons.schedule_rounded),
            label: Text(_scheduledAt == null
                ? 'Chọn thời gian'
                : _dateTimeLabel(_scheduledAt!)),
          ),
        ],
        if (_serviceType == 'vehicle_inspection_assist') ...[
          const SizedBox(height: 10),
          const _InfoBox(
            icon: Icons.description_outlined,
            text:
                'Khi nhận xe, FlashX sẽ xác nhận xe và giấy tờ bàn giao trước khi mang đi đăng kiểm.',
          ),
        ],
        if (trips.error != null) ...[
          const SizedBox(height: 10),
          _ErrorBox(message: trips.error!),
        ],
        const SizedBox(height: 16),
        FilledButton.icon(
          onPressed: trips.busy || compatible.isEmpty
              ? null
              : () => _showEstimateAndBook(destinationEntries),
          icon: trips.busy
              ? const SizedBox.square(
                  dimension: 18,
                  child: CircularProgressIndicator(strokeWidth: 2))
              : const Icon(Icons.bolt_rounded),
          label: Text(trips.busy ? 'Đang xử lý...' : 'Xem giá trước'),
        ),
      ],
    );
  }

  void _queuePlaceSearch(String value) {
    _placeSearchDebounce?.cancel();
    final query = value.trim();
    if (query.runes.length < 3) {
      trips.clearPlaceSearch();
      setState(() {});
      return;
    }
    _placeSearchDebounce = Timer(const Duration(milliseconds: 400), () {
      if (mounted) trips.searchPlaces(query);
    });
    setState(() {});
  }

  void _selectPlaceResult(Map<String, dynamic> place) {
    final location = place['location'];
    if (location is! Map) return;
    final lat = (location['lat'] as num?)?.toDouble();
    final lng = (location['lng'] as num?)?.toDouble();
    if (lat == null || lng == null) return;
    final name = place['name']?.toString().trim();
    final address = place['address']?.toString().trim();
    final label = name?.isNotEmpty == true
        ? name!
        : address?.isNotEmpty == true
            ? address!
            : 'Điểm đã tìm';
    trips.selectDestination({'lat': lat, 'lng': lng});
    trips.clearPlaceSearch();
    _destinationSearch.text = label;
    FocusScope.of(context).unfocus();
    setState(() {
      _destinationName = label;
      _customDestinationSelected = true;
      _pinDestinationMode = false;
    });
  }

  void _selectService(String service) {
    setState(() {
      _serviceType = service;
      trips.estimate = null;
      final compatible = trips.vehiclesForService(service);
      trips.selectVehicle(
          compatible.isEmpty ? null : compatible.first['id']?.toString());
      _placeSearchDebounce?.cancel();
      _destinationSearch.clear();
      trips.clearPlaceSearch();
      _customDestinationSelected = false;
      _pinDestinationMode = false;
      if (service == 'vehicle_inspection_assist') {
        _destinationName = 'Trung tâm đăng kiểm Thanh Hóa';
      } else {
        _destinationName = 'Vincom Plaza Thanh Hóa';
      }
      trips.selectDestination(
          RiderTripController.destinations[_destinationName]);
    });
  }

  Future<void> _pickSchedule() async {
    final initial =
        _scheduledAt ?? DateTime.now().add(const Duration(hours: 1));
    final date = await showDatePicker(
      context: context,
      initialDate: initial,
      firstDate: DateTime.now(),
      lastDate: DateTime.now().add(const Duration(days: 30)),
    );
    if (!mounted || date == null) return;
    final time = await showTimePicker(
      context: context,
      initialTime: TimeOfDay.fromDateTime(initial),
    );
    if (!mounted || time == null) return;
    setState(() {
      _bookingMode = 'scheduled';
      _scheduledAt =
          DateTime(date.year, date.month, date.day, time.hour, time.minute);
    });
  }

  Future<void> _showEstimateAndBook(
      List<MapEntry<String, Map<String, double>>> destinations) async {
    final destination = trips.selectedDestination ??
        destinations
            .firstWhere((entry) => entry.key == _destinationName,
                orElse: () => destinations.first)
            .value;
    final ok = await trips.estimateTo(destination, serviceType: _serviceType);
    if (!mounted || !ok || trips.estimate == null) return;
    final estimate = trips.estimate!;
    final fare =
        (estimate['fare'] as Map<String, dynamic>?)?['total_minor'] ?? 0;
    final distance = (estimate['distance_m'] as num?)?.toDouble() ?? 0;
    final duration = (estimate['duration_s'] as num?)?.toDouble() ?? 0;
    final vehicle = trips.vehicles
        .where((e) => e['id']?.toString() == trips.selectedVehicleID)
        .firstOrNull;

    final accepted = await showModalBottomSheet<bool>(
      context: context,
      isScrollControlled: true,
      builder: (context) => SafeArea(
        top: false,
        child: Padding(
          padding: const EdgeInsets.fromLTRB(20, 12, 20, 24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Center(child: _handle()),
              const SizedBox(height: 18),
              Text(_serviceLabel(_serviceType),
                  style: Theme.of(context).textTheme.headlineSmall),
              const SizedBox(height: 6),
              Text(
                  '${vehicle == null ? 'Xe của bạn' : _vehicleLabel(vehicle)} · $_destinationName'),
              const SizedBox(height: 18),
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                    color: const Color(0xFFF7F8FA),
                    borderRadius: BorderRadius.circular(18)),
                child: Row(
                  children: [
                    Expanded(
                        child: _Stat('Quãng đường',
                            '${(distance / 1000).toStringAsFixed(1)} km')),
                    Expanded(
                        child:
                            _Stat('Dự kiến', '${(duration / 60).ceil()} phút')),
                    Expanded(child: _Stat('Thanh toán', 'Tiền mặt')),
                  ],
                ),
              ),
              if (_bookingMode == 'scheduled') ...[
                const SizedBox(height: 10),
                _InfoBox(
                  icon: Icons.schedule_rounded,
                  text: _scheduledAt == null
                      ? 'Bạn chưa chọn giờ hẹn.'
                      : 'Thực hiện lúc ${_dateTimeLabel(_scheduledAt!)}',
                ),
              ],
              const SizedBox(height: 18),
              Row(
                children: [
                  const Text('Giá dự kiến',
                      style: TextStyle(fontWeight: FontWeight.w700)),
                  const Spacer(),
                  Text(_money(fare),
                      style: const TextStyle(
                          fontSize: 24,
                          fontWeight: FontWeight.w900,
                          color: FlashXTheme.navy)),
                ],
              ),
              const SizedBox(height: 16),
              FilledButton.icon(
                onPressed: _bookingMode == 'scheduled' && _scheduledAt == null
                    ? null
                    : () => Navigator.pop(context, true),
                icon: const Icon(Icons.check_circle_outline_rounded),
                label: const Text('Xác nhận đặt dịch vụ'),
              ),
            ],
          ),
        ),
      ),
    );
    if (accepted == true) {
      await trips.createTrip(
        destination,
        serviceType: _serviceType,
        bookingMode: _bookingMode,
        scheduledAt: _scheduledAt,
      );
    }
  }

  Future<void> _showAddVehicle() async {
    final isBike = _serviceType == 'designated_driver_bike';
    final plate = TextEditingController();
    final brand = TextEditingController();
    final model = TextEditingController();
    final color = TextEditingController();
    var transmission = isBike ? 'n/a' : 'automatic';
    final submit = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => StatefulBuilder(
        builder: (context, setDialogState) => AlertDialog(
          title: Text(isBike ? 'Thêm xe máy' : 'Thêm ô tô'),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextField(
                    controller: plate,
                    decoration: const InputDecoration(labelText: 'Biển số *')),
                const SizedBox(height: 10),
                TextField(
                    controller: brand,
                    decoration: const InputDecoration(labelText: 'Hãng xe')),
                const SizedBox(height: 10),
                TextField(
                    controller: model,
                    decoration: const InputDecoration(labelText: 'Dòng xe')),
                const SizedBox(height: 10),
                TextField(
                    controller: color,
                    decoration: const InputDecoration(labelText: 'Màu xe')),
                if (!isBike) ...[
                  const SizedBox(height: 10),
                  DropdownButtonFormField<String>(
                    initialValue: transmission,
                    decoration: const InputDecoration(labelText: 'Hộp số'),
                    items: const [
                      DropdownMenuItem(
                          value: 'automatic', child: Text('Số tự động')),
                      DropdownMenuItem(value: 'manual', child: Text('Số sàn')),
                    ],
                    onChanged: (value) => setDialogState(
                        () => transmission = value ?? transmission),
                  ),
                ],
              ],
            ),
          ),
          actions: [
            TextButton(
                onPressed: () => Navigator.pop(dialogContext, false),
                child: const Text('Đóng')),
            FilledButton(
              onPressed: () => Navigator.pop(dialogContext, true),
              child: const Text('Lưu xe'),
            ),
          ],
        ),
      ),
    );
    if (submit == true && plate.text.trim().isNotEmpty) {
      await trips.createVehicle(
        type: isBike ? 'motorbike' : 'car',
        licensePlate: plate.text,
        brand: brand.text,
        model: model.text,
        color: color.text,
        transmission: transmission,
      );
    }
    plate.dispose();
    brand.dispose();
    model.dispose();
    color.dispose();
  }

  Widget _activePanel(Map<String, dynamic> trip) {
    final status = trip['status']?.toString() ?? '';
    final service = trip['service_type']?.toString() ?? '';
    final incidentOpen = trip['incident_open'] == true;
    final canReportSupport = !incidentOpen &&
        const {
          'accepted',
          'arriving',
          'arriving_for_pickup',
          'arrived',
          'arrived_for_pickup',
          'vehicle_received',
          'in_progress',
          'en_route_to_inspection',
          'arrived_at_inspection_center',
          'inspection_in_progress',
          'inspection_completed',
          'returning_vehicle',
          'arrived_for_return',
          'handover',
        }.contains(status);
    final pickupEvidence = trips.custodyFor('pickup');
    final returnEvidence = trips.custodyFor('return');
    final payment = trips.activePayment;
    final paymentStatus = switch (payment?['status']?.toString()) {
      'paid' => 'Đã thu tiền mặt',
      'cancelled' => 'Đã hủy thanh toán',
      'failed' => 'Thanh toán lỗi',
      _ => 'Thanh toán khi hoàn tất',
    };
    final paymentAmount = payment?['amount_minor'] ??
        trip['final_fare_minor'] ??
        trip['estimated_fare_minor'] ??
        0;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Center(child: _handle()),
        const SizedBox(height: 14),
        Text(_serviceLabel(service),
            style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: 4),
        Text(_statusLabel(status),
            style: Theme.of(context).textTheme.headlineSmall),
        const SizedBox(height: 5),
        Text(_statusHint(status, service),
            style: Theme.of(context).textTheme.bodyMedium),
        const SizedBox(height: 14),
        _ProgressLine(
            status: status, inspection: service == 'vehicle_inspection_assist'),
        if (incidentOpen) ...[
          const SizedBox(height: 12),
          const _InfoBox(
            icon: Icons.support_agent_rounded,
            text:
                'FlashX Operations đang tiếp nhận sự cố và sẽ hỗ trợ tài xế/khách hàng.',
            warning: true,
          ),
        ],
        if (trip['driver_id']?.toString().isNotEmpty == true) ...[
          const SizedBox(height: 14),
          Container(
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(
              color: const Color(0xFFF7F8FA),
              borderRadius: BorderRadius.circular(18),
            ),
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
                      const Text('Tài xế/đối tác FlashX',
                          style: TextStyle(fontWeight: FontWeight.w800)),
                      Text(
                          'Mã ${_shortId(trip['driver_id']?.toString() ?? '')}'),
                    ],
                  ),
                ),
                const Icon(Icons.verified_rounded, color: FlashXTheme.success),
              ],
            ),
          ),
        ],
        if (trips.driverLocation != null) ...[
          const SizedBox(height: 10),
          const Text('● Vị trí đang cập nhật trực tiếp',
              style: TextStyle(
                  color: FlashXTheme.success, fontWeight: FontWeight.w700)),
        ],
        if (trips.liveRoutes != null) ...[
          const SizedBox(height: 10),
          _LiveRouteCard(snapshot: trips.liveRoutes!),
        ],
        const SizedBox(height: 10),
        _InfoBox(
          icon: Icons.payments_outlined,
          text: 'Tiền mặt · $paymentStatus · ${_money(paymentAmount)}',
        ),
        if (service == 'vehicle_inspection_assist' &&
            trips.inspectionChecklist != null) ...[
          const SizedBox(height: 12),
          _RiderInspectionChecklistCard(
            snapshot: trips.inspectionChecklist!,
            busy: trips.busy,
            editable: const {
              'scheduled',
              'searching',
              'accepted',
              'arriving_for_pickup',
              'arrived_for_pickup',
            }.contains(status),
            onEdit: _editInspectionChecklistItem,
          ),
        ],
        if (pickupEvidence != null) ...[
          const SizedBox(height: 12),
          _RiderCustodyCard(
            stage: 'pickup',
            snapshot: pickupEvidence,
            busy: trips.busy,
            onReview: () => _reviewCustodyEvidence('pickup', pickupEvidence),
          ),
        ],
        if (returnEvidence != null) ...[
          const SizedBox(height: 12),
          _RiderCustodyCard(
            stage: 'return',
            snapshot: returnEvidence,
            busy: trips.busy,
            onReview: () => _reviewCustodyEvidence('return', returnEvidence),
          ),
        ],
        if (canReportSupport) ...[
          const SizedBox(height: 12),
          OutlinedButton.icon(
            onPressed: trips.busy ? null : _reportIncident,
            icon: const Icon(Icons.support_agent_rounded),
            label: const Text('Báo sự cố / cần hỗ trợ'),
          ),
        ],
        if (trips.canCancel) ...[
          const SizedBox(height: 14),
          OutlinedButton.icon(
            onPressed: trips.busy ? null : trips.cancel,
            icon: const Icon(Icons.close_rounded),
            label: const Text('Hủy yêu cầu'),
          ),
        ],
      ],
    );
  }

  Future<void> _reportIncident() async {
    var type = 'support';
    final note = TextEditingController();
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => StatefulBuilder(
        builder: (context, setDialogState) => AlertDialog(
          title: const Text('Báo sự cố / cần hỗ trợ'),
          content: SingleChildScrollView(
            child: Column(mainAxisSize: MainAxisSize.min, children: [
              DropdownButtonFormField<String>(
                initialValue: type,
                decoration:
                    const InputDecoration(labelText: 'Bạn cần hỗ trợ về'),
                items: const [
                  DropdownMenuItem(
                      value: 'support', child: Text('Cần Operations hỗ trợ')),
                  DropdownMenuItem(
                      value: 'driver_issue', child: Text('Vấn đề với tài xế')),
                  DropdownMenuItem(
                      value: 'vehicle_issue', child: Text('Xe gặp vấn đề')),
                  DropdownMenuItem(
                      value: 'safety', child: Text('An toàn / va chạm')),
                  DropdownMenuItem(value: 'other', child: Text('Khác')),
                ],
                onChanged: (value) =>
                    setDialogState(() => type = value ?? type),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: note,
                maxLines: 4,
                maxLength: 1000,
                decoration: const InputDecoration(
                  labelText: 'Mô tả ngắn',
                  hintText: 'Nêu tình huống để Operations xử lý nhanh hơn',
                ),
              ),
              const SizedBox(height: 8),
              const _InfoBox(
                icon: Icons.info_outline_rounded,
                text:
                    'Khi gửi, công việc sẽ được đánh dấu có sự cố để Operations tiếp nhận trước khi tiếp tục các bước nhạy cảm.',
                warning: true,
              ),
            ]),
          ),
          actions: [
            TextButton(
                onPressed: () => Navigator.pop(dialogContext, false),
                child: const Text('Đóng')),
            FilledButton(
                onPressed: () => Navigator.pop(dialogContext, true),
                child: const Text('Gửi hỗ trợ')),
          ],
        ),
      ),
    );
    final noteText = note.text.trim();
    note.dispose();
    if (confirmed != true || !mounted) return;
    final ok = await trips.reportIncident(type, noteText);
    if (!mounted) return;
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(
        content: Text(ok
            ? 'Đã gửi sự cố tới FlashX Operations.'
            : trips.error ?? 'Không thể gửi yêu cầu hỗ trợ.'),
      ));
  }

  Future<void> _editInspectionChecklistItem(Map<String, dynamic> item) async {
    var status = item['customer_status']?.toString() ?? 'pending';
    if (status == 'pending') status = 'present';
    final required = item['required'] == true;
    final note =
        TextEditingController(text: item['customer_note']?.toString() ?? '');
    final saved = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => StatefulBuilder(
        builder: (context, setDialogState) => AlertDialog(
          title: Text(item['label']?.toString() ?? 'Giấy tờ đăng kiểm'),
          content: SingleChildScrollView(
            child: Column(mainAxisSize: MainAxisSize.min, children: [
              DropdownButtonFormField<String>(
                initialValue: status,
                decoration:
                    const InputDecoration(labelText: 'Tình trạng bàn giao'),
                items: [
                  const DropdownMenuItem(
                      value: 'present',
                      child: Text('Tôi sẽ bàn giao giấy này')),
                  const DropdownMenuItem(
                      value: 'not_available',
                      child: Text('Hiện không có giấy này')),
                  if (!required)
                    const DropdownMenuItem(
                        value: 'not_applicable', child: Text('Không áp dụng')),
                ],
                onChanged: (value) =>
                    setDialogState(() => status = value ?? status),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: note,
                maxLines: 3,
                maxLength: 500,
                decoration: const InputDecoration(
                  labelText: 'Ghi chú · không bắt buộc',
                  hintText: 'Ví dụ: Bản gốc để trong túi hồ sơ màu xanh',
                ),
              ),
              if (required) ...[
                const SizedBox(height: 8),
                const _InfoBox(
                  icon: Icons.info_outline_rounded,
                  text:
                      'Đây là giấy tờ bắt buộc. Tài xế phải xác nhận đã nhận thực tế trước khi FlashX được phép nhận xe.',
                ),
              ],
            ]),
          ),
          actions: [
            TextButton(
                onPressed: () => Navigator.pop(dialogContext, false),
                child: const Text('Đóng')),
            FilledButton(
                onPressed: () => Navigator.pop(dialogContext, true),
                child: const Text('Lưu khai báo')),
          ],
        ),
      ),
    );
    final noteText = note.text.trim();
    note.dispose();
    if (saved != true || !mounted) return;
    final ok = await trips.updateInspectionChecklistItem(
      item['key']?.toString() ?? '',
      status,
      noteText,
    );
    if (!mounted) return;
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(
        content: Text(ok
            ? 'Đã cập nhật giấy tờ đăng kiểm.'
            : trips.error ?? 'Không thể cập nhật giấy tờ.'),
      ));
  }

  Future<void> _reviewCustodyEvidence(
      String stage, Map<String, dynamic> snapshot) async {
    final evidence = Map<String, dynamic>.from(
      snapshot['evidence'] as Map? ?? const {},
    );
    final photos = (snapshot['photos'] as List? ?? const [])
        .whereType<Map>()
        .map((item) => Map<String, dynamic>.from(item))
        .toList(growable: false);
    final driverConfirmed = evidence['driver_confirmed_at'] != null;
    final riderConfirmed = evidence['rider_confirmed_at'] != null;
    final canConfirm = driverConfirmed && !riderConfirmed;
    final metrics = <String>[
      if (evidence['odometer_km'] != null) '${evidence['odometer_km']} km',
      if (evidence['fuel_percent'] != null) 'Xăng ${evidence['fuel_percent']}%',
      if (evidence['battery_percent'] != null)
        'Pin ${evidence['battery_percent']}%',
    ];

    final confirmed = await showModalBottomSheet<bool>(
      context: context,
      isScrollControlled: true,
      showDragHandle: true,
      builder: (sheetContext) => SafeArea(
        top: false,
        child: SingleChildScrollView(
          padding: EdgeInsets.fromLTRB(
            20,
            4,
            20,
            24 + MediaQuery.viewInsetsOf(sheetContext).bottom,
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            mainAxisSize: MainAxisSize.min,
            children: [
              Row(children: [
                Container(
                  width: 46,
                  height: 46,
                  decoration: BoxDecoration(
                    color: const Color(0xFFECFDF3),
                    borderRadius: BorderRadius.circular(15),
                  ),
                  child: Icon(
                    stage == 'pickup'
                        ? Icons.key_rounded
                        : Icons.handshake_rounded,
                    color: FlashXTheme.success,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        stage == 'pickup'
                            ? 'Xác nhận tình trạng khi nhận xe'
                            : 'Xác nhận tình trạng khi trả xe',
                        style: Theme.of(context).textTheme.titleLarge,
                      ),
                      const SizedBox(height: 2),
                      Text(
                        riderConfirmed
                            ? 'Bạn đã xác nhận bộ bằng chứng này.'
                            : driverConfirmed
                                ? 'Tài xế đã xác nhận. Hãy kiểm tra trước khi đồng ý.'
                                : 'Tài xế chưa hoàn tất xác nhận.',
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                    ],
                  ),
                ),
              ]),
              const SizedBox(height: 18),
              Container(
                padding: const EdgeInsets.all(14),
                decoration: BoxDecoration(
                  color: const Color(0xFFF7F8FA),
                  borderRadius: BorderRadius.circular(18),
                ),
                child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text('Tình trạng phương tiện',
                          style: TextStyle(fontWeight: FontWeight.w900)),
                      const SizedBox(height: 6),
                      Text(
                        evidence['condition_note']
                                    ?.toString()
                                    .trim()
                                    .isNotEmpty ==
                                true
                            ? evidence['condition_note'].toString()
                            : 'Chưa có mô tả tình trạng xe.',
                        style: Theme.of(context).textTheme.bodyMedium,
                      ),
                      if (metrics.isNotEmpty) ...[
                        const SizedBox(height: 10),
                        Wrap(
                          spacing: 7,
                          runSpacing: 7,
                          children: metrics
                              .map((metric) => Chip(
                                    visualDensity: VisualDensity.compact,
                                    label: Text(metric),
                                  ))
                              .toList(),
                        ),
                      ],
                    ]),
              ),
              const SizedBox(height: 16),
              Row(children: [
                Text('Ảnh bằng chứng',
                    style: Theme.of(context).textTheme.titleMedium),
                const Spacer(),
                Text('${photos.length} ảnh',
                    style: Theme.of(context).textTheme.bodySmall),
              ]),
              const SizedBox(height: 9),
              if (photos.isEmpty)
                const _InfoBox(
                  icon: Icons.photo_library_outlined,
                  text: 'Chưa có ảnh bằng chứng được tải lên.',
                )
              else
                GridView.builder(
                  shrinkWrap: true,
                  physics: const NeverScrollableScrollPhysics(),
                  gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                    crossAxisCount: 2,
                    crossAxisSpacing: 8,
                    mainAxisSpacing: 8,
                    childAspectRatio: 1.12,
                  ),
                  itemCount: photos.length,
                  itemBuilder: (context, index) {
                    final item = photos[index];
                    final photo = Map<String, dynamic>.from(
                      item['photo'] as Map? ?? const {},
                    );
                    final view = Map<String, dynamic>.from(
                      item['view'] as Map? ?? const {},
                    );
                    final url = view['url']?.toString() ?? '';
                    final contentType = photo['content_type']?.toString() ?? '';
                    final previewable = url.isNotEmpty &&
                        const {'image/jpeg', 'image/png', 'image/webp'}
                            .contains(contentType);
                    return Container(
                      clipBehavior: Clip.antiAlias,
                      decoration: BoxDecoration(
                        color: const Color(0xFFF1F3F7),
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(color: FlashXTheme.border),
                      ),
                      child: previewable
                          ? Stack(fit: StackFit.expand, children: [
                              Image.network(
                                url,
                                fit: BoxFit.cover,
                                errorBuilder: (_, __, ___) => const Center(
                                  child: Icon(Icons.broken_image_outlined),
                                ),
                              ),
                              Align(
                                alignment: Alignment.bottomCenter,
                                child: Container(
                                  width: double.infinity,
                                  color: const Color(0xAA0B132B),
                                  padding: const EdgeInsets.symmetric(
                                      horizontal: 8, vertical: 5),
                                  child: Text(
                                    photo['photo_type']?.toString() ?? 'Ảnh xe',
                                    maxLines: 1,
                                    overflow: TextOverflow.ellipsis,
                                    style: const TextStyle(
                                      color: Colors.white,
                                      fontSize: 10,
                                      fontWeight: FontWeight.w700,
                                    ),
                                  ),
                                ),
                              ),
                            ])
                          : Padding(
                              padding: const EdgeInsets.all(12),
                              child: Column(
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  const Icon(Icons.image_outlined,
                                      color: FlashXTheme.navy),
                                  const SizedBox(height: 6),
                                  Text(
                                    photo['filename']?.toString() ??
                                        'Ảnh bằng chứng',
                                    maxLines: 2,
                                    overflow: TextOverflow.ellipsis,
                                    textAlign: TextAlign.center,
                                    style: const TextStyle(
                                        fontSize: 10,
                                        fontWeight: FontWeight.w700),
                                  ),
                                  if (contentType == 'image/heic' ||
                                      contentType == 'image/heif')
                                    const Padding(
                                      padding: EdgeInsets.only(top: 3),
                                      child: Text('HEIC/HEIF',
                                          style: TextStyle(
                                              fontSize: 9,
                                              color:
                                                  FlashXTheme.textSecondary)),
                                    ),
                                ],
                              ),
                            ),
                    );
                  },
                ),
              const SizedBox(height: 16),
              _InfoBox(
                icon: Icons.verified_user_outlined,
                text: riderConfirmed
                    ? 'Bạn đã xác nhận mốc này. Nếu tài xế thay đổi ghi chú hoặc thêm ảnh, backend sẽ tự yêu cầu xác nhận lại.'
                    : canConfirm
                        ? 'Khi bấm xác nhận, bạn đồng ý bộ ảnh và tình trạng trên phản ánh phương tiện tại mốc ${stage == 'pickup' ? 'nhận xe' : 'trả xe'}.'
                        : 'Chờ tài xế hoàn tất tối thiểu 2 ảnh và xác nhận trước.',
              ),
              const SizedBox(height: 14),
              if (canConfirm)
                FilledButton.icon(
                  onPressed: trips.busy
                      ? null
                      : () => Navigator.pop(sheetContext, true),
                  icon: const Icon(Icons.verified_rounded),
                  label: Text(stage == 'pickup'
                      ? 'Tôi xác nhận tình trạng khi nhận xe'
                      : 'Tôi xác nhận tình trạng khi trả xe'),
                )
              else
                OutlinedButton(
                  onPressed: () => Navigator.pop(sheetContext, false),
                  child: const Text('Đóng'),
                ),
            ],
          ),
        ),
      ),
    );

    if (confirmed == true && mounted) {
      final ok = await trips.confirmCustody(stage);
      if (!mounted) return;
      ScaffoldMessenger.of(context)
        ..hideCurrentSnackBar()
        ..showSnackBar(SnackBar(
          content: Text(ok
              ? 'Đã xác nhận tình trạng phương tiện.'
              : trips.error ?? 'Không thể xác nhận bằng chứng.'),
        ));
    }
  }

  Future<void> _showHistoryTrip(Map<String, dynamic> trip) async {
    final tripID = trip['id']?.toString() ?? '';
    final completed = trip['status'] == 'completed';
    Map<String, dynamic>? rating;
    if (completed && tripID.isNotEmpty) {
      try {
        rating = await trips.loadTripRating(tripID);
      } catch (_) {
        // History detail remains usable if the optional rating lookup fails.
      }
    }
    if (!mounted) return;

    var stars = (rating?['stars'] as num?)?.toInt() ?? 5;
    final comment =
        TextEditingController(text: rating?['comment']?.toString() ?? '');
    final submitted = await showModalBottomSheet<bool>(
      context: context,
      isScrollControlled: true,
      showDragHandle: true,
      builder: (sheetContext) => StatefulBuilder(
        builder: (context, setSheetState) => SafeArea(
          top: false,
          child: SingleChildScrollView(
            padding: EdgeInsets.fromLTRB(
              20,
              4,
              20,
              24 + MediaQuery.viewInsetsOf(sheetContext).bottom,
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              mainAxisSize: MainAxisSize.min,
              children: [
                Row(children: [
                  _ServiceIcon(service: trip['service_type']?.toString() ?? ''),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          _serviceLabel(trip['service_type']?.toString() ?? ''),
                          style: Theme.of(context).textTheme.titleLarge,
                        ),
                        Text(
                          _statusLabel(trip['status']?.toString() ?? ''),
                          style: Theme.of(context).textTheme.bodySmall,
                        ),
                      ],
                    ),
                  ),
                  Text(
                    _money(trip['final_fare_minor'] ??
                        trip['estimated_fare_minor'] ??
                        0),
                    style: const TextStyle(
                        fontSize: 20, fontWeight: FontWeight.w900),
                  ),
                ]),
                const SizedBox(height: 18),
                Container(
                  padding: const EdgeInsets.all(14),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF7F8FA),
                    borderRadius: BorderRadius.circular(18),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('Mã công việc ${_shortId(tripID)}',
                          style: const TextStyle(fontWeight: FontWeight.w800)),
                      if ((trip['inspection_result']?.toString() ?? '')
                          .isNotEmpty) ...[
                        const SizedBox(height: 6),
                        Text(
                          'Kết quả đăng kiểm: ${_inspectionResult(trip['inspection_result'].toString())}',
                          style: Theme.of(context).textTheme.bodyMedium,
                        ),
                      ],
                      if (trip['scheduled_at'] != null) ...[
                        const SizedBox(height: 6),
                        Text(
                          'Lịch hẹn: ${_dateTimeLabel(DateTime.tryParse(trip['scheduled_at'].toString())?.toLocal() ?? DateTime.now())}',
                          style: Theme.of(context).textTheme.bodySmall,
                        ),
                      ],
                    ],
                  ),
                ),
                if (completed &&
                    trip['driver_id']?.toString().isNotEmpty == true) ...[
                  const SizedBox(height: 20),
                  Text(rating == null ? 'Đánh giá tài xế' : 'Đánh giá của bạn',
                      style: Theme.of(context).textTheme.titleMedium),
                  const SizedBox(height: 8),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: List.generate(5, (index) {
                      final selected = index < stars;
                      return IconButton(
                        onPressed: rating == null
                            ? () => setSheetState(() => stars = index + 1)
                            : null,
                        tooltip: '${index + 1} sao',
                        icon: Icon(
                          selected
                              ? Icons.star_rounded
                              : Icons.star_border_rounded,
                          color: selected
                              ? FlashXTheme.warning
                              : FlashXTheme.textSecondary,
                          size: 34,
                        ),
                      );
                    }),
                  ),
                  const SizedBox(height: 8),
                  TextField(
                    controller: comment,
                    enabled: rating == null,
                    maxLines: 3,
                    maxLength: 1000,
                    decoration: InputDecoration(
                      labelText: rating == null
                          ? 'Nhận xét · không bắt buộc'
                          : 'Nhận xét',
                      hintText: 'Chia sẻ trải nghiệm của bạn',
                    ),
                  ),
                  if (rating == null) ...[
                    const SizedBox(height: 10),
                    FilledButton.icon(
                      onPressed: trips.busy
                          ? null
                          : () => Navigator.pop(sheetContext, true),
                      icon: const Icon(Icons.star_rounded),
                      label: const Text('Gửi đánh giá'),
                    ),
                  ] else
                    const _InfoBox(
                      icon: Icons.verified_rounded,
                      text: 'Đánh giá đã được ghi nhận cho công việc này.',
                    ),
                ],
              ],
            ),
          ),
        ),
      ),
    );
    final commentText = comment.text.trim();
    comment.dispose();
    if (submitted == true && mounted) {
      final ok = await trips.rateTrip(tripID, stars, commentText);
      if (!mounted) return;
      ScaffoldMessenger.of(context)
        ..hideCurrentSnackBar()
        ..showSnackBar(SnackBar(
          content: Text(ok
              ? 'Cảm ơn bạn đã đánh giá tài xế.'
              : trips.error ?? 'Không thể gửi đánh giá.'),
        ));
    }
  }

  Widget _activity() {
    return SafeArea(
      child: RefreshIndicator(
        onRefresh: trips.loadHistory,
        child: ListView(
          padding: const EdgeInsets.fromLTRB(20, 24, 20, 32),
          children: [
            const _PageHeader(
                title: 'Hoạt động', subtitle: 'Lịch sử dịch vụ của bạn'),
            const SizedBox(height: 18),
            if (trips.history.isEmpty)
              const _InfoBox(
                  icon: Icons.history_rounded,
                  text: 'Chưa có dịch vụ nào được đặt.'),
            ...trips.history.map((trip) => Card(
                  margin: const EdgeInsets.only(bottom: 10),
                  clipBehavior: Clip.antiAlias,
                  child: InkWell(
                    onTap: () => _showHistoryTrip(trip),
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Row(
                        children: [
                          _ServiceIcon(
                              service: trip['service_type']?.toString() ?? ''),
                          const SizedBox(width: 12),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                    _serviceLabel(
                                        trip['service_type']?.toString() ?? ''),
                                    style: const TextStyle(
                                        fontWeight: FontWeight.w800)),
                                const SizedBox(height: 3),
                                Text(
                                    _statusLabel(
                                        trip['status']?.toString() ?? ''),
                                    style:
                                        Theme.of(context).textTheme.bodySmall),
                                if ((trip['inspection_result']?.toString() ??
                                        '')
                                    .isNotEmpty)
                                  Text(
                                      'Kết quả đăng kiểm: ${_inspectionResult(trip['inspection_result'].toString())}',
                                      style: Theme.of(context)
                                          .textTheme
                                          .bodySmall),
                              ],
                            ),
                          ),
                          Column(
                            crossAxisAlignment: CrossAxisAlignment.end,
                            children: [
                              Text(
                                  _money(trip['final_fare_minor'] ??
                                      trip['estimated_fare_minor'] ??
                                      0),
                                  style: const TextStyle(
                                      fontWeight: FontWeight.w900)),
                              const SizedBox(height: 4),
                              const Icon(Icons.chevron_right_rounded,
                                  size: 18, color: FlashXTheme.textSecondary),
                            ],
                          ),
                        ],
                      ),
                    ),
                  ),
                )),
          ],
        ),
      ),
    );
  }

  Widget _account() {
    final phone = widget.auth.phone ?? '';
    return SafeArea(
      child: ListView(
        padding: const EdgeInsets.fromLTRB(20, 24, 20, 32),
        children: [
          const _PageHeader(title: 'Tài khoản', subtitle: 'Khách hàng FlashX'),
          const SizedBox(height: 18),
          Container(
            padding: const EdgeInsets.all(18),
            decoration: BoxDecoration(
                color: FlashXTheme.navy,
                borderRadius: BorderRadius.circular(24)),
            child: Row(
              children: [
                const CircleAvatar(
                    radius: 26,
                    backgroundColor: FlashXTheme.yellow,
                    foregroundColor: FlashXTheme.navy,
                    child: Icon(Icons.person_rounded)),
                const SizedBox(width: 14),
                Expanded(
                  child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text('Khách hàng FlashX',
                            style: TextStyle(
                                color: Colors.white,
                                fontWeight: FontWeight.w800,
                                fontSize: 18)),
                        const SizedBox(height: 3),
                        Text(phone,
                            style: const TextStyle(color: Color(0xFFB9C1D5))),
                      ]),
                ),
              ],
            ),
          ),
          const SizedBox(height: 20),
          Row(children: [
            Expanded(
                child: Text('Xe của tôi',
                    style: Theme.of(context).textTheme.titleLarge)),
            TextButton.icon(
                onPressed: _showAddVehicle,
                icon: const Icon(Icons.add_rounded),
                label: const Text('Thêm')),
          ]),
          if (trips.vehicles.isEmpty)
            const _InfoBox(
                icon: Icons.directions_car_outlined,
                text: 'Chưa lưu phương tiện.')
          else
            ...trips.vehicles.map((vehicle) => ListTile(
                  contentPadding: EdgeInsets.zero,
                  leading: CircleAvatar(
                    backgroundColor: const Color(0xFFF1F3F7),
                    child: Icon(vehicle['type'] == 'motorbike'
                        ? Icons.two_wheeler_rounded
                        : Icons.directions_car_rounded),
                  ),
                  title: Text(_vehicleLabel(vehicle)),
                  subtitle: Text(vehicle['transmission'] == 'manual'
                      ? 'Số sàn'
                      : vehicle['transmission'] == 'automatic'
                          ? 'Số tự động'
                          : 'Xe máy'),
                  trailing: const Icon(Icons.verified_outlined,
                      color: FlashXTheme.success),
                )),
          const Divider(height: 30),
          const ListTile(
            contentPadding: EdgeInsets.zero,
            leading: Icon(Icons.payments_outlined),
            title: Text('Thanh toán'),
            subtitle: Text('Tiền mặt trong bản demo'),
          ),
          const ListTile(
            contentPadding: EdgeInsets.zero,
            leading: Icon(Icons.shield_outlined),
            title: Text('An toàn & hỗ trợ'),
            subtitle: Text('Theo dõi vị trí, giao nhận xe và hỗ trợ sự cố'),
          ),
          const SizedBox(height: 16),
          OutlinedButton.icon(
              onPressed: widget.auth.logout,
              icon: const Icon(Icons.logout_rounded),
              label: const Text('Đăng xuất')),
        ],
      ),
    );
  }

  static String _serviceLabel(String service) => switch (service) {
        'car' || 'designated_driver_car' => 'Lái hộ ô tô',
        'bike' || 'designated_driver_bike' => 'Lái hộ xe máy',
        'vehicle_inspection_assist' => 'Đăng kiểm hộ',
        _ => 'Dịch vụ FlashX',
      };

  static String _statusLabel(String status) => switch (status) {
        'scheduled' => 'Đã hẹn lịch',
        'searching' => 'Đang tìm tài xế phù hợp',
        'accepted' => 'Tài xế đã nhận yêu cầu',
        'arriving' || 'arriving_for_pickup' => 'Tài xế đang đến nhận xe',
        'arrived' || 'arrived_for_pickup' => 'Tài xế đã đến điểm nhận',
        'vehicle_received' => 'FlashX đã nhận xe',
        'in_progress' => 'Đang thực hiện hành trình',
        'en_route_to_inspection' => 'Đang đưa xe tới nơi đăng kiểm',
        'arrived_at_inspection_center' => 'Đã tới nơi đăng kiểm',
        'inspection_in_progress' => 'Đang thực hiện đăng kiểm',
        'inspection_completed' => 'Đã có kết quả đăng kiểm',
        'returning_vehicle' => 'Đang đưa xe về cho bạn',
        'arrived_for_return' => 'Đã tới điểm trả xe',
        'handover' => 'Đang bàn giao xe',
        'completed' => 'Dịch vụ đã hoàn thành',
        'cancelled' => 'Yêu cầu đã hủy',
        _ => 'Đang cập nhật',
      };

  static String _statusHint(String status, String service) {
    if (status == 'vehicle_received') {
      return 'Tình trạng xe đã được xác nhận. Sau bước này yêu cầu không còn hủy theo cách thông thường.';
    }
    if (status == 'inspection_completed') {
      return 'Kết quả đăng kiểm được lưu riêng; FlashX tiếp tục đưa xe về và bàn giao cho bạn.';
    }
    if (status == 'scheduled') {
      return 'FlashX sẽ bắt đầu tìm người thực hiện gần giờ hẹn.';
    }
    if (status == 'searching') {
      return 'Hệ thống đang kết nối với tài xế/đối tác đủ điều kiện.';
    }
    if (status == 'completed') return 'Cảm ơn bạn đã sử dụng FlashX.';
    return service == 'vehicle_inspection_assist'
        ? 'Theo dõi từng bước nhận xe, đăng kiểm và trả xe tại đây.'
        : 'Bạn có thể theo dõi tài xế và trạng thái xe theo thời gian thực.';
  }

  static String _inspectionResult(String value) => switch (value) {
        'passed' => 'Đạt',
        'failed' => 'Không đạt',
        'deferred' => 'Cần thực hiện lại',
        'unavailable' => 'Chưa có kết quả',
        _ => value,
      };

  static String _vehicleLabel(Map<String, dynamic> vehicle) {
    final plate = vehicle['license_plate']?.toString() ?? '';
    final brand = vehicle['brand']?.toString() ?? '';
    final model = vehicle['model']?.toString() ?? '';
    final name = [brand, model].where((e) => e.trim().isNotEmpty).join(' ');
    return name.isEmpty ? plate : '$name · $plate';
  }

  static String _money(dynamic value) {
    final amount =
        value is num ? value.toInt() : int.tryParse(value.toString()) ?? 0;
    return '${amount.toString().replaceAllMapped(RegExp(r'(?=(\d{3})+(?!\d))'), (m) => '.')}₫';
  }

  static String _shortId(String value) => value.length <= 8
      ? value
      : value.substring(value.length - 8).toUpperCase();

  static String _pinnedDestinationLabel(Map<String, double>? value) {
    if (value == null) return 'Chưa ghim điểm đến';
    final lat = value['lat']?.toStringAsFixed(5) ?? '—';
    final lng = value['lng']?.toStringAsFixed(5) ?? '—';
    return 'Điểm đã ghim · $lat, $lng';
  }

  static String _dateTimeLabel(DateTime value) {
    final h = value.hour.toString().padLeft(2, '0');
    final m = value.minute.toString().padLeft(2, '0');
    return '$h:$m · ${value.day.toString().padLeft(2, '0')}/${value.month.toString().padLeft(2, '0')}/${value.year}';
  }

  Widget _handle() => Container(
        width: 42,
        height: 4,
        decoration: BoxDecoration(
            color: const Color(0xFFD7DAE0),
            borderRadius: BorderRadius.circular(99)),
      );
}

class _BrandPill extends StatelessWidget {
  const _BrandPill();
  @override
  Widget build(BuildContext context) => Container(
        padding: const EdgeInsets.fromLTRB(9, 7, 13, 7),
        decoration: BoxDecoration(
            color: FlashXTheme.navy, borderRadius: BorderRadius.circular(18)),
        child: const Row(mainAxisSize: MainAxisSize.min, children: [
          Icon(Icons.bolt_rounded, color: FlashXTheme.yellow),
          SizedBox(width: 4),
          Text('FlashX',
              style:
                  TextStyle(color: Colors.white, fontWeight: FontWeight.w900)),
        ]),
      );
}

class _MapButton extends StatelessWidget {
  const _MapButton(
      {required this.icon, required this.tooltip, required this.onPressed});
  final IconData icon;
  final String tooltip;
  final VoidCallback? onPressed;
  @override
  Widget build(BuildContext context) => Material(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        child: IconButton(
            onPressed: onPressed,
            tooltip: tooltip,
            icon: Icon(icon, color: FlashXTheme.navy)),
      );
}

class _ActivePill extends StatelessWidget {
  const _ActivePill({required this.label, required this.service});
  final String label;
  final String service;
  @override
  Widget build(BuildContext context) => Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
        decoration: BoxDecoration(
            color: FlashXTheme.navy, borderRadius: BorderRadius.circular(18)),
        child: Row(children: [
          const Icon(Icons.bolt_rounded, color: FlashXTheme.yellow, size: 18),
          const SizedBox(width: 8),
          Expanded(
              child: Text('$service · $label',
                  style: const TextStyle(
                      color: Colors.white, fontWeight: FontWeight.w700))),
        ]),
      );
}

class _ServiceCard extends StatelessWidget {
  const _ServiceCard(
      {required this.selected,
      required this.icon,
      required this.title,
      required this.onTap});
  final bool selected;
  final IconData icon;
  final String title;
  final VoidCallback onTap;
  @override
  Widget build(BuildContext context) => Material(
        color: selected ? FlashXTheme.navy : const Color(0xFFF7F8FA),
        borderRadius: BorderRadius.circular(18),
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(18),
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 13),
            child: Column(children: [
              Icon(icon,
                  color: selected ? FlashXTheme.yellow : FlashXTheme.navy,
                  size: 28),
              const SizedBox(height: 7),
              Text(title,
                  textAlign: TextAlign.center,
                  maxLines: 2,
                  style: TextStyle(
                      color: selected ? Colors.white : FlashXTheme.navy,
                      fontWeight: FontWeight.w800,
                      fontSize: 12)),
            ]),
          ),
        ),
      );
}

class _RiderInspectionChecklistCard extends StatelessWidget {
  const _RiderInspectionChecklistCard({
    required this.snapshot,
    required this.busy,
    required this.editable,
    required this.onEdit,
  });

  final Map<String, dynamic> snapshot;
  final bool busy;
  final bool editable;
  final ValueChanged<Map<String, dynamic>> onEdit;

  @override
  Widget build(BuildContext context) {
    final items = (snapshot['items'] as List? ?? const [])
        .whereType<Map>()
        .map((item) => Map<String, dynamic>.from(item))
        .toList(growable: false);
    final ready = snapshot['ready'] == true;
    final customerComplete = snapshot['customer_complete'] == true;
    final checklist = Map<String, dynamic>.from(
      snapshot['checklist'] as Map? ?? const {},
    );
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: ready ? const Color(0xFFECFDF3) : const Color(0xFFF7F8FA),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(
          color: ready ? const Color(0xFFABEFC6) : FlashXTheme.border,
        ),
      ),
      child: Column(crossAxisAlignment: CrossAxisAlignment.stretch, children: [
        Row(children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(13),
            ),
            child: Icon(Icons.description_rounded,
                color: ready ? FlashXTheme.success : FlashXTheme.navy),
          ),
          const SizedBox(width: 10),
          Expanded(
            child:
                Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              const Text('Giấy tờ đăng kiểm',
                  style: TextStyle(fontWeight: FontWeight.w900)),
              const SizedBox(height: 2),
              Text(
                ready
                    ? 'Khách và tài xế đã xác minh đủ giấy tờ bắt buộc'
                    : customerComplete
                        ? 'Bạn đã khai báo · chờ tài xế đối chiếu thực tế'
                        : 'Khai báo giấy tờ sẽ bàn giao cho FlashX',
                style: TextStyle(
                  color:
                      ready ? FlashXTheme.success : FlashXTheme.textSecondary,
                  fontSize: 12,
                  fontWeight: FontWeight.w700,
                ),
              ),
            ]),
          ),
          Text('v${checklist['template_version'] ?? '—'}',
              style: Theme.of(context).textTheme.bodySmall),
        ]),
        const SizedBox(height: 12),
        ...items.map((item) {
          final required = item['required'] == true;
          final status = item['customer_status']?.toString() ?? 'pending';
          final declared = status != 'pending';
          final label = switch (status) {
            'present' => 'Sẽ bàn giao',
            'not_available' => 'Không có',
            'not_applicable' => 'Không áp dụng',
            _ => 'Chưa khai báo',
          };
          return Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: Material(
              color: Colors.white,
              borderRadius: BorderRadius.circular(14),
              child: InkWell(
                onTap: editable && !busy ? () => onEdit(item) : null,
                borderRadius: BorderRadius.circular(14),
                child: Padding(
                  padding: const EdgeInsets.all(12),
                  child: Row(children: [
                    Icon(
                      declared
                          ? Icons.check_circle_rounded
                          : Icons.radio_button_unchecked_rounded,
                      size: 19,
                      color: status == 'present'
                          ? FlashXTheme.success
                          : status == 'not_available'
                              ? FlashXTheme.warning
                              : FlashXTheme.textSecondary,
                    ),
                    const SizedBox(width: 9),
                    Expanded(
                      child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(item['label']?.toString() ?? '',
                                style: const TextStyle(
                                    fontWeight: FontWeight.w800)),
                            const SizedBox(height: 2),
                            Text('$label${required ? ' · Bắt buộc' : ''}',
                                style: Theme.of(context).textTheme.bodySmall),
                          ]),
                    ),
                    if (editable)
                      const Icon(Icons.chevron_right_rounded,
                          color: FlashXTheme.textSecondary),
                  ]),
                ),
              ),
            ),
          );
        }),
        if (!editable && !ready)
          const _InfoBox(
            icon: Icons.lock_outline_rounded,
            text:
                'Khai báo đã khóa ở giai đoạn hiện tại. Nếu giấy tờ có vấn đề, tài xế sẽ báo sự cố để Operations xử lý.',
          ),
      ]),
    );
  }
}

class _RiderCustodyCard extends StatelessWidget {
  const _RiderCustodyCard({
    required this.stage,
    required this.snapshot,
    required this.busy,
    required this.onReview,
  });

  final String stage;
  final Map<String, dynamic> snapshot;
  final bool busy;
  final VoidCallback onReview;

  @override
  Widget build(BuildContext context) {
    final evidence = Map<String, dynamic>.from(
      snapshot['evidence'] as Map? ?? const {},
    );
    final photos = snapshot['photos'] as List? ?? const [];
    final driverConfirmed = evidence['driver_confirmed_at'] != null;
    final riderConfirmed = evidence['rider_confirmed_at'] != null;
    final ready = snapshot['ready'] == true;
    final needsAction = driverConfirmed && !riderConfirmed;
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: needsAction
            ? const Color(0xFFFFF8E8)
            : ready
                ? const Color(0xFFECFDF3)
                : const Color(0xFFF7F8FA),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(
          color: needsAction
              ? const Color(0xFFFEC84B)
              : ready
                  ? const Color(0xFFABEFC6)
                  : FlashXTheme.border,
        ),
      ),
      child: Column(crossAxisAlignment: CrossAxisAlignment.stretch, children: [
        Row(children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(13),
            ),
            child: Icon(
              stage == 'pickup' ? Icons.key_rounded : Icons.handshake_rounded,
              color: needsAction ? FlashXTheme.warning : FlashXTheme.success,
            ),
          ),
          const SizedBox(width: 10),
          Expanded(
            child:
                Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Text(
                stage == 'pickup' ? 'Bằng chứng nhận xe' : 'Bằng chứng trả xe',
                style: const TextStyle(fontWeight: FontWeight.w900),
              ),
              const SizedBox(height: 2),
              Text(
                ready
                    ? 'Hai bên đã xác nhận'
                    : needsAction
                        ? 'Tài xế đã xác nhận · đến lượt bạn'
                        : 'Đang được tài xế hoàn thiện',
                style: TextStyle(
                  color: needsAction
                      ? FlashXTheme.warning
                      : ready
                          ? FlashXTheme.success
                          : FlashXTheme.textSecondary,
                  fontSize: 12,
                  fontWeight: FontWeight.w700,
                ),
              ),
            ]),
          ),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 5),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(99),
            ),
            child: Text(
              '${photos.length} ảnh',
              style: const TextStyle(fontSize: 10, fontWeight: FontWeight.w800),
            ),
          ),
        ]),
        const SizedBox(height: 10),
        OutlinedButton.icon(
          onPressed: busy ? null : onReview,
          icon: Icon(needsAction
              ? Icons.verified_user_outlined
              : Icons.visibility_outlined),
          label: Text(needsAction ? 'Kiểm tra & xác nhận' : 'Xem bằng chứng'),
        ),
      ]),
    );
  }
}

class _LiveRouteCard extends StatelessWidget {
  const _LiveRouteCard({required this.snapshot});
  final Map<String, dynamic> snapshot;

  @override
  Widget build(BuildContext context) {
    final primary = snapshot['primary'] is Map
        ? Map<String, dynamic>.from(snapshot['primary'] as Map)
        : snapshot['service'] is Map
            ? Map<String, dynamic>.from(snapshot['service'] as Map)
            : const <String, dynamic>{};
    final distance = (primary['distance_m'] as num?)?.toDouble() ?? 0;
    final duration = (primary['duration_s'] as num?)?.toDouble() ?? 0;
    final phase = snapshot['phase']?.toString() ?? 'service';
    final title = switch (phase) {
      'approach' => 'Tài xế đang tới điểm nhận',
      'return' => 'Đang đưa xe về điểm trả',
      _ => 'Lộ trình dịch vụ',
    };
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: const Color(0xFFECFDF3),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: const Color(0xFFABEFC6)),
      ),
      child: Row(children: [
        const Icon(Icons.route_rounded, color: FlashXTheme.success),
        const SizedBox(width: 10),
        Expanded(
          child:
              Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Text(title, style: const TextStyle(fontWeight: FontWeight.w800)),
            const SizedBox(height: 2),
            Text(
              distance > 0 && duration > 0
                  ? '${(distance / 1000).toStringAsFixed(1)} km · khoảng ${(duration / 60).ceil()} phút'
                  : 'Đang cập nhật khoảng cách và thời gian',
              style: Theme.of(context).textTheme.bodySmall,
            ),
          ]),
        ),
      ]),
    );
  }
}

class _InfoBox extends StatelessWidget {
  const _InfoBox(
      {required this.icon, required this.text, this.warning = false});
  final IconData icon;
  final String text;
  final bool warning;
  @override
  Widget build(BuildContext context) => Container(
        padding: const EdgeInsets.all(13),
        decoration: BoxDecoration(
          color: warning ? const Color(0xFFFFF5E6) : const Color(0xFFF4F7FB),
          borderRadius: BorderRadius.circular(16),
        ),
        child: Row(children: [
          Icon(icon, color: warning ? FlashXTheme.warning : FlashXTheme.navy),
          const SizedBox(width: 10),
          Expanded(child: Text(text, style: const TextStyle(fontSize: 13))),
        ]),
      );
}

class _ErrorBox extends StatelessWidget {
  const _ErrorBox({required this.message});
  final String message;
  @override
  Widget build(BuildContext context) => Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
            color: const Color(0xFFFFECEC),
            borderRadius: BorderRadius.circular(14)),
        child: Row(children: [
          const Icon(Icons.error_outline_rounded, color: FlashXTheme.danger),
          const SizedBox(width: 8),
          Expanded(
              child: Text(message,
                  style: const TextStyle(color: FlashXTheme.danger))),
        ]),
      );
}

class _Stat extends StatelessWidget {
  const _Stat(this.label, this.value);
  final String label;
  final String value;
  @override
  Widget build(BuildContext context) => Column(children: [
        Text(value,
            style: const TextStyle(
                fontWeight: FontWeight.w900, color: FlashXTheme.navy)),
        const SizedBox(height: 3),
        Text(label,
            textAlign: TextAlign.center,
            style: Theme.of(context).textTheme.bodySmall),
      ]);
}

class _ProgressLine extends StatelessWidget {
  const _ProgressLine({required this.status, required this.inspection});
  final String status;
  final bool inspection;
  @override
  Widget build(BuildContext context) {
    final stages = inspection
        ? const [
            'searching',
            'accepted',
            'vehicle_received',
            'inspection_in_progress',
            'returning_vehicle',
            'handover',
            'completed'
          ]
        : const [
            'searching',
            'accepted',
            'arrived',
            'vehicle_received',
            'in_progress',
            'handover',
            'completed'
          ];
    final aliases = <String, String>{
      'scheduled': 'searching',
      'arriving': 'accepted',
      'arriving_for_pickup': 'accepted',
      'arrived_for_pickup': 'arrived',
      'en_route_to_inspection': 'vehicle_received',
      'arrived_at_inspection_center': 'inspection_in_progress',
      'inspection_completed': 'returning_vehicle',
      'arrived_for_return': 'handover',
    };
    final normalized = aliases[status] ?? status;
    final current = stages.indexOf(normalized).clamp(0, stages.length - 1);
    return Row(
      children: List.generate(
          stages.length,
          (index) => Expanded(
                child: Container(
                  margin: EdgeInsets.only(
                      right: index == stages.length - 1 ? 0 : 4),
                  height: 5,
                  decoration: BoxDecoration(
                    color: index <= current
                        ? FlashXTheme.lime
                        : const Color(0xFFE5E7EB),
                    borderRadius: BorderRadius.circular(99),
                  ),
                ),
              )),
    );
  }
}

class _ServiceIcon extends StatelessWidget {
  const _ServiceIcon({required this.service});
  final String service;
  @override
  Widget build(BuildContext context) => CircleAvatar(
        backgroundColor: const Color(0xFFF1F3F7),
        foregroundColor: FlashXTheme.navy,
        child: Icon(service.contains('inspection')
            ? Icons.fact_check_rounded
            : service.contains('bike')
                ? Icons.two_wheeler_rounded
                : Icons.directions_car_rounded),
      );
}

class _PageHeader extends StatelessWidget {
  const _PageHeader({required this.title, required this.subtitle});
  final String title;
  final String subtitle;
  @override
  Widget build(BuildContext context) => Row(children: [
        const _BrandPill(),
        const SizedBox(width: 14),
        Expanded(
            child:
                Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text(title, style: Theme.of(context).textTheme.headlineSmall),
          Text(subtitle, style: Theme.of(context).textTheme.bodySmall),
        ])),
      ]);
}

extension _FirstOrNull<T> on Iterable<T> {
  T? get firstOrNull {
    final iterator = this.iterator;
    return iterator.moveNext() ? iterator.current : null;
  }
}
