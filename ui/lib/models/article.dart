// SPDX-License-Identifier: LGPL-2.1-or-later
class Article {
  final String id;
  final String feedId;
  final String title;
  final String? link;
  final String? content;
  final String? summary;
  final String? author;
  final String? publishedTime;
  final String? guid;
  final String? thumbnailUrl;
  final bool isRead;
  final bool isStarred;
  final bool isSavedForLater;
  final String? readAt;
  // Denormalized for display
  final String? feedTitle;
  final String? feedIconUrl;

  const Article({
    required this.id,
    required this.feedId,
    required this.title,
    this.link,
    this.content,
    this.summary,
    this.author,
    this.publishedTime,
    this.guid,
    this.thumbnailUrl,
    this.isRead = false,
    this.isStarred = false,
    this.isSavedForLater = false,
    this.readAt,
    this.feedTitle,
    this.feedIconUrl,
  });

  factory Article.fromJson(Map<String, dynamic> json) {
    return Article(
      id: json['id'] as String? ?? '',
      feedId: json['feedId'] as String? ?? '',
      title: json['title'] as String? ?? '(no title)',
      link: json['link'] as String?,
      content: json['content'] as String?,
      summary: json['summary'] as String?,
      author: json['author'] as String?,
      publishedTime: json['publishedTime'] as String?,
      guid: json['guid'] as String?,
      thumbnailUrl: json['thumbnailUrl'] as String?,
      isRead: json['isRead'] as bool? ?? false,
      isStarred: json['isStarred'] as bool? ?? false,
      isSavedForLater: json['isSavedForLater'] as bool? ?? false,
      readAt: json['readAt'] as String?,
      feedTitle: json['feedTitle'] as String?,
      feedIconUrl: json['feedIconUrl'] as String?,
    );
  }

  Map<String, dynamic> toJson() => {
    'id': id,
    'feedId': feedId,
    'title': title,
    if (link != null) 'link': link,
    if (content != null) 'content': content,
    if (summary != null) 'summary': summary,
    if (author != null) 'author': author,
    if (publishedTime != null) 'publishedTime': publishedTime,
    if (guid != null) 'guid': guid,
    if (thumbnailUrl != null) 'thumbnailUrl': thumbnailUrl,
    'isRead': isRead,
    'isStarred': isStarred,
    'isSavedForLater': isSavedForLater,
    if (readAt != null) 'readAt': readAt,
  };

  Article copyWith({
    String? id,
    String? feedId,
    String? title,
    String? link,
    String? content,
    String? summary,
    String? author,
    String? publishedTime,
    String? guid,
    String? thumbnailUrl,
    bool? isRead,
    bool? isStarred,
    bool? isSavedForLater,
    String? readAt,
    String? feedTitle,
    String? feedIconUrl,
  }) {
    return Article(
      id: id ?? this.id,
      feedId: feedId ?? this.feedId,
      title: title ?? this.title,
      link: link ?? this.link,
      content: content ?? this.content,
      summary: summary ?? this.summary,
      author: author ?? this.author,
      publishedTime: publishedTime ?? this.publishedTime,
      guid: guid ?? this.guid,
      thumbnailUrl: thumbnailUrl ?? this.thumbnailUrl,
      isRead: isRead ?? this.isRead,
      isStarred: isStarred ?? this.isStarred,
      isSavedForLater: isSavedForLater ?? this.isSavedForLater,
      readAt: readAt ?? this.readAt,
      feedTitle: feedTitle ?? this.feedTitle,
      feedIconUrl: feedIconUrl ?? this.feedIconUrl,
    );
  }

  @override
  bool operator ==(Object other) => other is Article && other.id == id;

  @override
  int get hashCode => id.hashCode;
}
