package ui

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/DanilNaum/secret-storage-client/internal/constants"
	"github.com/DanilNaum/secret-storage-client/internal/models"
)

func (a *App) showRecordForm(record *models.Record, isNew bool) {
	form := tview.NewForm()
	modal := tview.NewModal()

	if isNew {
		a.showTypeSelectionModal(modal, form)
		return
	}

	a.createRecordForm(form, record, false)
	a.pages.AddAndSwitchToPage(constants.FormPageName, form, true)
}

func (a *App) showTypeSelectionModal(modal *tview.Modal, form *tview.Form) {
	modal.SetText(constants.SelectRecordTypeMessage).
		AddButtons([]string{constants.CredentialsType, constants.TextType, constants.FileType, constants.CancelButton}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == constants.CancelButton {
				a.pages.SwitchToPage(constants.MainPageName)
				return
			}

			var recordType models.RecordType
			switch buttonLabel {
			case constants.CredentialsType:
				recordType = models.Credentials
			case constants.TextType:
				recordType = models.TextData
			case constants.FileType:
				recordType = models.File
			}

			a.createRecordForm(form, &models.Record{Type: recordType}, true)
			a.pages.AddAndSwitchToPage(constants.FormPageName, form, true)
		})

	a.pages.AddAndSwitchToPage(constants.TypeSelectPageName, modal, true)
}

func (a *App) createRecordForm(form *tview.Form, record *models.Record, isNew bool) {
	form.Clear(true)

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})

	form.AddInputField(constants.NameLabel, record.Name, constants.InputFieldWidth, nil, func(text string) {
		record.Name = text
	})

	var passwordField *tview.InputField
	var passwordVisible bool

	switch record.Type {
	case models.Credentials:
		form.AddInputField(constants.UsernameLabel, record.Username, constants.InputFieldWidth, nil, func(text string) {
			record.Username = text
		})

		passwordVisible = false
		passwordField = tview.NewInputField().SetLabel(constants.PasswordLabel).SetFieldWidth(constants.InputFieldWidth).SetMaskCharacter(constants.PasswordMask)
		passwordField.SetText(record.Password)
		passwordField.SetChangedFunc(func(text string) {
			record.Password = text
		})

		form.AddFormItem(passwordField)

	case models.TextData:
		form.AddTextArea(constants.ContentLabel, record.TextContent, constants.InputFieldWidth, constants.TextAreaHeight, 0, func(text string) {
			record.TextContent = text
		})

	case models.File:
		form.AddInputField(constants.FilePathLabel, record.FilePath, constants.InputFieldWidth, nil, func(text string) {
			record.FilePath = text
		})
	}

	form.AddButton(constants.SaveButton, func() {
		a.saveRecord(record, isNew)
	})

	form.AddButton(constants.CancelButton, func() {
		a.pages.SwitchToPage(constants.MainPageName)
	})

	if record.Type == models.Credentials {
		form.AddButton(constants.ShowPasswordButton, func() {
			passwordVisible = !passwordVisible
			if passwordVisible {
				passwordField.SetMaskCharacter(0)
				form.GetButton(2).SetLabel(constants.HidePasswordButton)
			} else {
				passwordField.SetMaskCharacter(constants.PasswordMask)
				form.GetButton(2).SetLabel(constants.ShowPasswordButton)
			}
		})
	}
}

func (a *App) saveRecord(record *models.Record, isNew bool) {
	var err error

	if isNew {
		err = a.localHandlers.CreateRecord(record)
	} else {
		if record.IsServer {
			err = a.serverHandlers.ServerRecordChange(record)
		} else {
			err = a.localHandlers.LocalRecordChange(record)
		}
	}

	if err != nil {
		a.showError(err.Error())
		return
	}

	a.pages.SwitchToPage(constants.MainPageName)
	a.updateLocalList()
	a.updateServerList()
}

func (a *App) downloadFileForm(record *models.Record) {
	form := tview.NewForm()
	form.Clear(true)

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})
	var path string
	filePathField := tview.NewInputField().SetLabel("file path").SetFieldWidth(constants.InputFieldWidth)
	filePathField.SetChangedFunc(func(text string) {
		path = text
	})
	form.AddFormItem(filePathField)

	progressFeeler := tview.NewTextView().SetLabel("Прогресс:")

	form.AddFormItem(progressFeeler)

	form.AddButton("download", func() {
		a.downloadFile(record, progressFeeler, path)
	})
	form.AddButton("cancel", func() {
		a.pages.SwitchToPage(constants.MainPageName)
	})
	a.pages.AddAndSwitchToPage(constants.FormPageName, form, true)

}

func (a *App) downloadFile(record *models.Record, progressFeeler *tview.TextView, path string) {
	progress, errChan := a.serverHandlers.DownloadFile(record.ServerID, path)
LOOP:
	for {
		select {
		case progress, ok := <-progress:
			if !ok {
				break LOOP
			}
			progressFeeler.SetText(strconv.Itoa(progress) + "%")
		case err, ok := <-errChan:
			if !ok {
				break LOOP
			}
			if err != nil {
				a.showError(err.Error())
				return
			}
		}
	}
	a.pages.SwitchToPage(constants.MainPageName)
}
