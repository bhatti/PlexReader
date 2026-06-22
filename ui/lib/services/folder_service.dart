// SPDX-License-Identifier: LGPL-2.1-or-later
import '../models/folder.dart';
import 'api_client.dart';

class FolderService {
  static const _service = 'plexreader.v1.FolderService';
  final ApiClient _client;

  FolderService(this._client);

  Future<List<Folder>> listFolders() async {
    final res = await _client.post(_service, 'ListFolders', {});
    final list = res['folders'] as List<dynamic>? ?? [];
    return list.map((e) => Folder.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<Folder> createFolder(String name, {String? parentId, int? position}) async {
    final res = await _client.post(_service, 'CreateFolder', {
      'folder': {
        'name': name,
        if (parentId != null) 'parentId': parentId,
        if (position != null) 'position': position,
      },
    });
    return Folder.fromJson(res);
  }

  Future<Folder> updateFolder(Folder folder) async {
    final res = await _client.post(_service, 'UpdateFolder', {
      'folder': folder.toJson(),
    });
    return Folder.fromJson(res);
  }

  Future<void> deleteFolder(String id) async {
    await _client.post(_service, 'DeleteFolder', {'id': id});
  }

  Future<List<Folder>> reorderFolders(List<String> folderIds) async {
    final res = await _client.post(_service, 'ReorderFolders', {
      'folderIds': folderIds,
    });
    final list = res['folders'] as List<dynamic>? ?? [];
    return list.map((e) => Folder.fromJson(e as Map<String, dynamic>)).toList();
  }
}
