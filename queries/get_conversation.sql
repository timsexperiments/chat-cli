SELECT
    c.id,
    completion_id,
    c.title,
    c.context,
    c.created_at,
    m.id AS message_id,
    m.body,
    m.sender,
    m.created_at AS message_created_at
FROM conversations c 
    LEFT JOIN messages m ON c.id = m.conversation_id
WHERE c.id = ?
ORDER BY m.created_at DESC;
