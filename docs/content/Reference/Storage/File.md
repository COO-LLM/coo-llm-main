---
sidebar_position: 6
tags: [reference, storage, file, json]
---

# File-based Storage

Simple JSON file storage for development and small-scale COO-LLM deployments.

## Configuration

```yaml
storage:
  runtime:
    type: "file"
    path: "./data/runtime.json"
```

## Features

- **Simple JSON Files**: Easy to read and edit manually
- **No External Dependencies**: Works without database servers
- **Development Friendly**: Quick setup and debugging
- **Backup/Restore**: Simple file copy operations
- **Version Control**: Can be committed to git
- **Cross-Platform**: Works on all operating systems

## Data Structure

JSON file structure:
```json
{
  "usage": {
    "openai:key1:req": 45,
    "openai:key1:tokens": 15000,
    "gemini:key2:req": 23
  },
  "cache": {
    "cache_key_1": {
      "value": "cached_response",
      "expiry": 1640995260
    }
  }
}
```

## Implementation Details

- **File Locking**: Prevents concurrent access corruption
- **Atomic Writes**: Temporary files for safe updates
- **JSON Serialization**: Human-readable data format
- **Memory Caching**: In-memory cache for performance
- **Auto-Create Directories**: Creates parent directories as needed
- **Error Handling**: Graceful handling of file system errors
