package database

import (
	"fmt"
	"time"

	"github.com/timsexperiments/chat-cli/internal/constants"
	"github.com/timsexperiments/chat-cli/internal/proto/chat"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (db *DB) CreateConversation(title string) (*chat.Conversation, error) {
	result, err := db.Exec(constants.CREATE_CONVERSATION_QUERY, title)
	if err != nil {
		return nil, fmt.Errorf("unable to create conversation: %w", err)
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
	return db.GetConversation(int(id))
}

func (db *DB) GetConversation(id int) (*chat.Conversation, error) {
	rows, err := db.Query(constants.GET_CONVERSATION_QUERY, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get conversation: %w", err)
	}

	defer rows.Close()
	var conversation *chat.Conversation
	var conversationId *int64
	messages := make([]*chat.Message, 0)
	for rows.Next() {
		var id int64
		var title string
		var context *string
		var createdAt time.Time
		var messageId *int64
		var messageBody *string
		var messageCreatedAt *time.Time
		var messageSender *string
		if err := rows.Scan(
			&id,
			&title,
			&context,
			&createdAt,
			&messageId,
			&messageBody,
			&messageSender,
			&messageCreatedAt,
		); err != nil {
			return nil, fmt.Errorf("unable to build conversation: %w", err)
		}
		if conversationId == nil {
			conversationId = &id
			conversation = &chat.Conversation{
				Id:        id,
				Title:     title,
				CreatedAt: timestamppb.New(createdAt),
				Messages:  nil,
			}
			if context != nil {
				conversation.Context = *context
			}
		}
		if id != *conversationId {
			return nil, fmt.Errorf("expected conversation with id [%d], got conversation with id [%d]", *conversationId, id)
		}
		if messageId == nil || messageBody == nil || messageSender == nil || messageCreatedAt == nil {
			continue
		}
		senderValue, ok := chat.Message_Sender_value[*messageSender]
		if !ok {
			return nil, fmt.Errorf("unable to parse sender one conversation message %d: %w", messageId, err)
		}
		messages = append(messages, &chat.Message{
			Id:        *messageId,
			Body:      *messageBody,
			Sender:    chat.Message_Sender(senderValue),
			CreatedAt: timestamppb.New(*messageCreatedAt),
		})
	}

	if conversation == nil {
		return nil, fmt.Errorf("conversation not found")
	}

	if len(messages) > 0 {
		conversation.Messages = messages
	}

	return conversation, nil
}

func (db *DB) ListConversations() ([]*chat.Conversation, error) {
	rows, err := db.Query(constants.LIST_CONVERSATIONS_QUERY)
	if err != nil {
		return nil, fmt.Errorf("unable to list conversations: %w", err)
	}

	defer rows.Close()
	conversations := []*chat.Conversation{}
	for rows.Next() {
		var id *int64
		var title, context *string
		var createdAt *time.Time
		if err := rows.Scan(&id, &title, &context, &createdAt); err != nil {
			return nil, fmt.Errorf("unable to build conversation: %w", err)
		}
		if id == nil || title == nil || createdAt == nil {
			return nil, fmt.Errorf("missing required fields. id = %v, title = %v, context = %v, createdAt = %v: %w", id, title, context, createdAt, err)
		}
		conversation := &chat.Conversation{
			Id:        *id,
			Title:     *title,
			CreatedAt: timestamppb.New(*createdAt),
			Messages:  nil,
		}
		if context != nil {
			conversation.Context = *context
		}
		conversations = append(conversations, conversation)
	}

	return conversations, nil
}
