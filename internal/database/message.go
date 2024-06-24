package database

import (
	"fmt"
	"time"

	"github.com/timsexperiments/chat-cli/internal/constants"
	"github.com/timsexperiments/chat-cli/internal/proto/chat"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (db *DB) CreateMessage(body string, sender chat.Message_Sender, conversationId int64) (*chat.Message, error) {
	result, err := db.Exec(constants.CREATE_MESSAGE_QUERY, body, sender.String(), conversationId)
	if err != nil {
		return nil, fmt.Errorf("unable to create message: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("unable to get rows affected: %w", err)
	}
	if rowsAffected != 1 {
		return nil, fmt.Errorf("expected 1 row to be affected, got %d", rowsAffected)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("unable to get last insert ID: %w", err)
	}
	return db.GetMessage(int(id))
}

func (db *DB) GetMessage(id int) (*chat.Message, error) {
	rows, err := db.Query(constants.GET_MESSAGE_QUERY, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get conversation: %w", err)
	}
	defer rows.Close()

	var message *chat.Message
	for rows.Next() {
		if message != nil {
			return nil, fmt.Errorf("expected only one message, got more than one")
		}
		var id, conversationId int64
		var body, senderStr string
		var createdAt time.Time
		if err := rows.Scan(&id, &body, &senderStr, &createdAt, &conversationId); err != nil {
			return nil, fmt.Errorf("unable to build message: %w", err)
		}

		sender, ok := chat.Message_Sender_value[senderStr]
		if !ok {
			return nil, fmt.Errorf("unable to parse sender: %w", err)
		}

		message = &chat.Message{
			Id:        id,
			Body:      body,
			Sender:    chat.Message_Sender(sender),
			CreatedAt: timestamppb.New(createdAt),
		}
	}

	return message, nil
}
