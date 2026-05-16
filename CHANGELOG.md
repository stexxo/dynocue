# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-05-16

### Added
- Real-time cue and action execution engine with support for delays and follow-throughs.
- "Elapsed" time column for cues and actions to monitor active playback.
- Visual progress bars and countdown timers for cues with active delay or follow-through times.
- Live-updating interface that reflects current playback state across all views.
- Real-time synchronization of execution states between the engine and the UI.

### Changed
- Enhanced Cue Table styling with distinct colors for active and selected cues for better visibility.
- Improved time formatting to be more robust and support various input formats.
- Optimized backend state transitions for smoother cue playback.
- Updated default logging level to Info to reduce terminal clutter.

### Fixed
- Fixed a build error that prevented the application from compiling on Windows.
- Resolved an issue where timing data could be lost during certain cue transitions.
- Improved reliability of tracking for running actions.

## [0.1.0] - 2026-05-14

### Added
- Core messaging system based on NATS for inter-subsystem communication.
- Subsystem management framework with logging and persistence support.
- Cue management system including CueLists, Cues, and Actions.
- Action template system for extensible cue behaviors.
- Audio subsystem for media playback and management.
- Desktop GUI built with Wails and Svelte for show control.
- Go client library for external integration with the cueing engine.
- Automated license management and project Taskfile for development.
