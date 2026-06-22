// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/folder.dart';
import '../services/folder_service.dart';
import 'api_client_provider.dart';

class FolderNotifier extends StateNotifier<AsyncValue<List<Folder>>> {
  final FolderService _service;

  FolderNotifier(this._service) : super(const AsyncValue.loading()) {
    loadFolders();
  }

  Future<void> loadFolders() async {
    state = const AsyncValue.loading();
    try {
      final folders = await _service.listFolders();
      folders.sort((a, b) => a.position.compareTo(b.position));
      state = AsyncValue.data(folders);
    } catch (e, st) {
      state = AsyncValue.error(e, st);
    }
  }

  Future<Folder?> createFolder(String name, {String? parentId}) async {
    try {
      final folder = await _service.createFolder(name, parentId: parentId);
      final current = state.valueOrNull ?? [];
      state = AsyncValue.data([...current, folder]);
      return folder;
    } catch (_) {
      return null;
    }
  }

  Future<bool> updateFolder(Folder folder) async {
    try {
      final updated = await _service.updateFolder(folder);
      final current = state.valueOrNull ?? [];
      state = AsyncValue.data(
        current.map((f) => f.id == updated.id ? updated : f).toList(),
      );
      return true;
    } catch (_) {
      return false;
    }
  }

  Future<bool> deleteFolder(String id) async {
    try {
      await _service.deleteFolder(id);
      final current = state.valueOrNull ?? [];
      state = AsyncValue.data(current.where((f) => f.id != id).toList());
      return true;
    } catch (_) {
      return false;
    }
  }

  void updateUnreadCount(String folderId, int delta) {
    final current = state.valueOrNull;
    if (current == null) return;
    state = AsyncValue.data(
      current.map((f) {
        if (f.id == folderId) {
          return f.copyWith(unreadCount: (f.unreadCount + delta).clamp(0, 999999));
        }
        return f;
      }).toList(),
    );
  }

  void decrementUnread(String folderId) => updateUnreadCount(folderId, -1);
}

final folderProvider =
    StateNotifierProvider<FolderNotifier, AsyncValue<List<Folder>>>((ref) {
  return FolderNotifier(ref.watch(folderServiceProvider));
});
