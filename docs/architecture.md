# TAWS Architecture

## Overview

TAWS (Termux Android Web Server) is a lightweight, self-managing local web development environment for Android built on Termux.

## Core Stack

- **Web Server:** Nginx
- **Language:** PHP
- **Database:** MariaDB
- **Package Manager:** Composer
- **DB Management:** Adminer

## Directory Layout

```
taws/
├── cmd/            # CLI entry points
│   └── taws/
├── internal/       # Private application code
├── pkg/            # Public library code
├── scripts/        # Build and utility scripts
├── templates/      # Configuration templates
├── docs/           # Documentation
└── tests/          # Test files
```

## Runtime Directories

```
~/web/              # User projects
~/.taws/            # TAWS runtime data
├── configs/
├── state/
├── logs/
├── backups/
├── certs/
├── plugins/
└── cache/
```

## Domain Convention

- Default: `project.test`
- Examples: `blog.test`, `api.test`, `shop.test`
