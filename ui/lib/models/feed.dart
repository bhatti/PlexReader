// SPDX-License-Identifier: LGPL-2.1-or-later
class Feed {
  final String id;
  final String title;
  final String xmlUrl;
  final String? htmlUrl;
  final String? folderId;
  final String? description;
  final String? iconUrl;
  final int refreshIntervalSeconds;
  final int unreadCount;
  final bool isFavorite;
  final String? lastFetchedTime;
  final String? lastError;
  final int errorCount;

  const Feed({
    required this.id,
    required this.title,
    required this.xmlUrl,
    this.htmlUrl,
    this.folderId,
    this.description,
    this.iconUrl,
    this.refreshIntervalSeconds = 3600,
    this.unreadCount = 0,
    this.isFavorite = false,
    this.lastFetchedTime,
    this.lastError,
    this.errorCount = 0,
  });

  bool get hasError => lastError != null && lastError!.isNotEmpty;

  factory Feed.fromJson(Map<String, dynamic> json) {
    return Feed(
      id: json['id'] as String? ?? '',
      title: json['title'] as String? ?? '',
      xmlUrl: json['xmlUrl'] as String? ?? '',
      htmlUrl: json['htmlUrl'] as String?,
      folderId: json['folderId'] as String?,
      description: json['description'] as String?,
      iconUrl: json['iconUrl'] as String?,
      refreshIntervalSeconds: (json['refreshIntervalSeconds'] as num?)?.toInt() ?? 3600,
      unreadCount: (json['unreadCount'] as num?)?.toInt() ?? 0,
      isFavorite: json['isFavorite'] as bool? ?? false,
      lastFetchedTime: json['lastFetchedTime'] as String?,
      lastError: json['lastError'] as String?,
      errorCount: (json['errorCount'] as num?)?.toInt() ?? 0,
    );
  }

  Map<String, dynamic> toJson() => {
    'id': id,
    'title': title,
    'xmlUrl': xmlUrl,
    if (htmlUrl != null) 'htmlUrl': htmlUrl,
    if (folderId != null) 'folderId': folderId,
    if (description != null) 'description': description,
    if (iconUrl != null) 'iconUrl': iconUrl,
    'refreshIntervalSeconds': refreshIntervalSeconds,
    'isFavorite': isFavorite,
  };

  static const _keep = Object();

  Feed copyWith({
    String? id,
    String? title,
    String? xmlUrl,
    Object? htmlUrl = _keep,
    Object? folderId = _keep,
    Object? description = _keep,
    Object? iconUrl = _keep,
    int? refreshIntervalSeconds,
    int? unreadCount,
    bool? isFavorite,
    Object? lastFetchedTime = _keep,
    Object? lastError = _keep,
    int? errorCount,
  }) {
    return Feed(
      id: id ?? this.id,
      title: title ?? this.title,
      xmlUrl: xmlUrl ?? this.xmlUrl,
      htmlUrl: identical(htmlUrl, _keep) ? this.htmlUrl : htmlUrl as String?,
      folderId: identical(folderId, _keep) ? this.folderId : folderId as String?,
      description: identical(description, _keep) ? this.description : description as String?,
      iconUrl: identical(iconUrl, _keep) ? this.iconUrl : iconUrl as String?,
      refreshIntervalSeconds: refreshIntervalSeconds ?? this.refreshIntervalSeconds,
      unreadCount: unreadCount ?? this.unreadCount,
      isFavorite: isFavorite ?? this.isFavorite,
      lastFetchedTime: identical(lastFetchedTime, _keep) ? this.lastFetchedTime : lastFetchedTime as String?,
      lastError: identical(lastError, _keep) ? this.lastError : lastError as String?,
      errorCount: errorCount ?? this.errorCount,
    );
  }

  @override
  bool operator ==(Object other) => other is Feed && other.id == id;

  @override
  int get hashCode => id.hashCode;
}
