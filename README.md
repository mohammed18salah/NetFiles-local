<p align="center">
  <h1 align="center">NetFiles</h1>
  <p align="center"><strong>Zero-Internet Offline LAN File-Sharing System for Windows</strong></p>
  <p align="center">Created by <strong>Mohammed Salah</strong></p>
</p>

<p align="center">
  <a href="https://github.com/mohammed18salah/NetFiles-local/releases/download/v1.0.0/netfiles.exe">
    <img src="https://img.shields.io/badge/Download-netfiles.exe%20(v1.0.0)-2ea44f?style=for-the-badge&logo=windows&logoColor=white" alt="Download netfiles.exe">
  </a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Release-v1.0.0-blue.svg?style=flat-square" alt="Version">
  <img src="https://img.shields.io/badge/Language-Go%201.21+-00ADD8.svg?style=flat-square" alt="Go">
  <img src="https://img.shields.io/badge/Platform-Windows%2010%20%7C%2011-0078D6.svg?style=flat-square" alt="Windows">
  <img src="https://img.shields.io/badge/Service-NetFilesSvc%20(LocalSystem)-green.svg?style=flat-square" alt="Service">
  <img src="https://img.shields.io/badge/Binary-Static%20(7.47%20MB)-blue.svg?style=flat-square" alt="Binary">
  <img src="https://img.shields.io/badge/Network-Zero%20Internet%20%7C%20Link--Local-orange.svg?style=flat-square" alt="Network">
  <img src="https://img.shields.io/badge/License-MIT-gray.svg?style=flat-square" alt="License">
</p>

---

<p align="center">
  <img src="demo.gif" alt="NetFiles Live Demonstration" width="640">
</p>
<p align="center">
  <sub>Full showcase video: <a href="IMG_1198.MP4">IMG_1198.MP4</a></sub>
</p>

---

## Overview

**NetFiles** is a high-performance, offline LAN file distribution and synchronization system engineered for Windows 10 and 11. It transforms computers connected to a local area network—via switch, unrouted Ethernet cable, or offline Wi-Fi—into discoverable peer nodes.

Every online peer surfaces directly inside native **Windows File Explorer** as a standard folder. Users transfer files across the network simply by dragging and dropping them into peer folders, without launching desktop software, opening terminal windows, or depending on cloud services or internet connectivity.

NetFiles is split into two specialized components packaged inside a single static executable (`netfiles.exe`):
1. **Background Service (`NetFilesSvc`)**: A headless Windows service managed by the Service Control Manager (SCM). It handles UDP broadcast discovery, TCP HTTP streaming, directory watching, peer synchronization, and offline queuing. It runs under `LocalSystem`, starts automatically at boot, and persists across user logoffs.
2. **Control Panel / Installer**: A lightweight setup and management interface with an interactive ASCII terminal interface. It installs the service, provisions ACL permissions, configures firewall rules, links File Explorer, and offers diagnostics.

---

## Core Capabilities

- **Zero-Internet Operation**: Fully functional on isolated switches, link-local addresses (`169.254.x.x`), or direct cross-cables without DHCP servers, DNS, or gateways.
- **Native File Explorer Integration**: Pinned to the File Explorer root navigation pane using Windows Shell CLSID namespace registration (`System.IsPinnedToNameSpaceTree`), alongside Desktop shortcuts and Network Shortcuts (`shell:NetHood`).
- **High-Throughput Chunked Streaming**: Transfers are streamed over TCP HTTP chunked sockets directly to disk with temporary `.part` staging and atomic renaming to prevent partial or corrupted file writes.
- **Public & Peer Synchronization**: Includes a synchronized `Public` workspace visible to all LAN nodes, direct private `Inbox` destinations for point-to-point transfers, and a `_Sent` ledger.
- **Offline Fault Tolerance**: Supports peer state awareness. If a target node disconnects, transfers are automatically queued and flushed when the node re-appears on the network.
- **Scalable Architecture**: Engineered to scale cleanly across computer labs and institutional classrooms of 40+ nodes without bandwidth degradation or centralized bottlenecks.
- **Single Self-Contained Binary**: Written in pure Go with zero runtime external dependencies or CGO requirements.

---

## Architecture

```
                             +-----------------------+
                             |    File Explorer      |
                             |   (Shell Namespace)   |
                             +-----------+-----------+
                                         |
                                         v
                             +-----------------------+
                             |     C:\NetFiles       |
                             | (NTFS ACL: Users:M)   |
                             +-----------+-----------+
                                         |
                                         v
                     +---------------------------------------+
                     |         NetFilesSvc Service           |
                     |       (LocalSystem, Auto-Start)       |
                     +-------------------+-------------------+
                                         |
            +----------------------------+---------------------------+
            |                                                        |
            v                                                        v
+-----------------------+                                +-----------------------+
|    Discovery Engine   |                                |    Transfer Server    |
|   UDP Broadcast 47832 |                                |    HTTP / TCP 47831   |
|   Heartbeat: 2000ms   |                                |    Chunked Streaming  |
+-----------------------+                                +-----------------------+
            |                                                        |
            +----------------------------+---------------------------+
                                         |
                                         v
                     +---------------------------------------+
                     |        Local Area Network (LAN)       |
                     |  (Switch / Direct Cable / Link-Local) |
                     +---------------------------------------+
```

---

## System Layout

NetFiles provisions a standardized workspace at `C:\NetFiles` (or a configured user directory) with full read/write inheritance:

```text
C:\NetFiles\
├── Public\             # Files dropped here sync across all LAN nodes
├── Inbox\              # Incoming files received from other peers
├── _Sent\              # Outgoing transfer history and delivery logs
├── _Online.txt         # Live status indicator tracking active LAN nodes
├── Lab-PC-02\          # Dynamic peer directory (auto-managed)
└── Lab-PC-03\          # Dynamic peer directory (auto-managed)
```

---

## Interactive Control Panel

Launching `netfiles.exe` without arguments opens the built-in management interface featuring the green ASCII terminal emblem:

```text
            ..oo$00ooo..                    ..ooo00$oo..
         .o$$$$$$$$$'                          '$$$$$$$$$o.
       .o$$$$$$$$"              .    .              "$$$$$$$$o.
     .o$$$$$$$$$~              /$    $\              ~$$$$$$$$$o.
    .{$$$$$$$$$$.              $\___/$              .$$$$$$$$$$}.
   o$$$$$$$$$$$$8              .$$$$$$.              8$$$$$$$$$$$$o
   $$$$$$$$$$$$$$              $$$$$$$$              $$$$$$$$$$$$$$
  o$$$$$$$$$$$$$$.            o$$$$$$$$o            .$$$$$$$$$$$$$$o
  $$$$$$$$$$$$$$$$.          o{$$$$$$$$}o          .$$$$$$$$$$$$$$$$
 ^$$$$$$$$$$$$$$$$$.         J$$$$$$$$$$L         .$$$$$$$$$$$$$$$$$^
 !$$$$$$$$$$$$$$$$$$oo..oo$$$$$$$$$$$$$$$$$$oo..oo$$$$$$$$$$$$$$$$$$!
 {$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$}
 6$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$?
 '$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$'
  o$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$o
    $$$$$$$$$$$$$; '~^Y$$$7^' 'o$$$$$$$$$$o' '^Y$$$7^~' ;$$$$$$$$$$$$$
    '$$$$$$$$$$$'      `$'     `'$$$$$$$$'`     `$'      '$$$$$$$$$$$'
     !$$$$$$$$$7        !        '$$$$$$'        !        V$$$$$$$$$!
      ^o$$$$$$!                   '$$$$'                   !$$$$$$o^
        ^$$$$$"                   $$$$$$                   "$$$$$^
          'o$$$`                  ^$$$'                  '$$$o'
            ~$$$.                  $$$.                  .$$$~
              '$;                  `$'                  ;$' 
               '.                   !                   .'
                .                                       .

       [ NETFILES LAN SYSTEM V1.0 ]
       [ CREATED BY MOHAMMED SALAH - 2026 ]

:: Existing NetFiles installation detected on this computer:
   Computer Name : PC-001
   Root Folder   : C:\NetFiles [ OK - Present ]
   Service       : [ Running ] (NetFilesSvc)
   Access Mode   : [ Administrator ]

+----------------------------------------------------------------------+
|                       NETFILES CONTROL PANEL                         |
+----------------------------------------------------------------------+
  [1] Rename this PC         - Change computer name on LAN
  [2] Background Service     - Start / Stop / Restart NetFilesSvc
  [3] Delete Folder & Files  - Permanently wipe NetFiles folder
  [4] Uninstall NetFiles     - Remove service, firewall & links
  [5] Repair Explorer Links  - Refresh navigation pane & Desktop links
  [6] System Diagnostics     - Run Doctor health checks
  [7] Re-run Full Setup      - Reinstall components & reset settings
  [8] Exit
+----------------------------------------------------------------------+
```

---

## Quick Start

### 1. Download Binary

Download the pre-compiled static executable directly from the latest release:
- **[Download netfiles.exe v1.0.0 (Windows 64-bit)](https://github.com/mohammed18salah/NetFiles-local/releases/download/v1.0.0/netfiles.exe)** (~7.47 MB, static binary)

### 2. Installation

Run the automated setup to provision the background service, configure firewall ports, and register Windows File Explorer links:

```powershell
.\netfiles.exe setup
```

Or run unattended setup with automatic confirmation:

```powershell
.\netfiles.exe setup -y
```

### Uninstallation

To cleanly remove the Windows service, delete firewall rules, and detach File Explorer shortcuts:

```powershell
.\netfiles.exe uninstall
```

Add `-force` to simultaneously remove the `C:\NetFiles` root directory and stored files:

```powershell
.\netfiles.exe uninstall -force
```

---

## CLI Command Reference

| Command | Description |
| :--- | :--- |
| `netfiles` | Opens the interactive Control Panel / auto-detects installation state |
| `netfiles setup [-y]` | Provisions service, directory layout, firewall rules, and Explorer links |
| `netfiles uninstall [-force]` | Stops and deletes service, removes firewall rules, and cleans shortcuts |
| `netfiles status` | Prints node identity, IP addresses, service state, and storage health |
| `netfiles doctor` | Runs 5-point system diagnostics (Service, Directory, Ports, Profile, CLSID) |
| `netfiles start` | Starts the `NetFilesSvc` background service |
| `netfiles stop` | Gracefully stops the `NetFilesSvc` background service |
| `netfiles rename <name>` | Renames the local node and updates network advertisement |
| `netfiles about` | Shows build version and authorship information |
| `netfiles service` | Internal entry point invoked directly by Windows SCM |

---

## System Diagnostics (`doctor`)

NetFiles includes a built-in diagnostic engine (`netfiles doctor`) that validates end-to-end operational readiness:

```text
:: Running NetFiles Diagnostics Doctor...

   [1/5] Checking Windows Service (NetFilesSvc)...         [ OK - Running ]
   [2/5] Checking root folder accessibility...            [ OK - C:\NetFiles ]
   [3/5] Checking TCP/UDP ports...                        [ OK - HTTP 47831 in use by service ]
   [4/5] Checking Windows Network profile...              [ OK - Private ]
   [5/5] Checking Explorer navigation pane registration... [ OK - Pinned & Active ]

:: Doctor diagnostics complete.
```

---

## Specifications

| Parameter | Specification |
| :--- | :--- |
| **Language** | Go 1.21+ (Static binary) |
| **Target OS** | Windows 10, Windows 11 (64-bit) |
| **Binary Footprint** | ~7.4 MB (symbols stripped via `-ldflags "-s -w"`) |
| **Discovery Protocol** | UDP Broadcast / Link-Local Multicast on port `47832` |
| **Transfer Protocol** | HTTP 1.1 / TCP Chunked Streaming on port `47831` |
| **Service Framework** | `golang.org/x/sys/windows/svc` |
| **Registry Integration** | HKLM CLSID `{D45E2001-A482-4E90-951E-478310000001}` |
| **File Permissions** | NTFS ACL `Users:(OI)(CI)M`, `Everyone:(OI)(CI)M` |
| **Explorer Notification** | `SHChangeNotify(0x08000000, 0, 0, 0)` |

---

## Building from Source

### Prerequisites

- Go 1.21 or higher
- Windows 10 / 11 with administrator privileges for testing service integration

### Compilation

Use the automated build script or compile directly with standard flags:

```powershell
.\build.bat
```

Or manually:

```powershell
go build -ldflags "-s -w" -o netfiles.exe .
```

The resulting `netfiles.exe` is completely static and can be deployed directly to target machines without prerequisites, runtimes, or installer frameworks.

---

## Author & Credits

- **Project**: NetFiles LAN System
- **Author**: **Mohammed Salah**
- **Year**: 2026
- **License**: MIT License
