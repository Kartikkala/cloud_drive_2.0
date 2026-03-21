## Cloud Drive 2.0

![Go Version](https://img.shields.io/badge/Go-1.18+-blue.svg)
![License](https://img.shields.io/badge/License-MIT-green.svg)

**Cloud Drive 2.0** is a self-hostable cloud storage system built with Go, designed with a focus on clean architecture, extensibility, and media processing.

> A distributed, event-driven cloud drive with a built-in video streaming pipeline.

---

## Related Services

This project is part of a distributed system:

* Artifacts Service (video processing):
  [https://github.com/Kartikkala/artifacts_svc](https://github.com/Kartikkala/artifacts_svc)

* Sidecar Service (artifact upload and orchestration):
  [https://github.com/Kartikkala/adapter_sidecar_artifactssvc](https://github.com/Kartikkala/adapter_sidecar_artifactssvc)

---

## Core Features

### Drive Layer

* Upload, download, and file and folder management
* Hierarchical structure with parent-child relationships
* Metadata stored in PostgreSQL

---

### Access Control

* Fine-grained permission system
* Automatic propagation of permissions across subdirectories
* Implemented using recursive SQL CTEs

---

### Video Processing Pipeline

* Upload triggers asynchronous processing via hooks (observer pattern)
* Jobs are dispatched using NATS
* Artifacts service processes video using FFmpeg
* Generates multi-bitrate HLS streams:

  * `.m3u8` playlists
  * `.ts` segments
* Sidecar service uploads processed artifacts to object storage

---

### Asynchronous Architecture

* Observer pattern for extensibility (hooks on storage operations)
* Event-driven pipeline using NATS
* Decoupled services (storage, processing, artifact upload)

---

## Technology Stack

* Backend: Go (Echo)
* Database: PostgreSQL (GORM)
* Object Storage: MinIO (S3-compatible)
* Messaging: NATS
* Media Processing: FFmpeg

---

## Getting Started

### Prerequisites

* Go (1.18 or newer)
* Docker and Docker Compose
* PostgreSQL
* MinIO (recommended)

---

### Setup

```sh
git clone https://github.com/sirkartik/cloud_drive_2.0.git
cd cloud_drive_2.0
```

Set environment variables:

```sh
export POSTGRES_USER="your_db_user"
export POSTGRES_PASSWORD="your_db_password"
export POSTGRES_DB="your_db_name"
export POSTGRES_HOST="localhost"
export POSTGRES_PORT="5432"
```

---

## Running

```sh
make dev    # development (with live reload)
make run    # production
make build  # build binary
```

---

## Roadmap

* Server-side downloads (HTTP, torrent, magnet)
* Fuzzy file search
* Public file sharing
* Improved permission model
* Image and music artifact processing

---

## Contributing

Contributions, issues, and feature requests are welcome.

---

## License

This project is licensed under the MIT License.

