// SPDX-License-Identifier: LGPL-2.1-or-later
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/folder_provider.dart';
import '../theme/app_theme.dart';

/// Dialog for creating a new folder (category) to organize feeds.
class CreateFolderDialog extends ConsumerStatefulWidget {
  const CreateFolderDialog({super.key});

  @override
  ConsumerState<CreateFolderDialog> createState() =>
      _CreateFolderDialogState();
}

class _CreateFolderDialogState extends ConsumerState<CreateFolderDialog> {
  final _formKey = GlobalKey<FormState>();
  final _nameController = TextEditingController();
  bool _isLoading = false;
  String? _error;

  @override
  void dispose() {
    _nameController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('Create New Folder'),
      content: SizedBox(
        width: 360,
        child: Form(
          key: _formKey,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextFormField(
                controller: _nameController,
                autofocus: true,
                decoration: const InputDecoration(
                  labelText: 'Folder name',
                  hintText: 'e.g. Tech, News, Sports',
                  prefixIcon: Icon(Icons.folder_outlined, size: 18),
                ),
                validator: (v) {
                  if (v == null || v.trim().isEmpty) return 'Name is required';
                  if (v.trim().length > 64) {
                    return 'Name must be 64 characters or fewer';
                  }
                  return null;
                },
                onFieldSubmitted: (_) => _submit(),
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
              : const Text('Create'),
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
      final folder = await ref
          .read(folderProvider.notifier)
          .createFolder(_nameController.text.trim());
      if (!mounted) return;
      if (folder != null) {
        Navigator.of(context).pop();
      } else {
        setState(
            () => _error = 'Failed to create folder. Please try again.');
      }
    } catch (e) {
      if (mounted) setState(() => _error = e.toString());
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }
}
