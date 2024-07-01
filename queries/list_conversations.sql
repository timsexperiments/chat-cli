SELECT
    id,
    completion_id,
    title,
    context,
    created_at
FROM conversations
ORDER BY created_at ASC;
