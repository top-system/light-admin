<div align="center">
  <img src="docs/logo.png" width="180" height="180" alt="Light Admin Logo" />

  <h1>Light Admin</h1>

  <p>
    <strong>A Lightweight and Elegant Backend Management Solution</strong>
  </p>

  <p>
    RBAC Admin Scaffolding built with Echo + GORM + Casbin + Uber-FX
  </p>

  <p>
    English | <a href="https://github.com/top-system/light-admin/blob/main/README.md">简体中文</a>
  </p>

  <p>
    <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go" alt="Go Version" />
    <img src="https://img.shields.io/badge/Echo-4.11+-00ADD8?style=flat-square" alt="Echo Version" />
    <img src="https://img.shields.io/badge/GORM-1.25+-red?style=flat-square" alt="GORM Version" />
    <img src="https://img.shields.io/badge/Casbin-2.77+-brightgreen?style=flat-square" alt="Casbin Version" />
    <img src="https://img.shields.io/badge/License-MIT-blue?style=flat-square" alt="License" />
  </p>
</div>

---

## ✨ Features

### Core Features
- 🔐 **Authentication** - JWT Token authentication with refresh support
- 👥 **User Management** - CRUD operations, status management, password reset
- 🎭 **Role Management** - Flexible role configuration, multi-role support
- 📋 **Menu Management** - Dynamic menu configuration, multi-level menus
- 🏢 **Department Management** - Tree-structured organization management
- 🔑 **Access Control** - Casbin-based RBAC access control
- 📝 **Operation Logs** - Complete audit logging
- 📢 **Announcements** - System notifications and announcements
- ⚙️ **System Config** - Dynamic system parameter configuration
- 📚 **Dictionary** - Data dictionary maintenance

### Extended Features
- 📤 **File Upload** - Local storage, MinIO, Aliyun OSS support
- ⏰ **Scheduled Tasks** - Flexible cron job scheduling
- 📥 **Task Queue** - Async task processing with retry mechanism
- ⬇️ **Download Manager** - aria2/qBittorrent integration

### Technical Features
- 🚀 **High Performance** - Based on Echo framework with efficient routing
- 📦 **Dependency Injection** - Uber-FX based dependency injection
- 📖 **API Documentation** - Integrated Swagger auto-generation
- 🔧 **Modular Design** - Clean code structure, easy to extend
- 🛡️ **Security** - Comprehensive security middleware support

---

## 📁 Project Structure

```
light-admin/
├── api/                    # API Layer
│   ├── middlewares/        # Middlewares
│   ├── platform/           # Platform module (file upload, etc.)
│   └── system/             # System module (user, role, menu, etc.)
├── bootstrap/              # Application bootstrap
├── cmd/                    # CLI entry points
├── config/                 # Configuration files
├── docs/                   # Documentation & Swagger
├── errors/                 # Error definitions
├── lib/                    # Core libraries
├── models/                 # Data models
│   ├── database/           # Database model base
│   ├── dto/                # Data transfer objects
│   ├── platform/           # Platform module models
│   └── system/             # System module models
├── pkg/                    # Utility packages
│   ├── crontab/            # Scheduled tasks
│   ├── downloader/         # Downloader (aria2/qBittorrent)
│   ├── queue/              # Task queue
│   └── ...                 # Other utilities
└── tests/                  # Test files
```

---

## 🚀 Quick Start

### Requirements

- Go 1.21+
- MySQL 5.7+ / PostgreSQL 12+
- Redis 6.0+
- Node.js 16+ (for frontend)

### Installation

```bash
# Clone the repository
git clone https://github.com/top-system/light-admin.git
cd light-admin

# Copy configuration file
cp config/config.yaml.default config/config.yaml

# Edit database and Redis configuration
vim config/config.yaml

# Initialize database
make migrate

# Setup menu data
make setup

# Start the service
make run
```

### Using Docker

```bash
# Build image
docker build -t light-admin .

# Run container
docker run -d -p 9999:9999 \
  -v ./config:/app/config \
  light-admin
```

---

## 📖 Documentation

| Document | Description |
|----------|-------------|
| [API Docs](docs/swagger.yaml) | Swagger API documentation |
| [Task Queue](docs/queue.md) | Async task queue guide |
| [Crontab](docs/crontab.md) | Scheduled tasks guide |
| [Downloader](docs/downloader.md) | aria2/qBittorrent integration guide |

---

## ⚙️ Configuration

### Basic Configuration

```yaml
Name: light-admin
Http:
  Host: 0.0.0.0
  Port: 9999

Database:
  Engine: mysql
  Host: 127.0.0.1
  Port: 3306
  Name: light_admin
  Username: root
  Password: your_password

Redis:
  Host: 127.0.0.1
  Port: 6379
```

### Extended Features Configuration

```yaml
# Task Queue
Queue:
  Enable: true
  WorkerNum: 4
  MaxRetry: 3

# Scheduled Tasks
Crontab:
  Enable: true

# Downloader
Downloader:
  Enable: false
  Type: aria2
  Aria2:
    Server: http://localhost:6800
    Token: your-secret
```

---

## 🛠️ Development Commands

```bash
# Build
make build

# Run
make run

# Generate Swagger docs
make swagger

# Database migration
make migrate

# Initialize data
make setup

# Run tests
make test
```

---

## 🗺️ Roadmap

- [x] User authentication & access control
- [x] Dynamic menus & role management
- [x] Department & organization structure
- [x] System configuration & dictionary
- [x] File upload (Local/OSS)
- [x] Async task queue
- [x] Scheduled task scheduling
- [x] Downloader integration
- [ ] Operation log auditing
- [ ] Workflow engine
- [ ] Message push
- [ ] Data import/export

---

## 🤝 Contributing

Contributions are welcome! Feel free to submit PRs and Issues.

1. Fork this repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## 📄 License

This project is licensed under the [MIT](LICENSE) License.

---

## 🔗 相关链接

- [Frontend Project](https://github.com/top-system/light-admin-ui)
