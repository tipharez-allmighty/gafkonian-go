// Package metadata handles the parsing and in-memory storage of Kafka cluster state.
package metadata

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/gafkonian-go/internal/utils"
)

const batchSize int = 61

var ClusterState = &ClusterMetadata{
	Topics: make(map[utils.UUID]*TopicMetadata),
}

type ClusterMetadata struct {
	Topics map[utils.UUID]*TopicMetadata
}

type TopicMetadata struct {
	Name       string
	UUID       utils.UUID
	Partitions []PartitionMetadata
}

type PartitionMetadata struct {
	ID             uint32
	LeaderID       uint32
	ReplicaNodes   []uint32
	IsrNodes       []uint32
	LeaderEpoch    uint32
	PartitionEpoch uint32
}

func Load(path string) error {
	logData, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read metadata log file: %w", err)
	}
	fmt.Println("Loaded metadata log from path:", path)
	if err := parseMetadatLog(logData); err != nil {
		return fmt.Errorf("failed to parse log metadata file: %w", err)
	}
	return nil
}

func parseMetadatLog(logData []byte) error {
	offset := 0
	for offset < len(logData) {
		if offset+batchSize > len(logData) {
			break
		}
		batchLength := int32(binary.BigEndian.Uint32(logData[offset+8 : offset+12]))
		recordsCount := int32(binary.BigEndian.Uint32(logData[offset+batchSize-4 : offset+batchSize]))

		recordStart := offset + batchSize
		recordEnd := offset + 12 + int(batchLength)

		if recordEnd > len(logData) {
			return fmt.Errorf("batch length exceeds log data")
		}
		recordData := logData[recordStart:recordEnd]
		if err := parseRecords(recordData, recordsCount); err != nil {
			return err
		}
		offset = recordEnd
	}
	return nil
}

func parseRecords(data []byte, count int32) error {
	offset := 0
	for i := 0; i < int(count); i++ {
		_, n := binary.Varint(data[offset:])
		offset += n
		offset++
		_, n = binary.Varint(data[offset:])
		offset += n
		_, n = binary.Varint(data[offset:])
		offset += n
		keyLen, n := binary.Varint(data[offset:])
		offset += n
		if keyLen > 0 {
			offset += int(keyLen)
		}

		valueLen, n := binary.Varint(data[offset:])
		offset += n
		if valueLen < 0 {
			continue
		}

		valueData := data[offset : offset+int(valueLen)]
		offset += int(valueLen)

		if err := parseMetadataValue(valueData); err != nil {
			return err
		}

		headersCount, n := binary.Varint(data[offset:])
		offset += n
		for j := 0; j < int(headersCount); j++ {
			hKeyLen, n := binary.Varint(data[offset:])
			offset += n
			offset += int(hKeyLen)
			hValLen, n := binary.Varint(data[offset:])
			offset += n
			offset += int(hValLen)
		}
	}
	return nil
}

func parseMetadataValue(data []byte) error {
	if len(data) < 3 {
		return nil
	}
	offset := 0
	_ = data[offset]
	offset++
	recordType := data[offset]
	offset++
	_ = data[offset]
	offset++

	switch recordType {
	case 2:
		parseTopicRecord(data[offset:], ClusterState)
	case 3:
		parsePartitionRecord(data[offset:], ClusterState)
	}
	return nil
}

func parseTopicRecord(data []byte, state *ClusterMetadata) {
	offset := 0
	topicName := readCompactString(data, &offset)
	var topicID utils.UUID
	copy(topicID[:], data[offset:offset+16])
	offset += 16

	if _, ok := state.Topics[topicID]; !ok {
		state.Topics[topicID] = &TopicMetadata{
			Name: topicName,
			UUID: topicID,
		}
	}
}

func parsePartitionRecord(data []byte, state *ClusterMetadata) {
	offset := 0
	partitionID := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	var topicID utils.UUID
	copy(topicID[:], data[offset:offset+16])
	offset += 16

	replicas := readCompactInt32Array(data, &offset)
	isr := readCompactInt32Array(data, &offset)
	_ = readCompactInt32Array(data, &offset)
	_ = readCompactInt32Array(data, &offset)

	leader := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4
	leaderEpoch := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4
	partitionEpoch := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	if topic, ok := state.Topics[topicID]; ok {
		topic.Partitions = append(topic.Partitions, PartitionMetadata{
			ID:             partitionID,
			LeaderID:       leader,
			ReplicaNodes:   replicas,
			IsrNodes:       isr,
			LeaderEpoch:    leaderEpoch,
			PartitionEpoch: partitionEpoch,
		})
	}
}

func readCompactString(data []byte, offset *int) string {
	length, n := binary.Uvarint(data[*offset:])
	*offset += n
	if length <= 1 {
		return ""
	}
	strLen := int(length - 1)
	result := string(data[*offset : *offset+strLen])
	*offset += strLen
	return result
}

func readCompactInt32Array(data []byte, offset *int) []uint32 {
	length, n := binary.Uvarint(data[*offset:])
	*offset += n
	if length <= 1 {
		return nil
	}
	arrLen := int(length - 1)
	result := make([]uint32, arrLen)
	for idx := range arrLen {
		result[idx] = binary.BigEndian.Uint32(data[*offset : *offset+4])
		*offset += 4
	}
	return result
}
