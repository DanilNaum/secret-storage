package entity

type recordType int32

func (r *recordType)String() string {
	switch *r {
	case RecordTypeText:
		return "text"
	case RecordTypeFile:
		return "file"
	case RecordTypeCredentials:
		return "credentials"
	default:
		return "unknown"
	}
}
const (
	RecordTypeUnknown recordType = iota
	RecordTypeText
	RecordTypeFile
	RecordTypeCredentials
)
func RecordTypeFromString(s string) recordType {
	switch s {
	case "text":
		return RecordTypeText
	case "file":
		return RecordTypeFile
	case "credentials":
		return RecordTypeCredentials
	default:
		return RecordTypeUnknown
	}
}

type RecordInfo struct{
	ID         string
	Name       string
	Type       recordType
}

type Record struct {
	*RecordInfo
	Credential *Credential
	Text       *Text

}

type Credential struct {
	Username string
	Password string
}

type Text struct {
	Content string
}
