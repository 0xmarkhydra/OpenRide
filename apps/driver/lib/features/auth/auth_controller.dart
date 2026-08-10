import 'dart:async';

import 'package:flutter/foundation.dart';

import '../../core/network/api_client.dart';
import '../../core/realtime/realtime_client.dart';
import '../../core/storage/session_store.dart';

class AuthController extends ChangeNotifier {
  AuthController({SessionStore? store})
      : _store = store ?? const SessionStore() {
    realtime = RealtimeClient();
    _api = ApiClient(
      accessTokenProvider: () async => (await _store.read())?.accessToken,
      onUnauthorized: _refreshAccessToken,
    );
  }

  final SessionStore _store;
  late final ApiClient _api;
  late final RealtimeClient realtime;
  SessionTokens? _tokens;
  Future<String?>? _refreshInFlight;
  Map<String, dynamic>? profile;
  String? challengeId;
  String? developmentCode;
  bool initialized = false;
  bool busy = false;
  String? error;

  bool get authenticated => _tokens != null;
  String? get accessToken => _tokens?.accessToken;
  ApiClient get api => _api;

  Future<void> initialize() async {
    try {
      _tokens = await _store.read();
    } catch (_) {
      _tokens = null;
    }
    if (_tokens != null) {
      try {
        await refreshProfile();
        final token = _tokens?.accessToken;
        if (token != null) await realtime.connect(token);
      } catch (_) {
        await _clearLocalSession();
      }
    }
    initialized = true;
    notifyListeners();
  }

  Future<void> refreshProfile() async {
    final response = await _api.get('/v1/driver/me');
    profile = response['data'] as Map<String, dynamic>?;
    notifyListeners();
  }

  Future<bool> requestOTP(String phone) => _guard(() async {
        final response = await _api.post(
          '/v1/auth/otp/request',
          body: {'phone': phone, 'role': 'driver'},
          authenticated: false,
          allowTokenRefresh: false,
        );
        final data = response['data'] as Map<String, dynamic>;
        challengeId = data['challenge_id'].toString();
        developmentCode = data['debug_code']?.toString();
      });

  Future<bool> verifyOTP(String code) async {
    final challenge = challengeId;
    if (challenge == null) return false;
    return _guard(() async {
      final response = await _api.post(
        '/v1/auth/otp/verify',
        body: {'challenge_id': challenge, 'role': 'driver', 'code': code},
        authenticated: false,
        allowTokenRefresh: false,
      );
      final data = response['data'] as Map<String, dynamic>;
      await _saveTokenResponse(data['tokens'] as Map<String, dynamic>);
      profile = data['actor'] as Map<String, dynamic>?;
      challengeId = null;
      developmentCode = null;
      await refreshProfile();
      await realtime.connect(_tokens!.accessToken);
    });
  }

  Future<void> logout() async {
    final refresh = _tokens?.refreshToken;
    if (refresh != null) {
      try {
        await _api.post(
          '/v1/auth/logout',
          body: {'refresh_token': refresh},
          authenticated: false,
          allowTokenRefresh: false,
        );
      } catch (_) {}
    }
    await _clearLocalSession();
    notifyListeners();
  }

  Future<String?> _refreshAccessToken() async {
    final inFlight = _refreshInFlight;
    if (inFlight != null) return inFlight;
    final future = _performRefresh();
    _refreshInFlight = future;
    try {
      return await future;
    } finally {
      _refreshInFlight = null;
    }
  }

  Future<String?> _performRefresh() async {
    final current = _tokens ?? await _store.read();
    if (current == null) return null;
    try {
      final response = await _api.post(
        '/v1/auth/refresh',
        body: {'refresh_token': current.refreshToken},
        authenticated: false,
        allowTokenRefresh: false,
      );
      await _saveTokenResponse(response['data'] as Map<String, dynamic>);
      final access = _tokens!.accessToken;
      try {
        await realtime.connect(access);
      } catch (_) {}
      return access;
    } catch (_) {
      await _clearLocalSession();
      notifyListeners();
      return null;
    }
  }

  Future<void> _saveTokenResponse(Map<String, dynamic> tokenData) async {
    _tokens = SessionTokens(
      accessToken: tokenData['access_token'].toString(),
      refreshToken: tokenData['refresh_token'].toString(),
    );
    await _store.save(_tokens!);
  }

  Future<void> _clearLocalSession() async {
    await realtime.disconnect();
    await _store.clear();
    _tokens = null;
    profile = null;
    challengeId = null;
    developmentCode = null;
  }

  Future<bool> _guard(Future<void> Function() fn) async {
    busy = true;
    error = null;
    notifyListeners();
    try {
      await fn();
      return true;
    } on ApiException catch (e) {
      error = e.message;
      return false;
    } catch (_) {
      error = 'Không thể kết nối máy chủ.';
      return false;
    } finally {
      busy = false;
      notifyListeners();
    }
  }

  @override
  void dispose() {
    unawaited(realtime.dispose());
    _api.close();
    super.dispose();
  }
}
