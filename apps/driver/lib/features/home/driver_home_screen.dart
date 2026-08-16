import 'dart:async';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';

import '../../core/config/app_config.dart';
import '../../core/theme/flashx_theme.dart';
import '../auth/auth_controller.dart';
import '../documents/driver_document_client.dart';
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
  late final DriverDocumentClient documents;
  Timer? _ticker;
  var _index = 0;
  List<Map<String, dynamic>> _documents = const [];
  bool _documentsBusy = false;
  String? _documentsError;

  @override
  void initState() {
    super.initState();
    trips = DriverTripController(auth: widget.auth)..addListener(_refresh);
    documents = DriverDocumentClient(api: widget.auth.api);
    trips.initialize();
    unawaited(_loadDocuments());
    _ticker = Timer.periodic(const Duration(seconds: 1), (_) {
      final status = trips.activeTrip?['status']?.toString() ?? '';
      if (mounted &&
          (trips.currentOffer != null ||
              status == 'arrived' ||
              status == 'arrived_for_pickup')) {
        setState(() {});
      }
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
    documents.close();
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
            liveRoutes: trips.liveRoutes,
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
    final inspectionPickupReady = service != 'vehicle_inspection_assist' ||
        custodyStage != 'pickup' ||
        trips.inspectionChecklist?['ready'] == true;
    final allowDemoBypass =
        AppConfig.demoMode || trips.custodyStorageUnavailable;
    final noShowRemaining = _remainingNoShowSeconds(trip);
    final payment = trips.activePayment;
    final paymentStatus = switch (payment?['status']?.toString()) {
      'paid' => 'Đã thu từ khách',
      'cancelled' => 'Đã hủy',
      'failed' => 'Có lỗi',
      _ => 'Chờ thu khi hoàn tất',
    };
    final paymentAmount = payment?['amount_minor'] ??
        trip['final_fare_minor'] ??
        trip['estimated_fare_minor'] ??
        0;
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
        if (trips.liveRoutes != null) ...[
          const SizedBox(height: 12),
          _DriverLiveRouteCard(snapshot: trips.liveRoutes!),
        ],
        const SizedBox(height: 12),
        _InfoCard(
          icon: Icons.payments_outlined,
          title: 'Khách thanh toán ${_money(paymentAmount)}',
          text:
              'Tiền mặt · $paymentStatus. Đây là giá trị khách thanh toán, không phải thu nhập tài xế.',
        ),
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
        if (!incidentOpen &&
            (status == 'arrived' || status == 'arrived_for_pickup')) ...[
          const SizedBox(height: 10),
          _InfoCard(
            icon: noShowRemaining > 0
                ? Icons.hourglass_top_rounded
                : Icons.person_off_outlined,
            title: noShowRemaining > 0
                ? 'Đang trong thời gian chờ miễn phí'
                : 'Có thể báo khách không xuất hiện',
            text: noShowRemaining > 0
                ? 'Còn ${_durationCountdown(noShowRemaining)} trước khi hệ thống mở thao tác no-show. Không tự ý rời điểm nhận sớm.'
                : 'Chỉ sử dụng khi đã cố gắng liên hệ nhưng khách vẫn không có mặt tại điểm nhận.',
            warning: noShowRemaining <= 0,
          ),
          if (noShowRemaining <= 0) ...[
            const SizedBox(height: 8),
            OutlinedButton.icon(
              onPressed: trips.busy ? null : _confirmCustomerNoShow,
              icon: const Icon(Icons.person_off_outlined),
              label: const Text('Khách không xuất hiện'),
            ),
          ],
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
        if (service == 'vehicle_inspection_assist' &&
            trips.inspectionChecklist != null) ...[
          const SizedBox(height: 12),
          _DriverInspectionChecklistCard(
            snapshot: trips.inspectionChecklist!,
            busy: trips.busy,
            editable: status == 'arrived_for_pickup',
            onEdit: _editInspectionChecklistItem,
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
            inspectionPickupReady &&
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

  Future<void> _editInspectionChecklistItem(Map<String, dynamic> item) async {
    var status = item['driver_status']?.toString() ?? 'pending';
    if (status == 'pending') status = 'received';
    final required = item['required'] == true;
    final note =
        TextEditingController(text: item['driver_note']?.toString() ?? '');
    final saved = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => StatefulBuilder(
        builder: (context, setDialogState) => AlertDialog(
          title: Text(item['label']?.toString() ?? 'Giấy tờ đăng kiểm'),
          content: SingleChildScrollView(
            child: Column(mainAxisSize: MainAxisSize.min, children: [
              Text(
                'Khách khai báo: ${_customerDocumentStatus(item['customer_status']?.toString() ?? 'pending')}',
                style: Theme.of(context).textTheme.bodyMedium,
              ),
              const SizedBox(height: 12),
              DropdownButtonFormField<String>(
                initialValue: status,
                decoration:
                    const InputDecoration(labelText: 'Đối chiếu thực tế'),
                items: [
                  const DropdownMenuItem(
                      value: 'received', child: Text('Đã nhận giấy tờ')),
                  const DropdownMenuItem(
                      value: 'missing', child: Text('Không nhận được / thiếu')),
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
                  hintText: 'Ví dụ: Đã nhận bản gốc trong túi hồ sơ xanh',
                ),
              ),
              if (required) ...[
                const SizedBox(height: 8),
                const _InfoCard(
                  icon: Icons.lock_outline_rounded,
                  title: 'Mục bắt buộc',
                  text:
                      'Phải xác nhận đã nhận thực tế trước khi hệ thống mở bước Nhận xe.',
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
                child: const Text('Lưu đối chiếu')),
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
    _showMessage(ok
        ? 'Đã cập nhật đối chiếu giấy tờ.'
        : trips.error ?? 'Không thể cập nhật giấy tờ.');
  }

  static String _customerDocumentStatus(String status) => switch (status) {
        'present' => 'Sẽ bàn giao',
        'not_available' => 'Không có',
        'not_applicable' => 'Không áp dụng',
        _ => 'Chưa khai báo',
      };

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

  Future<void> _confirmCustomerNoShow() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Xác nhận khách không xuất hiện'),
        content: const Text(
          'Chỉ xác nhận sau khi đã chờ đủ thời gian miễn phí và cố gắng liên hệ khách. Công việc sẽ được hủy với lý do customer_no_show và bạn trở lại trạng thái sẵn sàng.',
        ),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(dialogContext, false),
              child: const Text('Tiếp tục chờ')),
          FilledButton(
              onPressed: () => Navigator.pop(dialogContext, true),
              child: const Text('Xác nhận no-show')),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;
    final ok = await trips.customerNoShow();
    if (!mounted) return;
    _showMessage(ok
        ? 'Đã đóng công việc do khách không xuất hiện.'
        : trips.error ?? 'Chưa thể xác nhận no-show.');
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

  Future<void> _loadDocuments() async {
    if (!mounted) return;
    setState(() {
      _documentsBusy = true;
      _documentsError = null;
    });
    try {
      final items = await documents.list();
      if (!mounted) return;
      setState(() => _documents = items);
    } catch (error) {
      if (!mounted) return;
      setState(() => _documentsError = error.toString());
    } finally {
      if (mounted) setState(() => _documentsBusy = false);
    }
  }

  Map<String, dynamic>? _documentFor(String type) {
    for (final document in _documents.reversed) {
      if (document['document_type']?.toString() == type) return document;
    }
    return null;
  }

  Future<void> _uploadDriverDocument(String type, String label) async {
    final source = await showModalBottomSheet<ImageSource>(
      context: context,
      showDragHandle: true,
      builder: (sheetContext) => SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(16, 6, 16, 18),
          child: Column(mainAxisSize: MainAxisSize.min, children: [
            Text(label, style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 12),
            ListTile(
              leading: const Icon(Icons.photo_camera_rounded),
              title: const Text('Chụp ảnh mới'),
              subtitle: const Text('Đảm bảo giấy tờ rõ nét, đủ bốn góc'),
              onTap: () => Navigator.pop(sheetContext, ImageSource.camera),
            ),
            ListTile(
              leading: const Icon(Icons.photo_library_outlined),
              title: const Text('Chọn từ thư viện'),
              onTap: () => Navigator.pop(sheetContext, ImageSource.gallery),
            ),
          ]),
        ),
      ),
    );
    if (source == null || !mounted) return;
    final picked = await ImagePicker().pickImage(
      source: source,
      imageQuality: 88,
      maxWidth: 2200,
    );
    if (picked == null || !mounted) return;
    final contentType = picked.mimeType ?? 'image/jpeg';
    if (!const {'image/jpeg', 'image/png', 'image/webp'}
        .contains(contentType)) {
      _showMessage(
          'FlashX hiện hỗ trợ ảnh JPG, PNG hoặc WebP cho hồ sơ tài xế.');
      return;
    }
    setState(() {
      _documentsBusy = true;
      _documentsError = null;
    });
    try {
      await documents.uploadDocument(
        file: File(picked.path),
        documentType: type,
        contentType: contentType,
      );
      _documents = await documents.list();
      if (!mounted) return;
      _showMessage('Đã tải $label. Hồ sơ đang chờ Operations duyệt.');
    } catch (error) {
      if (!mounted) return;
      setState(() => _documentsError = error.toString());
      _showMessage('Không thể tải $label.');
    } finally {
      if (mounted) setState(() => _documentsBusy = false);
    }
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
          Row(children: [
            Expanded(
              child: Text('Hồ sơ KYC',
                  style: Theme.of(context).textTheme.titleLarge),
            ),
            IconButton(
              onPressed: _documentsBusy ? null : _loadDocuments,
              tooltip: 'Làm mới hồ sơ',
              icon: _documentsBusy
                  ? const SizedBox.square(
                      dimension: 18,
                      child: CircularProgressIndicator(strokeWidth: 2))
                  : const Icon(Icons.refresh_rounded),
            ),
          ]),
          const SizedBox(height: 6),
          Text(
            'Tải trực tiếp lên kho bảo mật của FlashX. Operations sẽ duyệt từng tài liệu trước khi tài xế được phép nhận việc.',
            style: Theme.of(context).textTheme.bodySmall,
          ),
          if (_documentsError != null) ...[
            const SizedBox(height: 10),
            _ErrorCard(text: _documentsError!),
          ],
          const SizedBox(height: 12),
          ...const [
            ('identity_front', 'CCCD · mặt trước', Icons.badge_outlined),
            ('identity_back', 'CCCD · mặt sau', Icons.badge_outlined),
            ('driver_license', 'Giấy phép lái xe', Icons.drive_eta_outlined),
            ('portrait', 'Ảnh chân dung', Icons.account_circle_outlined),
          ].map((requirement) {
            final document = _documentFor(requirement.$1);
            return Padding(
              padding: const EdgeInsets.only(bottom: 9),
              child: _DriverDocumentCard(
                icon: requirement.$3,
                label: requirement.$2,
                document: document,
                busy: _documentsBusy,
                onUpload: () =>
                    _uploadDriverDocument(requirement.$1, requirement.$2),
              ),
            );
          }),
          const SizedBox(height: 10),
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

  int _remainingNoShowSeconds(Map<String, dynamic> trip) {
    final raw = trip['arrived_at']?.toString();
    final arrived = raw == null ? null : DateTime.tryParse(raw)?.toUtc();
    if (arrived == null) return trips.pickupGracePeriodSeconds;
    final availableAt =
        arrived.add(Duration(seconds: trips.pickupGracePeriodSeconds));
    return availableAt
        .difference(DateTime.now().toUtc())
        .inSeconds
        .clamp(0, trips.pickupGracePeriodSeconds)
        .toInt();
  }

  static String _durationCountdown(int seconds) {
    final minutes = seconds ~/ 60;
    final remain = seconds % 60;
    return '${minutes.toString().padLeft(2, '0')}:${remain.toString().padLeft(2, '0')}';
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

class _DriverInspectionChecklistCard extends StatelessWidget {
  const _DriverInspectionChecklistCard({
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
    final driverComplete = snapshot['driver_complete'] == true;
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
            child: Icon(Icons.fact_check_rounded,
                color: ready ? FlashXTheme.success : FlashXTheme.navy),
          ),
          const SizedBox(width: 10),
          Expanded(
            child:
                Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              const Text('Đối chiếu giấy tờ đăng kiểm',
                  style: TextStyle(fontWeight: FontWeight.w900)),
              const SizedBox(height: 2),
              Text(
                ready
                    ? 'Đủ giấy tờ bắt buộc · có thể tiếp tục nhận xe'
                    : !customerComplete
                        ? 'Khách chưa khai báo đủ giấy tờ bắt buộc'
                        : driverComplete
                            ? 'Đã đối chiếu · còn mục bắt buộc chưa hợp lệ'
                            : editable
                                ? 'Đối chiếu từng giấy với hồ sơ khách bàn giao'
                                : 'Chờ tới điểm nhận để đối chiếu thực tế',
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
          final customerStatus =
              item['customer_status']?.toString() ?? 'pending';
          final driverStatus = item['driver_status']?.toString() ?? 'pending';
          final driverLabel = switch (driverStatus) {
            'received' => 'Đã nhận',
            'missing' => 'Thiếu',
            'not_applicable' => 'Không áp dụng',
            _ => 'Chưa đối chiếu',
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
                      driverStatus == 'received'
                          ? Icons.check_circle_rounded
                          : driverStatus == 'missing'
                              ? Icons.error_outline_rounded
                              : Icons.radio_button_unchecked_rounded,
                      size: 19,
                      color: driverStatus == 'received'
                          ? FlashXTheme.success
                          : driverStatus == 'missing'
                              ? FlashXTheme.danger
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
                            Text(
                              'Khách: ${_customerLabel(customerStatus)} · Tài xế: $driverLabel${required ? ' · Bắt buộc' : ''}',
                              style: Theme.of(context).textTheme.bodySmall,
                            ),
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
        if (!ready && customerComplete && !editable)
          const _InfoCard(
            icon: Icons.info_outline_rounded,
            title: 'Chưa mở bước Nhận xe',
            text:
                'Tài xế chỉ được đối chiếu giấy tờ khi đã tới điểm nhận. Sau khi các mục bắt buộc được xác nhận, hệ thống mới mở bước tiếp theo.',
          ),
      ]),
    );
  }

  static String _customerLabel(String status) => switch (status) {
        'present' => 'Có',
        'not_available' => 'Không có',
        'not_applicable' => 'Không áp dụng',
        _ => 'Chưa khai báo',
      };
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

class _DriverLiveRouteCard extends StatelessWidget {
  const _DriverLiveRouteCard({required this.snapshot});
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
      'approach' => 'Đường tới điểm nhận',
      'return' => 'Đường trả xe cho khách',
      _ => 'Đường thực hiện dịch vụ',
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
                  : 'Đang cập nhật khoảng cách và ETA',
              style: Theme.of(context).textTheme.bodySmall,
            ),
          ]),
        ),
      ]),
    );
  }
}

class _DriverDocumentCard extends StatelessWidget {
  const _DriverDocumentCard({
    required this.icon,
    required this.label,
    required this.document,
    required this.busy,
    required this.onUpload,
  });

  final IconData icon;
  final String label;
  final Map<String, dynamic>? document;
  final bool busy;
  final VoidCallback onUpload;

  @override
  Widget build(BuildContext context) {
    final status = document?['review_status']?.toString() ?? 'missing';
    final statusLabel = switch (status) {
      'approved' => 'Đã duyệt',
      'rejected' => 'Cần tải lại',
      'pending' => 'Chờ duyệt',
      _ => 'Chưa tải',
    };
    final statusColor = switch (status) {
      'approved' => FlashXTheme.success,
      'rejected' => FlashXTheme.danger,
      'pending' => FlashXTheme.warning,
      _ => FlashXTheme.textSecondary,
    };
    final note = document?['review_note']?.toString().trim() ?? '';
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: FlashXTheme.border),
      ),
      child: Column(crossAxisAlignment: CrossAxisAlignment.stretch, children: [
        Row(children: [
          Container(
            width: 42,
            height: 42,
            decoration: BoxDecoration(
              color: const Color(0xFFF4F7FB),
              borderRadius: BorderRadius.circular(14),
            ),
            child: Icon(icon, color: FlashXTheme.navy),
          ),
          const SizedBox(width: 10),
          Expanded(
            child:
                Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Text(label, style: const TextStyle(fontWeight: FontWeight.w800)),
              const SizedBox(height: 2),
              Text(statusLabel,
                  style: TextStyle(
                      color: statusColor,
                      fontSize: 12,
                      fontWeight: FontWeight.w700)),
            ]),
          ),
          OutlinedButton(
            onPressed: busy || status == 'approved' ? null : onUpload,
            child: Text(document == null || status == 'missing'
                ? 'Tải lên'
                : status == 'rejected'
                    ? 'Tải lại'
                    : 'Thay ảnh'),
          ),
        ]),
        if (note.isNotEmpty) ...[
          const SizedBox(height: 10),
          Text('Operations: $note',
              style: Theme.of(context).textTheme.bodySmall),
        ],
      ]),
    );
  }
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
