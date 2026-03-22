# LazyAzure

A TUI application for Azure resource management, inspired by lazydocker. Browse Azure subscriptions, resource groups, and resources with an interactive terminal interface.

## MVP Status (Phase 1 Complete)

### Current Features
- [x] Azure CLI authentication integration (via DefaultAzureCredential)
- [x] List and browse subscriptions
- [x] View subscription details in Summary and JSON tabs
- [x] Keyboard navigation (arrow keys, j/k)
- [x] Tab switching between Summary and JSON views ([ and ])
- [x] Real-time status bar
- [x] Clean exit (q or Ctrl+C)

### Architecture

```
lazyazure/
├── main.go                       # Entry point
├── PLAN.md                       # Full implementation plan
├── pkg/
│   ├── azure/
│   │   ├── client.go            # Azure SDK wrapper with DefaultAzureCredential
│   │   └── subscriptions.go     # Subscription operations
│   ├── domain/
│   │   ├── user.go              # User domain model
│   │   └── subscription.go      # Subscription domain model
│   ├── gui/
│   │   ├── gui.go               # Main TUI controller
│   │   └── panels/
│   │       ├── filtered_list.go # Generic filtered list with generics
│   │       └── ...
│   └── tasks/
│       └── tasks.go             # Async task management
```

## Usage

### Prerequisites
- Azure CLI installed and authenticated (`az login`)
- Go 1.26.1+ installed

### Building
```bash
go build .
```

### Running
```bash
./lazyazure
```

### Controls
- **Arrow Up/Down** or **j/k**: Navigate subscriptions list
- **[ / ]**: Switch between Summary and JSON tabs
- **r**: Refresh subscriptions list
- **q** or **Ctrl+C**: Quit

## Authentication

LazyAzure uses Azure's `DefaultAzureCredential` which automatically:
1. Checks environment variables
2. Checks for managed identity
3. Falls back to Azure CLI credentials (primary method for this app)

To authenticate:
```bash
az login
```

## Project Status

- **Phase 1 (MVP)**: ✅ Complete - Auth & subscriptions working
- **Phase 2**: 📝 Planned - Resource groups navigation
- **Phase 3**: 📝 Planned - Resources viewer
- **Phase 4**: 📝 Planned - Polish & advanced features

See `PLAN.md` for the full implementation roadmap.

## Dependencies

- [gocui](https://github.com/jesseduffield/gocui) - TUI framework
- Azure SDK for Go:
  - `azidentity` - Authentication
  - `azcore` - Core types
  - `armsubscriptions` - Subscription management
- [lo](https://github.com/samber/lo) - Utility functions
- [logrus](https://github.com/sirupsen/logrus) - Logging

## License

MIT
