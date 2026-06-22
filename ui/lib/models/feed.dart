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
  };

  Feed copyWith({
    String? id,
    String? title,
    String? xmlUrl,
    String? htmlUrl,
    String? folderId,
    String? description,
    String? iconUrl,
    int? refreshIntervalSeconds,
    int? unreadCount,
    String? lastFetchedTime,
    String? lastError,
    int? errorCount,
  }) {
    return Feed(
      id: id ?? this.id,
      title: title ?? this.title,
      xmlUrl: xmlUrl ?? this.xmlUrl,
      htmlUrl: htmlUrl ?? this.htmlUrl,
      folderId: folderId ?? this.folderId,
      description: description ?? this.description,
      iconUrl: iconUrl ?? this.iconUrl,
      refreshIntervalSeconds: refreshIntervalSeconds ?? this.refreshIntervalSeconds,
      unreadCount: unreadCount ?? this.unreadCount,
      lastFetchedTime: lastFetchedTime ?? this.lastFetchedTime,
      lastError: lastError ?? this.lastError,
      errorCount: errorCount ?? this.errorCount,
    );
  }

  @override
  bool operator ==(Object other) => other is Feed && other.id == id;

  @override
  int get hashCode => id.hashCode;
}
