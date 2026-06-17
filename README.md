# TAWS - Termux Android Web Server

A lightweight, self-managing local web development environment for Android built on Termux.

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Android%20%7C%20Linux-lightgrey)](#installation)

## Overview

## Table of Contents

**Getting Started** — [Overview](#overview) · [Features](#features) · [Requirements](#requirements) · [Installation](#installation)

**Usage** — [Setup](#setup) · [Commands](#commands) · [Project Management](#project-management) · [Databases](#databases)

**Features** — [Updates](#updates) · [Backups](#backups) · [Rollback](#rollback) · [SSL](#ssl) · [Plugins](#plugins)

**Reference** — [Uninstalling TAWS](#uninstalling-taws) · [Security](#security) · [Troubleshooting](#troubleshooting) · [Directory Layout](#directory-layout)

**Development** — [Development Guide](#development-guide) · [Contribution Guide](#contribution-guide) · [FAQ](#faq) · [License](#license) · [Acknowledgments](#acknowledgments)

TAWS provides a XAMPP-like experience while remaining lightweight, scriptable, terminal-friendly, and native to the Android/Termux ecosystem. Users should only need to install TAWS - it automatically detects missing dependencies, installs required packages, configures services, generates project configurations, manages updates, provides diagnostics, and offers a rich terminal dashboard.

## Features

- **Automatic Setup**: Detects and installs all required dependencies
- **Service Management**: Start, stop, restart, and monitor Nginx, PHP-FPM, MariaDB with automatic port conflict resolution
- **Project Management**: Create, delete, and list web projects with automatic virtual host generation
- **Database Management**: Create and manage MariaDB databases with Adminer web interface
- **Configuration Management**: Automatic configuration generation with override support
- **Backup & Rollback**: Snapshot and restore system configurations
- **Plugin System**: Extend functionality with plugins
- **Terminal Dashboard**: Real-time monitoring of all services and system resources
- **Diagnostics**: Comprehensive system health checks with `taws doctor`

---

## Requirements

- Android device with Termux installed
- Minimum 1GB RAM
- 500MB free storage
- Internet connection (for initial setup)

---

## Installation

TAWS can be installed on Termux (Android) or standard Linux distributions.

### Termux Installation (Android)

```bash
# Install Termux from F-Droid or Google Play Store

# Update Termux packages
pkg update && pkg upgrade

# Download and install TAWS
wget https://github.com/mnxhh/taws/releases/latest/download/taws-linux-arm64
chmod +x taws-linux-arm64
mv taws-linux-arm64 $PREFIX/bin/taws

# Run initial setup
taws setup
```

### Linux Installation

TAWS can also be used on standard Linux distributions for development and testing.

#### Ubuntu/Debian

<details>
<summary>Expand Ubuntu/Debian instructions</summary>

```bash
# Install dependencies
sudo apt update
sudo apt install -y golang git make gcc nginx php php-fpm php-mysql mariadb-server

# Clone and build
git clone https://github.com/mnxhh/taws.git
cd taws
go build -o taws ./cmd/taws

# Run setup
./taws setup
```

</details>

#### Fedora/RHEL

<details>
<summary>Expand Fedora/RHEL instructions</summary>

```bash
# Install dependencies
sudo dnf install -y golang git make gcc nginx php php-fpm php-mysqlnd mariadb-server

# Clone and build
git clone https://github.com/mnxhh/taws.git
cd taws
go build -o taws ./cmd/taws

# Run setup
./taws setup
```

</details>

#### Arch Linux

<details>
<summary>Expand Arch Linux instructions</summary>

```bash
# Install dependencies
sudo pacman -S go git make gcc nginx php php-fpm mariadb

# Clone and build
git clone https://github.com/mnxhh/taws.git
cd taws
go build -o taws ./cmd/taws

# Run setup
./taws setup
```

</details>

#### Linux Mint

<details>
<summary>Expand Linux Mint instructions</summary>

> **Note:** Linux Mint is based on Ubuntu, so use the Ubuntu instructions above.

```bash
# Install dependencies
sudo apt update
sudo apt install -y golang git make gcc nginx php php-fpm php-mysql mariadb-server

# Clone and build
git clone https://github.com/mnxhh/taws.git
cd taws
go build -o taws ./cmd/taws

# Run setup
./taws setup
```

</details>

### Development Installation

```bash
# Clone the repository
git clone https://github.com/mnxhh/taws.git
cd taws

# Build from source
go build -o taws ./cmd/taws

# Run setup
./taws setup
```

## Setup

The `taws setup` command performs the following:

1. Verifies the environment (Termux or standard Linux)
2. Detects the package manager
3. Checks for required packages (Nginx, PHP, MariaDB, Composer)
4. Creates TAWS directory structure (~/.taws/)
5. Generates default configuration files
6. Provides installation instructions for missing packages

```bash
taws setup
```

## Commands

Every TAWS command at your fingertips.

| Command | Description | Category |
|---------|-------------|----------|
| `taws setup` | Initial setup and installation | System |
| `taws doctor` | Check system dependencies and configuration | System |
| `taws dashboard` | Launch interactive terminal dashboard | System |
| `taws start` | Start all services with automatic port conflict resolution | Service |
| `taws stop` | Stop all services | Service |
| `taws restart` | Restart all services | Service |
| `taws status` | Show service status | Service |
| `taws project create <name>` | Create a new project | Project |
| `taws project delete <name>` | Delete a project | Project |
| `taws project list` | List all projects | Project |
| `taws mariadb list` | List all databases | Database |
| `taws mariadb create <name>` | Create a new database | Database |
| `taws mariadb drop <name>` | Drop a database | Database |
| `taws mariadb tables <db>` | List tables in a database | Database |
| `taws mariadb generate` | Generate MariaDB configuration | Database |
| `taws adminer status` | Show Adminer status | Web UI |
| `taws adminer install` | Download and install Adminer | Web UI |
| `taws backup` | Create a backup snapshot | Backup |
| `taws backup list` | List all snapshots | Backup |
| `taws backup delete <id>` | Delete a snapshot | Backup |
| `taws rollback <id>` | Restore from a snapshot | Backup |
| `taws plugin install <name>` | Install a plugin | Plugin |
| `taws plugin remove <name>` | Remove a plugin | Plugin |
| `taws plugin enable <name>` | Enable a plugin | Plugin |
| `taws plugin disable <name>` | Disable a plugin | Plugin |
| `taws plugin list` | List installed plugins | Plugin |
| `taws plugin available` | List available plugins | Plugin |
| `taws nginx list` | List Nginx virtual hosts | Config |
| `taws nginx test` | Test Nginx configuration | Config |
| `taws nginx reload` | Reload Nginx configuration | Config |
| `taws php info` | Show PHP information | Config |
| `taws php modules` | List PHP modules | Config |
| `taws php generate` | Generate PHP configuration | Config |
| `taws uninstall` | Completely remove TAWS and all its files | Maintenance |

---

## Project Management

### Creating a Project

```bash
taws project create mywebsite
```

This creates:
- Project directory: ~/web/mywebsite/
- Default index.php file
- Nginx virtual host configuration
- Domain: mywebsite.test

### Listing Projects

```bash
taws project list
```

### Deleting a Project

```bash
taws project delete mywebsite
```

## Databases

### Creating a Database

```bash
taws mariadb create mydb
```

### Listing Databases

```bash
taws mariadb list
```

### Accessing Adminer

1. Install Adminer: `taws adminer install`
2. Access at: http://127.0.0.1:8081
3. Integrity is verified via SHA-256 checksum on download.

## Updates

> **Note:** TAWS manages updates through the backup and rollback system.

1. Create a backup: `taws backup`
2. Update TAWS
3. If issues occur: `taws rollback <snapshot-id>`

## Backups

### Creating a Backup

```bash
taws backup
```

### Listing Backups

```bash
taws backup list
```

### Restoring a Backup

```bash
taws rollback <snapshot-id>
```

## Rollback

Rollback restores:
- Configuration files
- State files
- Metadata

> **Tip:** Rollback preserves your user project files in ~/web/.

## SSL

TASSL supports SSL certificates:

- Preferred: mkcert
- Fallback: Self-signed certificates
- Certificate location: ~/.taws/certs/

## Plugins

### Available Plugins

- apache
- postgresql
- redis
- nodejs
- phpmyadmin
- mailhog

### Installing a Plugin

```bash
taws plugin install redis
taws plugin enable redis
```

## Uninstalling TAWS

To completely remove TAWS and all its files, run:

```bash
taws uninstall
```

This will remove:
- All TAWS configuration files (~/.taws/configs/)
- All state files (~/.taws/state/)
- All log files (~/.taws/logs/)
- All backup snapshots (~/.taws/backups/)
- All SSL certificates (~/.taws/certs/)
- All installed plugins (~/.taws/plugins/)
- All cache files (~/.taws/cache/)
- The TAWS binary (if found in current directory)

> **Note:** By default, your projects in ~/web/ are preserved. To also remove your projects:

```bash
taws uninstall --keep-projects=false
```

### Manual Uninstallation

If you prefer to manually remove TAWS:

```bash
# Remove TAWS directories
rm -rf ~/.taws/

# Remove projects (optional)
rm -rf ~/web/

# Remove TAWS binary (if installed)
rm -f /usr/local/bin/taws
# or
rm -f $PREFIX/bin/taws
```

## Troubleshooting

### Common Issues

**Services won't start**
```bash
taws doctor
taws status
```

**Port already in use**
```bash
taws start
# TAWS will automatically detect the conflict and assign an available port.
# Run 'taws status' to verify the resolved ports.
```

**Configuration errors**
```bash
taws nginx test
```

**Permission denied**
```bash
# Ensure proper permissions on TAWS directories
chmod -R 755 ~/.taws/
```

### Logs

Check logs in:
- ~/.taws/logs/

## Directory Layout

```
~/.taws/
├── configs/          # Configuration files
├── state/            # State files
├── logs/             # Log files
├── backups/          # Backup snapshots
├── certs/            # SSL certificates
├── plugins/          # Installed plugins
└── cache/            # Cache files

~/web/                # User projects
├── blog/
├── api/
└── shop/
```

## Security

TAWS applies defense-in-depth security practices throughout:

- **Input Validation** — All user-facing inputs (database names, project names,
  plugin names, template strings, snapshot IDs) are validated with strict regex
  and length limits before processing. No raw user input reaches the shell or
  database.
- **File Permissions** — Sensitive state, config, metadata, and export files use
  `0600` permissions.
- **Backup Integrity** — Archive extraction guards against path traversal (zip
  slip), symlink attacks, hardlink attacks, and decompression bombs (10,000
  file / 1 GB size limits).
- **Plugin Safety** — All public plugin methods validate names and prevent path
  traversal beyond the plugin directory.
- **Template Safety** — Template-generated configuration files are bounded (no
  newlines, control characters, max 1024 characters per value).
- **Secure Downloads** — Adminer downloads are verified against a SHA-256
  checksum before installation.

## Development Guide

Get started building and contributing to TAWS.

### Prerequisites

- Go 1.24+
- Git
- Make
- GCC

### Building

```bash
go build -o taws ./cmd/taws
```

### Testing

```bash
go test ./...
```

### Project Structure

```
cmd/taws/          # CLI entry points
internal/          # Private packages
├── adminer/       # Adminer integration
├── backup/        # Backup system
├── config/        # Configuration management
├── doctor/        # Diagnostics
├── generator/     # Template generation
├── installer/     # Setup installer
├── mariadb/       # MariaDB integration
├── nginx/         # Nginx integration
├── php/           # PHP integration
├── plugin/        # Plugin framework
├── projects/      # Project management
├── services/      # Service management
├── state/         # State management
├── tui/           # Terminal UI
└── validate/      # Input validation & sanitization
plugins/           # Example plugins
templates/         # Configuration templates
docs/              # Documentation
tests/             # Test files
```

## Contribution Guide

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run `go test ./...`
6. Run `go vet ./...`
7. Submit a pull request

## FAQ

**Q: What is TAWS?**
A: TAWS is a lightweight web development environment for Android/Termux.

**Q: Do I need root access?**
A: No, TAWS runs without root privileges.

**Q: Can I use TAWS on Linux?**
A: Yes, TAWS supports standard Linux distributions for development.

**Q: How do I access my projects?**
A: Open a browser and go to http://projectname.test

**Q: How do I backup my data?**
A: Run `taws backup` to create a snapshot.

**Q: How do I restore from a backup?**
A: Run `taws rollback <snapshot-id>` to restore.

**Q: Can I install plugins?**
A: Yes, use `taws plugin install <name>`.

**Q: How do I check system health?**
A: Run `taws doctor` for diagnostics.

---

## License

MIT License

## Acknowledgments

- Built with Go
- Terminal UI powered by Bubble Tea and Lip Gloss
- CLI framework: Cobra
- Inspired by XAMPP and Laragon
