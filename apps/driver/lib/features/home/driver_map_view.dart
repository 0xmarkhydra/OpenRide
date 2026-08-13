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
  });

  final bool online;
  final Position? position;
  final Map<String, dynamic>? trip;

  @override
  Widget build(BuildContext context) {
    if (!AppConfig.mapsEnabled) return _MapFallback(online: online);

    final driver = position == null
        ? const LatLng(19.8067, 105.7852)
        : LatLng(position!.latitude, position!.longitude);
    final pickup = _point(trip?['pickup']);
    final destination = _point(trip?['destination']);
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
