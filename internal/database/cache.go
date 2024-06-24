package database

import (
	"fmt"
	"os"
	"sync"
)

const (
	SCRIPTS_PATH = "queries"
)

type queryCache struct {
	queries map[string]string
	mutex   sync.Mutex
}

func newQueryCache() *queryCache {
	return &queryCache{
		queries: make(map[string]string),
	}
}

func (q *queryCache) GetQuery(name string) (string, error) {
	query, ok := q.queries[name]
	if !ok {
		return q.cacheQuery(name, fmt.Sprintf("%s/%s.sql", SCRIPTS_PATH, name))
	}
	return query, nil
}

func (q *queryCache) cacheQuery(name, scriptFile string) (string, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	contents, err := os.ReadFile(scriptFile)
	if err != nil {
		return "", fmt.Errorf("unable to read script file %s: %w", scriptFile, err)
	}
	query := string(contents)
	q.queries[name] = query
	return query, nil
}
