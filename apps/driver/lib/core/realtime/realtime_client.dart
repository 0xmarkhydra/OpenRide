import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:math';

import '../config/app_config.dart';

class RealtimeClient {
  WebSocket? _socket;
  final _events = StreamController<Map<String, dynamic>>.broadcast();
  Timer? _retry;
  String? _accessToken;
  bool _shouldReconnect = false;
  bool _connecting = false;
  int _attempt = 0;

  Stream<Map<String, dynamic>> get events => _events.stream;

  Future<void> connect(String accessToken) async {
    _accessToken = accessToken;
    _shouldReconnect = true;
    _retry?.cancel();
    await _closeSocket();
    await _open();
  }

  Future<void> _open() async {
    if (!_shouldReconnect || _connecting || _socket != null) return;
    final token = _accessToken;
    if (token == null || token.isEmpty) return;
    _connecting = true;
    try {
      final socket = await WebSocket.connect(
        AppConfig.realtimeUri().toString(),
        headers: {HttpHeaders.authorizationHeader: 'Bearer $token'},
      );
      if (!_shouldReconnect) {
        await socket.close();
        return;
      }
      _socket = socket;
      _attempt = 0;
      socket.listen(
        (message) {
          if (message is! String) return;
          try {
            final decoded = jsonDecode(message);
            if (decoded is Map<String, dynamic> && !_events.isClosed) {
              _events.add(decoded);
            }
          } catch (_) {}
        },
        onError: (Object error, StackTrace stack) {
          if (!_events.isClosed) _events.addError(error, stack);
        },
        onDone: () {
          if (identical(_socket, socket)) _socket = null;
          _scheduleReconnect();
        },
        cancelOnError: false,
      );
    } catch (_) {
      _socket = null;
      _scheduleReconnect();
    } finally {
      _connecting = false;
    }
  }

  void _scheduleReconnect() {
    if (!_shouldReconnect || _events.isClosed || _retry?.isActive == true) {
      return;
    }
    final seconds = min(15, 1 << min(_attempt, 4));
    _attempt++;
    _retry = Timer(Duration(seconds: seconds), () => unawaited(_open()));
  }

  Future<void> _closeSocket() async {
    final socket = _socket;
    _socket = null;
    if (socket != null) {
      try {
        await socket.close();
      } catch (_) {}
    }
  }

  Future<void> disconnect() async {
    _shouldReconnect = false;
    _accessToken = null;
    _retry?.cancel();
    _retry = null;
    _attempt = 0;
    await _closeSocket();
  }

  Future<void> dispose() async {
    await disconnect();
    await _events.close();
  }
}
