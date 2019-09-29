package gocelery

import (
	"dbms/lib/mongodbdao"

	"dbms/lib/mgo/bson"
)

// MgoBackend is CeleryBackend for Redis
type MgoBackend struct {
	md    *mongodbdao.MongoDao
	table string
}

// NewMgoBackend creates new MgoBackend
func NewMgoBackend(host string, collection string) (*MgoBackend, error) {
	mb := &MgoBackend{
		md:    &mongodbdao.MongoDao{},
		table: collection,
	}
	if err := mb.md.Init(host); err != nil {
		return nil, err
	}
	return mb, nil
}

func (mg *MgoBackend) Init(host string) error {
	return mg.md.Init(host)
}

func (mg *MgoBackend) Close() {
	mg.md.Close()
}

// GetResult calls API to get asynchronous result
// Should be called by AsyncResult

func (cb *MgoBackend) GetResult(taskID string) (*ResultMessage, error) {

	var msg ResultMessage
	err := cb.md.Conn().Collection(cb.table).FindOne(bson.M{"_id": taskID}, &msg)
	if err != nil {
		var merr mongodbdao.DocumentNotFoundError
		if err.Error() == merr.Error() {
			return nil, ERR_NilObj
		} else {
			return nil, err
		}
	}
	//log.Debugf("mgo ret:%#v", msg)
	return &msg, nil
}

// SetResult pushes result back into backend
func (cb *MgoBackend) SetResult(taskID string, result *ResultMessage) error {
	result.ID = taskID
	if err := cb.md.Conn().Collection(cb.table).Save(result); err != nil {
		return err
	}
	return nil
}
