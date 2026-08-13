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
  List<Map<String, dynamic>> vehicles = const [];
  String? selectedVehicleID;
  Map<String, dynamic>? driverLocation;
  final Set<String> ratedTripIds = <String>{};

  // Thanh Hoa city fallback keeps the demo deterministic when location access
  // is denied; a real device GPS location replaces this immediately.
  Map<String, double> pickup = const {'lat': 19.8067, 'lng': 105.7852};
  Map<String, double>? selectedDestination;
  bool busy = false;
  String? error;

  String get status => activeTrip?['status']?.toString() ?? '';
  bool get hasActiveTrip =>
      activeTrip != null && status != 'completed' && status != 'cancelled';

  bool get canCancel {
    final actions = activeTrip?['allowed_actions'];
    if (actions is List)
      return actions.map((e) => e.toString()).contains('cancel');
    return const {
      'scheduled',
      'searching',
      'accepted',
      'arriving',
      'arrived',
      'arriving_for_pickup',
      'arrived_for_pickup',
    }.contains(status);
  }

  static const destinations = <String, Map<String, double>>{
    'Quảng trường Lam Sơn': {'lat': 19.8079, 'lng': 105.7768},
    'Vincom Plaza Thanh Hóa': {'lat': 19.8045, 'lng': 105.7779},
    'Sầm Sơn': {'lat': 19.7412, 'lng': 105.9020},
    'Trung tâm đăng kiểm Thanh Hóa': {'lat': 19.8330, 'lng': 105.7810},
  };

  Future<void> initialize() async {
    await Future.wait([loadVehicles(), loadHistory(), refreshPickupLocation()]);
  }

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
      // Keep the deterministic Thanh Hoa fallback for demo/offline development.
    }
  }

  Future<void> loadVehicles() async {
    try {
      final response = await api.get('/v1/rider/vehicles');
      final list = response['data'] as List<dynamic>? ?? const [];
      vehicles = list.cast<Map<String, dynamic>>();
      if (selectedVehicleID == null && vehicles.isNotEmpty) {
        selectedVehicleID = vehicles.first['id']?.toString();
      }
      notifyListeners();
    } catch (_) {}
  }

  List<Map<String, dynamic>> vehiclesForService(String serviceType) {
    final wanted =
        serviceType == 'designated_driver_bike' ? 'motorbike' : 'car';
    return vehicles.where((vehicle) => vehicle['type'] == wanted).toList();
  }

  void selectVehicle(String? id) {
    selectedVehicleID = id;
    estimate = null;
    notifyListeners();
  }

  Future<bool> createVehicle({
    required String type,
    required String licensePlate,
    String brand = '',
    String model = '',
    String color = '',
    String transmission = 'n/a',
  }) async {
    return _guard(() async {
      final response = await api.post('/v1/rider/vehicles', body: {
        'type': type,
        'license_plate': licensePlate,
        'brand': brand,
        'model': model,
        'color': color,
        'transmission': transmission,
        'seats': type == 'car' ? 5 : 0,
      });
      final vehicle = response['data'] as Map<String, dynamic>;
      selectedVehicleID = vehicle['id']?.toString();
      await loadVehicles();
    });
  }

  Future<void> loadHistory() async {
    try {
      final response = await api.get('/v1/trips?limit=20');
      final list = response['data'] as List<dynamic>? ?? const [];
      history = list.cast<Map<String, dynamic>>();
      final candidates = history
          .where((e) => !['completed', 'cancelled'].contains(e['status']))
          .toList();
      activeTrip = candidates.isNotEmpty ? candidates.first : null;
      notifyListeners();
    } catch (_) {}
  }

  Future<bool> estimateTo(Map<String, double> destination,
      {String serviceType = 'designated_driver_car'}) async {
    return _guard(() async {
      selectedDestination = destination;
      final response = await api.post('/v1/trips/estimate', body: {
        'pickup': pickup,
        'destination': destination,
        'service_type': serviceType,
        if (selectedVehicleID != null) 'customer_vehicle_id': selectedVehicleID,
      });
      estimate = response['data'] as Map<String, dynamic>;
    });
  }

  Future<bool> createTrip(Map<String, double> destination,
      {String serviceType = 'designated_driver_car',
      String bookingMode = 'immediate',
      DateTime? scheduledAt}) async {
    if (selectedVehicleID == null) {
      error = 'Hãy chọn xe của bạn trước khi đặt dịch vụ.';
      notifyListeners();
      return false;
    }
    return _guard(() async {
      final response = await api.post(
        '/v1/trips',
        headers: {'Idempotency-Key': _idempotencyKey()},
        body: {
          'pickup': pickup,
          'destination': destination,
          'service_type': serviceType,
          'customer_vehicle_id': selectedVehicleID,
          'booking_mode': bookingMode,
          if (scheduledAt != null)
            'scheduled_at': scheduledAt.toUtc().toIso8601String(),
        },
      );
      activeTrip = response['data'] as Map<String, dynamic>;
      estimate = null;
      await loadHistory();
    });
  }

  Future<bool> cancel() async {
    final id = activeTrip?['id']?.toString();
    if (id == null || !canCancel) return false;
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
