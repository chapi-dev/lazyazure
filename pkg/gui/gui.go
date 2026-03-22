package gui

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jesseduffield/gocui"
	"github.com/matsest/lazyazure/pkg/azure"
	"github.com/matsest/lazyazure/pkg/domain"
	"github.com/matsest/lazyazure/pkg/gui/panels"
	"github.com/matsest/lazyazure/pkg/tasks"
	"github.com/matsest/lazyazure/pkg/utils"
)

// Gui is the main GUI controller
type Gui struct {
	g           *gocui.Gui
	azureClient *azure.Client
	subClient   *azure.SubscriptionsClient
	rgClient    *azure.ResourceGroupsClient
	taskManager *tasks.TaskManager

	// Views
	sideView   *gocui.View
	mainView   *gocui.View
	statusView *gocui.View

	// Navigation state
	viewMode    string // "subscriptions" or "resourcegroups"
	selectedSub *domain.Subscription
	selectedRG  *domain.ResourceGroup

	// Data
	subscriptions  []*domain.Subscription
	resourceGroups []*domain.ResourceGroup

	// UI state
	tabIndex int // 0 = summary, 1 = json

	mu sync.RWMutex
}

// NewGui creates a new GUI instance
func NewGui(azureClient *azure.Client) (*Gui, error) {
	return &Gui{
		azureClient: azureClient,
		taskManager: tasks.NewTaskManager(),
		tabIndex:    0,
		viewMode:    "subscriptions",
	}, nil
}

// Run starts the GUI event loop
func (gui *Gui) Run() error {
	utils.Log("Gui.Run: Creating gocui...")
	g, err := gocui.NewGui(gocui.NewGuiOpts{
		OutputMode:       gocui.OutputTrue,
		RuneReplacements: map[rune]string{},
	})
	if err != nil {
		utils.Log("Gui.Run: ERROR creating gocui: %v", err)
		return err
	}
	defer g.Close()

	gui.g = g
	utils.Log("Gui.Run: gocui created successfully")

	// Set up initial layout
	utils.Log("Gui.Run: Setting up views...")
	maxX, maxY := g.Size()
	if err := gui.setupViews(maxX, maxY); err != nil {
		utils.Log("Gui.Run: ERROR setting up views: %v", err)
		return err
	}
	utils.Log("Gui.Run: Views set up successfully")

	// Set up keybindings
	utils.Log("Gui.Run: Setting up keybindings...")
	if err := gui.setupKeybindings(); err != nil {
		utils.Log("Gui.Run: ERROR setting up keybindings: %v", err)
		return err
	}
	utils.Log("Gui.Run: Keybindings set up successfully")

	// Initialize Azure clients
	utils.Log("Gui.Run: Initializing Azure clients...")
	subClient, err := gui.azureClient.InitSubscriptionsClient()
	if err != nil {
		utils.Log("Gui.Run: ERROR initializing subscription client: %v", err)
		return fmt.Errorf("failed to initialize subscription client: %w", err)
	}
	gui.subClient = subClient
	utils.Log("Gui.Run: Azure clients initialized")

	// Load initial data
	utils.Log("Gui.Run: Loading initial data...")
	gui.loadSubscriptions()

	// Start the main loop
	utils.Log("Gui.Run: Starting MainLoop...")
	return g.MainLoop()
}

func (gui *Gui) setupViews(maxX, maxY int) error {
	// Ensure minimum dimensions
	if maxX < 30 || maxY < 10 {
		return fmt.Errorf("terminal too small: need at least 30x10, got %dx%d", maxX, maxY)
	}

	sideWidth := maxX / 3
	if sideWidth < 10 {
		sideWidth = 10
	}

	// Side panel (left) - shows list
	if v, err := gui.g.SetView("side", 0, 0, sideWidth-1, maxY-2, 0); err != nil && !gocui.IsUnknownView(err) {
		return err
	} else {
		v.Title = " Subscriptions "
		v.Highlight = true
		v.SelBgColor = gocui.ColorBlue
		v.SelFgColor = gocui.ColorWhite
		gui.sideView = v
	}

	// Main panel (right) - shows details
	if v, err := gui.g.SetView("main", sideWidth, 0, maxX-1, maxY-2, 0); err != nil && !gocui.IsUnknownView(err) {
		return err
	} else {
		v.Title = " Details "
		v.Wrap = true
		gui.mainView = v
	}

	// Status bar (bottom) - at least 2 rows
	statusY := maxY - 2
	if statusY < 0 {
		statusY = 0
	}
	if v, err := gui.g.SetView("status", 0, statusY, maxX-1, maxY, 0); err != nil && !gocui.IsUnknownView(err) {
		return err
	} else {
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Frame = false
		gui.statusView = v
	}

	// Set initial focus if no view is currently focused
	if gui.g.CurrentView() == nil && gui.sideView != nil {
		if _, err := gui.g.SetCurrentView("side"); err != nil {
			return err
		}
	}

	gui.updateStatus()
	gui.refreshSidePanel()
	gui.refreshMainPanel()

	return nil
}

func (gui *Gui) setupKeybindings() error {
	utils.Log("setupKeybindings: Setting up keybindings...")

	// Quit - bind to ALL views including side, main, and status
	quitKeys := []struct {
		view string
		key  interface{}
	}{
		{"", gocui.KeyCtrlC},
		{"side", gocui.KeyCtrlC},
		{"main", gocui.KeyCtrlC},
		{"status", gocui.KeyCtrlC},
		{"", 'q'},
		{"side", 'q'},
	}

	for _, binding := range quitKeys {
		if err := gui.g.SetKeybinding(binding.view, binding.key, gocui.ModNone, gui.quit); err != nil {
			utils.Log("setupKeybindings: ERROR setting quit binding for %s: %v", binding.view, err)
			return err
		}
	}
	utils.Log("setupKeybindings: Quit keybindings set")

	// Navigation - Arrow keys (only on side view)
	utils.Log("setupKeybindings: Setting up navigation keys...")
	if err := gui.g.SetKeybinding("side", gocui.KeyArrowDown, gocui.ModNone, gui.nextLine); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding("side", gocui.KeyArrowUp, gocui.ModNone, gui.prevLine); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding("side", 'j', gocui.ModNone, gui.nextLine); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding("side", 'k', gocui.ModNone, gui.prevLine); err != nil {
		return err
	}

	// Drill down and back navigation
	utils.Log("setupKeybindings: Setting up Enter/Esc keys...")
	if err := gui.g.SetKeybinding("side", gocui.KeyEnter, gocui.ModNone, gui.enterPressed); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding("side", gocui.KeyEsc, gocui.ModNone, gui.goBack); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding("side", 'h', gocui.ModNone, gui.goBack); err != nil {
		return err
	}

	// Tab switching - global
	utils.Log("setupKeybindings: Setting up tab keys...")
	if err := gui.g.SetKeybinding("", '[', gocui.ModNone, gui.prevTab); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding("", ']', gocui.ModNone, gui.nextTab); err != nil {
		return err
	}

	// Refresh
	utils.Log("setupKeybindings: Setting up refresh key...")
	if err := gui.g.SetKeybinding("", 'r', gocui.ModNone, gui.refresh); err != nil {
		return err
	}

	utils.Log("setupKeybindings: All keybindings set successfully")
	return nil
}

func (gui *Gui) quit(g *gocui.Gui, v *gocui.View) error {
	utils.Log("quit: Ctrl+C or q pressed - quitting application")
	gui.taskManager.StopAll()
	utils.Log("quit: Task manager stopped")
	return gocui.ErrQuit
}

func (gui *Gui) nextLine(g *gocui.Gui, v *gocui.View) error {
	cx, cy := v.Cursor()

	// Read state without holding lock during select calls
	gui.mu.RLock()
	viewMode := gui.viewMode
	subCount := len(gui.subscriptions)
	rgCount := len(gui.resourceGroups)
	gui.mu.RUnlock()

	utils.Log("nextLine: cursor at (%d, %d), viewMode=%s, subCount=%d, rgCount=%d", cx, cy, viewMode, subCount, rgCount)

	if viewMode == "subscriptions" {
		if cy < subCount-1 {
			v.SetCursor(cx, cy+1)
			gui.selectSubscription(cy + 1)
			utils.Log("nextLine: moved to subscription %d", cy+1)
		} else {
			utils.Log("nextLine: already at last subscription (cy=%d, count=%d)", cy, subCount)
		}
	} else if viewMode == "resourcegroups" {
		if cy < rgCount-1 {
			v.SetCursor(cx, cy+1)
			gui.selectResourceGroup(cy + 1)
			utils.Log("nextLine: moved to resource group %d", cy+1)
		} else {
			utils.Log("nextLine: already at last resource group (cy=%d, count=%d)", cy, rgCount)
		}
	}
	return nil
}

func (gui *Gui) prevLine(g *gocui.Gui, v *gocui.View) error {
	cx, cy := v.Cursor()

	// Read state without holding lock during select calls
	gui.mu.RLock()
	viewMode := gui.viewMode
	gui.mu.RUnlock()

	utils.Log("prevLine: cursor at (%d, %d), viewMode=%s", cx, cy, viewMode)

	if cy > 0 {
		v.SetCursor(cx, cy-1)
		if viewMode == "subscriptions" {
			gui.selectSubscription(cy - 1)
			utils.Log("prevLine: moved to subscription %d", cy-1)
		} else if viewMode == "resourcegroups" {
			gui.selectResourceGroup(cy - 1)
			utils.Log("prevLine: moved to resource group %d", cy-1)
		}
	} else {
		utils.Log("prevLine: already at first item")
	}
	return nil
}

func (gui *Gui) selectResourceGroup(index int) {
	utils.Log("selectResourceGroup: selecting index %d", index)
	gui.mu.Lock()
	if index >= 0 && index < len(gui.resourceGroups) {
		gui.selectedRG = gui.resourceGroups[index]
	}
	gui.mu.Unlock()

	// Call refresh outside the lock to avoid deadlock
	gui.refreshMainPanel()
	utils.Log("selectResourceGroup: selection complete")
}

func (gui *Gui) enterPressed(g *gocui.Gui, v *gocui.View) error {
	utils.Log("enterPressed: Enter key pressed")
	if gui.viewMode == "subscriptions" && gui.selectedSub != nil {
		utils.Log("enterPressed: Loading resource groups for sub %s", gui.selectedSub.Name)
		gui.loadResourceGroups(gui.selectedSub.ID)
	} else {
		utils.Log("enterPressed: Not in subscriptions mode or no sub selected (viewMode=%s)", gui.viewMode)
	}
	return nil
}

func (gui *Gui) goBack(g *gocui.Gui, v *gocui.View) error {
	if gui.viewMode == "resourcegroups" {
		gui.viewMode = "subscriptions"
		gui.rgClient = nil
		gui.resourceGroups = nil
		gui.selectedRG = nil
		gui.sideView.Title = " Subscriptions "
		gui.refreshSidePanel()
		gui.refreshMainPanel()
		gui.updateStatus()
	}
	return nil
}

func (gui *Gui) selectSubscription(index int) {
	utils.Log("selectSubscription: selecting index %d", index)
	gui.mu.Lock()
	if index >= 0 && index < len(gui.subscriptions) {
		gui.selectedSub = gui.subscriptions[index]
	}
	gui.mu.Unlock()

	// Call refresh outside the lock to avoid deadlock
	gui.refreshMainPanel()
	utils.Log("selectSubscription: selection complete")
}

func (gui *Gui) nextTab(g *gocui.Gui, v *gocui.View) error {
	gui.tabIndex = (gui.tabIndex + 1) % 2
	gui.refreshMainPanel()
	return nil
}

func (gui *Gui) prevTab(g *gocui.Gui, v *gocui.View) error {
	gui.tabIndex = (gui.tabIndex - 1 + 2) % 2
	gui.refreshMainPanel()
	return nil
}

func (gui *Gui) refresh(g *gocui.Gui, v *gocui.View) error {
	if gui.viewMode == "subscriptions" {
		gui.loadSubscriptions()
	} else if gui.viewMode == "resourcegroups" && gui.selectedSub != nil {
		gui.loadResourceGroups(gui.selectedSub.ID)
	}
	return nil
}

func (gui *Gui) loadSubscriptions() {
	gui.taskManager.NewTask(func(ctx context.Context) {
		subs, err := gui.subClient.ListSubscriptions(ctx)
		if err != nil {
			gui.g.UpdateAsync(func(g *gocui.Gui) error {
				gui.updateStatusMessage(fmt.Sprintf("Error: %v", err))
				return nil
			})
			return
		}

		gui.mu.Lock()
		gui.subscriptions = subs
		if len(subs) > 0 && gui.selectedSub == nil {
			gui.selectedSub = subs[0]
		}
		gui.mu.Unlock()

		// Update UI using UpdateAsync
		gui.g.UpdateAsync(func(g *gocui.Gui) error {
			gui.refreshSidePanel()
			gui.refreshMainPanel()
			gui.updateStatusMessage(fmt.Sprintf("Loaded %d subscriptions", len(subs)))
			return nil
		})
	})
}

func (gui *Gui) loadResourceGroups(subscriptionID string) {
	// Show loading indicator
	gui.updateStatusMessage("Loading resource groups...")

	// Run the actual loading in a goroutine
	go func() {
		utils.Log("loadResourceGroups: goroutine started")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		utils.Log("loadResourceGroups: context created with 30s timeout")

		utils.Log("loadResourceGroups: creating resource groups client...")
		rgClient, err := azure.NewResourceGroupsClient(gui.azureClient, subscriptionID)
		if err != nil {
			utils.Log("loadResourceGroups: ERROR creating client: %v", err)
			gui.g.UpdateAsync(func(g *gocui.Gui) error {
				gui.updateStatusMessage(fmt.Sprintf("Error: %v", err))
				return nil
			})
			return
		}
		utils.Log("loadResourceGroups: client created successfully")

		utils.Log("loadResourceGroups: calling ListResourceGroups...")
		rgs, err := rgClient.ListResourceGroups(ctx)
		if err != nil {
			utils.Log("loadResourceGroups: ERROR listing RGs: %v", err)
			gui.g.UpdateAsync(func(g *gocui.Gui) error {
				gui.updateStatusMessage(fmt.Sprintf("Error: %v", err))
				return nil
			})
			return
		}
		utils.Log("loadResourceGroups: successfully loaded %d resource groups", len(rgs))

		// Update state
		utils.Log("loadResourceGroups: updating state...")
		gui.mu.Lock()
		gui.rgClient = rgClient
		gui.resourceGroups = rgs
		gui.viewMode = "resourcegroups"
		if len(rgs) > 0 {
			gui.selectedRG = rgs[0]
		}
		gui.mu.Unlock()
		utils.Log("loadResourceGroups: state updated")

		// Update UI - use UpdateAsync to avoid blocking
		utils.Log("loadResourceGroups: queuing UI update...")
		gui.g.UpdateAsync(func(g *gocui.Gui) error {
			utils.Log("loadResourceGroups: UI update callback executing...")
			gui.sideView.Title = " Resource Groups "
			gui.refreshSidePanel()
			gui.refreshMainPanel()
			gui.updateStatusMessage(fmt.Sprintf("Loaded %d resource groups", len(rgs)))
			gui.updateStatus()

			// Set cursor after refresh
			if len(rgs) > 0 {
				gui.sideView.SetCursor(0, 0)
			}
			utils.Log("loadResourceGroups: UI update complete")
			return nil
		})
		utils.Log("loadResourceGroups: goroutine finished")
	}()
}

func (gui *Gui) refreshSidePanel() {
	if gui.sideView == nil {
		return
	}

	gui.sideView.Clear()
	gui.mu.RLock()
	defer gui.mu.RUnlock()

	if gui.viewMode == "subscriptions" {
		for _, sub := range gui.subscriptions {
			fmt.Fprintln(gui.sideView, sub.DisplayString())
		}
	} else if gui.viewMode == "resourcegroups" {
		for _, rg := range gui.resourceGroups {
			fmt.Fprintln(gui.sideView, rg.DisplayString())
		}
	}
}

func (gui *Gui) refreshMainPanel() {
	if gui.mainView == nil {
		return
	}

	gui.mainView.Clear()
	gui.mu.RLock()
	defer gui.mu.RUnlock()

	if gui.viewMode == "subscriptions" {
		gui.renderSubscriptionDetails()
	} else if gui.viewMode == "resourcegroups" {
		gui.renderResourceGroupDetails()
	}
}

func (gui *Gui) renderSubscriptionDetails() {
	if gui.selectedSub == nil {
		fmt.Fprintln(gui.mainView, "No subscription selected")
		return
	}

	if gui.tabIndex == 0 {
		// Summary tab
		gui.mainView.Title = " Details [Summary] "
		fmt.Fprintf(gui.mainView, "Name: %s\n", gui.selectedSub.Name)
		fmt.Fprintf(gui.mainView, "ID: %s\n", gui.selectedSub.ID)
		fmt.Fprintf(gui.mainView, "State: %s\n", gui.selectedSub.State)
		fmt.Fprintf(gui.mainView, "Tenant ID: %s\n", gui.selectedSub.TenantID)
	} else {
		// JSON tab
		gui.mainView.Title = " Details [JSON] "
		data, err := json.MarshalIndent(gui.selectedSub, "", "  ")
		if err != nil {
			fmt.Fprintf(gui.mainView, "Error: %v\n", err)
			return
		}
		gui.mainView.Write(data)
	}
}

func (gui *Gui) renderResourceGroupDetails() {
	if gui.selectedRG == nil {
		fmt.Fprintln(gui.mainView, "No resource group selected")
		return
	}

	if gui.tabIndex == 0 {
		// Summary tab
		gui.mainView.Title = " Resource Group [Summary] "
		fmt.Fprintf(gui.mainView, "Name: %s\n", gui.selectedRG.Name)
		fmt.Fprintf(gui.mainView, "Location: %s\n", gui.selectedRG.Location)
		fmt.Fprintf(gui.mainView, "ID: %s\n", gui.selectedRG.ID)
		fmt.Fprintf(gui.mainView, "Provisioning State: %s\n", gui.selectedRG.ProvisioningState)
		if len(gui.selectedRG.Tags) > 0 {
			fmt.Fprintln(gui.mainView, "Tags:")
			for k, v := range gui.selectedRG.Tags {
				fmt.Fprintf(gui.mainView, "  %s: %s\n", k, v)
			}
		}
	} else {
		// JSON tab
		gui.mainView.Title = " Resource Group [JSON] "
		data, err := json.MarshalIndent(gui.selectedRG, "", "  ")
		if err != nil {
			fmt.Fprintf(gui.mainView, "Error: %v\n", err)
			return
		}
		gui.mainView.Write(data)
	}
}

func (gui *Gui) updateStatus() {
	if gui.statusView == nil {
		return
	}

	gui.statusView.Clear()
	gui.mu.RLock()
	defer gui.mu.RUnlock()

	if gui.viewMode == "subscriptions" {
		if gui.selectedSub != nil {
			fmt.Fprintf(gui.statusView, " Sub: %s | Enter to view RGs | Tab: %s | q:quit, r:refresh",
				gui.selectedSub.Name, gui.getTabName())
		} else {
			fmt.Fprint(gui.statusView, " q:quit, r:refresh")
		}
	} else if gui.viewMode == "resourcegroups" {
		if gui.selectedRG != nil {
			fmt.Fprintf(gui.statusView, " RG: %s | Esc/h:back | Tab: %s | q:quit, r:refresh",
				gui.selectedRG.Name, gui.getTabName())
		} else {
			fmt.Fprint(gui.statusView, " Esc/h:back to subs | q:quit, r:refresh")
		}
	}
}

func (gui *Gui) updateStatusMessage(msg string) {
	if gui.statusView == nil {
		return
	}

	gui.statusView.Clear()
	fmt.Fprint(gui.statusView, msg)
}

func (gui *Gui) getTabName() string {
	if gui.tabIndex == 0 {
		return "Summary"
	}
	return "JSON"
}

// Ensure FilteredList is used
var _ = panels.NewFilteredList[string]
