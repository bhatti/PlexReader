// SPDX-License-Identifier: LGPL-2.1-or-later
import '../models/article.dart';
import '../models/preferences.dart';
import 'api_client.dart';

class ArticleListResult {
  final List<Article> articles;
  final String? nextPageToken;
  final int? totalCount;

  const ArticleListResult({
    required this.articles,
    this.nextPageToken,
    this.totalCount,
  });
}

class ArticleService {
  static const _service = 'plexreader.v1.ArticleService';
  final ApiClient _client;

  ArticleService(this._client);

  Future<ArticleListResult> listArticles({
    String? feedId,
    String? folderId,
    bool? unreadOnly,
    bool? readOnly,
    bool? starredOnly,
    bool? savedForLaterOnly,
    bool? todayOnly,
    String? query,
    SortOrder? sortOrder,
    int pageSize = 50,
    String? pageToken,
  }) async {
    final res = await _client.post(_service, 'ListArticles', {
      if (feedId != null) 'feedId': feedId,
      if (folderId != null) 'folderId': folderId,
      if (unreadOnly != null) 'unreadOnly': unreadOnly,
      if (readOnly != null) 'readOnly': readOnly,
      if (starredOnly != null) 'starredOnly': starredOnly,
      if (savedForLaterOnly != null) 'savedForLaterOnly': savedForLaterOnly,
      if (todayOnly != null) 'todayOnly': todayOnly,
      if (query != null) 'query': query,
      if (sortOrder != null) 'sortOrder': sortOrder.value,
      'pageSize': pageSize,
      if (pageToken != null) 'pageToken': pageToken,
    });
    final list = res['articles'] as List<dynamic>? ?? [];
    return ArticleListResult(
      articles: list.map((e) => Article.fromJson(e as Map<String, dynamic>)).toList(),
      nextPageToken: res['nextPageToken'] as String?,
      totalCount: (res['totalCount'] as num?)?.toInt(),
    );
  }

  Future<Article> getArticle(String id) async {
    final res = await _client.post(_service, 'GetArticle', {'id': id});
    return Article.fromJson(res);
  }

  Future<Article> markAsRead(String id) async {
    final res = await _client.post(_service, 'MarkAsRead', {'id': id});
    return Article.fromJson(res);
  }

  Future<Article> markAsUnread(String id) async {
    final res = await _client.post(_service, 'MarkAsUnread', {'id': id});
    return Article.fromJson(res);
  }

  Future<int> markAllAsRead({
    List<String>? articleIds,
    String? feedId,
    String? folderId,
    bool? all,
  }) async {
    // Connect-JSON serializes proto oneof fields at the top level, camelCase.
    Map<String, dynamic> body;
    if (articleIds != null) {
      body = {'articleIds': articleIds};
    } else if (feedId != null) {
      body = {'feedId': feedId};
    } else if (folderId != null) {
      body = {'folderId': folderId};
    } else {
      body = {'all': true};
    }
    final res = await _client.post(_service, 'MarkAllAsRead', body);
    return (res['count'] as num?)?.toInt() ?? 0;
  }

  Future<Article> starArticle(String id, {required bool starred}) async {
    final res = await _client.post(_service, 'StarArticle', {'id': id, 'starred': starred});
    return Article.fromJson(res);
  }

  Future<Article> saveForLater(String id, {required bool saved}) async {
    final res = await _client.post(_service, 'SaveForLater', {'id': id, 'saved': saved});
    return Article.fromJson(res);
  }
}
