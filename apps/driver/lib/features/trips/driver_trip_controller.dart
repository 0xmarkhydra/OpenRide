import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:geolocator/geolocator.dart';

import '../../core/network/api_client.dart';
import '../../core/realtime/realtime_client.dart';
import '../auth/auth_controller.dart';

class DriverTripController extends ChangeNotifier {
  DriverTripController({required this.auth})
      : api = auth.api,
        realtime = auth.realtime {
    _realtimeSub = realtime.events.listen(_onRealtime, onError: (_) {});
  }

  final AuthController auth;
  final ApiClient api;
  final RealtimeClient realtime;

  StreamSubscription<Map<String, dynamic>>? _realtimeSub;
  StreamSubscription<Position>? _positionSub;
  Map<String, dynamic>? currentOffer;
  Map<String, dynamic>? offeredTrip;
  Map<String, dynamic>? activeTrip;
  List<Map<String, dynamic>> history = const [];
  Position? lastPosition;
  bool online = false;
  bool busy = false;
  String? error;

  String get approval =>
      auth.profile?['approval_status']?.toString() ?? 'pending';
  bool get approved => approval == 'approved';
  int get completedTrips =>
      history.where((trip) => trip['status'] == 'completed').length;
  int get totalEarningsMinor => history
      .where((trip) => trip['status'] == 'completed')
      .fold<int>(
        0,
        (total, trip) =>
            total +
            ((trip['final_fare_minor'] ?? trip['estimated_fare_minor'] ?? 0)
                    as num)
                .toInt(),
      );

  Future<void> initialize() async {
    final profile = auth.profile;
    online = profile?['availability_status'] == 'online';
    await refreshOffer();
    await loadHistory();
    if (online) await _startLocationStream();
    notifyListeners();
  }

  Future<bool> setOnline(bool value) async {
    if (value && !approved) {
      error = 'Hồ sơ tài xế cần được Admin duyệt trước khi online.';
      notifyListeners();
      return false;
    }
    return _guard(() async {
      if (value) await _ensureLocationPermission();
      final response = await api.post('/v1/driver/availability',
          body: {'status': value ? 'online' : 'offline'});
      final data = response['data'] as Map<String, dynamic>;
      online = data['availability_status'] == 'online';
      auth.profile = data;
      if (online) {
        await _startLocationStream();
      } else {
        await _stopLocationStream();
        currentOffer = null;
        offeredTrip = null;
      }
    });
  }

  Future<void> loadHistory() async {
    try {
      final response = await api.get('/v1/driver/trips?limit=50');
      final items = response['data'] as List<dynamic>? ?? const [];
      history = items.cast<Map<String, dynamic>>();
      notifyListeners();
    } catch (_) {}
  }

  Future<void> refreshOffer() async {
    if (!approved) return;
    try {
      final response = await api.get('/v1/driver/offers/current');
      final data = response['data'] as Map<String, dynamic>?;
      currentOffer = data?['offer'] as Map<String, dynamic>?;
      offeredTrip = data?['trip'] as Map<String, dynamic>?;
      notifyListeners();
    } on ApiException catch (e) {
      if (e.statusCode == 404) {
        currentOffer = null;
        notifyListeners();
      }
    }
  }

  Future<bool> acceptOffer() async {
    final id = currentOffer?['id']?.toString();
    if (id == null) return false;
    return _guard(() async {
      final response = await api.post('/v1/driver/offers/$id/accept');
      final data = response['data'] as Map<String, dynamic>;
      activeTrip = data['trip'] as Map<String, dynamic>?;
      currentOffer = null;
      offeredTrip = null;
    });
  }

  Future<bool> rejectOffer() async {
    final id = currentOffer?['id']?.toString();
    if (id == null) return false;
    return _guard(() async {
      await api.post('/v1/driver/offers/$id/reject');
      currentOffer = null;
      offeredTrip = null;
    });
  }

  Future<bool> arriving() => _tripCommand('arriving');
  Future<bool> arrived() => _tripCommand('arrived');
  Future<bool> startTrip() => _tripCommand('start');
  Future<bool> completeTrip() => _tripCommand('complete');

  Future<bool> _tripCommand(String action) async {
    final id = activeTrip?['id']?.toString();
    if (id == null) return false;
    return _guard(() async {
      final response = await api.post('/v1/driver/trips/$id/$action');
      activeTrip = response['data'] as Map<String, dynamic>?;
      if (activeTrip?['status'] == 'completed') {
        activeTrip = null;
        online = true;
        await auth.refreshProfile();
        await loadHistory();
      }
    });
  }

  Future<void> _ensureLocationPermission() async {
    if (!await Geolocator.isLocationServiceEnabled()) {
      throw Exception('Location services are disabled');
    }
    var permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
    }
    if (permission == LocationPermission.denied ||
        permission == LocationPermission.deniedForever) {
      throw Exception('Location permission denied');
    }
  }

  Future<void> _startLocationStream() async {
    await _ensureLocationPermission();
    await _positionSub?.cancel();
    const settings =
        LocationSettings(accuracy: LocationAccuracy.high, distanceFilter: 15);
    _positionSub = Geolocator.getPositionStream(locationSettings: settings)
        .listen((position) {
      unawaited(_sendPosition(position));
    }, onError: (_) {});
    try {
      await _sendPosition(await Geolocator.getCurrentPosition());
    } catch (_) {}
  }

  Future<void> _sendPosition(Position position) async {
    lastPosition = position;
    notifyListeners();
    if (!online) return;
    try {
      await api.post('/v1/driver/location', body: {
        'lat': position.latitude,
        'lng': position.longitude,
        'accuracy_m': position.accuracy,
        'heading_deg': position.heading,
        'speed_mps': position.speed,
        'captured_at': position.timestamp.toUtc().toIso8601String(),
      });
    } catch (_) {}
  }

  Future<void> _stopLocationStream() async {
    await _positionSub?.cancel();
    _positionSub = null;
  }

  void _onRealtime(Map<String, dynamic> event) {
    final type = event['type']?.toString() ?? '';
    final data = event['data'];
    if (type == 'dispatch.offer' && data is Map<String, dynamic>) {
      currentOffer = data['offer'] as Map<String, dynamic>?;
      notifyListeners();
      return;
    }
    if (type.startsWith('trip.') && data is Map<String, dynamic>) {
      final nested = data['trip'];
      final trip = nested is Map<String, dynamic> ? nested : data;
      activeTrip = trip;
      if (trip['status'] == 'completed' || trip['status'] == 'cancelled') {
        activeTrip = null;
      }
      notifyListeners();
    }
  }

  Future<bool> _guard(Future<void> Function() fn) async {
    busy = true;
    error = null;
    notifyListeners();
    try {
      await fn();
      return true;
    } on ApiException catch (e) {
      error = e.message;
      return false;
    } catch (e) {
      final text = e.toString();
      error = text.contains('permission') || text.contains('Location')
          ? 'Cần bật GPS và cấp quyền vị trí để nhận chuyến.'
          : 'Không thể kết nối hệ thống.';
      return false;
    } finally {
      busy = false;
      notifyListeners();
    }
  }

  @override
  void dispose() {
    unawaited(_positionSub?.cancel());
    unawaited(_realtimeSub?.cancel());
    super.dispose();
  }
}
