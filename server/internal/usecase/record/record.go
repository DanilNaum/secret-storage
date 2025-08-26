package record

import (
	"context"

	"github.com/DanilNaum/secret-storage-server/internal/entity"
	"github.com/google/uuid"
)

type credentialsStorage interface {
	CreateCredential(context.Context, string, *entity.Credential) error
	GetCredential(context.Context, string) (*entity.Credential, error)
	UpdateCredential(context.Context, string, *entity.Credential) error
	DeleteCredential(context.Context, string) error
}
type textStorage interface {
	CreateText(ctx context.Context, recordId string, text *entity.Text) error
	GetText(ctx context.Context, recordId string) (*entity.Text, error)
	UpdateText(ctx context.Context, recordId string, text *entity.Text) error
	DeleteText(ctx context.Context, recordId string) error
}

type repository interface {
	CreateRecord(context.Context, int, *entity.RecordInfo) error
	DeleteRecord(context.Context, int, string) error
	GetRecord(context.Context, int, string) (*entity.RecordInfo, error)
	ListRecords(context.Context, int) ([]*entity.RecordInfo, error)
	UpdateRecord(context.Context, int, *entity.RecordInfo) error
}

type recordUsecase struct {
	repo               repository
	credentialsStorage credentialsStorage
	textStorage        textStorage
}

// NewRecordUsecase will create new an recordUsecase object representation of record.Usecase interface
func NewRecordUsecase(repo repository, credentialsStorage credentialsStorage, textStorage textStorage) *recordUsecase {
	return &recordUsecase{repo: repo, credentialsStorage: credentialsStorage, textStorage: textStorage}
}

func (u *recordUsecase) CreateRecord(ctx context.Context, userID int, record *entity.Record) (string, error) {
	if record.RecordInfo.ID == "" {
		record.RecordInfo.ID = uuid.NewString()
	}
	err := u.repo.CreateRecord(ctx, userID, record.RecordInfo)
	if err != nil {
		return "", err
	}

	switch record.Type {
	case entity.RecordTypeCredentials:
		err = u.credentialsStorage.CreateCredential(ctx, record.RecordInfo.ID, record.Credential)
		if err != nil {
			return "", err
		}
	case entity.RecordTypeText:
		err = u.textStorage.CreateText(ctx, record.RecordInfo.ID, record.Text)
		if err != nil {
			return "", err
		}
	}

	return record.ID, nil
}
func (u *recordUsecase) DeleteRecord(ctx context.Context, userID int, recordID string) error {
	err := u.repo.DeleteRecord(ctx, userID, recordID)
	if err != nil {
		return err
	}
	return nil
}
func (u *recordUsecase) GetRecord(ctx context.Context, userID int, recordID string) (*entity.Record, error) {
	recordInfo, err := u.repo.GetRecord(ctx, userID, recordID)
	if err != nil {
		return nil, err
	}
	record := &entity.Record{RecordInfo: recordInfo}

	switch record.Type {
	case entity.RecordTypeCredentials:
		cred, err := u.credentialsStorage.GetCredential(ctx, recordID)
		if err != nil {
			return nil, err
		}
		record.Credential = cred
	case entity.RecordTypeText:
		text, err := u.textStorage.GetText(ctx, recordID)
		if err != nil {
			return nil, err
		}
		record.Text = text
	}

	return record, nil
}
func (u *recordUsecase) ListRecords(ctx context.Context, userID int) ([]*entity.RecordInfo, error) {
	return u.repo.ListRecords(ctx, userID)
}
func (u *recordUsecase) UpdateRecord(ctx context.Context, userID int, record *entity.Record) error {
	err:=  u.repo.UpdateRecord(ctx, userID, record.RecordInfo)
	if err != nil {
		return err
	}
	switch record.Type {
	case entity.RecordTypeCredentials:
		err = u.credentialsStorage.UpdateCredential(ctx, record.RecordInfo.ID, record.Credential)
		if err != nil {
			return err
		}
	case entity.RecordTypeText:
		err = u.textStorage.UpdateText(ctx, record.RecordInfo.ID, record.Text)
		if err != nil {
			return err
		}
	}
	return nil
}
