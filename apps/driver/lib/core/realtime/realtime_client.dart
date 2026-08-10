import 'dart:async';
import 'dart:convert';
import 'dart:io';

import '../config/app_config.dart';

class RealtimeClient {
  WebSocket? _socket;
  final _events = StreamController<Map<String, dynamic>>.broadcast();
  Stream<Map<String, dynamic>> get events => _events.stream;

  Future<void> connect(String accessToken) async {
    await disconnect();
    final socket = await WebSocket.connect(
      AppConfig.realtimeUri().toString(),
      headers: {HttpHeaders.authorizationHeader: 'Bearer $accessToken'},
    );
    _socket = socket;
    socket.listen((message) {
      if (message is! String) return;
      try {
        final decoded = jsonDecode(message);
        if (decoded is Map<String, dynamic>) _events.add(decoded);
      } catch (_) {}
    }, onError: _events.addError, onDone: () => _socket = null);
  }

  Future<void> disconnect() async {
    final socket = _socket;
    _socket = null;
    if (socket != null) await socket.close();
  }

  Future<void> dispose() async {
    await disconnect();
    await _events.close();
  }
}
