package handler

import (
	"encoding/binary"
	"fmt"

	customerr "github.com/gafkonian-go/internal/custom_error"
	"github.com/gafkonian-go/internal/metadata"
	"github.com/gafkonian-go/internal/utils"
)

type baseResponse struct {
	CorrelationID uint32
	ErrorCode     uint16
}

// Error Response hadnling
type errorResponse struct {
	baseResponse
}

type responseEncoder interface {
	encode() []byte
}

func (e *errorResponse) encode() []byte {
	buf := make([]byte, 10)
	binary.BigEndian.PutUint32(buf[0:4], uint32(len(buf)-4))
	binary.BigEndian.PutUint32(buf[4:8], e.CorrelationID)
	binary.BigEndian.PutUint16(buf[8:10], e.ErrorCode)
	return buf
}

func getErrorResponse(errorCode uint16, correlationID uint32) responseEncoder {
	response := &errorResponse{
		baseResponse: baseResponse{
			CorrelationID: correlationID,
			ErrorCode:     errorCode,
		},
	}
	return response
}

// API version response handling
type apiVersionResponse struct {
	baseResponse
	APIKeysArrayLength uint8
	APIKeys            []apiVersionKey
	ThrottleTimeMs     uint32
	TagBuffer          uint8
}

func (a *apiVersionResponse) encode() []byte {
	body := make([]byte, 0)

	buf16 := make([]byte, 2)
	buf32 := make([]byte, 4)
	utils.AppendBuf32(&buf32, a.CorrelationID, &body)
	utils.AppendBuf16(&buf16, 0, &body)

	body = append(body, uint8(len(a.APIKeys)+1))
	for _, apiKey := range a.APIKeys {
		utils.AppendBuf16(&buf16, apiKey.APIKey, &body)
		binary.BigEndian.PutUint16(buf16, apiKey.MinVersion)
		utils.AppendBuf16(&buf16, apiKey.MinVersion, &body)
		utils.AppendBuf16(&buf16, apiKey.MaxVersion, &body)
		body = append(body, apiKey.TagBuffer)
	}
	utils.AppendBuf32(&buf32, a.ThrottleTimeMs, &body)
	body = append(body, a.TagBuffer)
	size := uint32(len(body))
	encodedResponse := make([]byte, 4)
	binary.BigEndian.PutUint32(encodedResponse, size)
	encodedResponse = append(encodedResponse, body...)
	return encodedResponse
}

func getAPIVersionResponse(correlationID uint32) responseEncoder {
	response := &apiVersionResponse{
		baseResponse: baseResponse{
			CorrelationID: correlationID,
			ErrorCode:     0,
		},
		APIKeysArrayLength: uint8(len(avaliableAPI)),
		APIKeys:            avaliableAPI,
		ThrottleTimeMs:     0,
		TagBuffer:          0,
	}
	return response
}

// Topic Version Response handling
type topicPartitionsResponse struct {
	CorrelationID     uint32
	TagBufferHeader   uint8
	ThrottleTimeMs    uint32
	Topics            []Topic
	NextCursor        int8
	TagBufferResponse uint8
}

type Topic struct {
	ErrorCode                 uint16
	TopicName                 string
	TopicID                   utils.UUID
	IsInternal                bool
	Partitions                []Partition
	TopicAuthorizedOperations uint32
	TagBuffer                 uint8
}

type Partition struct {
	ErrorCode uint16
	metadata.PartitionMetadata
	EligibleLeaderReplicas []uint32
	LastKnownElr           []uint32
	OfflineReplicas        []uint32
}

func (t *topicPartitionsResponse) encode() []byte {
	body := make([]byte, 0)
	buf16 := make([]byte, 2)
	buf32 := make([]byte, 4)
	utils.AppendBuf32(&buf32, t.CorrelationID, &body)
	utils.AppendBuf16(&buf16, 0, &body)
	utils.AppendBuf32(&buf32, t.ThrottleTimeMs, &body)
	body = append(body, uint8(len(t.Topics)+1))
	for _, topic := range t.Topics {
		utils.AppendBuf16(&buf16, topic.ErrorCode, &body)
		topicNameSize := uint8(len(topic.TopicName) + 1)
		body = append(body, topicNameSize)
		body = append(body, topic.TopicName...)
		body = append(body, topic.TopicID[:]...)
		isInternal := 0
		if topic.IsInternal {
			isInternal = 1
		}
		body = append(body, byte(isInternal))
		body = append(body, uint8(len(topic.Partitions)+1))
		for _, partition := range topic.Partitions {
			utils.AppendBuf16(&buf16, partition.ErrorCode, &body)
			utils.AppendBuf32(&buf32, partition.ID, &body)
			utils.AppendBuf32(&buf32, partition.LeaderID, &body)
			utils.AppendBuf32(&buf32, partition.LeaderEpoch, &body)
			utils.AppendBuf32(&buf32, partition.PartitionEpoch, &body)
			body = append(body, uint8(len(partition.ReplicaNodes)+1))
			for _, node := range partition.ReplicaNodes {
				utils.AppendBuf32(&buf32, node, &body)
			}
			body = append(body, uint8(len(partition.IsrNodes)+1))
			for _, node := range partition.IsrNodes {
				utils.AppendBuf32(&buf32, node, &body)
			}

			body = append(body, uint8(len(partition.EligibleLeaderReplicas)+1))
			for _, node := range partition.EligibleLeaderReplicas {
				utils.AppendBuf32(&buf32, node, &body)
			}

			body = append(body, uint8(len(partition.LastKnownElr)+1))
			for _, node := range partition.LastKnownElr {
				utils.AppendBuf32(&buf32, node, &body)
			}

			body = append(body, uint8(len(partition.OfflineReplicas)+1))
			for _, node := range partition.OfflineReplicas {
				utils.AppendBuf32(&buf32, node, &body)
			}
			body = append(body, 0)
		}
		utils.AppendBuf32(&buf32, topic.TopicAuthorizedOperations, &body)
		body = append(body, 0)
	}
	body = append(body, byte(t.NextCursor))
	body = append(body, 0)
	size := uint32(len(body))
	encodedResponse := make([]byte, 4)
	binary.BigEndian.PutUint32(encodedResponse, size)
	encodedResponse = append(encodedResponse, body...)
	return encodedResponse
}

func topicNotFoundResponse(topicName string) Topic {
	return Topic{
		ErrorCode:                 3,
		TopicName:                 topicName,
		TopicID:                   utils.UUID{},
		IsInternal:                false,
		Partitions:                []Partition{},
		TopicAuthorizedOperations: 0,
		TagBuffer:                 0,
	}
}

func getTopicPartitionsResponse(correlationID uint32, body *DescribeTopicRequestBody) responseEncoder {
	topics := []Topic{}
	for _, topic := range body.Topics {
		if topicUUID, ok := metadata.ClusterState.TopicNameIndex[topic.TopicName]; ok {
			topicData := metadata.ClusterState.Topics[topicUUID]
			partitions := []Partition{}
			for _, partitionMetadata := range topicData.Partitions {
				partitions = append(partitions, Partition{
					ErrorCode:              0,
					PartitionMetadata:      partitionMetadata,
					EligibleLeaderReplicas: []uint32{},
					LastKnownElr:           []uint32{},
					OfflineReplicas:        []uint32{},
				})
			}
			topics = append(topics, Topic{
				ErrorCode:                 0,
				TopicName:                 topicData.Name,
				TopicID:                   topicData.UUID,
				IsInternal:                false,
				Partitions:                partitions,
				TopicAuthorizedOperations: 0,
				TagBuffer:                 0,
			})
		} else {
			topics = append(topics, topicNotFoundResponse(topic.TopicName))
		}
	}
	response := &topicPartitionsResponse{
		CorrelationID:     correlationID,
		TagBufferHeader:   0,
		ThrottleTimeMs:    0,
		Topics:            topics,
		NextCursor:        -1,
		TagBufferResponse: 0,
	}
	return response
}

type ProduceResponse struct {
	CorrelationID   uint32
	TagBufferHeader uint8
	ProduceBody
}

func (p *ProduceResponse) encode() []byte {
	body := make([]byte, 0)
	buf32 := make([]byte, 4)
	utils.AppendBuf32(&buf32, p.CorrelationID, &body)
	utils.AppendBuf32(&buf32, uint32(len(p.TopicName)), &body)
	body = append(body, p.TopicName...)
	utils.AppendBuf32(&buf32, p.PartitionID, &body)
	utils.AppendBuf32(&buf32, uint32(len(p.Records)), &body)
	for _, r := range p.Records {
		record := r.Encode()
		body = append(body, record...)
	}
	size := uint32(len(body))
	fullResponse := make([]byte, 4)
	binary.BigEndian.PutUint32(fullResponse, size)
	fullResponse = append(fullResponse, body...)
	return fullResponse
}

func getProduceResponse(correlationID uint32, body *ProduceBody, path string) (responseEncoder, error) {
	topicUUID, ok := metadata.ClusterState.TopicNameIndex[body.TopicName]
	if !ok {
		return nil, customerr.RaiseError(customerr.TopicNotFoundError)
	}
	topic := metadata.ClusterState.Topics[topicUUID]
	var recordPartition *metadata.PartitionMetadata
	for _, partition := range topic.Partitions {
		if body.PartitionID == partition.ID {
			recordPartition = &partition
			break
		}
	}
	if recordPartition == nil {
		return nil, customerr.RaiseError(customerr.PartitionNotFoundError)
	}
	var records []metadata.Record
	for _, recordRequest := range body.Records {
		newRecord, err := recordRequest.Append(path, topicUUID.String(), body.TopicName, recordPartition.ID)
		if err != nil {
			return nil, customerr.RaiseError(customerr.SystemError, err.Error())
		}
		records = append(records, *newRecord)
	}
	response := &ProduceResponse{
		CorrelationID:   correlationID,
		TagBufferHeader: 0,
		ProduceBody: ProduceBody{
			TopicName:   body.TopicName,
			PartitionID: body.PartitionID,
			Records:     records,
		},
	}
	fmt.Println(response)
	return response, nil
}
