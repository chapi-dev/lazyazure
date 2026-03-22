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

## Phase 1: MVP - Authentication & Subscription List

**Goal:** Working TUI showing Azure auth status and subscription picker

### Implementation Steps:

1. **Project Setup**
   - Initialize Go module: `go mod init github.com/matsest/lazyazure`
   - Add dependencies:
     - `github.com/jesseduffield/gocui` (TUI)
     - `github.com/Azure/azure-sdk-for-go/sdk/azidentity`
     - `github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions`
     - `github.com/samber/lo` (utility functions)
     - `github.com/jesseduffield/lazycore/pkg/boxlayout` (layout)
     - `github.com/sirupsen/logrus` (logging)

2. **Azure Client Layer** (`pkg/azure/`)
   - `client.go`: Wrapper around Azure SDK clients
   - `auth.go`: Authentication using `DefaultAzureCredential`
     - Automatically uses Azure CLI credentials if available
     - Gets current user info
   - `subscriptions.go`: List subscriptions with name and ID

3. **Domain Models** (`pkg/domain/`)
   - `subscription.go`: Subscription struct with Name, ID, State
   - `user.go`: User info (name, email, tenant)

4. **TUI Foundation** (`pkg/gui/`)
   - `gui.go`: Main GUI struct, event loop
   - `layout.go`: View arrangement (left sidebar + right main)
   - `views.go`: View creation (side, main, status bar)
   - `keybindings.go`: Navigation (arrow keys, tabs, quit)

5. **Panels** (`pkg/gui/panels/`)
   - Reusable `SideListPanel[T]` with generics
   - `subscriptions_panel.go`: List subscriptions
   - `auth_panel.go`: Show current user info

6. **Right Panel Viewer** (`pkg/gui/viewers/`)
   - `json_viewer.go`: Pretty-printed JSON view
   - `summary_viewer.go`: Key-value table format
   - Tab switching between views

7. **Main Entry** (`main.go`)
   - Initialize Azure client with DefaultAzureCredential
   - Start GUI event loop

---

## Phase 2: Resource Groups & Navigation 🔄 IN PROGRESS (Hang Issue Fixed, Layout Changes Needed)

**Goal:** Interactive hierarchy: Subscriptions → Resource Groups

### Current Issues Fixed:
- ✅ **Hang Issue**: Fixed race condition in async task updates - all UI updates now properly queued through `gui.g.Update()`

### Layout Changes Needed:
Current implementation uses a single dynamic side panel. Need to change to lazydocker-style stacked panels:

**New Layout Design:**
```
Left Sidebar (stacked panels):
┌─────────────────────┐
│ Auth                │ <- Show current user (compact)
├─────────────────────┤
│ Subscriptions       │ <- List of subscriptions
│ • Sub A             │
│ • Sub B  [selected] │
├─────────────────────┤
│ Resource Groups     │ <- List of RGs for selected sub
│ • RG 1              │
│ • RG 2  [selected] │
├─────────────────────┤
│ Resources           │ <- Future: list resources
└─────────────────────┘

Right Panel (viewer):
┌──────────────────────────────────┐
│ Details [Summary] [JSON]        │
│                                  │
│ Shows info for selected item    │
│ from any left panel             │
└──────────────────────────────────┘
```

### Implementation Steps:

1. **Azure Client Updates** ✅
   - Add `resourcegroups.go`: List resource groups with name, location, state

2. **Domain Models** ✅
   - `resourcegroup.go`: ResourceGroup struct with Name, Location, ID, ProvisioningState, Tags

3. **Navigation System** ✅
   - Context-aware left panel (switch between subscriptions and resource groups view)
   - `viewMode` state tracks current view ("subscriptions" or "resourcegroups")
   - Navigation with arrow keys (j/k or ↑/↓)
   - Enter key to drill down from subscriptions to resource groups
   - Escape or 'h' key to go back to subscriptions

4. **Updated Panels** ✅
   - Dynamic side panel title changes based on view mode
   - `refreshSidePanel()` displays subscriptions or resource groups based on viewMode
   - Navigation methods updated to work with both data types

5. **Right Panel** ✅
   - `renderSubscriptionDetails()`: Shows subscription details (name, ID, state, tenant)
   - `renderResourceGroupDetails()`: Shows resource group details (name, location, ID, provisioning state, tags)
   - Both support Summary and JSON tabs

6. **Status Bar Updates** ✅
   - Shows different messages for subscriptions vs resource groups view
   - Context-sensitive help text

### Known Issues (Being Fixed):
- ~~Hang when loading resource groups~~ ✅ Fixed: Race condition in async UI updates
- Layout needs to change to stacked panels (see New Layout Design above)

### Next Steps:
1. Redesign side panel layout to show auth, subs, RGs stacked
2. Click/select in subscriptions panel loads RGs panel
3. Each panel independently scrollable
4. Right panel shows details for whatever is selected in left panels

4. **Updated Panels**
   - `resourcegroups_panel.go`: List groups for selected subscription
   - Enhance `subscriptions_panel.go`: On select, load resource groups

5. **Right Panel**
   - Show resource group details (name, location, provisioning state, tags)
   - JSON and summary tabs

---

## Phase 3: Resources & Deep Viewing

**Goal:** Full hierarchy: Subscriptions → Resource Groups → Resources

### Implementation Steps:

1. **Azure Client**
   - Add `resources.go`: List resources with type, name, location
   - Generic resource client using `armresources.NewClient()`

2. **Domain Models**
   - `resource.go`: Generic Azure resource (ID, Name, Type, Location, Tags)

3. **Updated Panels**
   - `resources_panel.go`: List resources in selected resource group
   - Context switching between all three levels

4. **Enhanced Viewer**
   - Full JSON representation of any selected resource
   - Summary view with formatted key fields
   - Tag display

---

## Phase 4: Polish & Advanced Features

**Goal:** Production-ready with UX improvements

### Implementation Steps:

1. **Search & Filtering**
   - Real-time search/filter in all panels
   - Fuzzy matching
   - Case-insensitive search

2. **Keyboard Shortcuts**
   - `/` for search
   - `q` or `Ctrl+C` to quit
   - Arrow keys for navigation
   - `Tab` for switching right panel tabs
   - `Enter` to drill down, `Esc` or `h` to go back

3. **Caching**
   - In-memory cache for API responses
   - Refresh with `r` key
   - Expire cache after time interval

4. **Configuration**
   - Config file support (`~/.config/lazyazure/config.yml`)
   - Theme customization
   - Default subscription preference

5. **Error Handling**
   - Graceful handling of auth failures
   - Retry logic for API calls
   - Status bar messages

6. **Performance**
   - Lazy loading (fetch on demand)
   - Pagination for large resource lists
   - Background refresh

---

## Project Structure

```
lazyazure/
├── main.go
├── go.mod
├── go.sum
├── pkg/
│   ├── app/
│   │   └── app.go              # Bootstrap & DI
│   ├── azure/
│   │   ├── client.go           # Azure SDK wrapper
│   │   ├── auth.go             # Authentication
│   │   ├── subscriptions.go    # Subscription operations
│   │   ├── resourcegroups.go   # Resource group operations
│   │   └── resources.go        # Resource operations
│   ├── domain/
│   │   ├── user.go
│   │   ├── subscription.go
│   │   ├── resourcegroup.go
│   │   └── resource.go
│   ├── gui/
│   │   ├── gui.go              # Main GUI & event loop
│   │   ├── layout.go           # View arrangement
│   │   ├── views.go            # View definitions
│   │   ├── keybindings.go
│   │   ├── panels/
│   │   │   ├── side_list_panel.go   # Generic panel
│   │   │   ├── context_state.go     # Tab management
│   │   │   ├── filtered_list.go     # Searchable list
│   │   │   ├── auth_panel.go
│   │   │   ├── subscriptions_panel.go
│   │   │   ├── resourcegroups_panel.go
│   │   │   └── resources_panel.go
│   │   └── viewers/
│   │       ├── json_viewer.go
│   │       └── summary_viewer.go
│   ├── config/
│   │   └── config.go
│   └── tasks/
│       └── tasks.go            # Async operations
├── internal/
│   └── utils/
│       └── helpers.go
└── README.md
```

---

## Key Technical Decisions

1. **TUI Library**: `gocui` - Same as lazydocker, proven, battle-tested
2. **Azure Auth**: `DefaultAzureCredential` - Seamlessly uses Azure CLI auth
3. **Generic Panels**: Go generics for type-safe, reusable UI components
4. **Async Tasks**: Background loading to keep UI responsive
5. **Layout**: Box-based responsive layout system from lazycore

---

## MVP Success Criteria

- [ ] User can launch and see current Azure CLI identity
- [ ] Left panel shows list of subscriptions (name, ID)
- [ ] Can navigate subscriptions with arrow keys
- [ ] Right panel shows subscription details in JSON and summary tabs
- [ ] Can switch tabs with `[` and `]`
- [ ] App gracefully handles Azure CLI not being logged in
- [ ] Clean exit with `q` or `Ctrl+C`
