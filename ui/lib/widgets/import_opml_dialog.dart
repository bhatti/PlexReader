// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:file_picker/file_picker.dart';
import '../providers/feed_provider.dart';
import '../providers/folder_provider.dart';
import '../theme/app_theme.dart';

/// Dialog for importing an OPML file to bulk-add feeds and folders.
class ImportOpmlDialog extends ConsumerStatefulWidget {
  const ImportOpmlDialog({super.key});

  @override
  ConsumerState<ImportOpmlDialog> createState() => _ImportOpmlDialogState();
}

class _ImportOpmlDialogState extends ConsumerState<ImportOpmlDialog> {
  PlatformFile? _selectedFile;
  bool _isLoading = false;
  String? _error;
  bool _success = false;

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('Import OPML'),
      content: SizedBox(
        width: 400,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Import an OPML file to add all your feeds and folders at once.',
              style: TextStyle(color: AppColors.textSecondary, fontSize: 14),
            ),
            const SizedBox(height: 20),
            OutlinedButton.icon(
              onPressed: _isLoading || _success ? null : _pickFile,
              icon: const Icon(Icons.upload_file, size: 18),
              label: Text(
                _selectedFile != null
                    ? _selectedFile!.name
                    : 'Choose OPML file',
              ),
            ),
            if (_error != null) ...[
              const SizedBox(height: 12),
              Text(
                _error!,
                style: const TextStyle(color: AppColors.error, fontSize: 13),
              ),
            ],
            if (_success) ...[
              const SizedBox(height: 12),
              const Row(
                children: [
                  Icon(Icons.check_circle_outline,
                      color: Colors.green, size: 16),
                  SizedBox(width: 8),
                  Text(
                    'Import successful! Feeds have been added.',
                    style: TextStyle(color: Colors.green, fontSize: 13),
                  ),
                ],
              ),
            ],
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: Text(_success ? 'Done' : 'Cancel'),
        ),
        if (!_success)
          ElevatedButton(
            onPressed: (_isLoading || _selectedFile == null) ? null : _import,
            child: _isLoading
                ? const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(
                        strokeWidth: 2, color: Colors.white),
                  )
                : const Text('Import'),
          ),
      ],
    );
  }

  Future<void> _pickFile() async {
    final result = await FilePicker.platform.pickFiles(
      type: FileType.custom,
      allowedExtensions: ['opml', 'xml'],
      withData: true,
    );
    if (result != null && result.files.isNotEmpty) {
      setState(() {
        _selectedFile = result.files.first;
        _error = null;
      });
    }
  }

  Future<void> _import() async {
    if (_selectedFile?.bytes == null) {
      setState(() => _error = 'No file selected or file has no content');
      return;
    }
    setState(() {
      _isLoading = true;
      _error = null;
    });
    try {
      final bytes = _selectedFile!.bytes!.toList();
      final ok = await ref.read(feedProvider.notifier).importOPML(bytes);
      if (!mounted) return;
      if (ok) {
        // Reload folders so newly imported folder groups appear in the sidebar.
        ref.read(folderProvider.notifier).loadFolders();
        setState(() {
          _isLoading = false;
          _success = true;
        });
      } else {
        setState(() {
          _isLoading = false;
          _error =
              'Import failed. Please check the file format and try again.';
        });
      }
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _isLoading = false;
        _error = 'Import failed: $e';
      });
    }
  }
}
