package gui

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/jesseduffield/gocui"
	"github.com/matsest/lazyazure/pkg/utils"
)

// focusCLI switches focus to the CLI input panel
func (gui *Gui) focusCLI(g *gocui.Gui, v *gocui.View) error {
	utils.Log("focusCLI: Focusing CLI input")

	if _, err := gui.g.SetCurrentView("cli_input"); err != nil {
		utils.Log("focusCLI: ERROR setting current view: %v", err)
		return err
	}

	gui.mu.Lock()
	gui.activePanel = "cli_input"
	gui.cliActive = true
	gui.mu.Unlock()

	gui.updatePanelTitles()
	gui.updateStatus()

	return nil
}

// unfocusCLI returns focus from CLI to previous panel
func (gui *Gui) unfocusCLI(g *gocui.Gui, v *gocui.View) error {
	utils.Log("unfocusCLI: Leaving CLI input")

	if _, err := gui.g.SetCurrentView("subscriptions"); err != nil {
		utils.Log("unfocusCLI: ERROR setting current view: %v", err)
		return err
	}

	gui.mu.Lock()
	gui.activePanel = "subscriptions"
	gui.cliActive = false
	gui.mu.Unlock()

	gui.updatePanelTitles()
	gui.updateStatus()

	return nil
}

// executeCLI runs the copilot command with the input text
func (gui *Gui) executeCLI(g *gocui.Gui, v *gocui.View) error {
	input := strings.TrimSpace(v.Buffer())
	if input == "" {
		return nil
	}

	utils.Log("executeCLI: Running copilot with input: %s", input)

	// Clear input
	v.Clear()
	v.SetCursor(0, 0)

	// Save to history
	gui.mu.Lock()
	gui.cliHistory = append(gui.cliHistory, input)
	gui.cliHistoryIdx = len(gui.cliHistory)
	gui.mu.Unlock()

	// Show the command in output
	gui.cliOutputView.Clear()
	fmt.Fprintf(gui.cliOutputView, "\x1b[32m> %s\x1b[0m\n", input)

	// Execute in background with streaming output
	go func() {
		cmd := exec.Command("copilot", "-p", input)

		// Get stdout and stderr pipes for streaming
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			gui.g.UpdateAsync(func(g *gocui.Gui) error {
				fmt.Fprintf(gui.cliOutputView, "\x1b[31mError: %v\x1b[0m\n", err)
				return nil
			})
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			gui.g.UpdateAsync(func(g *gocui.Gui) error {
				fmt.Fprintf(gui.cliOutputView, "\x1b[31mError: %v\x1b[0m\n", err)
				return nil
			})
			return
		}

		if err := cmd.Start(); err != nil {
			gui.g.UpdateAsync(func(g *gocui.Gui) error {
				fmt.Fprintf(gui.cliOutputView, "\x1b[31mError starting copilot: %v\x1b[0m\n", err)
				return nil
			})
			return
		}

		// Stream stdout and stderr line by line
		go gui.streamReader(stdout)
		go gui.streamReader(stderr)

		// Wait for command to finish
		if err := cmd.Wait(); err != nil {
			gui.g.UpdateAsync(func(g *gocui.Gui) error {
				fmt.Fprintf(gui.cliOutputView, "\n\x1b[31mExit: %v\x1b[0m\n", err)
				return nil
			})
		}
	}()

	return nil
}

// streamReader reads from a reader line by line and streams to the CLI output
func (gui *Gui) streamReader(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		gui.g.UpdateAsync(func(g *gocui.Gui) error {
			fmt.Fprintln(gui.cliOutputView, line)
			return nil
		})
	}
}

// cliHistoryUp navigates to previous command in history
func (gui *Gui) cliHistoryUp(g *gocui.Gui, v *gocui.View) error {
	gui.mu.Lock()
	defer gui.mu.Unlock()

	if len(gui.cliHistory) == 0 || gui.cliHistoryIdx <= 0 {
		return nil
	}

	gui.cliHistoryIdx--
	cmd := gui.cliHistory[gui.cliHistoryIdx]

	v.Clear()
	fmt.Fprint(v, cmd)
	v.SetCursor(len(cmd), 0)

	return nil
}

// cliHistoryDown navigates to next command in history
func (gui *Gui) cliHistoryDown(g *gocui.Gui, v *gocui.View) error {
	gui.mu.Lock()
	defer gui.mu.Unlock()

	if len(gui.cliHistory) == 0 || gui.cliHistoryIdx >= len(gui.cliHistory)-1 {
		// At end of history, clear input
		gui.cliHistoryIdx = len(gui.cliHistory)
		v.Clear()
		v.SetCursor(0, 0)
		return nil
	}

	gui.cliHistoryIdx++
	cmd := gui.cliHistory[gui.cliHistoryIdx]

	v.Clear()
	fmt.Fprint(v, cmd)
	v.SetCursor(len(cmd), 0)

	return nil
}

// scrollCLIUp scrolls the CLI output up
func (gui *Gui) scrollCLIUp(g *gocui.Gui, v *gocui.View) error {
	ox, oy := gui.cliOutputView.Origin()
	if oy > 0 {
		gui.cliOutputView.SetOrigin(ox, oy-1)
	}
	return nil
}

// scrollCLIDown scrolls the CLI output down
func (gui *Gui) scrollCLIDown(g *gocui.Gui, v *gocui.View) error {
	ox, oy := gui.cliOutputView.Origin()
	gui.cliOutputView.SetOrigin(ox, oy+1)
	return nil
}
