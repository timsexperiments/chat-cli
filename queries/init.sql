CREATE TABLE IF NOT EXISTS conversations (
  id INTEGER PRIMARY KEY ASC,
  title TEXT NOT NULL UNIQUE,
  context TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS messages (
  id INTEGER PRIMARY KEY ASC,
  completion_id TEXT UNIQUE,
  body TEXT NOT NULL,
  sender VARCHAR(4) NOT NULL CHECK (sender IN ('USER', 'BOT')),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  conversation_id INT NOT NULL,
  CONSTRAINT fk__messages__convesations__id FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);
