package metadata

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gafkonian-go/internal/utils"
)

type Record struct {
	Offset    uint64
	Timestamp uint64
	Key       []byte
	Value     []byte
}

func (r *Record) Encode() []byte {
	body := make([]byte, 0)
	buf32 := make([]byte, 4)
	buf64 := make([]byte, 8)

	utils.AppendBuf64(&buf64, r.Offset, &body)
	utils.AppendBuf64(&buf64, r.Timestamp, &body)
	utils.AppendBuf32(&buf32, uint32(len(r.Key)), &body)
	body = append(body, r.Key...)
	utils.AppendBuf32(&buf32, uint32(len(r.Value)), &body)
	body = append(body, r.Value...)
	return body
}

// InitPartitionLog specifically done this way to empty all logs on every run
func InitPartitionLog(path string) error {
	for key, topicMetadata := range ClusterState.Topics {
		dirName := fmt.Sprintf("%v-%v", key, topicMetadata.Name)
		dirPath := filepath.Join(path, dirName)
		if err := os.MkdirAll(dirPath, 0o755); err != nil {
			return fmt.Errorf("failed to create topic directory for path %v: %w", dirPath, err)
		}
		for _, partition := range topicMetadata.Partitions {
			partitionName := fmt.Sprintf("%v.log", partition.ID)
			partitionLogPath := filepath.Join(dirPath, partitionName)
			f, err := os.Create(partitionLogPath)
			if err != nil {
				return fmt.Errorf("failed to create partition file for %v: %w", partitionName, err)
			}
			utils.CloseResource(f)
		}
	}
	return nil
}

func (r Record) Append(path, topicUUID, topicName string, partitionID uint32) (*Record, error) {
	logName := fmt.Sprintf("%v-%v/%v.log", topicUUID, topicName, partitionID)
	logPath := filepath.Join(path, logName)
	f, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log while appending: %w", err)
	}
	defer utils.CloseResource(f)
	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get the offest: %w", err)
	}
	r.Offset = uint64(info.Size())
	recordBytes := r.Encode()
	_, err = f.Write(recordBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to append file: %w", err)
	}
	return &r, nil
}
