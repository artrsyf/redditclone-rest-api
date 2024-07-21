package redis

import (
	"encoding/json"
	"redditclone/pkg/models"
	"strconv"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

type SessionRedisManager struct {
	redisConn redis.Conn
	mu        *sync.Mutex
}

func NewSessionRedisManager(conn redis.Conn) *SessionRedisManager {
	return &SessionRedisManager{
		redisConn: conn,
		mu:        &sync.Mutex{},
	}
}

func (sm *SessionRedisManager) Create(JWTToken string, userID int) error {
	mkey := "sessions:" + strconv.Itoa(userID)
	newSession := models.Session{
		ID:        1,
		JWT:       JWTToken,
		UserID:    userID,
		ExpiresAt: time.Now().AddDate(0, 0, 4),
	}

	dataSerialized, err := json.Marshal(newSession)
	if err != nil {
		return err
	}

	sm.mu.Lock()
	result, err := redis.String(sm.redisConn.Do("SET", mkey, dataSerialized, "EX", 4*24*60*60))
	sm.mu.Unlock()
	if err != nil || result != "OK" {
		return err
	}

	return nil
}

func (sm *SessionRedisManager) Check(userID int) (*models.Session, error) {
	mkey := "sessions:" + strconv.Itoa(userID)
	sm.mu.Lock()
	data, err := redis.Bytes(sm.redisConn.Do("GET", mkey))
	sm.mu.Unlock()
	if err != nil {
		if err == redis.ErrNil {
			return nil, models.ErrNoSession
		}

		return nil, err
	}
	session := &models.Session{}
	err = json.Unmarshal(data, session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (sm *SessionRedisManager) Delete(userID int) error {
	mkey := "sessions:" + strconv.Itoa(userID)
	sm.mu.Lock()
	_, err := redis.Int(sm.redisConn.Do("DEL", mkey))
	sm.mu.Unlock()
	if err != nil {
		if err == redis.ErrNil {
			return models.ErrNoSession
		}

		return err
	}

	return nil
}
