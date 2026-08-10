import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class SessionTokens {
  const SessionTokens({required this.accessToken, required this.refreshToken});
  final String accessToken;
  final String refreshToken;
}

class SessionStore {
  const SessionStore([this._storage = const FlutterSecureStorage()]);
  final FlutterSecureStorage _storage;

  Future<SessionTokens?> read() async {
    final access = await _storage.read(key: 'driver_access_token');
    final refresh = await _storage.read(key: 'driver_refresh_token');
    if (access == null || refresh == null) return null;
    return SessionTokens(accessToken: access, refreshToken: refresh);
  }

  Future<void> save(SessionTokens tokens) async {
    await Future.wait([
      _storage.write(key: 'driver_access_token', value: tokens.accessToken),
      _storage.write(key: 'driver_refresh_token', value: tokens.refreshToken),
    ]);
  }

  Future<void> clear() async {
    await Future.wait([
      _storage.delete(key: 'driver_access_token'),
      _storage.delete(key: 'driver_refresh_token'),
    ]);
  }
}
