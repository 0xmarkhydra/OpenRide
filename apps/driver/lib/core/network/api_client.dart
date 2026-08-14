import 'dart:convert';
import 'dart:io';

import '../config/app_config.dart';

class ApiException implements Exception {
  ApiException(this.statusCode, this.code, this.message);
  final int statusCode;
  final String code;
  final String message;
  @override
  String toString() => message;
}

typedef TokenRefreshCallback = Future<String?> Function();

class ApiClient {
  ApiClient({required this.accessTokenProvider, this.onUnauthorized});

  final Future<String?> Function() accessTokenProvider;
  final TokenRefreshCallback? onUnauthorized;
  final HttpClient _client = HttpClient()
    ..connectionTimeout = const Duration(seconds: 10);

  Future<Map<String, dynamic>> get(
    String path, {
    bool authenticated = true,
  }) =>
      _request('GET', path, authenticated: authenticated);

  Future<Map<String, dynamic>> post(
    String path, {
    Object? body,
    Map<String, String>? headers,
    bool authenticated = true,
    bool allowTokenRefresh = true,
  }) =>
      _request(
        'POST',
        path,
        body: body,
        headers: headers,
        authenticated: authenticated,
        allowTokenRefresh: allowTokenRefresh,
      );

  Future<Map<String, dynamic>> patch(
    String path, {
    Object? body,
    bool authenticated = true,
  }) =>
      _request('PATCH', path, body: body, authenticated: authenticated);

  Future<Map<String, dynamic>> put(
    String path, {
    Object? body,
    bool authenticated = true,
  }) =>
      _request('PUT', path, body: body, authenticated: authenticated);

  Future<Map<String, dynamic>> _request(
    String method,
    String path, {
    Object? body,
    Map<String, String>? headers,
    bool authenticated = true,
    bool allowTokenRefresh = true,
    String? tokenOverride,
  }) async {
    final request = await _client.openUrl(method, AppConfig.apiUri(path));
    request.headers.set(HttpHeaders.acceptHeader, 'application/json');
    request.headers
        .set(HttpHeaders.contentTypeHeader, 'application/json; charset=utf-8');
    if (authenticated) {
      final token = tokenOverride ?? await accessTokenProvider();
      if (token != null && token.isNotEmpty) {
        request.headers.set(HttpHeaders.authorizationHeader, 'Bearer $token');
      }
    }
    headers?.forEach(request.headers.set);
    if (body != null) request.write(jsonEncode(body));

    final response = await request.close();
    final raw = await utf8.decoder.bind(response).join();
    final decoded = raw.isEmpty
        ? <String, dynamic>{}
        : jsonDecode(raw) as Map<String, dynamic>;

    if (response.statusCode == HttpStatus.unauthorized &&
        authenticated &&
        allowTokenRefresh &&
        onUnauthorized != null) {
      final refreshedToken = await onUnauthorized!();
      if (refreshedToken != null && refreshedToken.isNotEmpty) {
        return _request(
          method,
          path,
          body: body,
          headers: headers,
          authenticated: true,
          allowTokenRefresh: false,
          tokenOverride: refreshedToken,
        );
      }
    }

    if (response.statusCode < 200 || response.statusCode >= 300) {
      final error = decoded['error'] as Map<String, dynamic>?;
      throw ApiException(
        response.statusCode,
        error?['code']?.toString() ?? 'HTTP_${response.statusCode}',
        error?['message']?.toString() ?? 'Có lỗi kết nối hệ thống.',
      );
    }
    return decoded;
  }

  void close() => _client.close(force: true);
}
