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
  Map<String, dynamic>? activePayment;
  List<Map<String, dynamic>> custodyEvidence = const [];
  Map<String, dynamic>? inspectionChecklist;
  List<Map<String, dynamic>> history = const [];
  List<Map<String, dynamic>> vehicles = const [];
  String? selectedVehicleID;
  Map<String, dynamic>? driverLocation;
  Map<String, dynamic>? liveRoutes;
  List<Map<String, dynamic>> placeResults = const [];
  bool placesLoading = false;
  bool placeSearchUnavailable = false;
  DateTime? _lastRouteRefreshAt;
  final Map<String, Map<String, dynamic>> ratingsByTripID =
      <String, Map<String, dynamic>>{};

  // Thanh Hoa city fallback keeps the demo deterministic when location access
  // is denied; a real device GPS location replaces this immediately.
  Map<String, double> pickup = const {'lat': 19.8067, 'lng': 105.7852};
  Map<String, double>? selectedDestination;
  bool busy = false;
  String? error;

  String get status => activeTrip?['status']?.toString() ?? '';
  bool get hasActiveTrip =>
      activeTrip != null && status != 'completed' && status != 'cancelled';

  Map<String, dynamic>? custodyFor(String stage) {
    for (final snapshot in custodyEvidence) {
      final evidence = snapshot['evidence'];
      if (evidence is Map && evidence['stage']?.toString() == stage) {
        return snapshot;
      }
    }
    return null;
  }

  bool get canCancel {
    final actions = activeTrip?['allowed_actions'];
    if (actions is List) {
      return actions.map((e) => e.toString()).contains('cancel');
    }
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

  void selectDestination(Map<String, double>? destination) {
    selectedDestination = destination == null
        ? null
        : {'lat': destination['lat']!, 'lng': destination['lng']!};
    estimate = null;
    notifyListeners();
  }

  Future<void> searchPlaces(String query) async {
    final normalized = query.trim();
    if (normalized.runes.length < 3) {
      placeResults = const [];
      placeSearchUnavailable = false;
      notifyListeners();
      return;
    }
    placesLoading = true;
    placeSearchUnavailable = false;
    notifyListeners();
    try {
      final params = Uri(queryParameters: {
        'q': normalized,
        'lat': pickup['lat']!.toString(),
        'lng': pickup['lng']!.toString(),
      }).query;
      final response = await api.get('/v1/places/search?$params');
      final list = response['data'] as List<dynamic>? ?? const [];
      placeResults = list.cast<Map<String, dynamic>>();
    } catch (_) {
      placeResults = const [];
      placeSearchUnavailable = true;
    } finally {
      placesLoading = false;
      notifyListeners();
    }
  }

  void clearPlaceSearch() {
    placeResults = const [];
    placeSearchUnavailable = false;
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
      if (activeTrip != null) {
        await loadPayment();
        await loadCustodyEvidence();
        await loadInspectionChecklist();
        await loadActiveRoutes(force: true);
      } else {
        activePayment = null;
        custodyEvidence = const [];
        inspectionChecklist = null;
      }
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

  Future<Map<String, dynamic>?> loadTripRating(String tripID) async {
    final cached = ratingsByTripID[tripID];
    if (cached != null) return cached;
    try {
      final response = await api.get('/v1/trips/$tripID/rating');
      final rating = response['data'] as Map<String, dynamic>?;
      if (rating != null) ratingsByTripID[tripID] = rating;
      return rating;
    } on ApiException catch (e) {
      if (e.statusCode == 404) return null;
      rethrow;
    }
  }

  Future<bool> rateTrip(String tripID, int stars, String comment) async {
    return _guard(() async {
      final response = await api.post('/v1/trips/$tripID/rating', body: {
        'stars': stars,
        'comment': comment,
      });
      final rating = response['data'] as Map<String, dynamic>?;
      if (rating != null) ratingsByTripID[tripID] = rating;
    });
  }

  Future<bool> reportIncident(String incidentType, String note) async {
    final id = activeTrip?['id']?.toString();
    if (id == null) return false;
    return _guard(() async {
      final response = await api.post('/v1/trips/$id/incident', body: {
        'type': incidentType,
        'note': note,
      });
      activeTrip = response['data'] as Map<String, dynamic>?;
    });
  }

  Future<void> loadPayment() async {
    final id = activeTrip?['id']?.toString();
    if (id == null || id.isEmpty) {
      activePayment = null;
      notifyListeners();
      return;
    }
    try {
      final response = await api.get('/v1/trips/$id/payment');
      activePayment = response['data'] as Map<String, dynamic>?;
      notifyListeners();
    } catch (_) {
      // Payment visibility is secondary to the trip lifecycle; keep the trip usable.
    }
  }

  Future<void> loadCustodyEvidence() async {
    final id = activeTrip?['id']?.toString();
    if (id == null || id.isEmpty) {
      custodyEvidence = const [];
      notifyListeners();
      return;
    }
    try {
      final response = await api.get('/v1/trips/$id/custody');
      final data = response['data'];
      custodyEvidence = data is List
          ? data
              .whereType<Map>()
              .map((item) => Map<String, dynamic>.from(item))
              .toList(growable: false)
          : const [];
      notifyListeners();
    } on ApiException catch (e) {
      if (e.statusCode == 404) {
        custodyEvidence = const [];
        notifyListeners();
      }
    }
  }

  Future<bool> confirmCustody(String stage) async {
    final id = activeTrip?['id']?.toString();
    if (id == null) return false;
    return _guard(() async {
      await api.post('/v1/trips/$id/custody/$stage/confirm');
      await loadCustodyEvidence();
    });
  }

  Future<void> loadActiveRoutes({bool force = false}) async {
    final id = activeTrip?['id']?.toString();
    if (id == null || id.isEmpty || !hasActiveTrip) {
      liveRoutes = null;
      notifyListeners();
      return;
    }
    final now = DateTime.now();
    if (!force &&
        _lastRouteRefreshAt != null &&
        now.difference(_lastRouteRefreshAt!) < const Duration(seconds: 15)) {
      return;
    }
    _lastRouteRefreshAt = now;
    try {
      final response = await api.get('/v1/trips/$id/routes');
      liveRoutes = response['data'] as Map<String, dynamic>?;
      notifyListeners();
    } catch (_) {
      // Route refresh is secondary to trip state; keep the last good snapshot.
    }
  }

  Future<void> loadInspectionChecklist() async {
    final trip = activeTrip;
    final id = trip?['id']?.toString();
    if (id == null || trip?['service_type'] != 'vehicle_inspection_assist') {
      inspectionChecklist = null;
      notifyListeners();
      return;
    }
    try {
      final response = await api.get('/v1/trips/$id/inspection-checklist');
      inspectionChecklist = response['data'] as Map<String, dynamic>?;
      notifyListeners();
    } on ApiException catch (e) {
      if (e.statusCode == 404 || e.statusCode == 422) {
        inspectionChecklist = null;
        notifyListeners();
      }
    }
  }

  Future<bool> updateInspectionChecklistItem(
      String itemKey, String status, String note) async {
    final id = activeTrip?['id']?.toString();
    if (id == null) return false;
    return _guard(() async {
      final response = await api.patch(
        '/v1/trips/$id/inspection-checklist/$itemKey',
        body: {'status': status, 'note': note},
      );
      inspectionChecklist = response['data'] as Map<String, dynamic>?;
    });
  }

  Future<void> refreshActive() async {
    final id = activeTrip?['id']?.toString();
    if (id == null) return;
    try {
      final response = await api.get('/v1/trips/$id');
      activeTrip = response['data'] as Map<String, dynamic>;
      if (hasActiveTrip) {
        await loadPayment();
        await loadCustodyEvidence();
        await loadInspectionChecklist();
        await loadActiveRoutes(force: true);
      } else {
        activePayment = null;
        custodyEvidence = const [];
        inspectionChecklist = null;
      }
      notifyListeners();
    } catch (_) {}
  }

  void _onRealtime(Map<String, dynamic> event) {
    final type = event['type']?.toString() ?? '';
    if (type == 'driver.location') {
      final data = event['data'];
      if (data is Map<String, dynamic>) driverLocation = data;
      notifyListeners();
      unawaited(loadActiveRoutes());
      return;
    }
    if (type == 'trip.custody_evidence_updated') {
      final data = event['data'];
      final evidence = data is Map ? data['evidence'] : null;
      final tripID = evidence is Map ? evidence['trip_id']?.toString() : null;
      if (tripID != null && tripID == activeTrip?['id']?.toString()) {
        unawaited(loadCustodyEvidence());
      }
      return;
    }
    if (type == 'trip.inspection_checklist_updated') {
      final data = event['data'];
      final tripID = data is Map ? data['trip_id']?.toString() : null;
      if (tripID != null && tripID == activeTrip?['id']?.toString()) {
        unawaited(loadInspectionChecklist());
      }
      return;
    }
    if (type.startsWith('trip.')) {
      final data = event['data'];
      if (data is Map<String, dynamic>) {
        final nested = data['trip'];
        final trip = nested is Map<String, dynamic> ? nested : data;
        // Auxiliary trip.* realtime events (location/evidence/etc.) must never
        // replace the current job unless they actually carry a trip record.
        if (trip['id'] == null || trip['status'] == null) return;
        activeTrip = trip;
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
