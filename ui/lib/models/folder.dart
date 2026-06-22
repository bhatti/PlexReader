// SPDX-License-Identifier: LGPL-2.1-or-later
class Folder {
  final String id;
  final String name;
  final String? parentId;
  final int position;
  final int unreadCount;

  const Folder({
    required this.id,
    required this.name,
    this.parentId,
    this.position = 0,
    this.unreadCount = 0,
  });

  factory Folder.fromJson(Map<String, dynamic> json) {
    return Folder(
      id: json['id'] as String? ?? '',
      name: json['name'] as String? ?? '',
      parentId: json['parentId'] as String?,
      position: (json['position'] as num?)?.toInt() ?? 0,
      unreadCount: (json['unreadCount'] as num?)?.toInt() ?? 0,
    );
  }

  Map<String, dynamic> toJson() => {
    'id': id,
    'name': name,
    if (parentId != null) 'parentId': parentId,
    'position': position,
    'unreadCount': unreadCount,
  };

  Folder copyWith({
    String? id,
    String? name,
    String? parentId,
    int? position,
    int? unreadCount,
  }) {
    return Folder(
      id: id ?? this.id,
      name: name ?? this.name,
      parentId: parentId ?? this.parentId,
      position: position ?? this.position,
      unreadCount: unreadCount ?? this.unreadCount,
    );
  }

  @override
  bool operator ==(Object other) => other is Folder && other.id == id;

  @override
  int get hashCode => id.hashCode;
}
