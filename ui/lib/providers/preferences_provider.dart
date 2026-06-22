// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/preferences.dart';
import '../services/preferences_service.dart';
import 'api_client_provider.dart';

class PreferencesNotifier extends StateNotifier<AsyncValue<Preferences>> {
  final PreferencesService _service;

  PreferencesNotifier(this._service) : super(const AsyncValue.loading()) {
    load();
  }

  Future<void> load() async {
    state = const AsyncValue.loading();
    try {
      final prefs = await _service.getPreferences();
      state = AsyncValue.data(prefs);
    } catch (e, _) {
      // Default prefs on error
      state = const AsyncValue.data(Preferences());
    }
  }

  Future<void> update(Preferences prefs) async {
    final prev = state;
    state = AsyncValue.data(prefs);
    try {
      final updated = await _service.updatePreferences(prefs);
      state = AsyncValue.data(updated);
    } catch (_) {
      state = prev;
    }
  }
}

final preferencesProvider =
    StateNotifierProvider<PreferencesNotifier, AsyncValue<Preferences>>((ref) {
  return PreferencesNotifier(ref.watch(preferencesServiceProvider));
});
