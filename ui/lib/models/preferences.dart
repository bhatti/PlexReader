// SPDX-License-Identifier: LGPL-2.1-or-later

enum AppThemeMode {
  dark('THEME_DARK'),
  light('THEME_LIGHT');

  final String value;
  const AppThemeMode(this.value);

  static AppThemeMode fromValue(String v) =>
      AppThemeMode.values.firstWhere((e) => e.value == v, orElse: () => AppThemeMode.dark);
}

enum StartPage {
  today('START_PAGE_TODAY'),
  firstFolder('START_PAGE_FIRST_FOLDER'),
  all('START_PAGE_ALL');

  final String value;
  const StartPage(this.value);

  static StartPage fromValue(String v) =>
      StartPage.values.firstWhere((e) => e.value == v, orElse: () => StartPage.today);
}

enum ViewMode {
  magazine('VIEW_MODE_MAGAZINE'),
  titleOnly('VIEW_MODE_TITLE_ONLY'),
  cards('VIEW_MODE_CARDS'),
  article('VIEW_MODE_ARTICLE');

  final String value;
  const ViewMode(this.value);

  static ViewMode fromValue(String v) =>
      ViewMode.values.firstWhere((e) => e.value == v, orElse: () => ViewMode.magazine);
}

enum SortOrder {
  newestFirst('SORT_ORDER_NEWEST_FIRST'),
  oldestFirst('SORT_ORDER_OLDEST_FIRST');

  final String value;
  const SortOrder(this.value);

  static SortOrder fromValue(String v) =>
      SortOrder.values.firstWhere((e) => e.value == v, orElse: () => SortOrder.newestFirst);
}

class Preferences {
  final StartPage startPage;
  final ViewMode defaultView;
  final SortOrder defaultSort;
  final bool hideReadArticles;
  final AppThemeMode theme;

  const Preferences({
    this.startPage = StartPage.today,
    this.defaultView = ViewMode.magazine,
    this.defaultSort = SortOrder.newestFirst,
    this.hideReadArticles = true,
    this.theme = AppThemeMode.dark,
  });

  factory Preferences.fromJson(Map<String, dynamic> json) {
    final prefs = json['preferences'] as Map<String, dynamic>? ?? json;
    return Preferences(
      startPage: StartPage.fromValue(prefs['startPage'] as String? ?? ''),
      defaultView: ViewMode.fromValue(prefs['defaultView'] as String? ?? ''),
      defaultSort: SortOrder.fromValue(prefs['defaultSort'] as String? ?? ''),
      hideReadArticles: prefs['hideReadArticles'] as bool? ?? true,
      theme: AppThemeMode.fromValue(prefs['theme'] as String? ?? ''),
    );
  }

  Map<String, dynamic> toJson() => {
    'startPage': startPage.value,
    'defaultView': defaultView.value,
    'defaultSort': defaultSort.value,
    'hideReadArticles': hideReadArticles,
    'theme': theme.value,
  };

  Preferences copyWith({
    StartPage? startPage,
    ViewMode? defaultView,
    SortOrder? defaultSort,
    bool? hideReadArticles,
    AppThemeMode? theme,
  }) {
    return Preferences(
      startPage: startPage ?? this.startPage,
      defaultView: defaultView ?? this.defaultView,
      defaultSort: defaultSort ?? this.defaultSort,
      hideReadArticles: hideReadArticles ?? this.hideReadArticles,
      theme: theme ?? this.theme,
    );
  }
}
