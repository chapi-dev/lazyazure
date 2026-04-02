# LazyAzure Implementation Plan

## Overview
A TUI application for Azure resource management, inspired by lazydocker. It provides an interactive terminal interface for browsing Azure subscriptions, resource groups, and resources with detailed viewers.

## Architecture

### Inspiration from lazydocker
- TUI Library: `gocui` (same as lazydocker)
- Generic panel system with Go generics
- Async task management
- Tab-based right panel viewers
- Filterable/sortable lists
- Box layout system

### Azure SDK Stack
- Authentication: `github.com/Azure/azure-sdk-for-go/sdk/azidentity` (DefaultAzureCredential)
- Subscriptions: `github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions`
- Resource Groups: `github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources`
- Generic Resources: `github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources`

---

## Phase 1: MVP - Authentication & Subscription List ✅ COMPLETE

**Goal:** Working TUI showing Azure auth status and subscription picker

### Implementation Steps:

1. **Project Setup** ✅
   - Initialize Go module: `go mod init github.com/matsest/lazyazure`
   - Add dependencies:
     - `github.com/jesseduffield/gocui` (TUI)
     - `github.com/Azure/azure-sdk-for-go/sdk/azidentity`
     - `github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions`

2. **Azure Client Layer** (`pkg/azure/`) ✅
   - `client.go`: Wrapper around Azure SDK clients
   - `subscriptions.go`: List subscriptions with name and ID

3. **Domain Models** (`pkg/domain/`) ✅
   - `subscription.go`: Subscription struct with Name, ID, State
   - `user.go`: User info (name)

4. **TUI Foundation** (`pkg/gui/`) ✅
   - `gui.go`: Main GUI struct, event loop, view arrangement

5. **Panels** (`pkg/gui/panels/`) ✅
   - `filtered_list.go`: Generic filtered list component

6. **Right Panel Viewer** ✅
   - JSON and Summary tabs implemented directly in gui.go

7. **Main Entry** (`main.go`) ✅
   - Initialize Azure client with DefaultAzureCredential
   - Start GUI event loop

---

## Phase 2: Resource Groups & Navigation ✅ COMPLETE

**Goal:** Interactive hierarchy: Subscriptions → Resource Groups

### Implementation Summary:

1. **Azure Client Updates** ✅
   - Added `resourcegroups.go`: List resource groups with name, location, state
   - Uses `armresources` SDK

2. **Domain Models** ✅
   - `resourcegroup.go`: ResourceGroup struct with Name, Location, ID, ProvisioningState, Tags

3. **Layout Redesign** ✅
   - Stacked panel layout (inspired by lazydocker)
   - Auth panel (5 lines fixed) - shows current user
   - Subscriptions panel (20% of remaining sidebar)
   - Resource Groups panel (30% of remaining sidebar)
   - Resources panel (remaining ~50% of sidebar)
   - All panels visible simultaneously

4. **Navigation System** ✅
   - Tab key switches between subscriptions and resource groups panels
   - Enter key on subscription loads resource groups
   - Arrow keys navigate within active panel
   - Visual indicator (▶) shows active panel

5. **Right Panel** ✅
   - Shows subscription details (name, ID, state, tenant)
   - Shows resource group details (name, location, ID, provisioning state, tags)
   - Summary and JSON tabs

6. **Status Bar** ✅
   - Context-sensitive help text
   - Shows current panel and available actions

---

## Phase 3: Resources & Deep Viewing ✅ COMPLETE

**Goal:** Full hierarchy: Subscriptions → Resource Groups → Resources

### Implementation Summary:

1. **Azure Client** ✅
   - Added `resources.go`: List resources within resource groups using filter
   - Uses `armresources.NewClient()` with resource group filter

2. **Domain Models** ✅
   - Added `resource.go`: Generic Azure resource (ID, Name, Type, Location, Tags, Properties)

3. **Updated Panels** ✅
   - Added resources panel to stacked layout
   - Redistributed panel heights for 4-panel layout (auth, subs, RGs, resources)
   - Context switching between all three levels with Tab key
   - Enter key drills down: Subscriptions → RGs → Resources

4. **Enhanced Viewer** ✅
   - Main panel shows appropriate details for selected item type
   - Resource details include: Name, Type, Location, ID, Resource Group, Tags
   - JSON representation for all resource types
   - Summary view with formatted key fields

---

## Phase 4: Polish & Advanced Features

**Goal:** Production-ready with UX improvements

**Status:** Partially implemented - core features complete, advanced features in backlog

**CLI Arguments**
   - ✅ `--help` / `-h` - Show help message with environment variables
   - ✅ `--version` / `-v` - Show version and commit information
   - ✅ `--check-update` - Check for available updates (non-interactive)

**Keyboard Shortcuts**
   - ✅ `/` for search (real-time filter across all panels)
   - ✅ `n` / `N` to navigate between matches (main panel search)
   - ✅ `q` or `Ctrl+C` to quit
   - ✅ Arrow keys for navigation
   - ✅ `Tab` for switching right panel tabs
   - ✅ `Enter` to drill down, `Escape` to clear filter
   - ✅ `?` for version information and update checking

**Visual Polish**
- ✅ Color-coded keys in Summary view (green)
- ✅ JSON syntax highlighting with Chroma
- ✅ Bold text for list indicators
- ✅ Green border for focused panel
- ✅ Gray suffix formatting (subscription ID, location, resource type)
- ✅ Human-readable resource type names (e.g., "Virtual Machine" not "virtualMachines")
- ✅ Sorted tags and properties
- ✅ Case-insensitive resource type lookup

**Main Panel Search** ✅
- ✅ `/` to search within main/details panel (highlights matching lines)
- ✅ `n` / `N` to navigate between matches
- ✅ Case-insensitive search with yellow highlight for current match
- ✅ Light grey highlight for other matches
- ✅ Works on both Summary and JSON views

**Navigation**
- ✅ `q` or `Ctrl+C` to quit
- ✅ Arrow keys and `j`/`k` for navigation
- ✅ `Tab` / `Shift+Tab` for panel switching
- ✅ `Enter` to drill down hierarchy
- ✅ `[` / `]` for Summary/JSON tabs
- ✅ `r` for manual refresh
- ✅ `c` to copy portal link

**Performance**
- ✅ Lazy loading (fetch on demand)
- ✅ API version caching for resource providers
- ✅ Async task management (non-blocking UI)
- ✅ Background preloading with cache
  - Resource groups preloaded after subscription load
  - Top 5 RG resources preloaded after RG load
  - 5-minute TTL for RGs, 3-minute TTL for resources
  - Smart cache invalidation on refresh (panel-specific)

---

## Backlog: Future Enhancements

The following features are planned but not yet implemented:

### Search & Filtering ✅ COMPLETE
- ✅ Real-time search/filter in all panels (`/` key)
- ✅ Case-insensitive search on displayed text (name + suffix)
- ✅ Backspace, Ctrl+U (clear), Ctrl+W (delete word) support
- ✅ Escape to cancel, Enter to confirm
- 📝 Fuzzy matching (future enhancement)

### Portal Integration ✅ COMPLETE
- ✅ `c` key copies Azure Portal link to clipboard
- ✅ Cross-platform clipboard support (Linux: xclip/xsel/wl-copy, macOS/Windows: native)
- ✅ Generates correct portal URLs for subscriptions, resource groups, and resources
- ✅ Portal link includes resource ID path

### Navigation Improvements
- `Esc` or `h` to navigate back up hierarchy
- ✅ Open portal link in browser (cross-platform)
- ✅ Mouse navigation (click to change focus between boxes and items)
- ✅ Click list items to trigger Enter action
- ✅ Click Summary/JSON tabs to switch views

### Caching ✅ COMPLETE
- ✅ Three-tier cache system (RG lists, resource lists, full resource details)
- ✅ In-memory cache for API responses with size limits (100/500 entries)
- ✅ Cache expiration/invalidation (TTL-based: 5min for RGs, 3min for resources)
- ✅ LRU eviction (50% when cache is full)
- ✅ Smart invalidation on refresh (panel-specific)
- ✅ Background preloading for top 10 RGs

### Configuration
- Config file support (`~/.config/lazyazure/config.yml`)
- Theme customization
- Default subscription preference
- Custom keybindings

### Error Handling
- Retry logic with exponential backoff
- Better error messages in UI
- Connection status indicator

### Performance
- ✅ Background loading (preloading)
  - Resource groups preloaded after subscription selection
  - Top 10 RG resources preloaded in background
  - Silent operation (no UI updates if user navigates away)
- UI-level pagination controls
- Virtual scrolling for large lists
- Optimistic updates
- ✅ Loading indicators (already implemented)

---

## Project Structure

See [AGENTS.md](AGENTS.md) for the complete file organization and architecture details.

High-level overview:
```
lazyazure/
├── main.go              # Entry point (refactored for testability)
├── main_test.go         # CLI tests (argument parsing, version checking)
├── pkg/
│   ├── azure/           # Azure SDK clients with caching
│   ├── demo/            # Demo mode (mock Azure data)
│   ├── domain/          # Domain models (Subscription, ResourceGroup, Resource)
│   ├── resources/       # Resource type display names
│   ├── gui/             # TUI implementation with panels
│   ├── tasks/           # Async task management
│   └── utils/           # Utilities (logging, clipboard, browser, portal URLs)
├── scripts/             # TUI integration test scripts (tmux-based)
└── tools/               # Development tools (API version updater)
```

---

## Key Technical Decisions

1. **TUI Library**: `gocui` - Same as lazydocker, proven, battle-tested
2. **Azure Auth**: `DefaultAzureCredential` - Supports multiple auth methods (CLI, env vars, managed identity, etc.)
3. **Generic Panels**: Go generics for type-safe, reusable UI components
4. **Async Tasks**: Background loading to keep UI responsive
5. **Layout**: Box-based responsive layout system from lazycore

---

## MVP Success Criteria

- [x] User can launch and see current Azure identity
- [x] Left panel shows list of subscriptions (name, ID)
- [x] Can navigate subscriptions with arrow keys
- [x] Right panel shows subscription details in JSON and summary tabs
- [x] Can switch tabs with `[` and `]`
- [x] App gracefully handles authentication failures
- [x] Clean exit with `q` or `Ctrl+C`
- [x] Navigate to resource groups within subscriptions
- [x] View resource group details (name, location, provisioning state, tags)

## Phase 3 Success Criteria

- [x] Browse resources within resource groups
- [x] View full resource details with provider-specific properties
- [x] Dynamic API version lookup for resource types
- [x] Color-coded UI with syntax highlighting
- [x] Scrollable content in details panel
- [x] Sorted and formatted display of tags and properties

## Phase 4 Success Criteria (Mostly Complete)

- [x] Resource type display names (human-readable)
- [x] Gray suffix formatting for all sidebar items
- [x] Copy portal link to clipboard
- [x] Search/filter functionality (real-time, case-insensitive)
- [x] Main panel search with highlighting and navigation
- [x] CLI arguments (--help, --version, --check-update)
- [x] API response caching with size limits and TTL
- [x] Background preloading for resource groups and resources
- [ ] Configuration file support
- [ ] Background refresh (auto-refresh on timer)
- [ ] Further performance improvements
