package storage

import (
	"context"
	"errors"

	"github.com/DanilNaum/secret-storage-server/internal/entity"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrNotFound          = errors.New("not found")
)

type storage struct {
	conn *pgxpool.Pool
}

// NewStorage creates a new storage instance with the provided database connection pool.
// It returns a pointer to the storage struct.
func NewStorage(conn *pgxpool.Pool) *storage {
	return &storage{
		conn: conn,
	}
}

func (s *storage) CreateUser(ctx context.Context, user *entity.UserDTO) (int, error) {
	query := "INSERT INTO users (login, password_hash, salt) VALUES ($1, $2, $3) RETURNING uuid"

	var id int
	err := s.conn.QueryRow(ctx, query, user.Login, user.PasswordHash, user.Salt).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return 0, ErrUserAlreadyExists
			}
		}
		return 0, err
	}
	return id, nil
}

func (s *storage) GetUserByLogin(ctx context.Context, login string) (*entity.UserDTO, error) {
	query := "SELECT uuid, login, password_hash, salt FROM users WHERE login = $1"

	var user entity.UserDTO
	err := s.conn.QueryRow(ctx, query, login).Scan(&user.ID, &user.Login, &user.PasswordHash, &user.Salt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (s *storage) CreateRecord(ctx context.Context, userID int, record *entity.RecordInfo) (error) {
	query := "INSERT into records (id, name, type_id) VALUES ($1, $2, (SELECT id FROM record_types WHERE type_name = $3))"
	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return  err
	}
	_, err = tx.Exec(ctx, query, record.ID, record.Name, record.Type.String())
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	query = "INSERT INTO user_records (user_id, record_id) VALUES ($1, $2)"
	_, err = tx.Exec(ctx, query, userID, record.ID)
	if err != nil {
		tx.Rollback(ctx)
		return  err
	}
	tx.Commit(ctx)
	return nil
}

func (s *storage) DeleteRecord(ctx context.Context, userID int, recordID string) error {
	query := "DELETE FROM user_records WHERE user_id = $1 AND record_id = $2"
	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, query, userID, recordID)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}
	query = "DELETE FROM records WHERE id = $1"
	_, err = tx.Exec(ctx, query, recordID)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}
	tx.Commit(ctx)
	return nil
}
func (s *storage) GetRecord(ctx context.Context, userID int, recordID string) (*entity.RecordInfo, error) {
	query := `SELECT id, name, (SELECT type_name FROM record_types WHERE id = r.type_id) AS type_name 
          FROM records r 
          WHERE id = $1 AND id IN (SELECT record_id FROM user_records WHERE user_id = $2)`
	var recordType string
	var record entity.RecordInfo
	err := s.conn.QueryRow(ctx, query, recordID, userID).Scan(&record.ID, &record.Name, &recordType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	record.Type = entity.RecordTypeFromString(recordType)
	return &record, nil
}
func (s *storage) ListRecords(ctx context.Context, userID int) ([]*entity.RecordInfo, error) {
	query := "SELECT id, name, (SELECT type_name FROM record_types WHERE id = r.type_id) AS type_name FROM records r WHERE id IN (SELECT record_id FROM user_records WHERE user_id = $1)"
	rows, err := s.conn.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*entity.RecordInfo
	for rows.Next() {
		var record entity.RecordInfo
		var recordType string
		err := rows.Scan(&record.ID, &record.Name, &recordType)
		if err != nil {
			return nil, err
		}

		record.Type = entity.RecordTypeFromString(recordType)
		records = append(records, &record)
	}

	return records, nil
}

func (s *storage) UpdateRecord(ctx context.Context, userID int, record *entity.RecordInfo) error {
	query := "SELECT (SELECT type_name FROM record_types WHERE id = r.type_id) AS type_name FROM  records r WHERE id = $1"
	var currentRecordType string
	err := s.conn.QueryRow(ctx, query, record.ID).Scan(&currentRecordType)
	if err != nil {
		return err
	}

	// Check if the record type has changed
	if entity.RecordTypeFromString(currentRecordType) != record.Type {
		return errors.New("record type has changed, cannot update")
	}

	qwery := "UPDATE records SET name = $1 WHERE id = $2 AND id IN (SELECT record_id FROM user_records WHERE user_id = $3)"
	_, err = s.conn.Exec(ctx, qwery, record.Name, record.ID, userID)
	if err != nil {
		return err
	}
	return nil
}


func (s *storage) CreateCredential(ctx context.Context,recordId string,cred *entity.Credential)error{
	qwery := "INSERT INTO credentials (record_id, login, password) VALUES ($1, $2, $3) ON CONFLICT (record_id) DO UPDATE SET login = $2, password = $3"
	_, err := s.conn.Exec(ctx, qwery, recordId, cred.Username, cred.Password)
	if err != nil {
		return err
	}
	return nil
}
func (s *storage) GetCredential(ctx context.Context, recordId string)(*entity.Credential,error){
	qwery :=	"SELECT login, password FROM credentials WHERE record_id = $1"
	var cred entity.Credential
	err := s.conn.QueryRow(ctx, qwery, recordId).Scan(&cred.Username, &cred.Password)
	if err != nil {
		return nil, err
	}
	return &cred, nil
}
func (s *storage) UpdateCredential(ctx context.Context,recordId string,cred *entity.Credential)error{
	qwery := "UPDATE credentials SET login = $1, password = $2 WHERE record_id = $3"
	_, err := s.conn.Exec(ctx, qwery, cred.Username, cred.Password, recordId)
	if err != nil {
		return err
	}
	return nil
}
func (s *storage) DeleteCredential(ctx context.Context, recordId string)error{
	qwery := "DELETE FROM credentials WHERE record_id = $1"
	_, err := s.conn.Exec(ctx, qwery, recordId)
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) CreateText(ctx context.Context,recordId string,text *entity.Text)error{
	qwery := "INSERT INTO credentials (record_id, content) VALUES ($1, $2) ON CONFLICT (record_id) DO UPDATE SET credentials = $2"
	_, err := s.conn.Exec(ctx, qwery, recordId, text.Content)
	if err != nil {
		return err
	}
	return nil
}
func (s *storage) GetText(ctx context.Context, recordId string)(*entity.Text,error){
	qwery :=	"SELECT content FROM text WHERE record_id = $1"
	var text entity.Text
	err := s.conn.QueryRow(ctx, qwery, recordId).Scan(&text.Content)
	if err != nil {
		return nil, err
	}
	return &text, nil
}
func (s *storage) UpdateText(ctx context.Context,recordId string,text *entity.Text)error{
	qwery := "UPDATE text SET content = $1 WHERE record_id = $2"
	_, err := s.conn.Exec(ctx, qwery, text.Content, recordId)
	if err != nil {
		return err
	}
	return nil
}
func (s *storage) DeleteText(ctx context.Context, recordId string)error{
	qwery := "DELETE FROM text WHERE record_id = $1"
	_, err := s.conn.Exec(ctx, qwery, recordId)
	if err != nil {
		return err
	}
	return nil
}