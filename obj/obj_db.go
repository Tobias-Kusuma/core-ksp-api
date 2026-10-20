package obj

import "time"

// 🔽 list struct untuk resembles full returned object dari db
type LOG_HTTP struct {
	LOGID            int64      `json:"-" gorm:"column:logid;type:bigint;primaryKey;autoIncrement"`
	REQUEST_TIME     *time.Time `json:"request_time" gorm:"column:request_time"`
	REQUEST_ENDPOINT *string    `json:"request_endpoint" gorm:"column:request_endpoint"`
	REQUEST_TEXT     *string    `json:"request_text" gorm:"column:request_text"`
	RESPONSE_TIME    *time.Time `json:"response_time" gorm:"column:response_time"`
	RESPONSE_CODE    *string    `json:"response_code" gorm:"column:response_code"`
	RESPONSE_TEXT    *string    `json:"response_text" gorm:"column:response_text"`
	CLIENT_IP        *string    `json:"client_ip" gorm:"column:client_ip"`
}
