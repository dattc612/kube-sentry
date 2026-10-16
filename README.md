# KubeSentry - Go Through The Storm

> **KubeSentry** is a lightweight, developer-focused Kubernetes dashboard that enables teams to safely inspect application resources without requiring direct cluster shell access.

---

# Features

- Safe Resource Inspection: Monitor running workloads without using `kubectl`.

- Container File Explorer: No interactive shell access is exposed.

- Safe Configuration Editing: All editable files are controlled through configurable allow-list patterns.

- File Downloads

## Security First

KubeSentry is designed around the principle of **least privilege**.

# Use Cases

Ideal for:

- Application developers

Common tasks include:

- Verifying configuration, package
- Inspecting runtime files
- Downloading diagnostic files
- Editing approved configuration safely

KubeSentry intentionally limits what users can do.

Allowed:

- View pods
- Read logs
- Browse files
- Download files
- Edit approved files only

Not Allowed:

- Execute arbitrary shell commands
- Delete Kubernetes resources
- Modify protected system files
- Escape configured namespaces
- Bypass Kubernetes RBAC

# Getting Started

## Prerequisites

Before you begin, ensure you have the following installed:

- Go **1.26+**
- Node.js **24+**
- An active Kubernetes configuration (for dev env) or in-cluster access (prod)

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

## 3. Configuration

Create or edit `config.yaml`.

```yaml
auth:
  users: #list user with hashed password
    - group: admin
      password: $2a$12$enwsrQGq9kdQMpCXFzcvE.ByquMI.Zygp/oqrvj0/IGAkP2NLf29O
      username: admin
editableFiles:
  enabled: true
  patterns: #limit editable file
    - ^.*logback\.xml$
kube:
  kubeconfigPath: "" #only require in dev env
namespace:
  allowed:
    - stagingvodds #limit namespace
server:
  mode: debug
  port: 6120 #server port
jobs: #allowed job
  - jobID: 1
    jobName: Restart Java Process
    description: This job will restart java process inside container
    cmd: ["pkill", "java"]
    allowedPod:
      - ^(?!ae).*$
```

Description:

```
auth.users: Defines the list of authorized users and their hashed passwords.

editableFiles.enabled: Enables or disables file editing.
editableFiles.patterns: Specifies the file patterns that are allowed to be edited.

namespace.allowed: Lists the namespaces that developers are permitted to access.

jobs: Defines the list of custom jobs or tasks that can be configured.
```

---

## 4. Build & Run

### Build Local Binary

```bash
make build
```

### Build Docker Image

```bash
make docker-build
```

---

### Deploy with Helm

Install into the `operation` namespace:

```bash
helm upgrade --install kubesentry -n operation ./  -f values.yaml  -f ./env/stag/values.yaml
```

```text
Release "kubesentry" has been upgraded. Happy Helming!
NAME: kubesentry
LAST DEPLOYED: Wed Jul 29 17:24:04 2026
NAMESPACE: operation
STATUS: deployed
REVISION: 23
DESCRIPTION: Upgrade complete
NOTES:
1. Get the application URL by running these commands:
  https://kube-sentry.stagingvodds.net/
```

![alt text](image.png)
