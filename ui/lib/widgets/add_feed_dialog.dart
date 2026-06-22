// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/feed_provider.dart';
import '../providers/folder_provider.dart';
import '../theme/app_theme.dart';

/// Dialog for adding a new RSS/Atom feed, with optional folder assignment.
class AddFeedDialog extends ConsumerStatefulWidget {
  const AddFeedDialog({super.key});

  @override
  ConsumerState<AddFeedDialog> createState() => _AddFeedDialogState();
}

class _AddFeedDialogState extends ConsumerState<AddFeedDialog> {
  final _formKey = GlobalKey<FormState>();
  final _urlController = TextEditingController();
  final _titleController = TextEditingController();
  String? _selectedFolderId;
  bool _isLoading = false;
  String? _error;

  @override
  void dispose() {
    _urlController.dispose();
    _titleController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final folders = ref.watch(folderProvider).valueOrNull ?? [];

    return AlertDialog(
      title: const Text('Follow a Source'),
      content: SizedBox(
        width: 400,
        child: Form(
          key: _formKey,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              TextFormField(
                controller: _urlController,
                autofocus: true,
                decoration: const InputDecoration(
                  labelText: 'Feed URL',
                  hintText: 'https://example.com/feed.xml',
                  prefixIcon: Icon(Icons.rss_feed, size: 18),
                ),
                keyboardType: TextInputType.url,
                validator: (v) {
                  if (v == null || v.trim().isEmpty) return 'URL is required';
                  final uri = Uri.tryParse(v.trim());
                  if (uri == null || !uri.hasScheme) return 'Enter a valid URL';
                  return null;
                },
                onFieldSubmitted: (_) => _submit(),
              ),
              const SizedBox(height: 12),
              TextFormField(
                controller: _titleController,
                decoration: const InputDecoration(
                  labelText: 'Title (optional)',
                  hintText: 'Leave blank to use feed title',
                ),
                onFieldSubmitted: (_) => _submit(),
              ),
              const SizedBox(height: 12),
              DropdownButtonFormField<String?>(
                value: _selectedFolderId,
                decoration: const InputDecoration(
                  labelText: 'Folder (optional)',
                  prefixIcon: Icon(Icons.folder_outlined, size: 18),
                ),
                dropdownColor: AppColors.surface,
                items: [
                  const DropdownMenuItem(
                      value: null, child: Text('No folder')),
                  ...folders.map(
                    (f) => DropdownMenuItem(value: f.id, child: Text(f.name)),
                  ),
                ],
                onChanged: (val) => setState(() => _selectedFolderId = val),
              ),
              if (_error != null) ...[
                const SizedBox(height: 12),
                Text(
                  _error!,
                  style: const TextStyle(
                      color: AppColors.error, fontSize: 13),
                ),
              ],
            ],
          ),
        ),
      ),
      actions: [
        TextButton(
          onPressed: _isLoading ? null : () => Navigator.of(context).pop(),
          child: const Text('Cancel'),
        ),
        ElevatedButton(
          onPressed: _isLoading ? null : _submit,
          child: _isLoading
              ? const SizedBox(
                  width: 16,
                  height: 16,
                  child: CircularProgressIndicator(
                      strokeWidth: 2, color: Colors.white),
                )
              : const Text('Follow'),
        ),
      ],
    );
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() {
      _isLoading = true;
      _error = null;
    });
    try {
      final title = _titleController.text.trim();
      final feed = await ref.read(feedProvider.notifier).createFeed(
            title: title.isNotEmpty ? title : _urlController.text.trim(),
            xmlUrl: _urlController.text.trim(),
            folderId: _selectedFolderId,
          );
      if (!mounted) return;
      if (feed != null) {
        Navigator.of(context).pop();
      } else {
        setState(() =>
            _error = 'Failed to add feed. Check the URL and try again.');
      }
    } catch (e) {
      if (mounted) setState(() => _error = e.toString());
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }
}
