import 'dart:io';

class AppConfig {
  const AppConfig._();
  static const _definedApi = String.fromEnvironment('API_BASE_URL');
  static const mapsEnabled =
      bool.fromEnvironment('MAPS_ENABLED', defaultValue: false);

  static String get apiBaseUrl {
    if (_definedApi.isNotEmpty) return _definedApi;
    if (Platform.isAndroid) return 'http://10.0.2.2:8080';
    return 'http://127.0.0.1:8080';
  }

  static Uri apiUri(String path) => Uri.parse('$apiBaseUrl$path');
  static Uri realtimeUri() {
    final uri = Uri.parse(apiBaseUrl);
    return uri.replace(
        scheme: uri.scheme == 'https' ? 'wss' : 'ws',
        path: '/v1/realtime',
        query: null);
  }
}
