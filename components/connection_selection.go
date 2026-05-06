package components

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/jorgerojas26/lazysql/app"
	"github.com/jorgerojas26/lazysql/commands"
	"github.com/jorgerojas26/lazysql/drivers"
	"github.com/jorgerojas26/lazysql/helpers"
	"github.com/jorgerojas26/lazysql/helpers/logger"
	"github.com/jorgerojas26/lazysql/models"
)

type ConnectionSelection struct {
	*tview.Flex
	StatusText *tview.TextView
}

func NewConnectionSelection(connectionForm *ConnectionForm, connectionPages *models.ConnectionPages) *ConnectionSelection {
	wrapper := tview.NewFlex()

	wrapper.SetDirection(tview.FlexColumnCSS)

	buttonsWrapper := tview.NewFlex().SetDirection(tview.FlexRowCSS)

	newButton := tview.NewButton("[yellow]N[dark]ew")
	newButton.SetStyle(tcell.StyleDefault.Background(app.Styles.PrimitiveBackgroundColor))
	newButton.SetBorder(true)

	buttonsWrapper.AddItem(newButton, 0, 1, false)
	buttonsWrapper.AddItem(nil, 1, 0, false)

	connectButton := tview.NewButton("[yellow]C[dark]onnect")
	connectButton.SetStyle(tcell.StyleDefault.Background(app.Styles.PrimitiveBackgroundColor))
	connectButton.SetBorder(true)

	buttonsWrapper.AddItem(connectButton, 0, 1, false)
	buttonsWrapper.AddItem(nil, 1, 0, false)

	editButton := tview.NewButton("[yellow]E[dark]dit")
	editButton.SetStyle(tcell.StyleDefault.Background(app.Styles.PrimitiveBackgroundColor))
	editButton.SetBorder(true)

	buttonsWrapper.AddItem(editButton, 0, 1, false)
	buttonsWrapper.AddItem(nil, 1, 0, false)

	deleteButton := tview.NewButton("[yellow]D[dark]elete")
	deleteButton.SetStyle(tcell.StyleDefault.Background(app.Styles.PrimitiveBackgroundColor))
	deleteButton.SetBorder(true)

	buttonsWrapper.AddItem(deleteButton, 0, 1, false)
	buttonsWrapper.AddItem(nil, 1, 0, false)

	quitButton := tview.NewButton("[yellow]Q[dark]uit")
	quitButton.SetStyle(tcell.StyleDefault.Background(app.Styles.PrimitiveBackgroundColor))
	quitButton.SetBorder(true)

	buttonsWrapper.AddItem(quitButton, 0, 1, false)
	buttonsWrapper.AddItem(nil, 1, 0, false)

	statusText := tview.NewTextView()
	statusText.SetBorderPadding(1, 1, 0, 0)

	wrapper.AddItem(NewConnectionsTable(), 0, 1, true)
	wrapper.AddItem(statusText, 4, 0, false)
	wrapper.AddItem(buttonsWrapper, 3, 0, false)

	cs := &ConnectionSelection{
		Flex:       wrapper,
		StatusText: statusText,
	}

	wrapper.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		connections := connectionsTable.GetConnections()

		command := app.Keymaps.Group(app.ConnectionGroup).Resolve(event)

		if len(connections) != 0 {
			row, _ := connectionsTable.GetSelection()
			selectedConnection := connections[row]

			switch command {
			case commands.Connect:
				go cs.Connect(selectedConnection)
			case commands.EditConnection:
				connectionPages.SwitchToPage(pageNameConnectionForm)
				connectionForm.SetConnectionData(selectedConnection)
				connectionForm.StatusText.SetText("")

				connectionForm.SetAction(actionEditConnection)
				return nil
			case commands.DeleteConnection:
				confirmationModal := NewConfirmationModal("")

				confirmationModal.SetDoneFunc(func(_ int, buttonLabel string) {
					mainPages.RemovePage(pageNameConfirmation)
					confirmationModal = nil

					if buttonLabel == "Yes" {
						newConnections := append(connections[:row], connections[row+1:]...)

						err := app.App.SaveConnections(newConnections)
						if err != nil {
							connectionsTable.SetError(err)
						} else {
							connectionsTable.SetConnections(newConnections)
						}

					}
				})

				mainPages.AddPage(pageNameConfirmation, confirmationModal, true, true)

				return nil
			}
		}

		switch command {
		case commands.NewConnection:
			connectionForm.SetAction(actionNewConnection)
			connectionForm.ResetForm()
			connectionForm.StatusText.SetText("")
			connectionPages.SwitchToPage(pageNameConnectionForm)
		case commands.Quit:
			if wrapper.HasFocus() {
				app.App.Stop()
			}
		}

		return event
	})

	return cs
}

func (cs *ConnectionSelection) Connect(connection models.Connection) *tview.Application {
	if len(connection.Profiles) > 1 {
		cs.ShowProfileSelectionModal(connection)
		return App.Draw()
	}

	return cs.doConnect(connection)
}

func (cs *ConnectionSelection) doConnect(connection models.Connection) *tview.Application {
	if mainPages.HasPage(connection.Name) {
		mainPages.SwitchToPage(connection.Name)
		return App.Draw()
	}

	if len(connection.Commands) > 0 {

		variables := map[string]string{}

		waitsForPort := strings.Contains(connection.URL, "${port}")
		waitsForPort = waitsForPort || slices.ContainsFunc(connection.Commands, func(command *models.Command) bool {
			return command.WaitForPort != ""
		})

		if waitsForPort {
			port, err := helpers.GetFreePort()
			if err != nil {
				cs.StatusText.SetText(err.Error()).SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
				return App.Draw()
			}
			variables["port"] = port
		}

		for i, command := range connection.Commands {
			message := fmt.Sprintf("Running command %d/%d...", i+1, len(connection.Commands))
			cs.StatusText.SetText(message).SetTextColor(app.Styles.TertiaryTextColor)
			App.Draw()

			cmd := command.Command
			for variable, value := range variables {
				cmd = strings.ReplaceAll(cmd, "${"+variable+"}", value)
			}

			markCommandComplete := App.Register()
			onCommandDone, waitToCaptureVariable := setupOutputVariableCommand(variables, command, markCommandComplete)

			timeout := time.Duration(command.Timeout) * time.Second
			if command.Timeout == 0 {
				timeout = 5 * time.Second
			}

			if err := helpers.RunCommand(App.Context(), cmd, timeout, onCommandDone); err != nil {
				cs.StatusText.SetText(err.Error()).SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
				return App.Draw()
			}

			waitToCaptureVariable()

			if command.WaitForPort != "" {
				interpolatedPort := command.WaitForPort
				for variable, value := range variables {
					interpolatedPort = strings.ReplaceAll(interpolatedPort, "${"+variable+"}", value)
				}

				if portInt, err := strconv.Atoi(interpolatedPort); err != nil || portInt < 0 || portInt >= 1<<16 {
					cs.StatusText.SetText("bad port: " + interpolatedPort).SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
					return App.Draw()
				}

				message := fmt.Sprintf("Waiting for port %s...", interpolatedPort)
				cs.StatusText.SetText(message).SetTextColor(app.Styles.TertiaryTextColor)
				App.Draw()

				if err := helpers.WaitForPort(App.Context(), interpolatedPort); err != nil {
					cs.StatusText.SetText(err.Error()).SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
					return App.Draw()
				}
			}
		}

		for variable, value := range variables {
			if variable == "" || value == "" {
				continue
			}
			connection.URL = strings.ReplaceAll(connection.URL, "${"+variable+"}", value)
		}
	}

	cs.StatusText.SetText("Connecting...").SetTextColor(app.Styles.TertiaryTextColor)
	App.Draw()

	var newDBDriver drivers.Driver

	switch connection.Provider {
	case drivers.DriverMySQL:
		newDBDriver = &drivers.MySQL{}
	case drivers.DriverPostgres:
		newDBDriver = &drivers.Postgres{}
	case drivers.DriverSqlite:
		newDBDriver = &drivers.SQLite{}
	case drivers.DriverMSSQL:
		newDBDriver = &drivers.MSSQL{}
	default:
		errorMsg := fmt.Sprintf("Unsupported database provider: '%s'. Valid providers are: mysql, postgres, sqlite3, sqlserver", connection.Provider)
		cs.StatusText.SetText(errorMsg).SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
		return App.Draw()
	}

	err := newDBDriver.Connect(connection.URL)
	if err != nil {
		cs.StatusText.SetText(err.Error()).SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
		return App.Draw()
	}

	selectedRow, selectedCol := connectionsTable.GetSelection()
	cell := connectionsTable.GetCell(selectedRow, selectedCol)
	cell.SetText(fmt.Sprintf("[green]* %s", cell.Text))
	cs.StatusText.SetText("")

	newHome := NewHomePage(connection, newDBDriver)
	newHome.Tree.SetCurrentNode(newHome.Tree.GetRoot())
	newHome.Tree.Wrapper.SetTitle(connection.Name)

	mainPages.AddAndSwitchToPage(connection.Name, newHome, true)
	App.SetFocus(newHome.Tree)

	return App.Draw()
}

func (cs *ConnectionSelection) ShowProfileSelectionModal(connection models.Connection) {
	modal := tview.NewModal()
	modal.SetText("Select a profile to connect with:")

	buttonLabels := make([]string, len(connection.Profiles))
	for i, profile := range connection.Profiles {
		profileName := profile.Name
		if profileName == "" {
			profileName = "default"
		}
		buttonLabels[i] = profileName
	}
	buttonLabels = append(buttonLabels, "Cancel")
	modal.AddButtons(buttonLabels)

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		mainPages.RemovePage("profileSelectionModal")

		if buttonLabel == "Cancel" || buttonIndex == -1 {
			return
		}

		for i, profile := range connection.Profiles {
			profileName := profile.Name
			if profileName == "" {
				profileName = "default"
			}
			if profileName == buttonLabel {
				cs.connectWithProfile(connection, i)
				break
			}
		}
	})

	mainPages.AddPage("profileSelectionModal", modal, true, true)
}

func (cs *ConnectionSelection) connectWithProfile(connection models.Connection, profileIndex int) {
	profile := connection.Profiles[profileIndex]

	connection.URL, _ = helpers.BuildConnectionURL(
		profile.Provider,
		profile.Username,
		profile.Password,
		profile.Hostname,
		profile.Port,
		profile.DBName,
		profile.SSLEnabled,
		profile.SSLCert,
		profile.SSLKey,
		profile.SSLCA,
	)

	cs.doConnect(connection)
}

func setupOutputVariableCommand(variables map[string]string, command *models.Command, markCommandComplete func()) (onCommandDone func(string), captureVariable func()) {
	if command.SaveOutputTo == "" {
		onCommandDone = func(_ string) { markCommandComplete() }
		return onCommandDone, func() {}
	}

	variableSaved := make(chan string)

	captureVariable = func() {
		output := <-variableSaved
		variables[command.SaveOutputTo] = output
		logger.Debug("Saved command output to variable", map[string]any{"Variable": command.SaveOutputTo, "Output": output, "Command": command.Command})
	}

	onCommandDone = func(output string) {
		variableSaved <- output
		close(variableSaved)
		markCommandComplete()
	}

	return onCommandDone, captureVariable
}
