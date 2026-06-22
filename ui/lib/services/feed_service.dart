// SPDX-License-Identifier: LGPL-2.1-or-later
import 'dart:convert';
import '../models/feed.dart';
import 'api_client.dart';

class FeedService {
  static const _service = 'plexreader.v1.FeedService';
  final ApiClient _client;

  FeedService(this._client);

  Future<List<Feed>> listFeeds({String? folderId, int pageSize = 200, String? pageToken}) async {
    final res = await _client.post(_service, 'ListFeeds', {
      if (folderId != null) 'folderId': folderId,
      'pageSize': pageSize,
      if (pageToken != null) 'pageToken': pageToken,
    });
    final list = res['feeds'] as List<dynamic>? ?? [];
    return list.map((e) => Feed.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<Feed> getFeed(String id) async {
    final res = await _client.post(_service, 'GetFeed', {'id': id});
    return Feed.fromJson(res);
  }

  Future<Feed> createFeed({
    required String title,
    required String xmlUrl,
    String? htmlUrl,
    String? folderId,
    String? iconUrl,
    int? refreshIntervalSeconds,
  }) async {
    final res = await _client.post(_service, 'CreateFeed', {
      'feed': {
        'title': title,
        'xmlUrl': xmlUrl,
        if (htmlUrl != null) 'htmlUrl': htmlUrl,
        if (folderId != null) 'folderId': folderId,
        if (iconUrl != null) 'iconUrl': iconUrl,
        if (refreshIntervalSeconds != null) 'refreshIntervalSeconds': refreshIntervalSeconds,
      },
    });
    return Feed.fromJson(res);
  }

  Future<Feed> updateFeed(Feed feed) async {
    final res = await _client.post(_service, 'UpdateFeed', {
      'feed': feed.toJson(),
    });
    return Feed.fromJson(res);
  }

  Future<void> deleteFeed(String id) async {
    await _client.post(_service, 'DeleteFeed', {'id': id});
  }

  Future<Map<String, dynamic>> importOPML(List<int> opmlContent) async {
    // Connect-JSON encodes proto `bytes` as standard base64.
    final encoded = base64Encode(opmlContent);
    return await _client.post(_service, 'ImportOPML', {
      'opmlContent': encoded,
    });
  }

  Future<List<int>> exportOPML() async {
    final res = await _client.post(_service, 'ExportOPML', {});
    final content = res['opmlContent'];
    // Connect-JSON returns proto `bytes` as base64.
    if (content is String && content.isNotEmpty) {
      return base64Decode(content);
    }
    return [];
  }

  Future<Feed> refreshFeed(String id) async {
    final res = await _client.post(_service, 'RefreshFeed', {'id': id});
    return Feed.fromJson(res);
  }
}
