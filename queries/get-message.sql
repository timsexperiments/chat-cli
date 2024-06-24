SELECT
    id,
    body,
    sender,
    created_at,
    conversation_id
FROM messages
WHERE id = ?;
