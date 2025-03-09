package model

import (
	"context"
	"fmt"
	"github.com/goccy/go-json"
	"net/http"
	"one-api/common"
	"time"
)

type Record struct {
	Id         int    `json:"id"`
	UserId     int    `json:"user_id" gorm:"index"`
	ModelName  string `json:"model_name" gorm:"index;index:index_username_model_name,priority:1;default:''"`
	Username   string `json:"username" gorm:"index:index_username_model_name,priority:2;default:''"`
	Method     string `json:"method" gorm:"index;type:varchar(8);"`
	RequestId  string `json:"request_id" gorm:"default:''"`
	Path       string `json:"path" gorm:"index;type:varchar(128);"`
	StatusCode int    `json:"status_code" gorm:"index;type:int(8);"`
	RemoteAddr string `json:"remote_addr" gorm:"index;type:varchar(128);"`
	//RequestHeaders string `json:"request_headers" gorm:"default:''"`
	RequestParams string `json:"request_params" gorm:"default:''"`
	RequestBody   string `json:"request_body" gorm:"type:mediumtext;"`
	ResponseBody  string `json:"response_body" gorm:"type:mediumtext;"`
	RequestTime   int64  `json:"request_time" gorm:"bigint;default:0"`
}

func AddRequestRecord(ctx context.Context, request *http.Request, requestId string, requestBody string) {
	milliseconds := time.Now().UnixMilli()
	//headersJson, err := json.Marshal(request.Header)
	//if err != nil {
	//	logger.Errorf(ctx, "Json marshal headers error: %v", err)
	//	return
	//}

	paramsJson, err := json.Marshal(request.URL.Query())
	if err != nil {
		common.LogError(ctx, fmt.Sprintf("Json marshal params error: %v", err))
		return
	}

	record := &Record{
		Method:     request.Method,
		Path:       request.URL.Path,
		RemoteAddr: request.RemoteAddr,
		//RequestHeaders: string(headersJson),
		RequestParams: string(paramsJson),
		RequestBody:   requestBody,
		RequestTime:   milliseconds,
		RequestId:     requestId,
	}
	saveRecord(ctx, record)
}

func AddResponseRecord(ctx context.Context, requestId string, userId int, modelName string, statusCode int, responseBody string) {
	var record Record
	err := LOG_DB.Where("request_id = ?", requestId).Order("request_time desc").Limit(1).Find(&record).Error
	if err != nil || record.Id == 0 {
		common.LogError(ctx, fmt.Sprintf("Query record by request_id-%s error: %v", requestId, err))
		return
	}

	username, err := GetUsernameById(userId, false)
	if err != nil {
		common.LogError(ctx, fmt.Sprintf("Get username by id(%d) error: %v", userId, err))
		return
	}

	record.UserId = userId
	record.Username = username
	record.ModelName = modelName
	record.StatusCode = statusCode
	record.ResponseBody = responseBody
	saveRecord(ctx, &record)
}

func saveRecord(ctx context.Context, record *Record) {
	var err error
	if record.Id == 0 {
		err = LOG_DB.Create(record).Error
	} else {
		err = LOG_DB.Updates(record).Error
	}
	if err != nil {
		common.LogError(ctx, fmt.Sprintf("Failed to save record: "+err.Error()))
		return
	}
	common.LogInfo(ctx, fmt.Sprintf("Save record: %+v", record))
}
