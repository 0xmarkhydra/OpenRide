import 'dart:async';
import 'dart:math';

import 'package:flutter/foundation.dart';
import 'package:geolocator/geolocator.dart';

import '../../core/network/api_client.dart';
import '../../core/realtime/realtime_client.dart';

class RiderTripController extends ChangeNotifier {
  RiderTripController({required this.api, required this.realtime}) {
    _subscription = realtime.events.listen(_onRealtime, onError: (_) {});
  }

  final ApiClient api;
  final RealtimeClient realtime;
  StreamSubscription<Map<String, dynamic>>? _subscription;

  Map<String, dynamic>? estimate;
  Map<String, dynamic>? activeTrip;
  List<Map<String, dynamic>> history = const [];
  Map<String, dynamic>? driverLocation;
  final Set<String> ratedTripIds = <String>{};
  Map<String, double> pickup = const {'lat': 21.0285, 'lng': 105.8542};
  Map<String, double>? selectedDestination;
  bool busy = false;
  String? error;

  String get status => activeTrip?['status']?.toString() ?? '';
  bool get hasActiveTrip =>
      activeTrip != null && status != 'completed' && status != 'cancelled';

  static const destinations = <String, Map<String, double>>{
    'Hồ Hoàn Kiếm': {'lat': 21.0288, 'lng': 105.8522},
    'Lotte Center': {'lat': 21.0322, 'lng': 105.8127},
    'Times City': {'lat': 20.9950, 'lng': 105.8682},
  };

  Future<void> refreshPickupLocation() async {
    try {
      if (!await Geolocator.isLocationServiceEnabled()) return;
      var permission = await Geolocator.checkPermission();
      if (permission == LocationPermission.denied) {
        permission = await Geolocator.requestPermission();
      }
      if (permission == LocationPermission.denied ||
          permission == LocationPermission.deniedForever) {
        return;
      }
      final position = await Geolocator.getCurrentPosition(
          locationSettings:
              const LocationSettings(accuracy: LocationAccuracy.high));
      pickup = {'lat': position.latitude, 'lng': position.longitude};
      notifyListeners();
    } catch (_) {
      // Keep the deterministic Hanoi fallback so development/demo remains usable.
    }
  }

  Future<void> loadHistory() async {
    try {
      final response = await api.get('/v1/trips?limit=20');
      final list = response['data'] as List<dynamic>? ?? const [];
      history = list.cast<Map<String, dynamic>>();
      final candidates = history
          .where((e) => !['completed', 'cancelled'].contains(e['status']))
          .toList();
      if (candidates.isNotEmpty) activeTrip = candidates.first;
      notifyListeners();
    } catch (_) {}
  }

  Future<bool> estimateTo(Map<String, double> destination,
      {String serviceType = 'bike'}) async {
    return _guard(() async {
      selectedDestination = destination;
      final response = await api.post('/v1/trips/estimate', body: {
        'pickup': pickup,
        'destination': destination,
        'service_type': serviceType,
      });
      estimate = response['data'] as Map<String, dynamic>;
    });
  }

  Future<bool> createTrip(Map<String, double> destination,
      {String serviceType = 'bike'}) async {
    return _guard(() async {
      final response = await api.post(
        '/v1/trips',
        headers: {'Idempotency-Key': _idempotencyKey()},
        body: {
          'pickup': pickup,
          'destination': destination,
          'service_type': serviceType
        },
      );
      activeTrip = response['data'] as Map<String, dynamic>;
      estimate = null;
    });
  }

  Future<bool> cancel() async {
    final id = activeTrip?['id']?.toString();
    if (id == null) return false;
    return _guard(() async {
      final response = await api
          .post('/v1/trips/$id/cancel', body: {'reason': 'rider_cancelled'});
      activeTrip = response['data'] as Map<String, dynamic>;
      await loadHistory();
    });
  }

  Future<bool> rateTrip(String tripID, int stars, String comment) async {
    return _guard(() async {
      await api.post('/v1/trips/$tripID/rating', body: {
        'stars': stars,
        'comment': comment,
      });
      ratedTripIds.add(tripID);
    });
  }

  Future<void> refreshActive() async {
    final id = activeTrip?['id']?.toString();
    if (id == null) return;
    try {
      final response = await api.get('/v1/trips/$id');
      activeTrip = response['data'] as Map<String, dynamic>;
      notifyListeners();
    } catch (_) {}
  }

  void _onRealtime(Map<String, dynamic> event) {
    final type = event['type']?.toString() ?? '';
    if (type == 'driver.location') {
      final data = event['data'];
      if (data is Map<String, dynamic>) driverLocation = data;
      notifyListeners();
      return;
    }
    if (type.startsWith('trip.')) {
      final data = event['data'];
      if (data is Map<String, dynamic>) {
        final trip = data['trip'];
        activeTrip = trip is Map<String, dynamic> ? trip : data;
      }
      unawaited(refreshActive());
    }
  }

  Future<bool> _guard(Future<void> Function() action) async {
    busy = true;
    error = null;
    notifyListeners();
    try {
      await action();
      return true;
    } on ApiException catch (e) {
      error = e.message;
      return false;
    } catch (_) {
      error = 'Không thể kết nối hệ thống.';
      return false;
    } finally {
      busy = false;
      notifyListeners();
    }
  }

  String _idempotencyKey() =>
      '${DateTime.now().microsecondsSinceEpoch}-${Random.secure().nextInt(1 << 31)}';

  @override
  void dispose() {
    unawaited(_subscription?.cancel());
    super.dispose();
  }
}
