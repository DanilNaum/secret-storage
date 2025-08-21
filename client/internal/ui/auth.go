package ui

import (
	"github.com/rivo/tview"

	"github.com/DanilNaum/secret-storage-client/internal/constants"
)

func (a *App) showAuthChoice() {
	modal := tview.NewModal().
		SetText(constants.WelcomeMessage).
		AddButtons([]string{constants.LoginButton, constants.RegisterButton, constants.ExitButton}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			switch buttonLabel {
			case constants.LoginButton:
				a.showLoginForm()
			case constants.RegisterButton:
				a.showRegistrationForm()
			case constants.ExitButton:
				a.app.Stop()
			}
		})

	a.pages.AddAndSwitchToPage(constants.AuthChoicePageName, modal, true)
}

func (a *App) showLoginForm() {
	form := tview.NewForm()

	usernameField := tview.NewInputField().SetLabel(constants.UsernameLabel).SetFieldWidth(20)
	passwordField := tview.NewInputField().SetLabel(constants.PasswordLabel).SetFieldWidth(20).SetMaskCharacter(constants.PasswordMask)

	form.AddFormItem(usernameField).
		AddFormItem(passwordField).
		AddButton(constants.LoginButton, func() {
			username := usernameField.GetText()
			password := passwordField.GetText()

			if username == "" || password == "" {
				form.SetTitle(constants.ErrorFillAllFields)
				return
			}

			err := a.serverHandlers.Authenticate(username, password)
			if err != nil {
				form.SetTitle(constants.ErrorTitle + ": " + err.Error())
				return
			}

			a.showMasterPasswordForm()
		}).
		AddButton(constants.BackButton, func() {
			a.pages.RemovePage(constants.LoginPageName)
			a.showAuthChoice()
		})

	form.SetBorder(true).SetTitle(constants.LoginTitle).SetTitleAlign(tview.AlignLeft)
	a.pages.AddAndSwitchToPage(constants.LoginPageName, form, true)
}

func (a *App) showRegistrationForm() {
	form := tview.NewForm()

	usernameField := tview.NewInputField().SetLabel(constants.UsernameLabel).SetFieldWidth(20)
	passwordField := tview.NewInputField().SetLabel(constants.PasswordLabel).SetFieldWidth(20).SetMaskCharacter(constants.PasswordMask)
	confirmPasswordField := tview.NewInputField().SetLabel(constants.ConfirmPasswordLabel).SetFieldWidth(20).SetMaskCharacter(constants.PasswordMask)

	form.AddFormItem(usernameField).
		AddFormItem(passwordField).
		AddFormItem(confirmPasswordField).
		AddButton(constants.RegisterButton, func() {
			username := usernameField.GetText()
			password := passwordField.GetText()
			confirmPassword := confirmPasswordField.GetText()

			if username == "" || password == "" || confirmPassword == "" {
				form.SetTitle(constants.ErrorFillAllFields)
				return
			}

			if password != confirmPassword {
				form.SetTitle(constants.ErrorPasswordsDoNotMatch)
				return
			}

			if len(password) < 6 {
				form.SetTitle(constants.ErrorPasswordTooShort)
				return
			}

			err := a.serverHandlers.Register(username, password)
			if err != nil {
				form.SetTitle(constants.ErrorTitle + ": " + err.Error())
				return
			}

			a.showMasterPasswordForm()
		}).
		AddButton(constants.BackButton, func() {
			a.pages.RemovePage(constants.RegisterPageName)
			a.showAuthChoice()
		})

	form.SetBorder(true).SetTitle(constants.RegistrationTitle).SetTitleAlign(tview.AlignLeft)
	a.pages.AddAndSwitchToPage(constants.RegisterPageName, form, true)
}



func (a *App) showMasterPasswordForm() {
	form := tview.NewForm()

	masterPasswordField := tview.NewInputField().SetLabel(constants.MasterPasswordLabel).SetFieldWidth(20).SetMaskCharacter(constants.PasswordMask)
	confirmMasterPasswordField := tview.NewInputField().SetLabel(constants.ConfirmMasterPasswordLabel).SetFieldWidth(20).SetMaskCharacter(constants.PasswordMask)

	form.AddFormItem(masterPasswordField).
		AddFormItem(confirmMasterPasswordField).
		AddButton(constants.ContinueButton, func() {
			masterPassword := masterPasswordField.GetText()
			confirmMasterPassword := confirmMasterPasswordField.GetText()

			if masterPassword == "" || confirmMasterPassword == "" {
				form.SetTitle(constants.ErrorFillAllFields)
				return
			}

			if masterPassword != confirmMasterPassword {
				form.SetTitle(constants.ErrorPasswordsDoNotMatch)
				return
			}

			if len(masterPassword) < 8 {
				form.SetTitle(constants.ErrorMasterPasswordTooShort)
				return
			}

			a.serverHandlers.SetMasterPassword(masterPassword)

			a.pages.RemovePage(constants.LoginPageName)
			a.pages.RemovePage(constants.RegisterPageName)
			a.pages.RemovePage(constants.AuthChoicePageName)
			a.pages.RemovePage(constants.RegisterSuccessPageName)
			a.pages.RemovePage(constants.MasterPasswordPageName)

			a.createMainUI()
		}).
		AddButton(constants.BackButton, func() {
			a.pages.RemovePage(constants.MasterPasswordPageName)
			a.showAuthChoice()
		})

	form.SetBorder(true).SetTitle(constants.MasterPasswordTitle).SetTitleAlign(tview.AlignLeft)
	a.pages.AddAndSwitchToPage(constants.MasterPasswordPageName, form, true)
}

func (a *App) logout() {
	a.serverHandlers.Logout()

	a.currentRecord = nil
	a.currentIsServer = false
	a.currentRecordID = ""
	a.passwordVisible = false

	a.pages.RemovePage(constants.MainPageName)
	a.showAuthChoice()
}