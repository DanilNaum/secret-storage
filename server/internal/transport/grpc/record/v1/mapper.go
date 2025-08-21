package record

import (
	"github.com/DanilNaum/secret-storage-server/internal/entity"
	pb "github.com/DanilNaum/secret-storage-server/pkg/proto"
)

func recordToEntityRecord(r *pb.Record) *entity.Record {
	record := &entity.Record{
		RecordInfo: &entity.RecordInfo{
			ID:   r.GetServerId(),
			Name: r.GetName(),
			Type: entity.RecordTypeUnknown,
		},
	}
	switch cont := r.GetContent().(type) {
	case *pb.Record_Credentials:
		record.Type = entity.RecordTypeCredentials
		record.Credential = &entity.Credential{
			Username: cont.Credentials.GetUsername(),
			Password: cont.Credentials.GetPassword(),
		}
	case *pb.Record_TextData:
		record.Type = entity.RecordTypeText
		record.Text = &entity.Text{
			Content: cont.TextData.GetTextContent(),
		}
	case *pb.Record_FileData:
		record.Type = entity.RecordTypeFile
	
	}
	return record
}

func recordFromEntityRecord(r *entity.Record) *pb.Record {
	record := &pb.Record{
		ServerId: r.ID,
		Name:     r.Name,
	}
	switch r.Type {
	case entity.RecordTypeCredentials:
		record.Content = &pb.Record_Credentials{
			Credentials: &pb.CredentialsData{
				Username: r.Credential.Username,
				Password: r.Credential.Password,
			},
		}
	case entity.RecordTypeText:
		record.Content = &pb.Record_TextData{
			TextData: &pb.TextData{
					// TextContent: r.Text.Content,
			},
		}
	case entity.RecordTypeFile:
		record.Content	= &pb.Record_FileData{
		    
		}

	}
	return record
}

func serverRecordFromEntityRecordInfo(r *entity.RecordInfo) *pb.ServerRecord {
	record := &pb.ServerRecord{
		Id: r.ID,
		Name:     r.Name,
	}
	switch r.Type {
	case entity.RecordTypeCredentials:
		record.Type = pb.RecordType_CREDENTIALS
	case entity.RecordTypeText:
		record.Type = pb.RecordType_TEXT_DATA
	case entity.RecordTypeFile:
		record.Type = pb.RecordType_FILE
	}
	return record
}

func serverRecordsFromEntityRecordsInfo(rs []*entity.RecordInfo) []*pb.ServerRecord {
    records := make([]*pb.ServerRecord, 0, len(rs)) 
	for _, r := range rs {
		records = append(records, serverRecordFromEntityRecordInfo(r))
	}
	return records
}