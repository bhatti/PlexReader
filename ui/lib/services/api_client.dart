// SPDX-License-Identifier: LGPL-2.1-or-later
import 'dart:convert';
import 'package:http/http.dart' as http;

// In dev: set via --dart-define=API_BASE_URL=http://localhost:8080
// In prod Docker: empty → same-origin (nginx proxies /plexreader.v1.* → backend)
const String _apiBase = String.fromEnvironment(
  'API_BASE_URL',
  defaultValue: 'http://localhost:8080',
);

class ConnectException implements Exception {
  final String code;
  final String message;
  const ConnectException(this.code, this.message);

  @override
  String toString() => 'ConnectException[$code]: $message';
}

class ApiClient {
  final http.Client _httpClient;
  final String baseUrl;

  ApiClient({http.Client? httpClient, String? baseUrl})
      : _httpClient = httpClient ?? http.Client(),
        baseUrl = baseUrl ?? _apiBase;

  Future<Map<String, dynamic>> post(
    String service,
    String method,
    Map<String, dynamic> body,
  ) async {
    final uri = Uri.parse('$baseUrl/$service/$method');
    try {
      final response = await _httpClient.post(
        uri,
        headers: {
          'Content-Type': 'application/json',
          'Connect-Protocol-Version': '1',
        },
        body: jsonEncode(body),
      );
      final decoded = jsonDecode(utf8.decode(response.bodyBytes)) as Map<String, dynamic>;
      if (response.statusCode == 200) {
        return decoded;
      }
      throw ConnectException(
        decoded['code'] as String? ?? 'unknown',
        decoded['message'] as String? ?? 'Request failed',
      );
    } catch (e) {
      if (e is ConnectException) rethrow;
      throw ConnectException('network_error', e.toString());
    }
  }

  void dispose() => _httpClient.close();
}
