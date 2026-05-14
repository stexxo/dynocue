# Dynocue

> **Note:** This project is currently **in progress** and is **not yet functional**.

Dynocue is a modern cueing system designed for live performances, theater, and events. It aims to provide a robust and flexible platform for managing complex show flows, integrating audio, lighting, and other show control elements.

For more information, visit [dynocue.com](https://dynocue.com).

## Project Overview

Dynocue is built with a modular architecture using Go for the backend and a modern web-based GUI. It leverages NATS for high-performance messaging between components.

### Key Components

- **Cues & Cuelists:** Advanced management of show cues, actions, and templates.
- **Audio Engine:** High-performance audio playback and routing (in development).
- **Show Management:** Tools for organizing and executing live shows.
- **Messaging Infrastructure:** Built-on NATS for reliable and low-latency communication.
- **Cross-platform GUI:** User-friendly interface for show programming and operation.

## Development

The project is structured into several core components:

- `cmd/`: Application entry points.
- `components/`: Core logic modules (cues, audio, show, system).
- `core/`: Fundamental infrastructure (logging, messaging).
- `gui/`: Frontend and backend for the user interface.
- `client/`: Client libraries for interacting with the Dynocue ecosystem.

---
© 2026 Dynocue
