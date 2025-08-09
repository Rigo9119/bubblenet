# Bubblenet

A terminal-based chat application that allows users to communicate through the command line by connecting two machines over a network.

## Overview

Bubblenet is a Go-based terminal chat system consisting of:
- **Server**: WebSocket server that handles connections and message routing
- **Client**: Terminal-based chat client that connects to the server
- **Protocol**: Message handling and communication protocol

## Project Structure

```
bubblenet/
├── cmd/
│   ├── client/          # Client application entry point
│   └── server/          # Server application entry point
├── internal/
│   ├── client/          # Client-side logic
│   ├── server/          # Server-side logic including WebSocket handling
│   └── ui/              # Terminal user interface components
├── pkg/
│   ├── config/          # Configuration management
│   └── protocol/        # Message protocol definitions
└── scripts/             # Build and utility scripts
```

## Prerequisites

- Go 1.24.4 or higher

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd bubblenet
```

2. Install dependencies:
```bash
go mod tidy
```

## Usage

### Starting the Server

```bash
go run cmd/server/main.go
```

### Connecting with a Client

```bash
go run cmd/client/main.go
```

## Building

To build both server and client:

```bash
# Build server
go build -o bin/server cmd/server/main.go

# Build client  
go build -o bin/client cmd/client/main.go
```

## Features

- Terminal-based chat interface
- WebSocket communication
- Multi-client support
- Real-time messaging

## Development

This project is structured as a modular Go application with clear separation between client, server, and shared components.