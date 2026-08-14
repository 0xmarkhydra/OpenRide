import 'dart:async';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';

import '../../core/config/app_config.dart';
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
  Timer? _ticker;
  var _index = 0;

  @override
  void initState() {
    super.initState();
    trips = DriverTripController(auth: widget.auth)..addListener(_refresh);
    trips.initialize();
    _ticker = Timer.periodic(const Duration(seconds: 1), (_) {
      if (mounted && trips.currentOffer != null) setState(() {});
    });
  }

  void _refresh() {
    if (mounted) setState(() {});
  }

  @override
  void dispose() {
    _ticker?.cancel();
    trips.removeListener(_refresh);
    trips.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: IndexedStack(
        index: _index,
        children: [_home(), _serviceValue(), _history(), _account()],
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
              icon: Icon(Icons.payments_outlined),
              selectedIcon: Icon(Icons.payments_rounded),
              label: 'Giá trị'),
          NavigationDestination(
              icon: Icon(Icons.history_rounded),
              selectedIcon: Icon(Icons.receipt_long_rounded),
              label: 'Lịch sử'),
          NavigationDestination(
              icon: Icon(Icons.person_outline_rounded),
              selectedIcon: Icon(Icons.person_rounded),
              label: 'Tài khoản'),
        ],
      ),
    );
  }

  Widget _home() {
    final mapTrip = trips.activeTrip ?? trips.offeredTrip;
    return Stack(
      children: [
        Positioned.fill(
          child: DriverMapView(
            online: trips.online,
            position: trips.lastPosition,
            trip: mapTrip,
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
              child: Row(children: [
                const _DriverBrand(),
                const Spacer(),
                _StatusPill(
                  text: trips.activeTrip != null
                      ? 'ĐANG LÀM VIỆC'
                      : trips.online
                          ? 'ONLINE'
                          : 'OFFLINE',
                  active: trips.activeTrip != null || trips.online,
                ),
              ]),
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
                    color: Color(0x280B132B),
                    blurRadius: 34,
                    offset: Offset(0, 16),
                  )
                ],
              ),
              child: SingleChildScrollView(
                padding: const EdgeInsets.fromLTRB(20, 10, 20, 20),
                child: trips.activeTrip != null
                    ? _activeJobPanel(trips.activeTrip!)
                    : trips.currentOffer != null && trips.offeredTrip != null
                        ? _offerPanel()
                        : _idlePanel(),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _idlePanel() {
    final theme = Theme.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Center(child: _handle()),
        const SizedBox(height: 14),
        Row(children: [
          Expanded(
            child:
                Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Text(trips.online ? 'Sẵn sàng nhận việc' : 'Bạn đang Offline',
                  style: theme.textTheme.headlineSmall),
              const SizedBox(height: 4),
              Text(
                trips.approved
                    ? 'Bật Online để nhận yêu cầu phù hợp quanh bạn.'
                    : 'Hồ sơ cần được FlashX duyệt trước khi nhận việc.',
                style: theme.textTheme.bodyMedium,
              ),
            ]),
          ),
          Switch.adaptive(
            value: trips.online,
            onChanged: trips.busy || !trips.approved ? null : trips.setOnline,
          ),
        ]),
        const SizedBox(height: 14),
        _InfoCard(
          icon: trips.approved
              ? Icons.verified_user_rounded
              : Icons.pending_actions_rounded,
          title:
              trips.approved ? 'Hồ sơ đã được duyệt' : 'Đang chờ duyệt hồ sơ',
          text: trips.approved
              ? 'FlashX chỉ gửi những công việc phù hợp với năng lực đã được phê duyệt.'
              : 'Admin sẽ kiểm tra giấy tờ và năng lực trước khi cho phép Online.',
        ),
        if (trips.error != null) ...[
          const SizedBox(height: 10),
          _ErrorCard(text: trips.error!),
        ],
        if (trips.online) ...[
          const SizedBox(height: 14),
          OutlinedButton.icon(
            onPressed: trips.busy ? null : trips.refreshOffer,
            icon: const Icon(Icons.refresh_rounded),
            label: const Text('Kiểm tra yêu cầu mới'),
          ),
        ],
      ],
    );
  }

  Widget _offerPanel() {
    final trip = trips.offeredTrip!;
    final offer = trips.currentOffer!;
    final vehicle = trips.offeredVehicle;
    final remaining = _remainingSeconds(offer['expires_at']?.toString());
    final service = trip['service_type']?.toString() ?? '';
    final fare = trip['estimated_fare_minor'] ?? 0;
    final distanceToPickup =
        (offer['distance_to_pickup_m'] as num?)?.toDouble() ?? 0;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Center(child: _handle()),
        const SizedBox(height: 12),
        Row(children: [
          _ServiceBadge(service: service),
          const SizedBox(width: 10),
          Expanded(
            child:
                Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Text(_serviceLabel(service),
                  style: Theme.of(context).textTheme.headlineSmall),
              Text('Yêu cầu mới · còn $remaining giây',
                  style: const TextStyle(color: FlashXTheme.textSecondary)),
            ]),
          ),
        ]),
        const SizedBox(height: 14),
        if (vehicle != null)
          _InfoCard(
            icon: vehicle['type'] == 'motorbike'
                ? Icons.two_wheeler_rounded
                : Icons.directions_car_rounded,
            title: _vehicleLabel(vehicle),
            text: vehicle['transmission'] == 'manual'
                ? 'Xe số sàn · xe của khách hàng'
                : vehicle['transmission'] == 'automatic'
                    ? 'Xe số tự động · xe của khách hàng'
                    : 'Xe của khách hàng',
          )
        else
          const _InfoCard(
            icon: Icons.key_rounded,
            title: 'Phương tiện của khách',
            text: 'Thông tin xe sẽ được xác nhận khi nhận việc.',
          ),
        const SizedBox(height: 10),
        Row(children: [
          Expanded(
              child: _Metric('Cách điểm nhận',
                  '${(distanceToPickup / 1000).toStringAsFixed(1)} km')),
          Expanded(child: _Metric('Giá trị dịch vụ', _money(fare))),
        ]),
        const SizedBox(height: 14),
        Row(children: [
          Expanded(
            child: OutlinedButton(
              onPressed: trips.busy ? null : trips.rejectOffer,
              child: const Text('Bỏ qua'),
            ),
          ),
          const SizedBox(width: 10),
          Expanded(
            flex: 2,
            child: FilledButton.icon(
              onPressed: trips.busy ? null : trips.acceptOffer,
              icon: const Icon(Icons.check_rounded),
              label: const Text('Nhận việc'),
            ),
          ),
        ]),
        if (trips.error != null) ...[
          const SizedBox(height: 10),
          _ErrorCard(text: trips.error!),
        ],
      ],
    );
  }

  Widget _activeJobPanel(Map<String, dynamic> trip) {
    final service = trip['service_type']?.toString() ?? '';
    final status = trip['status']?.toString() ?? '';
    final incidentOpen = trip['incident_open'] == true;
    final vehicle = trips.activeVehicle;
    final primary = _primaryAction();
    final custodyStage = trips.allowedActions.contains('vehicle_received')
        ? 'pickup'
        : trips.allowedActions.contains('handover')
            ? 'return'
            : null;
    final custodySnapshot =
        custodyStage == null ? null : trips.custodyFor(custodyStage);
    final driverCustodyConfirmed =
        custodyStage == null || trips.driverConfirmedCustody(custodyStage);
    final custodyReady =
        custodyStage == null || custodySnapshot?['ready'] == true;
    final allowDemoBypass =
        AppConfig.demoMode || trips.custodyStorageUnavailable;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Center(child: _handle()),
        const SizedBox(height: 12),
        Row(children: [
          _ServiceBadge(service: service),
          const SizedBox(width: 10),
          Expanded(
            child:
                Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Text(_serviceLabel(service),
                  style: Theme.of(context).textTheme.titleMedium),
              const SizedBox(height: 2),
              Text(_statusLabel(status),
                  style: Theme.of(context).textTheme.headlineSmall),
            ]),
          ),
        ]),
        const SizedBox(height: 8),
        Text(_statusHint(status),
            style: Theme.of(context).textTheme.bodyMedium),
        const SizedBox(height: 14),
        _JobProgress(
            status: status, inspection: service == 'vehicle_inspection_assist'),
        const SizedBox(height: 14),
        if (vehicle != null)
          _InfoCard(
            icon: vehicle['type'] == 'motorbike'
                ? Icons.two_wheeler_rounded
                : Icons.directions_car_rounded,
            title: _vehicleLabel(vehicle),
            text:
                'Đây là phương tiện của khách hàng. Kiểm tra trước khi xác nhận nhận xe.',
          ),
        if (incidentOpen) ...[
          const SizedBox(height: 10),
          const _InfoCard(
            icon: Icons.support_agent_rounded,
            title: 'Đang xử lý sự cố',
            text:
                'Operations đang tiếp quản. Không tiếp tục các bước tự động cho tới khi sự cố được xử lý.',
            warning: true,
          ),
        ],
        if (trip['inspection_result']?.toString().isNotEmpty == true) ...[
          const SizedBox(height: 10),
          _InfoCard(
            icon: Icons.fact_check_rounded,
            title:
                'Kết quả đăng kiểm: ${_inspectionResult(trip['inspection_result'].toString())}',
            text:
                'Tiếp tục đưa xe về và bàn giao cho khách dù kết quả đạt hay không đạt.',
          ),
        ],
        if (custodyStage != null && !incidentOpen) ...[
          const SizedBox(height: 12),
          _CustodyDriverCard(
            stage: custodyStage,
            snapshot: custodySnapshot,
          ),
          if (!driverCustodyConfirmed) ...[
            const SizedBox(height: 10),
            FilledButton.icon(
              onPressed: trips.busy
                  ? null
                  : () => _captureCustodyEvidence(custodyStage),
              icon: const Icon(Icons.add_a_photo_rounded),
              label: Text(custodyStage == 'pickup'
                  ? 'Ghi bằng chứng nhận xe'
                  : 'Ghi bằng chứng trả xe'),
            ),
          ] else if (!custodyReady && !trips.custodyStorageUnavailable) ...[
            const SizedBox(height: 10),
            const _InfoCard(
              icon: Icons.hourglass_top_rounded,
              title: 'Đang chờ khách xác nhận',
              text:
                  'Giữ nguyên bộ bằng chứng hiện tại. Khi khách xác nhận, bước tiếp theo sẽ tự mở.',
            ),
          ],
          if (trips.custodyStorageUnavailable) ...[
            const SizedBox(height: 10),
            const _InfoCard(
              icon: Icons.cloud_off_rounded,
              title: 'Kho ảnh demo chưa cấu hình',
              text:
                  'Ảnh chưa được lưu. Backend development không bật enforcement nên chỉ cho phép tiếp tục bằng nút có nhãn Demo; production vẫn bắt buộc evidence đầy đủ.',
              warning: true,
            ),
          ],
        ],
        if (primary != null &&
            !incidentOpen &&
            (custodyReady || allowDemoBypass)) ...[
          const SizedBox(height: 10),
          if (allowDemoBypass && custodyStage != null && !custodyReady)
            OutlinedButton.icon(
              onPressed: trips.busy ? null : primary.onPressed,
              icon: Icon(primary.icon),
              label: Text('${primary.label} · Demo'),
            )
          else
            FilledButton.icon(
              onPressed: trips.busy ? null : primary.onPressed,
              icon: Icon(primary.icon),
              label: Text(primary.label),
            ),
        ],
        if (trips.allowedActions.contains('report_incident') &&
            !incidentOpen) ...[
          const SizedBox(height: 8),
          OutlinedButton.icon(
            onPressed: trips.busy ? null : _reportIncident,
            icon: const Icon(Icons.warning_amber_rounded),
            label: const Text('Báo sự cố / cần hỗ trợ'),
          ),
        ],
        if (trips.error != null) ...[
          const SizedBox(height: 10),
          _ErrorCard(text: trips.error!),
        ],
      ],
    );
  }

  _Action? _primaryAction() {
    final actions = trips.allowedActions;
    if (actions.contains('arriving')) {
      return _Action(
          'Bắt đầu đến điểm nhận', Icons.navigation_rounded, trips.arriving);
    }
    if (actions.contains('arrived')) {
      return _Action(
          'Tôi đã đến điểm nhận', Icons.location_on_rounded, trips.arrived);
    }
    if (actions.contains('vehicle_received')) {
      return _Action(
          'Xác nhận đã nhận xe', Icons.key_rounded, trips.vehicleReceived);
    }
    if (actions.contains('start')) {
      return _Action(
          'Bắt đầu thực hiện', Icons.play_arrow_rounded, trips.startService);
    }
    if (actions.contains('arrive_inspection')) {
      return _Action('Đã tới nơi đăng kiểm', Icons.fact_check_outlined,
          trips.arriveInspection);
    }
    if (actions.contains('start_inspection')) {
      return _Action('Bắt đầu đăng kiểm', Icons.play_circle_outline_rounded,
          trips.startInspection);
    }
    if (actions.contains('complete_inspection')) {
      return _Action('Nhập kết quả đăng kiểm',
          Icons.assignment_turned_in_rounded, _completeInspection);
    }
    if (actions.contains('returning')) {
      return _Action(
          'Bắt đầu đưa xe về', Icons.route_rounded, trips.returningVehicle);
    }
    if (actions.contains('arrived_return')) {
      return _Action(
          'Đã tới điểm trả xe', Icons.location_on_rounded, trips.arrivedReturn);
    }
    if (actions.contains('handover')) {
      return _Action(
          'Xác nhận bàn giao xe', Icons.handshake_rounded, trips.handover);
    }
    if (actions.contains('complete')) {
      return _Action('Hoàn thành dịch vụ', Icons.check_circle_rounded,
          trips.completeService);
    }
    return null;
  }

  Future<void> _captureCustodyEvidence(String stage) async {
    final current = trips.custodyFor(stage);
    final evidence = Map<String, dynamic>.from(
      current?['evidence'] as Map? ?? const {},
    );
    final formKey = GlobalKey<FormState>();
    final note = TextEditingController(
        text: evidence['condition_note']?.toString() ?? '');
    final odometer =
        TextEditingController(text: evidence['odometer_km']?.toString() ?? '');
    final fuel =
        TextEditingController(text: evidence['fuel_percent']?.toString() ?? '');
    final battery = TextEditingController(
        text: evidence['battery_percent']?.toString() ?? '');

    final saved = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: Text(stage == 'pickup'
            ? 'Tình trạng khi nhận xe'
            : 'Tình trạng khi trả xe'),
        content: Form(
          key: formKey,
          child: SingleChildScrollView(
            child: Column(mainAxisSize: MainAxisSize.min, children: [
              TextFormField(
                controller: note,
                maxLines: 3,
                autofocus: true,
                decoration: const InputDecoration(
                  labelText: 'Ghi chú tình trạng xe *',
                  hintText:
                      'Ví dụ: Có vết xước cũ ở cản sau, ngoại thất khô ráo',
                ),
                validator: (value) => (value?.trim().length ?? 0) < 3
                    ? 'Mô tả tình trạng xe tối thiểu 3 ký tự'
                    : null,
              ),
              const SizedBox(height: 10),
              TextFormField(
                controller: odometer,
                keyboardType: TextInputType.number,
                decoration: const InputDecoration(
                    labelText: 'Odo (km) · không bắt buộc'),
                validator: (value) {
                  if (value == null || value.trim().isEmpty) return null;
                  final parsed = int.tryParse(value.trim());
                  return parsed == null || parsed < 0
                      ? 'Odo không hợp lệ'
                      : null;
                },
              ),
              const SizedBox(height: 10),
              Row(children: [
                Expanded(
                  child: TextFormField(
                    controller: fuel,
                    keyboardType: TextInputType.number,
                    decoration: const InputDecoration(labelText: 'Xăng %'),
                    validator: _percentValidator,
                  ),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: TextFormField(
                    controller: battery,
                    keyboardType: TextInputType.number,
                    decoration: const InputDecoration(labelText: 'Pin %'),
                    validator: _percentValidator,
                  ),
                ),
              ]),
              const SizedBox(height: 12),
              const _InfoCard(
                icon: Icons.verified_user_outlined,
                title: 'Xác nhận gắn với bộ evidence hiện tại',
                text:
                    'Nếu sửa tình trạng hoặc thêm ảnh sau khi xác nhận, hệ thống sẽ yêu cầu hai bên xác nhận lại.',
              ),
            ]),
          ),
        ),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(dialogContext, false),
              child: const Text('Đóng')),
          FilledButton(
            onPressed: () {
              if (formKey.currentState?.validate() != true) return;
              Navigator.pop(dialogContext, true);
            },
            child: const Text('Lưu & chụp ảnh'),
          ),
        ],
      ),
    );

    final conditionNote = note.text.trim();
    final odometerKm = _optionalInt(odometer.text);
    final fuelPercent = _optionalInt(fuel.text);
    final batteryPercent = _optionalInt(battery.text);
    note.dispose();
    odometer.dispose();
    fuel.dispose();
    battery.dispose();
    if (saved != true || !mounted) return;

    final metadataSaved = await trips.saveCustodyMetadata(
      stage: stage,
      conditionNote: conditionNote,
      odometerKm: odometerKm,
      fuelPercent: fuelPercent,
      batteryPercent: batteryPercent,
    );
    if (!metadataSaved || !mounted) {
      _showMessage(trips.error ?? 'Không thể lưu tình trạng xe.');
      return;
    }

    var snapshot = trips.custodyFor(stage);
    final existingTypes = <String>{};
    final existingPhotos = snapshot?['photos'];
    if (existingPhotos is List) {
      for (final item in existingPhotos) {
        if (item is Map && item['photo'] is Map) {
          final type = (item['photo'] as Map)['photo_type']?.toString();
          if (type != null && type.isNotEmpty) existingTypes.add(type);
        }
      }
    }
    var photoCount = trips.custodyPhotoCount(stage);
    const preferredPhotos = <(String, String)>[
      ('front', 'mặt trước xe'),
      ('rear', 'mặt sau xe'),
      ('left', 'bên trái xe'),
      ('right', 'bên phải xe'),
    ];
    for (final photo in preferredPhotos) {
      if (photoCount >= 2) break;
      if (existingTypes.contains(photo.$1)) continue;
      final picked = await _pickEvidenceImage(photo.$2);
      if (picked == null || !mounted) break;
      final uploaded = await trips.uploadCustodyPhoto(
        stage: stage,
        file: File(picked.path),
        photoType: photo.$1,
        contentType: picked.mimeType,
      );
      if (!uploaded || !mounted) {
        _showMessage(trips.error ?? 'Không thể tải ảnh bằng chứng.');
        return;
      }
      existingTypes.add(photo.$1);
      photoCount = trips.custodyPhotoCount(stage);
    }

    snapshot = trips.custodyFor(stage);
    if (trips.custodyPhotoCount(stage) < 2) {
      if (mounted) {
        _showMessage(
            'Đã lưu tình trạng. Cần tối thiểu 2 ảnh trước khi tài xế xác nhận.');
      }
      return;
    }
    final confirmed = await trips.confirmCustody(stage);
    if (!mounted) return;
    if (confirmed) {
      _showMessage(stage == 'pickup'
          ? 'Đã xác nhận bằng chứng nhận xe. Chờ khách xác nhận.'
          : 'Đã xác nhận bằng chứng trả xe. Chờ khách xác nhận.');
    } else {
      _showMessage(trips.error ?? 'Chưa thể xác nhận bằng chứng.');
    }
  }

  Future<XFile?> _pickEvidenceImage(String label) async {
    final source = await showModalBottomSheet<ImageSource>(
      context: context,
      showDragHandle: true,
      builder: (sheetContext) => SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(16, 6, 16, 18),
          child: Column(mainAxisSize: MainAxisSize.min, children: [
            Text('Ảnh $label', style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 12),
            ListTile(
              leading: const Icon(Icons.photo_camera_rounded),
              title: const Text('Chụp ảnh mới'),
              subtitle: const Text('Khuyến nghị khi đang nhận/bàn giao xe'),
              onTap: () => Navigator.pop(sheetContext, ImageSource.camera),
            ),
            ListTile(
              leading: const Icon(Icons.photo_library_outlined),
              title: const Text('Chọn từ thư viện'),
              subtitle: const Text('Hữu ích khi demo trên thiết bị giả lập'),
              onTap: () => Navigator.pop(sheetContext, ImageSource.gallery),
            ),
          ]),
        ),
      ),
    );
    if (source == null || !mounted) return null;
    return ImagePicker().pickImage(
      source: source,
      imageQuality: 84,
      maxWidth: 1800,
    );
  }

  static String? _percentValidator(String? value) {
    if (value == null || value.trim().isEmpty) return null;
    final parsed = int.tryParse(value.trim());
    return parsed == null || parsed < 0 || parsed > 100 ? '0–100' : null;
  }

  static int? _optionalInt(String value) {
    final trimmed = value.trim();
    return trimmed.isEmpty ? null : int.tryParse(trimmed);
  }

  void _showMessage(String message) {
    if (!mounted) return;
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(content: Text(message)));
  }

  Future<bool> _completeInspection() async {
    var result = 'passed';
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => StatefulBuilder(
        builder: (context, setDialogState) => AlertDialog(
          title: const Text('Kết quả đăng kiểm'),
          content: DropdownButtonFormField<String>(
            initialValue: result,
            decoration: const InputDecoration(labelText: 'Kết quả'),
            items: const [
              DropdownMenuItem(value: 'passed', child: Text('Đạt')),
              DropdownMenuItem(value: 'failed', child: Text('Không đạt')),
              DropdownMenuItem(
                  value: 'deferred', child: Text('Cần thực hiện lại')),
              DropdownMenuItem(
                  value: 'unavailable', child: Text('Chưa có kết quả')),
            ],
            onChanged: (value) =>
                setDialogState(() => result = value ?? result),
          ),
          actions: [
            TextButton(
                onPressed: () => Navigator.pop(dialogContext, false),
                child: const Text('Đóng')),
            FilledButton(
                onPressed: () => Navigator.pop(dialogContext, true),
                child: const Text('Lưu kết quả')),
          ],
        ),
      ),
    );
    if (confirmed != true) return false;
    return trips.completeInspection(result);
  }

  Future<void> _reportIncident() async {
    var type = 'vehicle_issue';
    final note = TextEditingController();
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => StatefulBuilder(
        builder: (context, setDialogState) => AlertDialog(
          title: const Text('Báo sự cố'),
          content: Column(mainAxisSize: MainAxisSize.min, children: [
            DropdownButtonFormField<String>(
              initialValue: type,
              decoration: const InputDecoration(labelText: 'Loại sự cố'),
              items: const [
                DropdownMenuItem(
                    value: 'vehicle_issue', child: Text('Xe gặp vấn đề')),
                DropdownMenuItem(
                    value: 'accident', child: Text('Va chạm / tai nạn')),
                DropdownMenuItem(
                    value: 'document_issue', child: Text('Giấy tờ có vấn đề')),
                DropdownMenuItem(
                    value: 'customer_unreachable',
                    child: Text('Không liên lạc được khách')),
                DropdownMenuItem(value: 'other', child: Text('Khác')),
              ],
              onChanged: (value) => setDialogState(() => type = value ?? type),
            ),
            const SizedBox(height: 10),
            TextField(
                controller: note,
                maxLines: 3,
                decoration: const InputDecoration(labelText: 'Mô tả ngắn')),
          ]),
          actions: [
            TextButton(
                onPressed: () => Navigator.pop(dialogContext, false),
                child: const Text('Đóng')),
            FilledButton(
                onPressed: () => Navigator.pop(dialogContext, true),
                child: const Text('Gửi Operations')),
          ],
        ),
      ),
    );
    if (confirmed == true) await trips.reportIncident(type, note.text);
    note.dispose();
  }

  Widget _serviceValue() {
    return SafeArea(
      child: RefreshIndicator(
        onRefresh: trips.loadHistory,
        child: ListView(
          padding: const EdgeInsets.fromLTRB(20, 24, 20, 32),
          children: [
            const _PageHeader(
                title: 'Giá trị dịch vụ', subtitle: 'Tổng hợp bản demo'),
            const SizedBox(height: 20),
            Container(
              padding: const EdgeInsets.all(22),
              decoration: BoxDecoration(
                  color: FlashXTheme.navy,
                  borderRadius: BorderRadius.circular(26)),
              child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text('Giá trị dịch vụ đã hoàn thành',
                        style: TextStyle(color: Color(0xFFB9C1D5))),
                    const SizedBox(height: 7),
                    Text(_money(trips.completedServiceValueMinor),
                        style: const TextStyle(
                            color: Colors.white,
                            fontSize: 32,
                            fontWeight: FontWeight.w900)),
                    const SizedBox(height: 8),
                    Text('${trips.completedTrips} công việc hoàn thành',
                        style: const TextStyle(
                            color: FlashXTheme.yellow,
                            fontWeight: FontWeight.w700)),
                  ]),
            ),
            const SizedBox(height: 14),
            const _InfoCard(
              icon: Icons.info_outline_rounded,
              title: 'Không phải thu nhập ròng',
              text:
                  'Giá khách trả, phần tài xế và phí FlashX được tách riêng. Bản demo chưa chốt chính sách đối soát cuối cùng.',
            ),
          ],
        ),
      ),
    );
  }

  Widget _history() {
    return SafeArea(
      child: RefreshIndicator(
        onRefresh: trips.loadHistory,
        child: ListView(
          padding: const EdgeInsets.fromLTRB(20, 24, 20, 32),
          children: [
            const _PageHeader(
                title: 'Lịch sử', subtitle: 'Các công việc gần đây'),
            const SizedBox(height: 18),
            if (trips.history.isEmpty)
              const _InfoCard(
                  icon: Icons.history_rounded,
                  title: 'Chưa có công việc',
                  text: 'Các việc đã nhận sẽ xuất hiện ở đây.'),
            ...trips.history.map((trip) => Card(
                  margin: const EdgeInsets.only(bottom: 10),
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Row(children: [
                      _ServiceBadge(
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
                            Text(_statusLabel(trip['status']?.toString() ?? ''),
                                style: Theme.of(context).textTheme.bodySmall),
                          ])),
                      Text(
                          _money(trip['final_fare_minor'] ??
                              trip['estimated_fare_minor'] ??
                              0),
                          style: const TextStyle(fontWeight: FontWeight.w900)),
                    ]),
                  ),
                )),
          ],
        ),
      ),
    );
  }

  Widget _account() {
    final profile = widget.auth.profile ?? const <String, dynamic>{};
    final capabilities =
        (profile['capabilities'] as List?)?.map((e) => e.toString()).toList() ??
            const <String>[];
    return SafeArea(
      child: ListView(
        padding: const EdgeInsets.fromLTRB(20, 24, 20, 32),
        children: [
          const _PageHeader(
              title: 'Tài khoản tài xế', subtitle: 'Hồ sơ & năng lực'),
          const SizedBox(height: 18),
          Container(
            padding: const EdgeInsets.all(18),
            decoration: BoxDecoration(
                color: FlashXTheme.navy,
                borderRadius: BorderRadius.circular(24)),
            child: Row(children: [
              const CircleAvatar(
                  radius: 27,
                  backgroundColor: FlashXTheme.yellow,
                  foregroundColor: FlashXTheme.navy,
                  child: Icon(Icons.person_rounded)),
              const SizedBox(width: 14),
              Expanded(
                  child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                    Text(
                        profile['full_name']?.toString().isNotEmpty == true
                            ? profile['full_name'].toString()
                            : 'Đối tác FlashX',
                        style: const TextStyle(
                            color: Colors.white,
                            fontSize: 18,
                            fontWeight: FontWeight.w800)),
                    const SizedBox(height: 3),
                    Text(
                        trips.approved
                            ? 'Đã xác minh & được phép nhận việc'
                            : 'Đang chờ phê duyệt',
                        style: TextStyle(
                            color: trips.approved
                                ? FlashXTheme.lime
                                : FlashXTheme.yellow)),
                  ])),
            ]),
          ),
          const SizedBox(height: 20),
          Text('Năng lực dịch vụ',
              style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: 10),
          if (capabilities.isEmpty)
            const _InfoCard(
                icon: Icons.rule_rounded,
                title: 'Chưa có năng lực',
                text: 'Admin sẽ cấu hình năng lực sau khi duyệt hồ sơ.')
          else
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: capabilities
                  .map((capability) => Chip(
                        avatar: const Icon(Icons.check_circle_rounded,
                            size: 18, color: FlashXTheme.success),
                        label: Text(_serviceLabel(capability)),
                      ))
                  .toList(),
            ),
          const SizedBox(height: 20),
          const ListTile(
            contentPadding: EdgeInsets.zero,
            leading: Icon(Icons.badge_outlined),
            title: Text('Hồ sơ & giấy phép lái xe'),
            subtitle: Text('Được quản lý và duyệt bởi FlashX Operations'),
          ),
          const ListTile(
            contentPadding: EdgeInsets.zero,
            leading: Icon(Icons.support_agent_rounded),
            title: Text('Hỗ trợ tài xế'),
            subtitle: Text('Báo sự cố trực tiếp từ công việc đang thực hiện'),
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

  static int _remainingSeconds(String? raw) {
    final expires = raw == null ? null : DateTime.tryParse(raw);
    if (expires == null) return 0;
    return expires
        .difference(DateTime.now().toUtc())
        .inSeconds
        .clamp(0, 999)
        .toInt();
  }

  static String _serviceLabel(String service) => switch (service) {
        'car' || 'designated_driver_car' => 'Lái hộ ô tô',
        'bike' || 'designated_driver_bike' => 'Lái hộ xe máy',
        'vehicle_inspection_assist' => 'Đăng kiểm hộ',
        _ => 'Dịch vụ FlashX',
      };

  static String _statusLabel(String status) => switch (status) {
        'scheduled' => 'Đã hẹn lịch',
        'searching' => 'Đang tìm tài xế',
        'accepted' => 'Đã nhận việc',
        'arriving' || 'arriving_for_pickup' => 'Đang đến điểm nhận',
        'arrived' || 'arrived_for_pickup' => 'Đã tới điểm nhận',
        'vehicle_received' => 'Đã nhận xe khách',
        'in_progress' => 'Đang lái xe',
        'en_route_to_inspection' => 'Đang tới nơi đăng kiểm',
        'arrived_at_inspection_center' => 'Đã tới nơi đăng kiểm',
        'inspection_in_progress' => 'Đang đăng kiểm',
        'inspection_completed' => 'Đã có kết quả',
        'returning_vehicle' => 'Đang đưa xe về',
        'arrived_for_return' => 'Đã tới điểm trả',
        'handover' => 'Đang bàn giao xe',
        'completed' => 'Đã hoàn thành',
        'cancelled' => 'Đã hủy',
        _ => 'Đang cập nhật',
      };

  static String _statusHint(String status) => switch (status) {
        'accepted' => 'Kiểm tra thông tin rồi bắt đầu di chuyển tới khách.',
        'arrived' ||
        'arrived_for_pickup' =>
          'Xác nhận tình trạng phương tiện trước khi nhận xe.',
        'vehicle_received' =>
          'Bạn đang chịu trách nhiệm với xe của khách. Không tự ý chuyển việc cho người khác.',
        'in_progress' => 'Lái xe an toàn tới điểm khách yêu cầu.',
        'inspection_in_progress' =>
          'Cập nhật kết quả sau khi hoàn tất thủ tục đăng kiểm.',
        'handover' => 'Bàn giao xe cho đúng người và xác nhận hoàn thành.',
        _ =>
          'Thực hiện đúng bước hiển thị để FlashX và khách theo dõi được tiến trình.',
      };

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

  Widget _handle() => Container(
        width: 42,
        height: 4,
        decoration: BoxDecoration(
            color: const Color(0xFFD7DAE0),
            borderRadius: BorderRadius.circular(99)),
      );
}

class _Action {
  const _Action(this.label, this.icon, this.onPressed);
  final String label;
  final IconData icon;
  final Future<bool> Function() onPressed;
}

class _DriverBrand extends StatelessWidget {
  const _DriverBrand();
  @override
  Widget build(BuildContext context) => Container(
        padding: const EdgeInsets.fromLTRB(9, 7, 13, 7),
        decoration: BoxDecoration(
            color: FlashXTheme.navy, borderRadius: BorderRadius.circular(18)),
        child: const Row(mainAxisSize: MainAxisSize.min, children: [
          Icon(Icons.bolt_rounded, color: FlashXTheme.yellow),
          SizedBox(width: 4),
          Text('FlashX Driver',
              style:
                  TextStyle(color: Colors.white, fontWeight: FontWeight.w900)),
        ]),
      );
}

class _StatusPill extends StatelessWidget {
  const _StatusPill({required this.text, required this.active});
  final String text;
  final bool active;
  @override
  Widget build(BuildContext context) => Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 9),
        decoration: BoxDecoration(
            color: Colors.white, borderRadius: BorderRadius.circular(18)),
        child: Row(mainAxisSize: MainAxisSize.min, children: [
          Container(
              width: 8,
              height: 8,
              decoration: BoxDecoration(
                  color: active ? FlashXTheme.success : Colors.grey,
                  shape: BoxShape.circle)),
          const SizedBox(width: 7),
          Text(text,
              style: const TextStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.w900,
                  color: FlashXTheme.navy)),
        ]),
      );
}

class _ServiceBadge extends StatelessWidget {
  const _ServiceBadge({required this.service});
  final String service;
  @override
  Widget build(BuildContext context) => Container(
        width: 48,
        height: 48,
        decoration: BoxDecoration(
            color: FlashXTheme.yellow, borderRadius: BorderRadius.circular(16)),
        child: Icon(
            service.contains('inspection')
                ? Icons.fact_check_rounded
                : service.contains('bike')
                    ? Icons.two_wheeler_rounded
                    : Icons.directions_car_filled_rounded,
            color: FlashXTheme.navy),
      );
}

class _CustodyDriverCard extends StatelessWidget {
  const _CustodyDriverCard({required this.stage, required this.snapshot});

  final String stage;
  final Map<String, dynamic>? snapshot;

  @override
  Widget build(BuildContext context) {
    final evidence = Map<String, dynamic>.from(
      snapshot?['evidence'] as Map? ?? const {},
    );
    final photos = snapshot?['photos'] as List? ?? const [];
    final driverConfirmed = evidence['driver_confirmed_at'] != null;
    final riderConfirmed = evidence['rider_confirmed_at'] != null;
    final ready = snapshot?['ready'] == true;
    final note = evidence['condition_note']?.toString().trim() ?? '';
    final metrics = <String>[
      if (evidence['odometer_km'] != null) '${evidence['odometer_km']} km',
      if (evidence['fuel_percent'] != null) 'Xăng ${evidence['fuel_percent']}%',
      if (evidence['battery_percent'] != null)
        'Pin ${evidence['battery_percent']}%',
    ];

    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: ready ? const Color(0xFFECFDF3) : const Color(0xFFF7F8FA),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(
          color: ready ? const Color(0xFFABEFC6) : FlashXTheme.border,
        ),
      ),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Row(children: [
          Container(
            width: 38,
            height: 38,
            decoration: BoxDecoration(
              color: ready ? const Color(0xFFD1FADF) : const Color(0xFFEFF1F5),
              borderRadius: BorderRadius.circular(13),
            ),
            child: Icon(
              stage == 'pickup' ? Icons.key_rounded : Icons.handshake_rounded,
              color: ready ? FlashXTheme.success : FlashXTheme.navy,
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
                    : driverConfirmed
                        ? 'Tài xế đã xác nhận · chờ khách'
                        : 'Cần tối thiểu 2 ảnh và xác nhận tài xế',
                style: TextStyle(
                  color:
                      ready ? FlashXTheme.success : FlashXTheme.textSecondary,
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ]),
          ),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 9, vertical: 5),
            decoration: BoxDecoration(
              color: photos.length >= 2
                  ? const Color(0xFFD1FADF)
                  : const Color(0xFFFFF3DF),
              borderRadius: BorderRadius.circular(99),
            ),
            child: Text(
              '${photos.length}/2 ảnh',
              style: TextStyle(
                color: photos.length >= 2
                    ? FlashXTheme.success
                    : FlashXTheme.warning,
                fontSize: 11,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ]),
        if (note.isNotEmpty) ...[
          const SizedBox(height: 12),
          Text(note, style: Theme.of(context).textTheme.bodySmall),
        ],
        if (metrics.isNotEmpty) ...[
          const SizedBox(height: 10),
          Wrap(
            spacing: 7,
            runSpacing: 7,
            children: metrics
                .map((metric) => Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 9, vertical: 5),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(99),
                        border: Border.all(color: FlashXTheme.border),
                      ),
                      child: Text(metric,
                          style: const TextStyle(
                              fontSize: 11, fontWeight: FontWeight.w700)),
                    ))
                .toList(),
          ),
        ],
        const SizedBox(height: 12),
        Row(children: [
          Expanded(
            child: _CustodyConfirmLine(
              label: 'Tài xế',
              confirmed: driverConfirmed,
            ),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: _CustodyConfirmLine(
              label: 'Khách',
              confirmed: riderConfirmed,
            ),
          ),
        ]),
      ]),
    );
  }
}

class _CustodyConfirmLine extends StatelessWidget {
  const _CustodyConfirmLine({required this.label, required this.confirmed});

  final String label;
  final bool confirmed;

  @override
  Widget build(BuildContext context) => Container(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 9),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: FlashXTheme.border),
        ),
        child: Row(children: [
          Icon(
            confirmed
                ? Icons.check_circle_rounded
                : Icons.radio_button_unchecked_rounded,
            size: 17,
            color: confirmed ? FlashXTheme.success : FlashXTheme.textSecondary,
          ),
          const SizedBox(width: 6),
          Expanded(
            child: Text(
              '$label · ${confirmed ? 'đã xác nhận' : 'chưa xác nhận'}',
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(fontSize: 10, fontWeight: FontWeight.w700),
            ),
          ),
        ]),
      );
}

class _InfoCard extends StatelessWidget {
  const _InfoCard(
      {required this.icon,
      required this.title,
      required this.text,
      this.warning = false});
  final IconData icon;
  final String title;
  final String text;
  final bool warning;
  @override
  Widget build(BuildContext context) => Container(
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          color: warning ? const Color(0xFFFFF3DF) : const Color(0xFFF5F7FA),
          borderRadius: BorderRadius.circular(18),
        ),
        child: Row(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Icon(icon, color: warning ? FlashXTheme.warning : FlashXTheme.navy),
          const SizedBox(width: 10),
          Expanded(
              child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                Text(title,
                    style: const TextStyle(fontWeight: FontWeight.w800)),
                const SizedBox(height: 3),
                Text(text, style: Theme.of(context).textTheme.bodySmall),
              ])),
        ]),
      );
}

class _Metric extends StatelessWidget {
  const _Metric(this.label, this.value);
  final String label;
  final String value;
  @override
  Widget build(BuildContext context) => Container(
        margin: const EdgeInsets.symmetric(horizontal: 4),
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
            color: const Color(0xFFF7F8FA),
            borderRadius: BorderRadius.circular(16)),
        child: Column(children: [
          Text(value,
              style: const TextStyle(
                  fontWeight: FontWeight.w900, color: FlashXTheme.navy)),
          const SizedBox(height: 3),
          Text(label,
              textAlign: TextAlign.center,
              style: Theme.of(context).textTheme.bodySmall),
        ]),
      );
}

class _ErrorCard extends StatelessWidget {
  const _ErrorCard({required this.text});
  final String text;
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
              child: Text(text,
                  style: const TextStyle(color: FlashXTheme.danger))),
        ]),
      );
}

class _JobProgress extends StatelessWidget {
  const _JobProgress({required this.status, required this.inspection});
  final String status;
  final bool inspection;
  @override
  Widget build(BuildContext context) {
    final stages = inspection
        ? const [
            'accepted',
            'arrived_for_pickup',
            'vehicle_received',
            'inspection_in_progress',
            'returning_vehicle',
            'handover',
            'completed'
          ]
        : const [
            'accepted',
            'arrived',
            'vehicle_received',
            'in_progress',
            'handover',
            'completed'
          ];
    final aliases = <String, String>{
      'arriving': 'accepted',
      'arriving_for_pickup': 'accepted',
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
                )));
  }
}

class _PageHeader extends StatelessWidget {
  const _PageHeader({required this.title, required this.subtitle});
  final String title;
  final String subtitle;
  @override
  Widget build(BuildContext context) => Row(children: [
        const _DriverBrand(),
        const SizedBox(width: 14),
        Expanded(
            child:
                Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text(title, style: Theme.of(context).textTheme.headlineSmall),
          Text(subtitle, style: Theme.of(context).textTheme.bodySmall),
        ])),
      ]);
}
