// SPDX-License-Identifier: LGPL-2.1-or-later
import '../models/preferences.dart';
import 'api_client.dart';

class PreferencesService {
  static const _service = 'plexreader.v1.PreferencesService';
  final ApiClient _client;

  PreferencesService(this._client);

  Future<Preferences> getPreferences() async {
    final res = await _client.post(_service, 'GetPreferences', {});
    return Preferences.fromJson(res);
  }

  Future<Preferences> updatePreferences(Preferences prefs) async {
    final res = await _client.post(_service, 'UpdatePreferences', {
      'preferences': prefs.toJson(),
    });
    return Preferences.fromJson(res);
  }
}
