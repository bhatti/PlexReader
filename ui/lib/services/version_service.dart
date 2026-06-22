// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:http/http.dart' as http;
import 'dart:convert';

class VersionService {
  final String _baseUrl;

  VersionService(this._baseUrl);

  Future<String> getVersion() async {
    try {
      final uri = Uri.parse('$_baseUrl/version');
      final response = await http.get(uri).timeout(const Duration(seconds: 5));
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        return data['version'] as String? ?? 'unknown';
      }
    } catch (_) {}
    return 'unknown';
  }
}
