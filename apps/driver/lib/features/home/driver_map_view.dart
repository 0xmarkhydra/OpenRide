import 'package:flutter/material.dart';
import 'package:geolocator/geolocator.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';

import '../../core/config/app_config.dart';

class DriverMapView extends StatelessWidget {
  const DriverMapView({
    super.key,
    required this.online,
    this.position,
    this.trip,
    this.liveRoutes,
  });

  final bool online;
  final Position? position;
  final Map<String, dynamic>? trip;
  final Map<String, dynamic>? liveRoutes;

  @override
  Widget build(BuildContext context) {
    if (!AppConfig.mapsEnabled) return _MapFallback(online: online);

    final driver = position == null
        ? const LatLng(19.8067, 105.7852)
        : LatLng(position!.latitude, position!.longitude);
    final pickup = _point(trip?['pickup']);
    final destination = _point(trip?['destination']);
    final primaryRoute = liveRoutes?['primary'];
    final serviceRoute = liveRoutes?['service'];
    final encoded = primaryRoute is Map &&
            (primaryRoute['encoded_polyline']?.toString().isNotEmpty ?? false)
        ? primaryRoute['encoded_polyline'].toString()
        : serviceRoute is Map
            ? serviceRoute['encoded_polyline']?.toString() ?? ''
            : '';
    final decodedRoute = _decodePolyline(encoded);
    final phase = liveRoutes?['phase']?.toString() ?? 'service';
    final target = phase == 'approach'
        ? pickup
        : phase == 'return'
            ? pickup
            : destination;
    final routePoints = decodedRoute.length >= 2
        ? decodedRoute
        : <LatLng>[driver, if (target != null) target];
    final markers = <Marker>{
      Marker(
        markerId: const MarkerId('driver'),
        position: driver,
        icon: BitmapDescriptor.defaultMarkerWithHue(BitmapDescriptor.hueYellow),
        infoWindow: const InfoWindow(title: 'Vị trí của bạn'),
      ),
      if (pickup != null)
        Marker(
          markerId: const MarkerId('pickup'),
          position: pickup,
          infoWindow: const InfoWindow(title: 'Điểm nhận xe'),
        ),
      if (destination != null)
        Marker(
          markerId: const MarkerId('destination'),
          position: destination,
          infoWindow: const InfoWindow(title: 'Điểm đến'),
        ),
    };

    return GoogleMap(
      initialCameraPosition: CameraPosition(target: driver, zoom: 14.5),
      markers: markers,
      polylines: routePoints.length >= 2
          ? {
              Polyline(
                polylineId: const PolylineId('flashx-driver-route'),
                points: routePoints,
                width: 5,
                color: const Color(0xFF079455),
                geodesic: true,
              ),
            }
          : const <Polyline>{},
      compassEnabled: true,
      mapToolbarEnabled: false,
      zoomControlsEnabled: false,
      myLocationButtonEnabled: false,
    );
  }

  static LatLng? _point(dynamic value) {
    if (value is! Map) return null;
    final lat = value['lat'];
    final lng = value['lng'];
    if (lat is! num || lng is! num) return null;
    return LatLng(lat.toDouble(), lng.toDouble());
  }

  static List<LatLng> _decodePolyline(String encoded) {
    if (encoded.isEmpty) return const [];
    final points = <LatLng>[];
    var index = 0;
    var lat = 0;
    var lng = 0;
    while (index < encoded.length) {
      int decodeValue() {
        var result = 0;
        var shift = 0;
        var byte = 0;
        do {
          if (index >= encoded.length) return 0;
          byte = encoded.codeUnitAt(index++) - 63;
          result |= (byte & 0x1f) << shift;
          shift += 5;
        } while (byte >= 0x20);
        return (result & 1) != 0 ? ~(result >> 1) : result >> 1;
      }

      lat += decodeValue();
      lng += decodeValue();
      points.add(LatLng(lat / 1e5, lng / 1e5));
    }
    return points;
  }
}

class _MapFallback extends StatelessWidget {
  const _MapFallback({required this.online});
  final bool online;

  @override
  Widget build(BuildContext context) => ColoredBox(
        color: const Color(0xFFE9EBF0),
        child: Center(
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 250),
            padding: const EdgeInsets.all(24),
            decoration: const BoxDecoration(
                color: Colors.white, shape: BoxShape.circle),
            child: Icon(
              online ? Icons.navigation_rounded : Icons.location_off_outlined,
              size: 48,
              color: online ? const Color(0xFF0B132B) : Colors.grey,
            ),
          ),
        ),
      );
}
