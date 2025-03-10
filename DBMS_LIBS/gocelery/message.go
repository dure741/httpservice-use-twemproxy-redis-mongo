package gocelery

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"reflect"
	"sync"
	"time"

	"dbms/lib/mgo/bson"

	"github.com/satori/go.uuid"
)

func NewV4() uuid.UUID {
	uuuid, _ := uuid.NewV4()
	return uuuid
}

// CeleryMessage is actual message to be sent to Redis
type CeleryMessage struct {
	Body            string                 `json:"body"`
	Headers         map[string]interface{} `json:"headers"`
	ContentType     string                 `json:"content-type"`
	Properties      CeleryProperties       `json:"properties"`
	ContentEncoding string                 `json:"content-encoding"`
}

func (cm *CeleryMessage) reset() {
	cm.Headers = nil
	cm.Body = ""
	cm.Properties.CorrelationID = NewV4().String()
	cm.Properties.ReplyTo = NewV4().String()
	cm.Properties.DeliveryTag = NewV4().String()
}

var celeryMessagePool = sync.Pool{
	New: func() interface{} {
		return &CeleryMessage{
			Body:        "",
			Headers:     nil,
			ContentType: "application/json",
			Properties: CeleryProperties{
				BodyEncoding:  "base64",
				CorrelationID: NewV4().String(),
				ReplyTo:       NewV4().String(),
				DeliveryInfo: CeleryDeliveryInfo{
					Priority:   0,
					RoutingKey: "celery",
					Exchange:   "celery",
				},
				DeliveryMode: 2,
				DeliveryTag:  NewV4().String(),
			},
			ContentEncoding: "utf-8",
		}
	},
}

func getCeleryMessage(encodedTaskMessage string) *CeleryMessage {
	msg := celeryMessagePool.Get().(*CeleryMessage)
	msg.Body = encodedTaskMessage
	return msg
}

func releaseCeleryMessage(v *CeleryMessage) {
	v.reset()
	celeryMessagePool.Put(v)
}

// CeleryProperties represents properties json
type CeleryProperties struct {
	BodyEncoding  string             `json:"body_encoding"`
	CorrelationID string             `json:"correlation_id"`
	ReplyTo       string             `json:"replay_to"`
	DeliveryInfo  CeleryDeliveryInfo `json:"delivery_info"`
	DeliveryMode  int                `json:"delivery_mode"`
	DeliveryTag   string             `json:"delivery_tag"`
}

// CeleryDeliveryInfo represents deliveryinfo json
type CeleryDeliveryInfo struct {
	Priority   int    `json:"priority"`
	RoutingKey string `json:"routing_key"`
	Exchange   string `json:"exchange"`
}

// GetTaskMessage retrieve and decode task messages from broker
func (cm *CeleryMessage) GetTaskMessage() *TaskMessage {
	// ensure content-type is 'application/json'
	if cm.ContentType != "application/json" {
		log.Println("unsupported content type " + cm.ContentType)
		return nil
	}
	// ensure body encoding is base64
	if cm.Properties.BodyEncoding != "base64" {
		log.Println("unsupported body encoding " + cm.Properties.BodyEncoding)
		return nil
	}
	// ensure content encoding is utf-8
	if cm.ContentEncoding != "utf-8" {
		log.Println("unsupported encoding " + cm.ContentEncoding)
		return nil
	}
	// decode body
	taskMessage, err := DecodeTaskMessage(cm.Body)
	if err != nil {
		log.Println("failed to decode task message")
		return nil
	}
	return taskMessage
}

// TaskMessage is celery-compatible message
type TaskMessage struct {
	ID      string                 `json:"id"`
	Task    string                 `json:"task"`
	Args    []interface{}          `json:"args"`
	Kwargs  map[string]interface{} `json:"kwargs"`
	Retries int                    `json:"retries"`
	ETA     string                 `json:"eta"`
}

func (tm *TaskMessage) reset() {
	tm.ID = NewV4().String()
	tm.Task = ""
	tm.Args = nil
	tm.Kwargs = nil
}

var taskMessagePool = sync.Pool{
	New: func() interface{} {
		return &TaskMessage{
			ID:      NewV4().String(),
			Retries: 0,
			Kwargs:  nil,
			ETA:     time.Now().Format(time.RFC3339),
		}
	},
}

func getTaskMessage(task string) *TaskMessage {
	msg := taskMessagePool.Get().(*TaskMessage)
	msg.Task = task
	msg.Args = make([]interface{}, 0)
	msg.Kwargs = make(map[string]interface{})
	msg.ETA = time.Now().Format(time.RFC3339)
	return msg
}

func releaseTaskMessage(v *TaskMessage) {
	v.reset()
	taskMessagePool.Put(v)
}

// DecodeTaskMessage decodes base64 encrypted body and return TaskMessage object
func DecodeTaskMessage(encodedBody string) (*TaskMessage, error) {
	body, err := base64.StdEncoding.DecodeString(encodedBody)
	if err != nil {
		return nil, err
	}
	message := taskMessagePool.Get().(*TaskMessage)
	err = json.Unmarshal(body, message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

// Encode returns base64 json encoded string
func (tm *TaskMessage) Encode() (string, error) {
	jsonData, err := json.Marshal(tm)
	if err != nil {
		return "", err
	}
	encodedData := base64.StdEncoding.EncodeToString(jsonData)
	return encodedData, err
}

// ResultMessage is return message received from broker
type ResultMessage struct {
	ID        string        `json:"task_id" bson:"_id"`
	Status    string        `json:"status" bson:"status"`
	Traceback interface{}   `json:"traceback" bson:"traceback"`
	Result    interface{}   `json:"result" bson:"result"`
	Children  []interface{} `json:"children" bson:"children"`
	DateDone  time.Time     `json:"date_done" bson:"date_done"`
	exists    bool          `json:"-" bson:"-"`
}

func (d *ResultMessage) SetIsNew(isNew bool) {
	d.exists = !isNew
}

func (d *ResultMessage) IsNew() bool {
	return !d.exists
}

// Satisfy the document interface
func (d *ResultMessage) GetId() bson.ObjectId {
	return bson.NewObjectId()
}

func (d *ResultMessage) SetId(id bson.ObjectId) {

}

func (rm *ResultMessage) reset() {
	rm.Result = nil
}

var resultMessagePool = sync.Pool{
	New: func() interface{} {
		return &ResultMessage{
			Status:    "SUCCESS",
			Traceback: nil,
			Children:  nil,
		}
	},
}

func getResultMessage(val interface{}) *ResultMessage {
	msg := resultMessagePool.Get().(*ResultMessage)
	msg.Result = val
	return msg
}

func getReflectionResultMessage(val *reflect.Value) *ResultMessage {
	msg := resultMessagePool.Get().(*ResultMessage)
	msg.Result = GetRealValue(val)
	return msg
}

func releaseResultMessage(v *ResultMessage) {
	v.reset()
	resultMessagePool.Put(v)
}
