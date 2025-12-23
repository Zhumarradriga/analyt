package domain

import "time"

type Event struct {
	Key       string                 `json:"key" binding:"required"`   // ключ нужон
	Value     map[string]interface{} `json:"value" binding:"required"` // JSON объект с данными
	Timestamp time.Time              `json:"timestamp"`                // время ивента
}