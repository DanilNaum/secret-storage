package ui

import (
	"time"

	"github.com/DanilNaum/secret-storage-client/internal/constants"
	"github.com/DanilNaum/secret-storage-client/internal/models"
	"github.com/gdamore/tcell/v2"
)

func (a *App) setupInputHandling() {
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentPageName, _ := a.pages.GetFrontPage()

		if currentPageName == constants.FormPageName || currentPageName == constants.TypeSelectPageName {
			return event
		}

		if currentPageName == constants.MainPageName {
			switch event.Key() {
			case tcell.KeyTab:
				a.switchActiveList()
				return nil

			case tcell.KeyF1:
				a.showRecordForm(nil, true)
				return nil

			case tcell.KeyF2:
				a.updateServerRecordList()
				return nil

			case tcell.KeyF3:
				if a.currentRecord != nil {
					a.showRecordForm(a.currentRecord, false)
				}
				return nil

			case tcell.KeyF4:
				a.copySelectedRecord()
				return nil

			case tcell.KeyF5:
				a.deleteSelectedRecord()
				return nil

			case tcell.KeyF6:
				if a.currentIsServer {
					a.refreshServerRecord()
				}
				return nil

			case tcell.KeyF7:
				if a.currentRecord != nil && a.currentRecord.Type == models.Credentials {
					a.passwordVisible = !a.passwordVisible
					if a.currentIsServer {
						cachedRecord, found, cachedAt, expired := a.serverHandlers.GetCachedRecord(a.currentRecordID)
						if found {
							a.displayRecordDetails(cachedRecord, true, cachedAt, expired)
						}
					} else {
						a.displayRecordDetails(a.currentRecord, false, time.Time{}, false)
					}
				}
				return nil

			case tcell.KeyF8:
				a.logout()
				return nil
			}
		}

		return event
	})
}