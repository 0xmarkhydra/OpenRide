import 'dart:io';

import '../../core/network/api_client.dart';
import '../../core/storage/direct_upload_client.dart';

class DriverDocumentClient {
  DriverDocumentClient({required this.api, DirectUploadClient? uploader})
      : uploader = uploader ?? DirectUploadClient();

  final ApiClient api;
  final DirectUploadClient uploader;

  Future<Map<String, dynamic>> uploadDocument({
    required File file,
    required String documentType,
    required String contentType,
  }) async {
    final size = await file.length();
    final filename =
        file.uri.pathSegments.isEmpty ? 'document' : file.uri.pathSegments.last;

    final prepareResponse = await api.post(
      '/v1/driver/documents/upload-url',
      body: {
        'document_type': documentType,
        'filename': filename,
        'content_type': contentType,
        'size_bytes': size,
      },
    );
    final ticket =
        Map<String, dynamic>.from(prepareResponse['data'] as Map? ?? const {});
    final upload =
        Map<String, dynamic>.from(ticket['upload'] as Map? ?? const {});

    await uploader.upload(
      file: file,
      method: upload['method']?.toString() ?? 'PUT',
      url: upload['url']?.toString() ?? '',
      signedHeaders: Map<String, dynamic>.from(
        upload['headers'] as Map? ?? const {},
      ),
    );

    final completeResponse = await api.post(
      '/v1/driver/documents/complete',
      body: {
        'document_id': ticket['document_id'],
        'document_type': ticket['document_type'],
        'object_key': ticket['object_key'],
        'filename': filename,
        'content_type': contentType,
        'size_bytes': size,
      },
    );
    return Map<String, dynamic>.from(
      completeResponse['data'] as Map? ?? const {},
    );
  }

  Future<List<Map<String, dynamic>>> list() async {
    final response = await api.get('/v1/driver/documents');
    final data = response['data'];
    if (data is! List) return const [];
    return data
        .whereType<Map>()
        .map((item) => Map<String, dynamic>.from(item))
        .toList(growable: false);
  }

  Future<Uri> viewUrl(String documentID) async {
    final response = await api.get('/v1/driver/documents/$documentID/view-url');
    final data = Map<String, dynamic>.from(
      response['data'] as Map? ?? const {},
    );
    final view = Map<String, dynamic>.from(data['view'] as Map? ?? const {});
    final raw = view['url']?.toString() ?? '';
    if (raw.isEmpty) {
      throw ApiException(500, 'SIGNED_URL_MISSING', 'Thiếu URL xem tài liệu.');
    }
    return Uri.parse(raw);
  }

  void close() => uploader.close();
}
