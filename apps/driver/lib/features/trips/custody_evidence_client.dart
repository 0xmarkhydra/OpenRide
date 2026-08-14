import 'dart:io';

import '../../core/network/api_client.dart';
import '../../core/storage/direct_upload_client.dart';

class CustodyEvidenceClient {
  CustodyEvidenceClient({required this.api, DirectUploadClient? uploader})
      : uploader = uploader ?? DirectUploadClient();

  final ApiClient api;
  final DirectUploadClient uploader;

  Future<List<Map<String, dynamic>>> list(String tripID) async {
    final response = await api.get('/v1/driver/trips/$tripID/custody');
    final data = response['data'];
    if (data is! List) return const [];
    return data
        .whereType<Map>()
        .map((item) => Map<String, dynamic>.from(item))
        .toList(growable: false);
  }

  Future<Map<String, dynamic>> saveMetadata({
    required String tripID,
    required String stage,
    required String conditionNote,
    int? odometerKm,
    int? fuelPercent,
    int? batteryPercent,
  }) async {
    final body = <String, dynamic>{
      'condition_note': conditionNote,
      if (odometerKm != null) 'odometer_km': odometerKm,
      if (fuelPercent != null) 'fuel_percent': fuelPercent,
      if (batteryPercent != null) 'battery_percent': batteryPercent,
    };
    final response = await api.put(
      '/v1/driver/trips/$tripID/custody/$stage',
      body: body,
    );
    return Map<String, dynamic>.from(
      response['data'] as Map? ?? const {},
    );
  }

  Future<Map<String, dynamic>> uploadPhoto({
    required String tripID,
    required String stage,
    required File file,
    required String photoType,
    String? contentType,
  }) async {
    final size = await file.length();
    final filename = file.uri.pathSegments.isEmpty
        ? 'photo.jpg'
        : file.uri.pathSegments.last;
    final mime = contentType ?? _contentType(filename);

    final prepare = await api.post(
      '/v1/driver/trips/$tripID/custody/$stage/photos/upload-url',
      body: {
        'photo_type': photoType,
        'filename': filename,
        'content_type': mime,
        'size_bytes': size,
      },
    );
    final ticket =
        Map<String, dynamic>.from(prepare['data'] as Map? ?? const {});
    final upload =
        Map<String, dynamic>.from(ticket['upload'] as Map? ?? const {});
    final url = upload['url']?.toString() ?? '';
    if (url.isEmpty) {
      throw ApiException(500, 'SIGNED_URL_MISSING', 'Thiếu URL tải ảnh.');
    }

    try {
      await uploader.upload(
        file: file,
        method: upload['method']?.toString() ?? 'PUT',
        url: url,
        signedHeaders:
            Map<String, dynamic>.from(upload['headers'] as Map? ?? const {}),
      );
    } on DirectUploadException catch (e) {
      throw ApiException(
          e.statusCode, 'CUSTODY_PHOTO_UPLOAD_FAILED', e.message);
    }

    final complete = await api.post(
      '/v1/driver/trips/$tripID/custody/$stage/photos/complete',
      body: {
        'photo_id': ticket['photo_id'],
        'photo_type': ticket['photo_type'],
        'object_key': ticket['object_key'],
        'filename': filename,
        'content_type': mime,
        'size_bytes': size,
      },
    );
    return Map<String, dynamic>.from(
      complete['data'] as Map? ?? const {},
    );
  }

  Future<Map<String, dynamic>> confirm({
    required String tripID,
    required String stage,
  }) async {
    final response = await api.post(
      '/v1/driver/trips/$tripID/custody/$stage/confirm',
    );
    return Map<String, dynamic>.from(
      response['data'] as Map? ?? const {},
    );
  }

  static String _contentType(String filename) {
    final lower = filename.toLowerCase();
    if (lower.endsWith('.png')) return 'image/png';
    if (lower.endsWith('.webp')) return 'image/webp';
    if (lower.endsWith('.heic')) return 'image/heic';
    if (lower.endsWith('.heif')) return 'image/heif';
    return 'image/jpeg';
  }

  void close() => uploader.close();
}
