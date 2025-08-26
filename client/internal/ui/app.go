package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/DanilNaum/secret-storage-client/internal/constants"
	"github.com/DanilNaum/secret-storage-client/internal/models"
)

type storageReader interface {
	GetLocalRecords() []*models.Record
	GetServerRecords() []*models.ServerRecord
}

type localHandlers interface {
	CreateRecord(record *models.Record) error
	DeleteLocalRecord(record *models.Record) error
	LocalRecordChange(record *models.Record) error
	LocalRecordMovedToServer(record *models.Record, serverID string) error
	ServerRecordMovedToLocal(record *models.Record) error
}

type serverHandlers interface {
	Authenticate(username, password string) error
	Register(username, password string) error
	Logout() error
	IsAuthenticated() bool

	SetMasterPassword(password string)

	DeleteServerRecord(record *models.Record) error
	ServerRecordChange(record *models.Record) error
	ServerRecordOpen(id string) (*models.Record, error)

	RefreshServerRecord(id string) (*models.Record, error)

	UpdateServerRecordList() error

	GetCachedRecord(id string) (*models.Record, bool, time.Time, bool)
	SetCachedRecord(id string, record *models.Record)

	LocalRecordMovedToServer(record *models.Record) (string, error)
	DownloadFile(recordID, savePath string) (chan int, chan error)
}

//go:generate moq -out pages_moq_test.go . pages
type pages interface {
	AddAndSwitchToPage(name string, item tview.Primitive, resize bool) *tview.Pages
	AddPage(name string, item tview.Primitive, resize bool, visible bool) *tview.Pages
	Blur()
	Draw(tcell.Screen)
	Focus(delegate func(p tview.Primitive))
	GetFrontPage() (name string, item tview.Primitive)
	GetRect() (int, int, int, int)
	HasFocus() bool
	InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive))
	MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive)
	RemovePage(name string) *tview.Pages
	SetRect(x int, y int, width int, height int)
	SetTitle(title string) *tview.Box
	SwitchToPage(name string) *tview.Pages
}

// App represents the main application UI controller.
type App struct {
	app   *tview.Application
	pages pages
	flex  *tview.Flex

	storage        storageReader
	localHandlers  localHandlers
	serverHandlers serverHandlers

	localList        *tview.List
	serverList       *tview.List
	detailsPanel     *tview.TextView
	detailsContainer *tview.Flex

	activeList      *tview.List
	currentRecord   *models.Record
	currentIsServer bool
	currentRecordID string
	passwordVisible bool
}

// NewApp creates a new application instance with the provided dependencies.
func NewApp(storage storageReader, localHandlers localHandlers, serverHandlers serverHandlers, pages pages) *App {
	return &App{
		app:            tview.NewApplication(),
		pages:          pages,
		storage:        storage,
		localHandlers:  localHandlers,
		serverHandlers: serverHandlers,
	}
}

// Run starts the application and enters the main event loop.
func (a *App) Run() error {
	a.showAuthChoice()
	return a.app.SetRoot(a.pages, true).EnableMouse(true).Run()
}

func (a *App) createMainUI() {
	a.flex = tview.NewFlex()

	localPanel := a.createLocalPanel()
	serverPanel := a.createServerPanel()
	a.createDetailsPanel()

	a.flex.AddItem(localPanel, 0, constants.ListColumnProportion, true)
	a.flex.AddItem(serverPanel, 0, constants.ListColumnProportion, false)
	a.flex.AddItem(a.detailsContainer, 0, constants.DetailColumnProportion, false)

	a.activeList = a.localList
	a.updateListTitles()

	a.pages.AddPage(constants.MainPageName, a.flex, true, true)
	a.setupInputHandling()
	a.app.SetFocus(a.activeList)
}

func (a *App) createLocalPanel() *tview.Flex {
	a.localList = tview.NewList().
		SetSelectedFunc(func(index int, name string, secondName string, shortcut rune) {
			a.clearOtherSelection(a.localList)
			a.showLocalRecordDetails(index)
		}).
		SetWrapAround(false).
		SetHighlightFullLine(true)

	a.localList.SetBorder(true).SetTitle(constants.LocalRecordsTitle)
	a.updateLocalList()

	localButtons := tview.NewFlex().SetDirection(tview.FlexColumn)
	newRecordBtn := tview.NewButton(constants.NewRecordButton).SetSelectedFunc(func() {
		a.showRecordForm(nil, true)
	})
	localButtons.AddItem(newRecordBtn, 0, 1, false)

	localPanel := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.localList, 0, 1, true).
		AddItem(localButtons, constants.ButtonRowHeight, 1, false)

	return localPanel
}

func (a *App) createServerPanel() *tview.Flex {
	a.serverList = tview.NewList().
		SetSelectedFunc(func(index int, name string, secondName string, shortcut rune) {
			a.clearOtherSelection(a.serverList)
			a.showServerRecordDetails(index)
		}).
		SetWrapAround(false).
		SetHighlightFullLine(true)

	a.serverList.SetBorder(true).SetTitle(constants.ServerRecordsTitle)
	a.updateServerList()

	serverButtons := tview.NewFlex().SetDirection(tview.FlexColumn)
	syncBtn := tview.NewButton(constants.UpdateServerRecordListButton).SetSelectedFunc(func() {
		a.updateServerRecordList()
	})
	serverButtons.AddItem(syncBtn, 0, 1, false)

	serverPanel := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.serverList, 0, 1, true).
		AddItem(serverButtons, constants.ButtonRowHeight, 1, false)

	return serverPanel
}

func (a *App) createDetailsPanel() {
	a.detailsPanel = tview.NewTextView()
	a.detailsPanel.SetDynamicColors(true).
		SetScrollable(true).
		SetBorder(true).
		SetTitle(constants.DetailsTitle)

	a.detailsPanel.SetText(constants.SelectRecordMessage)

	a.detailsContainer = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.detailsPanel, 0, 1, false)
}

// UpdateLocalList refreshes the local records list display.
func (a *App) updateLocalList() {
	a.localList.Clear()
	for _, record := range a.storage.GetLocalRecords() {
		displayName := record.Name
		if record.HasServerID() {
			displayName += " [green](synced)[-]"
		}
		a.localList.AddItem(displayName, fmt.Sprintf(constants.TypeFieldLabel, record.Type.String()), 0, nil)
	}

	title := fmt.Sprintf(constants.RecordCountFormat, constants.LocalRecordsTitle, len(a.storage.GetLocalRecords()), constants.ScrollHint)
	a.localList.SetTitle(title)
}

// UpdateServerList refreshes the server records list display.
func (a *App) updateServerList() {
	a.serverList.Clear()
	for _, serverRecord := range a.storage.GetServerRecords() {
		cachedRecord, found, _, expired := a.serverHandlers.GetCachedRecord(serverRecord.ID)

		var displayName, displayType string
		if found && !expired {
			displayName = cachedRecord.Name + " " + constants.CachedIndicator
			displayType = fmt.Sprintf(constants.TypeFieldLabel, cachedRecord.Type.String())
		} else if found && expired {
			displayName = cachedRecord.Name + " " + constants.ExpiredIndicator
			displayType = fmt.Sprintf(constants.TypeFieldLabel, cachedRecord.Type.String())
		} else {
			displayName = serverRecord.Name
			displayType = fmt.Sprintf(constants.TypeFieldLabel, serverRecord.Type.String())
		}

		a.serverList.AddItem(displayName, displayType, 0, nil)
	}

	title := fmt.Sprintf(constants.RecordCountFormat, constants.ServerRecordsTitle, len(a.storage.GetServerRecords()), constants.ScrollHint)
	a.serverList.SetTitle(title)
}

// UpdateListTitles updates list titles to show which list is currently active.
func (a *App) updateListTitles() {
	localTitle := fmt.Sprintf(constants.RecordCountFormat, constants.LocalRecordsTitle, len(a.storage.GetLocalRecords()), constants.ScrollHint)
	serverTitle := fmt.Sprintf(constants.RecordCountFormat, constants.ServerRecordsTitle, len(a.storage.GetServerRecords()), constants.ScrollHint)

	if a.activeList == a.localList {
		localTitle = fmt.Sprintf(constants.ActiveTitleFormat, localTitle, constants.ActiveIndicator)
	} else {
		serverTitle = fmt.Sprintf(constants.ActiveTitleFormat, serverTitle, constants.ActiveIndicator)
	}

	a.localList.SetTitle(localTitle)
	a.serverList.SetTitle(serverTitle)
}

func (a *App) showLocalRecordDetails(index int) {
	records := a.storage.GetLocalRecords()
	if index < 0 || index >= len(records) {
		return
	}

	record := records[index]
	a.currentRecord = record
	a.currentIsServer = false
	a.currentRecordID = record.ID

	a.displayRecordDetails(record, false, time.Time{}, false)
}

func (a *App) showServerRecordDetails(index int) {
	serverRecords := a.storage.GetServerRecords()
	if index < 0 || index >= len(serverRecords) {
		return
	}

	serverRecord := serverRecords[index]
	a.currentRecordID = serverRecord.ID

	cachedRecord, found, cachedAt, expired := a.serverHandlers.GetCachedRecord(serverRecord.ID)

	if found {
		a.currentRecord = cachedRecord
		a.currentIsServer = true
		a.displayRecordDetails(cachedRecord, true, cachedAt, expired)
		return
	}

	record, err := a.serverHandlers.ServerRecordOpen(serverRecord.ID)
	if err != nil {
		a.showError(err.Error())
		return
	}
	

	if record != nil {
		record.IsServer = true
		record.ServerID = serverRecord.ID
		if record.Type == models.File && record.FilePath == "" {
			a.downloadFileForm(record)
			
		}
		a.serverHandlers.SetCachedRecord(serverRecord.ID, record)
		a.currentRecord = record
		a.currentIsServer = true
		
		a.displayRecordDetails(record, true, time.Now(), false)
		a.updateServerList()
	}
	
}

func (a *App) displayRecordDetails(record *models.Record, isServer bool, cachedAt time.Time, expired bool) {
	var content strings.Builder

	if isServer && !cachedAt.IsZero() {
		if expired {
			content.WriteString("[red]CACHE EXPIRED[-]\n")
		} else {
			content.WriteString("[blue]CACHED[-]\n")
		}
		content.WriteString(fmt.Sprintf(constants.CacheTimeFormat, cachedAt.Format("15:04:05")))
		content.WriteString("\n")
	}

	content.WriteString(fmt.Sprintf(constants.DetailNameLabel, record.Name))
	content.WriteString(fmt.Sprintf(constants.DetailTypeLabel, record.Type.String()))

	if !isServer && record.HasServerID() {
		content.WriteString(fmt.Sprintf("[green]Server ID:[-] %s\n", record.ServerID))
	}

	switch record.Type {
	case models.Credentials:
		content.WriteString(fmt.Sprintf(constants.DetailUsernameLabel, record.Username))
		if a.passwordVisible {
			content.WriteString(fmt.Sprintf(constants.DetailPasswordLabel, record.Password))
		} else {
			content.WriteString(fmt.Sprintf(constants.DetailPasswordLabel, strings.Repeat("*", len(record.Password))))
		}
	case models.TextData:
		content.WriteString(fmt.Sprintf(constants.DetailContentLabel, record.TextContent))
	case models.File:
		if record.FilePath != "" {
			content.WriteString(fmt.Sprintf(constants.DetailFileLabel, record.FilePath))
		} else {
			a.downloadFileForm( record)
		}
	}

	a.detailsPanel.SetText(content.String())

	buttons := tview.NewFlex().SetDirection(tview.FlexColumn)

	editBtn := tview.NewButton(constants.EditButton).SetSelectedFunc(func() {
		a.showRecordForm(record, false)
	})
	buttons.AddItem(editBtn, 0, 1, false)

	var copyBtn *tview.Button
	if isServer {
		copyBtn = tview.NewButton(constants.CopyToLocalButton).SetSelectedFunc(func() {
			a.copySelectedRecord()
		})
	} else {
		copyBtn = tview.NewButton(constants.CopyToServerButton).SetSelectedFunc(func() {
			a.copySelectedRecord()
		})
	}
	buttons.AddItem(copyBtn, 0, 1, false)

	deleteBtn := tview.NewButton(constants.DeleteButton).SetSelectedFunc(func() {
		a.deleteSelectedRecord()
	})
	buttons.AddItem(deleteBtn, 0, 1, false)

	if record.Type == models.Credentials {
		var passwordBtnText string
		if a.passwordVisible {
			passwordBtnText = constants.HidePasswordButtonWithHotKey
		} else {
			passwordBtnText = constants.ShowPasswordButtonWithHotKey
		}

		passwordBtn := tview.NewButton(passwordBtnText).SetSelectedFunc(func() {
			a.passwordVisible = !a.passwordVisible
			a.displayRecordDetails(record, isServer, cachedAt, expired)
		})
		buttons.AddItem(passwordBtn, 0, 1, false)
	}

	if isServer {
		refreshBtn := tview.NewButton(constants.RefreshButton).SetSelectedFunc(func() {
			a.refreshServerRecord()
		})
		buttons.AddItem(refreshBtn, 0, 1, false)
	}

	logoutBtn := tview.NewButton(constants.LogoutButton).SetSelectedFunc(func() {
		a.logout()
	})
	buttons.AddItem(logoutBtn, 0, 1, false)

	a.detailsContainer.Clear()
	a.detailsContainer.AddItem(a.detailsPanel, 0, 1, false)
	a.detailsContainer.AddItem(buttons, constants.ButtonRowHeight, 1, false)
}

func (a *App) refreshServerRecord() {
	if !a.currentIsServer || a.currentRecordID == "" {
		return
	}

	record, err := a.serverHandlers.RefreshServerRecord(a.currentRecordID)
	if err != nil {
		a.showError(err.Error())
		return
	}

	if record != nil {
		record.IsServer = true
		record.ServerID = a.currentRecordID
		a.serverHandlers.SetCachedRecord(a.currentRecordID, record)
		a.currentRecord = record
		a.displayRecordDetails(record, true, time.Now(), false)
		a.updateServerList()
	}
}

func (a *App) clearDetails() {
	a.currentRecord = nil
	a.currentIsServer = false
	a.currentRecordID = ""
	a.passwordVisible = false

	a.detailsPanel.SetText(constants.SelectRecordMessage)

	a.detailsContainer.Clear()
	a.detailsContainer.AddItem(a.detailsPanel, 0, 1, false)

	a.app.SetFocus(a.activeList)
}

func (a *App) switchActiveList() {
	if a.activeList == a.localList {
		a.activeList = a.serverList
	} else {
		a.activeList = a.localList
	}
	a.updateListTitles()
	a.app.SetFocus(a.activeList)
}

func (a *App) clearOtherSelection(currentList *tview.List) {
	a.activeList = currentList
	a.updateListTitles()
	a.app.SetFocus(a.activeList)
}

func (a *App) copySelectedRecord() {
	if a.currentRecord == nil {
		return
	}

	var err error

	if a.currentIsServer {
		err = a.localHandlers.ServerRecordMovedToLocal(a.currentRecord)
		if err != nil {
			a.showError(err.Error())
			return
		}
	} else {
		serverID, err := a.serverHandlers.LocalRecordMovedToServer(a.currentRecord)
		if err != nil {
			a.showError(err.Error())
			return
		}
		
		err = a.localHandlers.LocalRecordMovedToServer(a.currentRecord, serverID)
		if err != nil {
			a.showError(err.Error())
			return
		}

	}

	a.updateLocalList()
	a.updateServerList()
	a.updateListTitles()
}

func (a *App) deleteSelectedRecord() {
	if a.currentRecord == nil {
		return
	}

	modal := tview.NewModal().
		SetText(constants.DeleteConfirmText).
		AddButtons([]string{constants.YesButton, constants.NoButton}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == constants.YesButton {
				a.performDelete()
			}
			a.pages.SwitchToPage(constants.MainPageName)
		})

	a.pages.AddAndSwitchToPage("deleteConfirm", modal, true)
}

func (a *App) performDelete() {
	if a.currentRecord == nil {
		return
	}

	var err error

	if a.currentIsServer {
		err = a.serverHandlers.DeleteServerRecord(a.currentRecord)
	} else {
		err = a.localHandlers.DeleteLocalRecord(a.currentRecord)
	}

	if err != nil {
		a.showError(err.Error())
	} else {
		a.clearDetails()
		a.updateLocalList()
		a.updateServerList()
		a.updateListTitles()
	}
}

// func (a *App) syncRecords() {
// 	localRecords := a.storage.GetLocalRecords()
// 	serverIDs, err := a.serverHandlers.Sync(localRecords)
// 	for _, localRecord := range localRecords{
// 		localRecord.ServerID = serverIDs[localRecord.ID]
// 	}
// 	if err != nil {
// 		a.showError(err.Error())
// 		return
// 	}

//		a.updateLocalList()
//		a.updateServerList()
//		a.updateListTitles()
//	}
func (a *App) updateServerRecordList() {
	err := a.serverHandlers.UpdateServerRecordList()
	if err != nil {
		a.showError(err.Error())
		return
	}

	a.updateServerList()

}

func (a *App) showError(message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{constants.OkButton}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.SwitchToPage(constants.MainPageName)
		})

	a.pages.AddAndSwitchToPage(constants.ErrorPageName, modal, true)
}
