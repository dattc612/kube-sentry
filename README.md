# 🌩️ KubeSentry - Go Through The Storm

> **KubeSentry** is a lightweight, developer-focused Kubernetes dashboard that enables teams to safely inspect application resources without requiring direct cluster shell access. Designed for development and troubleshooting, KubeSentry provides secure, read-first access to Kubernetes workloads while allowing controlled editing of approved configuration files.

---

# ✨ Features

## 🔍 Multi-Namespace Browsing

Browse pods across one or more Kubernetes namespaces that your account is authorized to access.

* Namespace selector
* Pod and container discovery
* RBAC-aware visibility
* Fast filtering and searching

---

## 📊 Safe Resource Inspection

Monitor running workloads without using `kubectl`.

Inspect:

* Pod status
* Container status
* Events
* Deep dive into a K8s Pod (examining files or resources)
Perfect for developers who need operational visibility without cluster-admin permissions.

---

## 📁 Container File Explorer

Browse container filesystems safely through an intuitive file explorer.

Supported operations:

* Navigate directories
* Preview files
* View file metadata
* Download files
* Read-only browsing for protected paths

No interactive shell access is exposed.

---

## ✏️ Safe Configuration Editing

Modify approved configuration files directly from the web interface.

Typical editable files include:

* application configuration files
* custom developer configuration

All editable files are controlled through configurable allow-list patterns.

Example:

```yaml
editable_files:
  patterns:
    - "^/app/.env$"
    - "^/app/config/.*\\.ya?ml$"
    - "^/home/appuser/.profile$"
```

Critical system files remain protected.

---

## ⬇️ File Downloads

Download files before making changes.

Useful for:

* Backups
* Local debugging
* Configuration review

---

## 🔒 Security First

KubeSentry is designed around the principle of **least privilege**.

Security features include:

* Kubernetes RBAC integration
* Namespace isolation
* Configurable editable file allow-list
* Regex-based path protection
* No unrestricted shell terminal
* No arbitrary command execution
* Read-first workflow
* Audit-friendly architecture

Developers can inspect applications safely without unnecessary cluster privileges.

---

# 🎯 Use Cases

KubeSentry is ideal for:

* Application developers

Common tasks include:

* Checking application logs
* Verifying mounted configuration
* Inspecting runtime files
* Reviewing environment configuration
* Downloading diagnostic files
* Editing approved configuration safely

---

# 🔐 Security Model

KubeSentry intentionally limits what users can do.

✅ Allowed

* View pods
* Read logs
* Browse files
* Download files
* Edit approved files only

❌ Not Allowed

* Execute arbitrary shell commands
* Delete Kubernetes resources
* Modify protected system files
* Escape configured namespaces
* Bypass Kubernetes RBAC

---

# ❓ Why KubeSentry?

Many developers only need to inspect workloads—not administer the cluster.

Instead of granting broad `kubectl` access, KubeSentry provides a secure web interface for common operational tasks while reducing the risk of accidental or unauthorized changes.

It enables teams to troubleshoot applications quickly while maintaining strong security boundaries.

---

# 🚀 Getting Started

## Prerequisites

Before you begin, ensure you have the following installed:

* Go **1.20+**
* Node.js **18+**
* Yarn
* Docker (optional)
* An active Kubernetes configuration (`~/.kube/config`) or in-cluster access

---

## 1. Setup

Install both frontend and backend dependencies.

```bash
make init
```

---

## 2. Development

Run the frontend and backend concurrently with hot reloading.

```bash
make run-dev
```

After startup:

| Service | URL |
|---------|-----|
| Frontend | <http://localhost:3000> *(or <http://localhost:5173> depending on your frontend configuration)* |
| Backend API | <http://localhost:6120> |

The backend uses **fresh** for automatic recompilation and restart during development.

---

## 3. Configuration

Create or edit `config.yaml`.

```yaml
server:
  port: 6120

namespace:
  allowed:
    - default
    - cicd

editable_files:
  enabled: true
  patterns:
    - '^.*\.profile$'
    - '^.*\.env.*$'
```

### Configuration Options

| Key | Description |
| ------ | ------------- |
| `server.port` | HTTP server listening port |
| `namespace.allowed` | Kubernetes namespaces users are allowed to browse |
| `editable_files.enabled` | Enable or disable the built-in file editor |
| `editable_files.patterns` | Regex allow-list for editable files |

> **Security:** Only files matching the configured regex patterns can be modified through the UI.

---

## 4. Build & Run

### Build Local Binary

```bash
make build
```

Run the application:

```bash
./bin/app
```

---

### Build Docker Image

```bash
make docker-build
```

Run the container:

```bash
docker run \
  -p 6120:6120 \
  -v ~/.kube/config:/.kube/config:ro \
  KubeSentry:latest
```

---

### Deploy with Helm

Install into the `application` namespace:

```bash
cd helm
helm upgrade --install  gtts ./  -f values.yaml -f ./env/values.yaml  -n application
```
