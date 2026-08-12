import 'package:flutter/material.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';

import '../../core/config/app_config.dart';

class RiderMapView extends StatelessWidget {
  const RiderMapView({
    super.key,
    required this.pickup,
    this.activeTrip,
    this.driverLocation,
  });

  final Map<String, double> pickup;
  final Map<String, dynamic>? activeTrip;
  final Map<String, dynamic>? driverLocation;

  @override
  Widget build(BuildContext context) {
    if (!AppConfig.mapsEnabled) return const _MapFallback();

    final pickupPoint = _point(pickup) ?? const LatLng(19.8067, 105.7852);
    final tripPickup = _point(activeTrip?['pickup']) ?? pickupPoint;
    final destination = _point(activeTrip?['destination']);
    final driverPayload = driverLocation?['location'] ?? driverLocation;
    final driver = _point(driverPayload);

    final markers = <Marker>{
      Marker(
        markerId: const MarkerId('pickup'),
        position: tripPickup,
        infoWindow: const InfoWindow(title: 'Điểm nhận xe'),
      ),
      if (destination != null)
        Marker(
          markerId: const MarkerId('destination'),
          position: destination,
          infoWindow: const InfoWindow(title: 'Điểm đến'),
        ),
      if (driver != null)
        Marker(
          markerId: const MarkerId('driver'),
          position: driver,
          icon:
              BitmapDescriptor.defaultMarkerWithHue(BitmapDescriptor.hueYellow),
          infoWindow: const InfoWindow(title: 'Tài xế'),
        ),
    };

    return GoogleMap(
      initialCameraPosition: CameraPosition(target: tripPickup, zoom: 14.5),
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
  const _MapFallback();

  @override
  Widget build(BuildContext context) => ColoredBox(
        color: const Color(0xFFE9EBF0),
        child: CustomPaint(
          painter: _MapGridPainter(),
          child: const Center(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(Icons.location_on_rounded,
                    size: 48, color: Color(0xFF0B132B)),
                SizedBox(height: 8),
                Text('Bản đồ theo dõi dịch vụ FlashX'),
              ],
            ),
          ),
        ),
      );
}

class _MapGridPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = Colors.white.withValues(alpha: .65)
      ..strokeWidth = 8;
    for (var i = 0; i < 8; i++) {
      canvas.drawLine(
          Offset(0, i * 100), Offset(size.width, i * 100 + 90), paint);
    }
    for (var i = 0; i < 6; i++) {
      canvas.drawLine(Offset(i * 110, 0), Offset(i * 85, size.height), paint);
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}
