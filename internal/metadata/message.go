package metadata

import (
	"encoding/binary"
	"fmt"
	"io"
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

func ReadRecords(path, topicUUID, topicName string, partitionID uint32, offset uint64) ([]Record, error) {
	logName := fmt.Sprintf("%v-%v/%v.log", topicUUID, topicName, partitionID)
	logPath := filepath.Join(path, logName)
	f, err := os.Open(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log while reading: %w", err)
	}
	defer utils.CloseResource(f)

	newOffset, err := f.Seek(int64(offset), 0)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to offset: %w", err)
	}
	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get the length of the partition log: %w", err)
	}
	logLength := info.Size()
	currentOffset := int(newOffset)
	var records []Record
	for currentOffset < int(logLength) {
		header := make([]byte, 20)
		n, err := io.ReadFull(f, header)
		if err != nil {
			return nil, fmt.Errorf("failed to read record metadata: %w", err)
		}
		currentOffset += n
		record := Record{
			Offset:    binary.BigEndian.Uint64(header[0:8]),
			Timestamp: binary.BigEndian.Uint64(header[8:16]),
		}
		keyLength := binary.BigEndian.Uint32(header[16:20])

		record.Key = make([]byte, keyLength)
		n, err = io.ReadFull(f, record.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to read key: %w", err)
		}
		currentOffset += n
		valLengthBuf := make([]byte, 4)
		n, err = io.ReadFull(f, valLengthBuf)
		if err != nil {
			return nil, fmt.Errorf("failed to read value length: %w", err)
		}
		currentOffset += n
		valLength := binary.BigEndian.Uint32(valLengthBuf)
		record.Value = make([]byte, valLength)
		n, err = io.ReadFull(f, record.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to read value: %w", err)
		}
		currentOffset += n
		records = append(records, record)
	}

	return records, nil
}
