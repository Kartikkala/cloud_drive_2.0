# 🚀 Cloud Drive 2.0

![Go Version](https://img.shields.io/badge/Go-1.18+-blue.svg)
![License](https://img.shields.io/badge/License-MIT-green.svg)

**Cloud Drive 2.0** is a modern, powerful, and self-hostable cloud storage solution built with Go. It's designed to be your all-in-one private cloud for files, collaboration, and media, giving you complete control over your data on your own hardware.

---

## ✨ Core Features

Cloud Drive 2.0 is more than just a place to store files. It's a rich ecosystem of tools designed for productivity and seamless media access.

#### 🗄️ Advanced File Management
- **Standard Operations:** Fast and reliable uploads, downloads, and file/folder organization.
- **In-Browser Previews:** Preview images, videos, PDFs, and text-based files directly in your browser without downloading them.
- **Shareable Links:** Create public, password-protected, or time-limited links to share files and folders with anyone.
- **Full-Text Search:** A powerful search engine that indexes the content of your documents (PDF, TXT, MD), not just the filenames.

#### 🎬 Superior Media Streaming
- **Adaptive Bitrate Streaming (ABS):** Enjoy smooth video and audio playback on any device and connection speed. The server automatically adjusts the stream quality in real-time to prevent buffering.
- **Wide Format Support:** Built on powerful streaming technology to handle a vast array of video and audio codecs.

#### 📥 Server-Sided Downloads
- **Download from URL:** Paste a direct HTTP link, and the server will download the file directly to your drive, saving your local bandwidth.
- **Torrent & Magnet Support:** Start torrent or magnet link downloads directly on the server. Your files will be downloaded and ready for you without needing a torrent client on your local machine.

#### 🔐 Security & Access Control
- **Role-Based Access Control (RBAC):** A granular permissions system to control exactly who can see, edit, or share files and folders. Perfect for teams or families.

#### 🤝 Collaboration & Productivity
- **TODO Lists:** A built-in task manager to create and track personal or to-do lists.
- **Meetings:** Integrated video and text chat functionality to host secure, private meetings with other users on your drive.
- **Markdown Editor:** A beautiful, built-in Markdown editor with a live preview for creating and editing documents.

---

## 🛠️ Technology Stack

- **Backend:** Go
- **Framework:** Echo
- **Database:** PostgreSQL with GORM
- **Authentication:** JWT-based
- **Real-time:** WebSockets (for Meetings & TODOs)

---

## 🚀 Getting Started

### Prerequisites
- [Go](https://golang.org/doc/install) (version 1.18 or newer)
- [Docker](https://www.docker.com/get-started) and Docker Compose (for running the required PostgreSQL database)
- A C compiler (like `gcc`) for certain Go dependencies

### Installation & Setup

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/sirkartik/cloud_drive_2.0.git
    cd cloud_drive_2.0
    ```

2.  **Configure your environment:**
    The application is configured using environment variables. You will need to set up variables for your database connection, application ports, and any third-party services. A good place to manage these is in a `.env` file that you source into your shell.

    Refer to `internal/config/config.go` for all available options. The essential ones are:
    ```sh
    export POSTGRES_USER="your_db_user"
    export POSTGRES_PASSWORD="your_db_password"
    export POSTGRES_DB="your_db_name"
    export POSTGRES_HOST="localhost"
    export POSTGRES_PORT="5432"
    ```

3.  **Install dependencies:**
    The first time you build or run the project, Go will automatically download all the necessary modules listed in `go.mod`.

---

## 🏗️ Building and Running

This project uses a `Makefile` to simplify common tasks.

- **Run the server in development mode (with live reload):**
  This requires `air` to be installed (`go install github.com/cosmtrek/air@latest`).
  ```sh
  make dev
  ```

- **Build and run the server for production:**
  ```sh
  make run
  ```

- **Build the binary only:**
  The compiled binary will be placed in the `bin/` directory.
  ```sh
  make build
  ```

- **Clean up generated files:**
  Removes the `bin/` directory.
  ```sh
  make clean
  ```

---

## 🗺️ Roadmap

- [ ] Two-Factor Authentication (2FA) for enhanced security.
- [ ] Photo gallery features (e.g., timeline view, albums).

---

## 🤝 Contributing

Contributions, issues, and feature requests are welcome! Feel free to check the [issues page](https://github.com/sirkartik/cloud_drive_2.0/issues).

---

## 📜 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
