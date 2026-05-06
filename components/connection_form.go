package components

import (
	"fmt"
	"net/url"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/jorgerojas26/lazysql/app"
	"github.com/jorgerojas26/lazysql/drivers"
	"github.com/jorgerojas26/lazysql/helpers"
	"github.com/jorgerojas26/lazysql/models"
)

type ConnectionForm struct {
	*tview.Flex
	StatusText      *tview.TextView
	Action         string
	form           *tview.Form
	showAdvanced   bool
	urlFieldAdded  bool
	profiles       []models.Profile
	activeProfile  int
	profileButtons *tview.Flex
	providerIndex  int
	providerRow    *tview.Flex
}

type FormFields struct {
	Name     *tview.InputField
	Hostname *tview.InputField
	Port     *tview.InputField
	Username *tview.InputField
	Password *tview.InputField
	Database *tview.InputField
	SSL      *tview.Checkbox
	ReadOnly *tview.Checkbox
	URL      *tview.InputField
	SSLCert  *tview.InputField
	SSLKey   *tview.InputField
	SSLCA    *tview.InputField
}

var providers = []string{"MySQL", "PostgreSQL", "SQLite", "MSSQL"}

var formFields *FormFields
var connectionsFormInstance *ConnectionForm

var providerButtons []*tview.Button

func providerToDriver(p int) string {
	switch p {
	case 0:
		return drivers.DriverMySQL
	case 1:
		return drivers.DriverPostgres
	case 2:
		return drivers.DriverSqlite
	case 3:
		return drivers.DriverMSSQL
	default:
		return drivers.DriverMySQL
	}
}

func providerToDefaultPort(p int) string {
	switch p {
	case 0:
		return "3306"
	case 1:
		return "5432"
	case 2:
		return ""
	case 3:
		return "1433"
	default:
		return ""
	}
}

func providerLabel(idx int) string {
	if idx >= 0 && idx < len(providers) {
		return providers[idx]
	}
	return providers[0]
}

func NewConnectionForm(connectionPages *models.ConnectionPages) *ConnectionForm {
	wrapper := tview.NewFlex()
	wrapper.SetDirection(tview.FlexColumnCSS)

	statusText := tview.NewTextView()
	statusText.SetBorderPadding(1, 1, 0, 0)

	form := &ConnectionForm{
		Flex:       wrapper,
		StatusText: statusText,
	}

	connectionsFormInstance = form
	formFields = &FormFields{}

	addForm := tview.NewForm().
		SetFieldBackgroundColor(app.Styles.InverseTextColor).
		SetButtonBackgroundColor(tview.Styles.InverseTextColor).
		SetLabelColor(tview.Styles.PrimaryTextColor).
		SetFieldTextColor(app.Styles.ContrastSecondaryTextColor)

	addForm.AddInputField("Name", "", 30, nil, nil)
	addForm.AddInputField("Hostname", "localhost", 30, nil, nil)
	addForm.AddInputField("Port", "", 30, nil, nil)
	addForm.AddInputField("Username", "", 30, nil, nil)
	addForm.AddInputField("Password", "", 30, nil, nil)
	addForm.AddInputField("Database", "", 30, nil, nil)
	addForm.AddCheckbox("SSL", false, func(checked bool) {
		connectionsFormInstance.handleSSLChange(checked)
	})
	addForm.AddCheckbox("Read-Only", false, nil)

	formFields.Name = addForm.GetFormItemByLabel("Name").(*tview.InputField)
	formFields.Hostname = addForm.GetFormItemByLabel("Hostname").(*tview.InputField)
	formFields.Port = addForm.GetFormItemByLabel("Port").(*tview.InputField)
	formFields.Username = addForm.GetFormItemByLabel("Username").(*tview.InputField)
	formFields.Password = addForm.GetFormItemByLabel("Password").(*tview.InputField)
	formFields.Database = addForm.GetFormItemByLabel("Database").(*tview.InputField)
	formFields.SSL = addForm.GetFormItemByLabel("SSL").(*tview.Checkbox)
	formFields.ReadOnly = addForm.GetFormItemByLabel("Read-Only").(*tview.Checkbox)

	formFields.Password.SetMaskCharacter('*')

	form.form = addForm

	providerRow := tview.NewFlex().SetDirection(tview.FlexColumn)
	providerLabel := tview.NewTextView().SetText("Provider:").SetTextAlign(tview.AlignRight)
	providerLabel.SetTextColor(app.Styles.PrimaryTextColor)
	providerRow.AddItem(providerLabel, 10, 0, false)

	providerButtons = make([]*tview.Button, len(providers))
	for i, name := range providers {
		idx := i
		btn := tview.NewButton(fmt.Sprintf(" %s ", name))
		btn.SetStyle(tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack))
		btn.SetSelectedFunc(func() {
			form.selectProvider(idx)
		})
		providerButtons[i] = btn
		providerRow.AddItem(btn, len(name)+3, 0, false)
		providerRow.AddItem(nil, 1, 0, false)
	}
	form.providerRow = providerRow
	form.selectProvider(0)

	buttonsWrapper := tview.NewFlex().SetDirection(tview.FlexColumn)

	saveButton := tview.NewButton("[yellow]F1 [dark]Save")
	saveButton.SetStyle(tcell.StyleDefault.Background(app.Styles.PrimaryTextColor))
	saveButton.SetBorder(true)
	buttonsWrapper.AddItem(saveButton, 0, 1, false)
	buttonsWrapper.AddItem(nil, 1, 0, false)

	testButton := tview.NewButton("[yellow]F2 [dark]Test")
	testButton.SetStyle(tcell.StyleDefault.Background(app.Styles.PrimaryTextColor))
	testButton.SetBorder(true)
	buttonsWrapper.AddItem(testButton, 0, 1, false)
	buttonsWrapper.AddItem(nil, 1, 0, false)

	connectButton := tview.NewButton("[yellow]F3 [dark]Connect")
	connectButton.SetStyle(tcell.StyleDefault.Background(app.Styles.PrimaryTextColor))
	connectButton.SetBorder(true)
	buttonsWrapper.AddItem(connectButton, 0, 1, false)
	buttonsWrapper.AddItem(nil, 1, 0, false)

	advancedButton := tview.NewButton("[yellow]F4 [dark]Advanced")
	advancedButton.SetStyle(tcell.StyleDefault.Background(app.Styles.PrimaryTextColor))
	advancedButton.SetBorder(true)
	buttonsWrapper.AddItem(advancedButton, 0, 1, false)
	buttonsWrapper.AddItem(nil, 1, 0, false)

	cancelButton := tview.NewButton("[yellow]Esc [dark]Cancel")
	cancelButton.SetStyle(tcell.StyleDefault.Background(tcell.Color(app.Styles.PrimaryTextColor)))
	cancelButton.SetBorder(true)
	buttonsWrapper.AddItem(cancelButton, 0, 1, false)

	wrapper.AddItem(form.createProfileButtonsArea(), 3, 0, false)
	wrapper.AddItem(providerRow, 3, 0, false)
	wrapper.AddItem(addForm, 0, 1, true)
	wrapper.AddItem(statusText, 2, 0, false)
	wrapper.AddItem(buttonsWrapper, 3, 0, false)

	wrapper.SetInputCapture(form.inputCapture(connectionPages))

	return form
}

func (form *ConnectionForm) selectProvider(idx int) {
	form.providerIndex = idx

	for i, btn := range providerButtons {
		if i == idx {
			btn.SetStyle(tcell.StyleDefault.Background(app.Styles.PrimaryTextColor).Foreground(app.Styles.ContrastSecondaryTextColor).Bold(true))
		} else {
			btn.SetStyle(tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack))
		}
	}

	provider := providers[idx]
	form.handleProviderChange(provider)

	formFields.Port.SetText(providerToDefaultPort(idx))

	if provider == "SQLite" {
		formFields.Hostname.SetText("")
		formFields.Username.SetText("")
		formFields.Password.SetText("")
	} else {
		formFields.Hostname.SetText("localhost")
	}

	app.App.ForceDraw()
}

func (form *ConnectionForm) handleProviderChange(provider string) {
	if formFields == nil || formFields.Hostname == nil {
		return
	}
	isSQLite := provider == "SQLite"

	formFields.Hostname.SetDisabled(isSQLite)
	formFields.Port.SetDisabled(isSQLite)
	formFields.Username.SetDisabled(isSQLite)
	formFields.Password.SetDisabled(isSQLite)
	formFields.SSL.SetDisabled(isSQLite)

	if isSQLite {
		formFields.SSL.SetChecked(false)
		form.handleSSLChange(false)
	}

	form.removeSSLFields()
}

func (form *ConnectionForm) handleSSLChange(enabled bool) {
	if formFields == nil {
		return
	}

	if enabled {
		form.addSSLFields()
	} else {
		form.removeSSLFields()
	}
}

func (form *ConnectionForm) addSSLFields() {
	if form.form.GetFormItemByLabel("SSL Cert") == nil {
		form.form.AddInputField("SSL Cert", "", 30, nil, nil)
		formFields.SSLCert = form.form.GetFormItemByLabel("SSL Cert").(*tview.InputField)
	}
	if form.form.GetFormItemByLabel("SSL Key") == nil {
		form.form.AddInputField("SSL Key", "", 30, nil, nil)
		formFields.SSLKey = form.form.GetFormItemByLabel("SSL Key").(*tview.InputField)
	}
	if form.form.GetFormItemByLabel("SSL CA") == nil {
		form.form.AddInputField("SSL CA", "", 30, nil, nil)
		formFields.SSLCA = form.form.GetFormItemByLabel("SSL CA").(*tview.InputField)
	}
}

func (form *ConnectionForm) removeSSLFields() {
	for _, label := range []string{"SSL Cert", "SSL Key", "SSL CA"} {
		idx := form.form.GetFormItemIndex(label)
		if idx >= 0 {
			form.form.RemoveFormItem(idx)
		}
	}
	formFields.SSLCert = nil
	formFields.SSLKey = nil
	formFields.SSLCA = nil
}

func (form *ConnectionForm) toggleAdvancedFields() {
	form.showAdvanced = !form.showAdvanced

	if form.showAdvanced {
		if !form.urlFieldAdded {
			form.form.AddInputField("URL", "", 50, nil, nil)
			formFields.URL = form.form.GetFormItemByLabel("URL").(*tview.InputField)
			form.urlFieldAdded = true
		}
	} else {
		if formFields.URL != nil {
			urlIndex := form.form.GetFormItemIndex("URL")
			if urlIndex >= 0 {
				form.form.RemoveFormItem(urlIndex)
			}
			form.urlFieldAdded = false
			formFields.URL = nil
		}
	}
}

func (form *ConnectionForm) inputCapture(connectionPages *models.ConnectionPages) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			connectionPages.SwitchToPage(pageNameConnectionSelection)
		} else if event.Key() == tcell.KeyF1 || event.Key() == tcell.KeyEnter {
			form.profiles[form.activeProfile] = form.getCurrentFormProfile()

			connectionName := formFields.Name.GetText()
			if connectionName == "" {
				form.StatusText.SetText("Connection name is required").SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
				return event
			}

			provider := providerToDriver(form.providerIndex)

			activeProfile := form.profiles[form.activeProfile]
			hostname := activeProfile.Hostname
			port := activeProfile.Port
			username := activeProfile.Username
			password := activeProfile.Password
			database := activeProfile.DBName
			sslEnabled := activeProfile.SSLEnabled
			sslCert := activeProfile.SSLCert
			sslKey := activeProfile.SSLKey
			sslCA := activeProfile.SSLCA
			readOnly := formFields.ReadOnly.IsChecked()

			var connectionURL string
			var err error

			if form.showAdvanced && formFields.URL != nil && formFields.URL.GetText() != "" {
				connectionURL = formFields.URL.GetText()
			} else {
				connectionURL, err = helpers.BuildConnectionURL(provider, username, password, hostname, port, database, sslEnabled, sslCert, sslKey, sslCA)
				if err != nil {
					form.StatusText.SetText(err.Error()).SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
					return event
				}
			}

			newConnection := models.Connection{
				Name:     connectionName,
				Provider: provider,
				URL:      connectionURL,
				ReadOnly: readOnly,
				Profiles: form.profiles,
			}

			databases := app.App.Connections()
			var newDatabases []models.Connection

			switch form.Action {
			case actionNewConnection:
				newDatabases = append(databases, newConnection)
				err := app.App.SaveConnections(newDatabases)
				if err != nil {
					form.StatusText.SetText(err.Error()).SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
					return event
				}
			case actionEditConnection:
				newDatabases = make([]models.Connection, len(databases))
				row, _ := connectionsTable.GetSelection()
				for i, database := range databases {
					if i == row {
						newDatabases[i] = newConnection
					} else {
						newDatabases[i] = database
					}
				}
				err := app.App.SaveConnections(newDatabases)
				if err != nil {
					form.StatusText.SetText(err.Error()).SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
					return event
				}
			}

			connectionsTable.SetConnections(newDatabases)
			connectionPages.SwitchToPage(pageNameConnectionSelection)

		} else if event.Key() == tcell.KeyF2 {
			provider := providerToDriver(form.providerIndex)
			hostname := formFields.Hostname.GetText()
			port := formFields.Port.GetText()
			username := formFields.Username.GetText()
			password := formFields.Password.GetText()
			database := formFields.Database.GetText()
			sslEnabled := formFields.SSL.IsChecked()
			sslCert := ""
			sslKey := ""
			sslCA := ""
			if formFields.SSLCert != nil {
				sslCert = formFields.SSLCert.GetText()
			}
			if formFields.SSLKey != nil {
				sslKey = formFields.SSLKey.GetText()
			}
			if formFields.SSLCA != nil {
				sslCA = formFields.SSLCA.GetText()
			}
			go form.testConnection(provider, username, password, hostname, port, database, sslEnabled, sslCert, sslKey, sslCA)
		} else if event.Key() == tcell.KeyF4 {
			form.toggleAdvancedFields()
		} else if event.Key() == tcell.KeyTab {
			form.cycleProvider()
			return nil
		}
		return event
	}
}

func (form *ConnectionForm) cycleProvider() {
	newIdx := (form.providerIndex + 1) % len(providers)
	form.selectProvider(newIdx)
}

func (form *ConnectionForm) testConnection(provider, username, password, hostname, port, dbname string, sslEnabled bool, sslCert, sslKey, sslCA string) {
	connectionURL, err := helpers.BuildConnectionURL(provider, username, password, hostname, port, dbname, sslEnabled, sslCert, sslKey, sslCA)
	if err != nil {
		form.StatusText.SetText(err.Error()).SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
		return
	}

	form.StatusText.SetText("Connecting...").SetTextColor(app.Styles.TertiaryTextColor)

	var db drivers.Driver

	switch provider {
	case drivers.DriverMySQL:
		db = &drivers.MySQL{}
	case drivers.DriverPostgres:
		db = &drivers.Postgres{}
	case drivers.DriverSqlite:
		db = &drivers.SQLite{}
	case drivers.DriverMSSQL:
		db = &drivers.MSSQL{}
	}

	err = db.TestConnection(connectionURL)

	if err != nil {
		form.StatusText.SetText(err.Error()).SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
	} else {
		form.StatusText.SetText("Connection success").SetTextColor(app.Styles.TertiaryTextColor)
	}
	app.App.ForceDraw()
}

func (form *ConnectionForm) SetAction(action string) {
	form.Action = action
}

func (form *ConnectionForm) ResetForm() {
	formFields.Name.SetText("")
	formFields.Hostname.SetText("localhost")
	formFields.Port.SetText("")
	formFields.Username.SetText("")
	formFields.Password.SetText("")
	formFields.Database.SetText("")
	formFields.SSL.SetChecked(false)
	formFields.ReadOnly.SetChecked(false)
	form.removeSSLFields()
	form.selectProvider(0)

	if form.showAdvanced {
		form.toggleAdvancedFields()
	}

	form.profiles = []models.Profile{{Name: "default"}}
	form.activeProfile = 0
	form.updateProfileButtons()
}

func (form *ConnectionForm) SetConnectionData(conn models.Connection) {
	formFields.Name.SetText(conn.Name)

	providerIndex := 0
	switch conn.Provider {
	case drivers.DriverMySQL:
		providerIndex = 0
	case drivers.DriverPostgres:
		providerIndex = 1
	case drivers.DriverSqlite:
		providerIndex = 2
	case drivers.DriverMSSQL:
		providerIndex = 3
	}
	form.selectProvider(providerIndex)

	var hostname, port, username, password, dbName string
	var sslEnabled bool
	var sslCert, sslKey, sslCA string

	if len(conn.Profiles) > 0 {
		profile := conn.Profiles[0]
		hostname = profile.Hostname
		port = profile.Port
		username = profile.Username
		password = profile.Password
		dbName = profile.DBName
		sslEnabled = profile.SSLEnabled
		sslCert = profile.SSLCert
		sslKey = profile.SSLKey
		sslCA = profile.SSLCA
	} else if conn.URL != "" {
		parsed, err := helpers.ParseConnectionString(conn.URL)
		if err == nil {
			hostname = parsed.Hostname()
			port = parsed.Port()
			if parsed.User != nil {
				username = parsed.User.Username()
				password, _ = parsed.User.Password()
			}
			dbName = parsed.Query().Get("dbname")
			if dbName == "" {
				dbName = parsed.Query().Get("database")
			}
			if dbName == "" {
				path := parsed.Path
				if path != "" && path != "/" {
					dbName = path[1:]
				}
			}
			sslStr := parsed.Query().Get("sslmode")
			sslEnabled = sslStr != "" && sslStr != "disable"
			tlsStr := parsed.Query().Get("tls")
			sslEnabled = sslEnabled || (tlsStr != "" && tlsStr != "false")
		}
	}

	formFields.Hostname.SetText(hostname)
	formFields.Port.SetText(port)
	formFields.Username.SetText(username)
	formFields.Password.SetText(password)
	formFields.Database.SetText(dbName)
	formFields.SSL.SetChecked(sslEnabled)
	form.handleSSLChange(sslEnabled)
	if sslEnabled {
		form.addSSLFields()
		if formFields.SSLCert != nil {
			formFields.SSLCert.SetText(sslCert)
		}
		if formFields.SSLKey != nil {
			formFields.SSLKey.SetText(sslKey)
		}
		if formFields.SSLCA != nil {
			formFields.SSLCA.SetText(sslCA)
		}
	}
	formFields.ReadOnly.SetChecked(conn.ReadOnly)

	if conn.URL != "" && len(conn.Profiles) == 0 {
		form.showAdvanced = true
		if !form.urlFieldAdded {
			form.form.AddInputField("URL", conn.URL, 50, nil, nil)
			formFields.URL = form.form.GetFormItemByLabel("URL").(*tview.InputField)
			form.urlFieldAdded = true
		} else if formFields.URL != nil {
			formFields.URL.SetText(conn.URL)
		}
	}

	form.loadProfilesFromConnection(conn)
}

func (form *ConnectionForm) createProfileButtonsArea() *tview.Flex {
	form.profileButtons = tview.NewFlex().SetDirection(tview.FlexColumn)

	if len(form.profiles) == 0 {
		form.profiles = append(form.profiles, models.Profile{Name: "default"})
		form.activeProfile = 0
	}
	form.updateProfileButtons()

	return form.profileButtons
}

func (form *ConnectionForm) updateProfileButtons() {
	form.profileButtons.Clear()

	addButton := tview.NewButton(" + ")
	addButton.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true))
	addButton.SetSelectedFunc(func() {
		form.showAddProfileModal()
	})

	deleteButton := tview.NewButton(" - ")
	deleteButton.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true))
	deleteButton.SetSelectedFunc(func() {
		form.deleteActiveProfile()
	})

	form.profileButtons.AddItem(addButton, 3, 0, false)
	form.profileButtons.AddItem(nil, 1, 0, false)
	form.profileButtons.AddItem(deleteButton, 3, 0, false)
	form.profileButtons.AddItem(nil, 1, 0, false)

	for i, profile := range form.profiles {
		indicator := "○ "
		textColor := app.Styles.SecondaryTextColor
		if i == form.activeProfile {
			indicator = "● "
			textColor = app.Styles.PrimaryTextColor
		}
		btn := tview.NewButton(fmt.Sprintf("%s%s", indicator, profile.Name))
		btn.SetStyle(tcell.StyleDefault.Foreground(textColor))
		idx := i
		btn.SetSelectedFunc(func() {
			form.selectProfile(idx)
		})
		form.profileButtons.AddItem(btn, 0, 1, false)
	}
}

func (form *ConnectionForm) showAddProfileModal() {
	modal := tview.NewModal()
	modal.SetText("Enter profile name:")
	modal.AddButtons([]string{"OK", "Cancel"})
	modal.SetBackgroundColor(app.Styles.PrimitiveBackgroundColor)
	modal.SetButtonActivatedStyle(tcell.StyleDefault.
		Background(app.Styles.InverseTextColor).
		Foreground(app.Styles.ContrastSecondaryTextColor),
	)
	modal.SetTextColor(app.Styles.PrimaryTextColor)

	inputField := tview.NewInputField()
	inputField.SetLabel("Name: ")
	inputField.SetFieldBackgroundColor(app.Styles.InverseTextColor)
	inputField.SetFieldTextColor(app.Styles.ContrastSecondaryTextColor)
	inputField.SetLabelColor(app.Styles.PrimaryTextColor)

	modalWrapper := tview.NewFlex().SetDirection(tview.FlexRow)
	modalWrapper.AddItem(inputField, 1, 0, true)
	modalWrapper.AddItem(modal, 0, 1, false)

	mainPages.AddPage("addProfileModal", modalWrapper, true, true)
	app.App.SetFocus(inputField)

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		mainPages.RemovePage("addProfileModal")
		if buttonLabel == "OK" {
			profileName := inputField.GetText()
			if profileName != "" {
				form.addProfile(profileName)
			}
		}
	})
}

func (form *ConnectionForm) addProfile(name string) {
	currentProfile := form.getCurrentFormProfile()
	currentProfile.Name = name
	form.profiles = append(form.profiles, currentProfile)
	form.activeProfile = len(form.profiles) - 1
	form.updateProfileButtons()
}

func (form *ConnectionForm) selectProfile(index int) {
	if index < 0 || index >= len(form.profiles) {
		return
	}

	profile := form.profiles[index]
	form.activeProfile = index

	formFields.Hostname.SetText(profile.Hostname)
	formFields.Port.SetText(profile.Port)
	formFields.Username.SetText(profile.Username)
	formFields.Password.SetText(profile.Password)
	formFields.Database.SetText(profile.DBName)
	formFields.SSL.SetChecked(profile.SSLEnabled)
	form.handleSSLChange(profile.SSLEnabled)
	if profile.SSLEnabled {
		form.addSSLFields()
		if formFields.SSLCert != nil {
			formFields.SSLCert.SetText(profile.SSLCert)
		}
		if formFields.SSLKey != nil {
			formFields.SSLKey.SetText(profile.SSLKey)
		}
		if formFields.SSLCA != nil {
			formFields.SSLCA.SetText(profile.SSLCA)
		}
	}

	providerName := driverToProvider(profile.Provider)
	for i, p := range providers {
		if p == providerName {
			form.selectProvider(i)
			break
		}
	}

	form.updateProfileButtons()
}

func (form *ConnectionForm) deleteActiveProfile() {
	if len(form.profiles) <= 1 {
		return
	}

	form.profiles = append(form.profiles[:form.activeProfile], form.profiles[form.activeProfile+1:]...)
	if form.activeProfile >= len(form.profiles) {
		form.activeProfile = len(form.profiles) - 1
	}

	form.selectProfile(form.activeProfile)
	form.updateProfileButtons()
}

func (form *ConnectionForm) getCurrentFormProfile() models.Profile {
	provider := providerToDriver(form.providerIndex)

	sslCert, sslKey, sslCA := "", "", ""
	if formFields.SSLCert != nil {
		sslCert = formFields.SSLCert.GetText()
	}
	if formFields.SSLKey != nil {
		sslKey = formFields.SSLKey.GetText()
	}
	if formFields.SSLCA != nil {
		sslCA = formFields.SSLCA.GetText()
	}

	return models.Profile{
		Name:       formFields.Name.GetText(),
		Provider:   provider,
		Hostname:   formFields.Hostname.GetText(),
		Port:       formFields.Port.GetText(),
		Username:   formFields.Username.GetText(),
		Password:   formFields.Password.GetText(),
		DBName:     formFields.Database.GetText(),
		SSLEnabled: formFields.SSL.IsChecked(),
		SSLCert:    sslCert,
		SSLKey:     sslKey,
		SSLCA:      sslCA,
	}
}

func (form *ConnectionForm) loadProfilesFromConnection(conn models.Connection) {
	form.profiles = nil
	form.activeProfile = 0

	if len(conn.Profiles) > 0 {
		for _, p := range conn.Profiles {
			form.profiles = append(form.profiles, p)
		}
		form.activeProfile = 0
	} else {
		profile := models.Profile{
			Name:     conn.Name,
			Provider: conn.Provider,
		}
		if conn.URL != "" {
			parsed, err := helpers.ParseConnectionString(conn.URL)
			if err == nil {
				profile.Hostname = parsed.Hostname()
				profile.Port = parsed.Port()
				if parsed.User != nil {
					profile.Username = parsed.User.Username()
					profile.Password, _ = parsed.User.Password()
				}
				profile.DBName = parsed.Query().Get("dbname")
				if profile.DBName == "" {
					profile.DBName = parsed.Query().Get("database")
				}
				if profile.DBName == "" {
					path := parsed.Path
					if path != "" && path != "/" {
						profile.DBName = path[1:]
					}
				}
			}
		} else {
			profile.Hostname = conn.Hostname
			profile.Port = conn.Port
			profile.Username = conn.Username
			profile.Password = conn.Password
			profile.DBName = conn.DBName
		}
		form.profiles = append(form.profiles, profile)
	}
}

func driverToProvider(d string) string {
	switch d {
	case drivers.DriverMySQL:
		return "MySQL"
	case drivers.DriverPostgres:
		return "PostgreSQL"
	case drivers.DriverSqlite:
		return "SQLite"
	case drivers.DriverMSSQL:
		return "MSSQL"
	default:
		return "MySQL"
	}
}

func parseURLParams(connURL string) (map[string]string, error) {
	result := make(map[string]string)
	if connURL == "" {
		return result, nil
	}

	u, err := url.Parse(connURL)
	if err != nil {
		return nil, err
	}

	if u.RawQuery != "" {
		for k, v := range u.Query() {
			if len(v) > 0 {
				result[k] = v[0]
			}
		}
	}

	return result, nil
}