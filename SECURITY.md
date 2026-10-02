# Security Policy

**NetFiles** is designed for offline, zero-internet local area network (LAN) file sharing. Because NetFiles operates as a background Windows Service (`NetFilesSvc`) handling direct TCP socket streaming and UDP discovery, security and integrity are core priorities.

---

## Supported Versions

Security updates and critical vulnerability patches are actively maintained for the following versions:

| Version | Supported | Notes |
| :--- | :--- | :--- |
| `1.0.x` | Yes | Current stable release line |
| `< 1.0.0` | No | Legacy and development previews |

---

## Scope & Security Model

NetFiles is intended to operate on trusted or isolated local network segments (switches, classrooms, lab environments, and direct link-local cables).

### 1. Network Exposure
- **TCP Port 47831**: Used strictly for chunked HTTP file streaming. Incoming requests must only access designated synchronization directories (`Public`, `Inbox`).
- **UDP Port 47832**: Used strictly for peer discovery heartbeats. Packet payloads are validated and size-limited to prevent buffer abuse.
- **Firewall Rules**: Rules created during setup (`NetFilesTool-TCP`, `NetFilesTool-UDP`) apply to Private network profiles by default.

### 2. Path Traversal & File Containment
- All incoming file transfers are strictly sanitized to prevent directory traversal attacks (`../`, relative paths, or absolute path overrides).
- Incoming data is staged in temporary `.part` files within the designated root folder (`C:\NetFiles`) and atomically renamed upon completion.

### 3. Windows Service Security
- `NetFilesSvc` runs under the `LocalSystem` security context to survive logoffs and manage file permissions.
- NTFS ACLs are configured on the root directory (`Users:(OI)(CI)M`) so that authenticated local users can access files without elevating privileges.

---

## Reporting a Vulnerability

If you discover a security vulnerability or potential exploit in NetFiles, please report it responsibly. **Do not disclose security vulnerabilities through public GitHub issues.**

### Reporting Channels

1. **GitHub Private Vulnerability Reporting (Preferred)**:
   - Navigate to the repository's **Security** tab.
   - Click on **Report a vulnerability** to open a confidential advisory draft.

2. **Direct Contact**:
   - If GitHub reporting is unavailable, contact the project maintainer directly:
   - Repository: `https://github.com/mohammed18salah/NetFiles-local`
   - Author: **Mohammed Salah**

### What to Include

To help us triage and resolve the issue quickly, please include:
- A clear description of the vulnerability and its potential impact.
- Step-by-step instructions or a minimal proof of concept (PoC) to reproduce the behavior.
- Operating system version (e.g., Windows 10 Pro 22H2, Windows 11 23H2).
- NetFiles version (`netfiles about`).
- Any suggested mitigations or patches, if available.

---

## Response Timeline

- **Initial Acknowledgment**: Within **48 hours** of receiving your report.
- **Triage & Assessment**: Within **5 business days**, detailing validity and estimated severity.
- **Resolution & Release**: Critical security fixes will be prioritized, packaged into a minor release, and documented in GitHub Security Advisories with credit to the reporter.

---

## Author & Attribution

- **Project**: NetFiles LAN System
- **Author**: **Mohammed Salah**
- **Year**: 2026
