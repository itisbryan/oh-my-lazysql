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
	StatusText     *tview.TextView
	Action         string
	form           *tview.Form
	showAdvanced   bool
	urlFieldAdded  bool
	profiles       []models.Profile
	activeProfile  int
	profileButtons *tview.Flex
}

type FormFields struct {
	Name      *tview.InputField
	Provider  *tview.DropDown
	Hostname  *tview.InputField
	Port      *tview.InputField
	Username  *tview.InputField
	Password  *tview.InputField
	Database  *tview.InputField
	SSL       *tview.Checkbox
	ReadOnly  *tview.Checkbox
	URL       *tview.InputField
	SSLCert   *tview.InputField
	SSLKey    *tview.InputField
	SSLCA     *tview.InputField
}

var providers = []string{"MySQL", "PostgreSQL", "SQLite", "MSSQL"}

var formFields *FormFields
var connectionsFormInstance *ConnectionForm

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

func NewConnectionForm(connectionPages *models.ConnectionPages) *ConnectionForm {
	wrapper := tview.NewFlex()
	wrapper.SetDirection(tview.FlexColumnCSS)

	addForm := tview.NewForm().
		SetFieldBackgroundColor(app.Styles.InverseTextColor).
		SetButtonBackgroundColor(tview.Styles.InverseTextColor).
		SetLabelColor(tview.Styles.PrimaryTextColor).
		SetFieldTextColor(tview.Styles.ContrastSecondaryTextColor)

	addForm.AddInputField("Name", "", 40, nil, nil)
	addForm.AddDropDown("Provider", providers, 0, func(option string, optionIndex int) {
		connectionsFormInstance.handleProviderChange(option)
	})
	addForm.AddInputField("Hostname", "", 40, nil, nil)
	addForm.AddInputField("Port", "", 40, nil, nil)
	addForm.AddInputField("Username", "", 40, nil, nil)
	addForm.AddInputField("Password", "", 40, nil, nil)
	addForm.AddInputField("Database", "", 40, nil, nil)
	addForm.AddCheckbox("SSL Enabled", false, func(checked bool) {
		connectionsFormInstance.handleSSLChange(checked)
	})
	addForm.AddCheckbox("Read-Only", false, nil)

	addForm.AddButton("Show Advanced", func() {
		connectionsFormInstance.toggleAdvancedFields()
	})

	addForm.AddInputField("SSL Cert", "", 40, nil, nil)
	addForm.AddInputField("SSL Key", "", 40, nil, nil)
	addForm.AddInputField("SSL CA", "", 40, nil, nil)

	formFields = &FormFields{
		Name:     addForm.GetFormItemByLabel("Name").(*tview.InputField),
		Provider: addForm.GetFormItemByLabel("Provider").(*tview.DropDown),
		Hostname: addForm.GetFormItemByLabel("Hostname").(*tview.InputField),
		Port:     addForm.GetFormItemByLabel("Port").(*tview.InputField),
		Username: addForm.GetFormItemByLabel("Username").(*tview.InputField),
		Password: addForm.GetFormItemByLabel("Password").(*tview.InputField),
		Database: addForm.GetFormItemByLabel("Database").(*tview.InputField),
		SSL:      addForm.GetFormItemByLabel("SSL Enabled").(*tview.Checkbox),
		ReadOnly: addForm.GetFormItemByLabel("Read-Only").(*tview.Checkbox),
		SSLCert:  addForm.GetFormItemByLabel("SSL Cert").(*tview.InputField),
		SSLKey:   addForm.GetFormItemByLabel("SSL Key").(*tview.InputField),
		SSLCA:    addForm.GetFormItemByLabel("SSL CA").(*tview.InputField),
	}

	formFields.Password.SetMaskCharacter('*')
	formFields.Hostname.SetDisabled(true)
	formFields.Port.SetDisabled(true)
	formFields.Username.SetDisabled(true)
	formFields.Password.SetDisabled(true)
	formFields.SSLCert.SetDisabled(true)
	formFields.SSLKey.SetDisabled(true)
	formFields.SSLCA.SetDisabled(true)

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

	cancelButton := tview.NewButton("[yellow]Esc [dark]Cancel")
	cancelButton.SetStyle(tcell.StyleDefault.Background(tcell.Color(app.Styles.PrimaryTextColor)))
	cancelButton.SetBorder(true)
	buttonsWrapper.AddItem(cancelButton, 0, 1, false)

	statusText := tview.NewTextView()
	statusText.SetBorderPadding(1, 1, 0, 0)

	form := &ConnectionForm{
		Flex:       wrapper,
		form:      addForm,
		StatusText: statusText,
	}

	connectionsFormInstance = form

	wrapper.AddItem(addForm, 0, 1, true)
	wrapper.AddItem(form.createProfileButtonsArea(), 3, 0, false)
	wrapper.AddItem(statusText, 4, 0, false)
	wrapper.AddItem(buttonsWrapper, 3, 0, false)

	wrapper.SetInputCapture(form.inputCapture(connectionPages))

	return form
}

func (form *ConnectionForm) handleProviderChange(provider string) {
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
}

func (form *ConnectionForm) handleSSLChange(enabled bool) {
	formFields.SSLCert.SetDisabled(!enabled)
	formFields.SSLKey.SetDisabled(!enabled)
	formFields.SSLCA.SetDisabled(!enabled)
}

func (form *ConnectionForm) toggleAdvancedFields() {
	form.showAdvanced = !form.showAdvanced

	if form.showAdvanced {
		if !form.urlFieldAdded {
			form.form.AddInputField("URL", "", 60, nil, nil)
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

	for i := 0; i < form.form.GetButtonCount(); i++ {
		btn := form.form.GetButton(i)
		if btn != nil && btn.GetLabel() == "Show Advanced" {
			if form.showAdvanced {
				btn.SetLabel("Hide Advanced")
			} else {
				btn.SetLabel("Show Advanced")
			}
		}
	}
}

func (form *ConnectionForm) inputCapture(connectionPages *models.ConnectionPages) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			connectionPages.SwitchToPage(pageNameConnectionSelection)
		} else if event.Key() == tcell.KeyF1 || event.Key() == tcell.KeyEnter {
			form.profiles[form.activeProfile] = form.getCurrentFormProfile()

			connectionName := form.profiles[form.activeProfile].Name
			if connectionName == "" {
				connectionName = formFields.Name.GetText()
			}
			if connectionName == "" {
				form.StatusText.SetText("Connection name is required").SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
				return event
			}

			providerOption, _ := formFields.Provider.GetCurrentOption()
			provider := providerToDriver(providerOption)

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

			connectionURL, err := helpers.BuildConnectionURL(provider, username, password, hostname, port, database, sslEnabled, sslCert, sslKey, sslCA)
			if err != nil {
				form.StatusText.SetText(err.Error()).SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
				return event
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
			providerOption, _ := formFields.Provider.GetCurrentOption()
			provider := providerToDriver(providerOption)
			hostname := formFields.Hostname.GetText()
			port := formFields.Port.GetText()
			username := formFields.Username.GetText()
			password := formFields.Password.GetText()
			database := formFields.Database.GetText()
			sslEnabled := formFields.SSL.IsChecked()
			sslCert := formFields.SSLCert.GetText()
			sslKey := formFields.SSLKey.GetText()
			sslCA := formFields.SSLCA.GetText()
			go form.testConnection(provider, username, password, hostname, port, database, sslEnabled, sslCert, sslKey, sslCA)
		}
		return event
	}
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
	formFields.Provider.SetCurrentOption(providerIndex)
	form.handleProviderChange(providers[providerIndex])

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
				dbName = parsed.Query().Get("parse")
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
	formFields.SSLCert.SetText(sslCert)
	formFields.SSLKey.SetText(sslKey)
	formFields.SSLCA.SetText(sslCA)
	formFields.ReadOnly.SetChecked(conn.ReadOnly)

	if conn.URL != "" {
		form.showAdvanced = true
		if !form.urlFieldAdded {
			form.form.AddInputField("URL", conn.URL, 60, nil, nil)
			formFields.URL = form.form.GetFormItemByLabel("URL").(*tview.InputField)
			form.urlFieldAdded = true
		} else if formFields.URL != nil {
			formFields.URL.SetText(conn.URL)
		}
		for i := 0; i < form.form.GetButtonCount(); i++ {
			btn := form.form.GetButton(i)
			if btn != nil && btn.GetLabel() == "Show Advanced" {
				btn.SetLabel("Hide Advanced")
			}
		}
	}

	form.loadProfilesFromConnection(conn)
}

func (form *ConnectionForm) createProfileButtonsArea() *tview.Flex {
	form.profileButtons = tview.NewFlex().SetDirection(tview.FlexColumn)

	addButton := tview.NewButton("[yellow]+[dark]")
	addButton.SetBorder(true)
	addButton.SetSelectedFunc(func() {
		form.showAddProfileModal()
	})

	deleteButton := tview.NewButton("[dark]-[dark]")
	deleteButton.SetBorder(true)
	deleteButton.SetSelectedFunc(func() {
		form.deleteActiveProfile()
	})

	form.profileButtons.AddItem(addButton, 3, 0, false)
	form.profileButtons.AddItem(deleteButton, 3, 0, false)

	if len(form.profiles) == 0 {
		form.profiles = append(form.profiles, models.Profile{Name: "default"})
		form.activeProfile = 0
	}
	form.updateProfileButtons()

	return form.profileButtons
}

func (form *ConnectionForm) updateProfileButtons() {
	form.profileButtons.Clear()

	addButton := tview.NewButton("[yellow]+[dark]")
	addButton.SetBorder(true)
	addButton.SetSelectedFunc(func() {
		form.showAddProfileModal()
	})

	deleteButton := tview.NewButton("[dark]-[dark]")
	deleteButton.SetBorder(true)
	deleteButton.SetSelectedFunc(func() {
		form.deleteActiveProfile()
	})

	form.profileButtons.AddItem(addButton, 3, 0, false)
	form.profileButtons.AddItem(deleteButton, 3, 0, false)

	for i, profile := range form.profiles {
		indicator := "○"
		textColor := app.Styles.SecondaryTextColor
		if i == form.activeProfile {
			indicator = "●"
			textColor = app.Styles.PrimaryTextColor
		}
		btn := tview.NewButton(fmt.Sprintf("%s %s", indicator, profile.Name))
		btn.SetStyle(tcell.StyleDefault.Foreground(textColor))
		btn.SetBorder(true)
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
	modalWrapper.AddItem(modal, 1, 0, false)

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
	formFields.SSLCert.SetText(profile.SSLCert)
	formFields.SSLKey.SetText(profile.SSLKey)
	formFields.SSLCA.SetText(profile.SSLCA)

	providerName := driverToProvider(profile.Provider)
	for i, p := range providers {
		if p == providerName {
			formFields.Provider.SetCurrentOption(i)
			form.handleProviderChange(p)
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
	providerOption, _ := formFields.Provider.GetCurrentOption()
	provider := providerToDriver(providerOption)

	return models.Profile{
		Name:       formFields.Name.GetText(),
		Provider:   provider,
		Hostname:   formFields.Hostname.GetText(),
		Port:       formFields.Port.GetText(),
		Username:   formFields.Username.GetText(),
		Password:   formFields.Password.GetText(),
		DBName:     formFields.Database.GetText(),
		SSLEnabled: formFields.SSL.IsChecked(),
		SSLCert:    formFields.SSLCert.GetText(),
		SSLKey:     formFields.SSLKey.GetText(),
		SSLCA:      formFields.SSLCA.GetText(),
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
		profile := models.Profile{Name: conn.Name}
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