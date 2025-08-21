// Package constants provides all string constants, configuration values,
// and UI-related constants used throughout the application.
// This centralization ensures consistency and makes localization easier.
package constants

import "time"

// UI Labels and Messages contains all user-visible strings.
const (
	// Window titles for main UI panels

	// LocalRecordsTitle is the title for the local records panel.
	LocalRecordsTitle = "Local Records"
	// ServerRecordsTitle is the title for the server records panel.
	ServerRecordsTitle = "Server Records"
	// DetailsTitle is the title for the record details panel.
	DetailsTitle = "Details"

	// Button labels for various UI actions

	// NewRecordButton is the label for creating new records.
	NewRecordButton = "F1: New Record"
	// SyncButton is the label for synchronizing all records.
	UpdateServerRecordListButton = "F2: Update Server Record Lists"
	// EditButton is the label for editing existing records.
	EditButton = "F3: Edit"
	// CopyToLocalButton is the label for copying server records to local storage.
	CopyToLocalButton = "F4: Copy to Local"
	// CopyToServerButton is the label for copying local records to server.
	CopyToServerButton = "F4: Copy to Server"
	// DeleteButton is the label for deleting records.
	DeleteButton = "F5: Delete"
	// RefreshButton is the label for refreshing server records.
	RefreshButton = "F6: Refresh from Server"
	// ShowPasswordButtonWithHotKey displays the show password button text with F7 hotkey indicator
	ShowPasswordButtonWithHotKey = "F7: Show Password"
	// HidePasswordButtonWithHotKey displays the hide password button text with F7 hotkey indicator
	HidePasswordButtonWithHotKey = "F7: Hide Password"
	// ShowPasswordButton displays the show password button text without hotkey indicator
	ShowPasswordButton = "Show Password"
	// HidePasswordButton displays the hide password button text without hotkey indicator
	HidePasswordButton = "Hide Password"
	// LogoutButton displays the logout button text with F8 hotkey indicator
	LogoutButton = "F8: Logout"

	// SaveButton is the label for saving form data.
	SaveButton = "Save"
	// CancelButton is the label for canceling operations.
	CancelButton = "Cancel"
	// YesButton is the label for confirmation dialogs.
	YesButton = "Yes"
	// NoButton is the label for rejection in dialogs.
	NoButton = "No"
	// OkButton is the label for acknowledgment dialogs.
	OkButton = "OK"
	// LoginButton is the label for login action.
	LoginButton = "Login"
	// RegisterButton is the label for register action.
	RegisterButton = "Register"
	// ExitButton is the label for exit action.
	ExitButton = "Exit"
	// BackButton is the label for back action.
	BackButton = "Back"
	// ContinueButton is the label for continue action.
	ContinueButton = "Continue"

	// Authentication-specific labels and messages

	// WelcomeMessage is the welcome text shown in auth choice dialog.
	WelcomeMessage = "Welcome to Secret Storage Client!\n\nChoose action:"
	// ConfirmPasswordLabel is the label for password confirmation field.
	ConfirmPasswordLabel = "Confirm Password"
	// MasterPasswordLabel is the label for master password field.
	MasterPasswordLabel = "Master Password"
	// ConfirmMasterPasswordLabel is the label for master password confirmation field.
	ConfirmMasterPasswordLabel = "Confirm Master Password"

	// Authentication form titles

	// LoginTitle is the title for login form.
	LoginTitle = "Login"
	// RegistrationTitle is the title for registration form.
	RegistrationTitle = "Registration"
	// MasterPasswordTitle is the title for master password setup form.
	MasterPasswordTitle = "Master Password Setup"

	// Authentication error messages

	// ErrorFillAllFields is shown when required fields are empty.
	ErrorFillAllFields = "Error: Fill all fields"
	// ErrorPasswordsDoNotMatch is shown when passwords don't match.
	ErrorPasswordsDoNotMatch = "Error: Passwords do not match"
	// ErrorPasswordTooShort is shown when password is too short.
	ErrorPasswordTooShort = "Error: Password must be at least 6 characters"
	// ErrorMasterPasswordTooShort is shown when master password is too short.
	ErrorMasterPasswordTooShort = "Error: Master password must be at least 8 characters"

	// Form field labels

	// NameLabel is the label for record name input field.
	NameLabel = "Name"
	// UsernameLabel is the label for username input field.
	UsernameLabel = "Username"
	// PasswordLabel is the label for password input field.
	PasswordLabel = "Password"
	// ContentLabel is the label for text content input field.
	ContentLabel = "Content"
	// FilePathLabel is the label for file path input field.
	FilePathLabel = "File Path"

	// User messages and prompts

	// SelectRecordMessage is shown when no record is selected.
	SelectRecordMessage = "Select a record to view details"
	// SelectRecordTypeMessage is shown in record type selection dialog.
	SelectRecordTypeMessage = "Select record type"
	// ScrollHint provides navigation instructions to users.
	ScrollHint = "Use ↑↓ to scroll"
	// ActiveIndicator shows which list is currently active.
	ActiveIndicator = "[ACTIVE]"
	// ServerErrorMessage is shown when server operations fail.
	ServerErrorMessage = "Server operation not available"
	// DeleteConfirmTitle is the title for delete confirmation dialog.
	DeleteConfirmTitle = "Confirm Delete"
	// DeleteConfirmText asks user to confirm record deletion.
	DeleteConfirmText = "Are you sure you want to delete this record?"
	// ErrorTitle is the title for error dialogs.
	ErrorTitle = "Error"

	// Cache status indicators

	// CachedIndicator shows that record data is cached and current.
	CachedIndicator = "[blue](cached)[-]"
	// ExpiredIndicator shows that cached record data has expired.
	ExpiredIndicator = "[red](expired)[-]"
	// LoadingIndicator shows that record data is being loaded.
	LoadingIndicator = "Loading..."
	// CacheTimeFormat is the format string for displaying cache timestamps.
	CacheTimeFormat = "[gray]Cached: %s[-]\n"

	// Record type display names

	// CredentialsType is the display name for credentials records.
	CredentialsType = "Credentials"
	// TextType is the display name for text data records.
	TextType = "Text"
	// FileType is the display name for file records.
	FileType = "File"
	// UnknownType is the display name for unrecognized record types.
	UnknownType = "Unknown"

	// Format strings for dynamic content

	// TypeFieldLabel formats the type information in record lists.
	TypeFieldLabel = "Type: %s"
	// RecordCountFormat formats the title with record count and hints.
	RecordCountFormat = "%s (%d) - %s"
	// ActiveTitleFormat formats titles for active panels.
	ActiveTitleFormat = "[green]%s %s[-]"

	// Detail view format strings

	// DetailNameLabel formats the name field in record details.
	DetailNameLabel = "[yellow]Name:[-] %s\n"
	// DetailTypeLabel formats the type field in record details.
	DetailTypeLabel = "[yellow]Type:[-] %s\n"
	// DetailUsernameLabel formats the username field in record details.
	DetailUsernameLabel = "[yellow]Username:[-] %s\n"
	// DetailPasswordLabel formats the password field in record details.
	DetailPasswordLabel = "[yellow]Password:[-] %s\n"
	// DetailContentLabel formats the content field in record details.
	DetailContentLabel = "[yellow]Content:[-]\n%s\n"
	// DetailFileLabel formats the file path field in record details.
	DetailFileLabel = "[yellow]File:[-] %s\n"
)

// Form configuration constants define the appearance and behavior of input forms.
const (
	// InputFieldWidth sets the width of input fields in characters.
	InputFieldWidth = 30
	// TextAreaHeight sets the height of text areas in lines.
	TextAreaHeight = 5
	// PasswordMask is the character used to mask password input.
	PasswordMask = '*'
)

// Layout proportions define the relative sizes of UI components.
const (
	// ListColumnProportion sets the relative width of list columns.
	ListColumnProportion = 1
	// DetailColumnProportion sets the relative width of the details column.
	DetailColumnProportion = 2
	// ButtonRowHeight sets the height of button rows.
	ButtonRowHeight = 1
)

// Page names define the identifiers for different UI pages.
const (
	// MainPageName is the identifier for the main application page.
	MainPageName = "main"
	// FormPageName is the identifier for record editing forms.
	FormPageName = "form"
	// TypeSelectPageName is the identifier for record type selection dialog.
	TypeSelectPageName = "typeSelect"
	// ErrorPageName is the identifier for error display dialogs.
	ErrorPageName = "error"
	// AuthChoicePageName is the identifier for the authentication choice page.
	AuthChoicePageName = "authChoice"
	// LoginPageName is the identifier for the login form page.
	LoginPageName = "login"
	// RegisterPageName is the identifier for the registration form page.
	RegisterPageName = "register"
	// RegisterSuccessPageName is the identifier for the registration success page.
	RegisterSuccessPageName = "registerSuccess"
	// MasterPasswordPageName is the identifier for the master password setup page.
	MasterPasswordPageName = "masterPassword"
)

// Cache configuration defines the behavior of the server record cache.
const (
	// DefaultCacheTTL is the default time-to-live for cached server records.
	// After this duration, cached records are considered expired and will
	// be refreshed from the server on next access.
	DefaultCacheTTL = 5 * time.Minute
)
