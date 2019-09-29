package mongodbdao

import (
	"dbms/lib/bongo"

	"dbms/lib/mgo/bson"

	mgo "dbms/lib/mgo"
)

/*
     if you want store data to mongo
	     make a embedded MongoBaseID in your struct

	     type YourSata {
	         MongoBaseID
	         your data ....
	     }
*/

type MongoBaseID struct {
	OrderID bson.ObjectId `bson:"_id" json:"order_id"` //
	exists  bool          `bson:"-" json:"-"`
}

func (d *MongoBaseID) GenID() {
	d.OrderID = bson.NewObjectId()
}

// Satisfy the new tracker interface
func (d *MongoBaseID) SetIsNew(isNew bool) {
	d.exists = !isNew
}

func (d *MongoBaseID) IsNew() bool {
	return !d.exists
}

// Satisfy the document interface
func (d *MongoBaseID) GetId() bson.ObjectId {
	return d.OrderID
}

func (d *MongoBaseID) SetId(id bson.ObjectId) {
	d.OrderID = id
}

type DocumentNotFoundError struct {
	bongo.DocumentNotFoundError
}

type MongoDao struct {
	conn       *bongo.Connection
	datasource string
}

func (m *MongoDao) Refresh() {
	m.conn.Session.Refresh()
}

func (m *MongoDao) Init(mongosrc string) error {
	var err error
	config := &bongo.Config{
		ConnectionString: mongosrc,
	}
	m.datasource = mongosrc
	//log.Debugf("will connect %s", mongosrc)
	m.conn, err = bongo.Connect(config)
	if err != nil {
		//log.Debugf("connect err:%#v", *config)
		return err
	}
	//log.Debugf("will connect %s", mongosrc)
	m.SetLogger()
	return nil
}

func (m *MongoDao) Close() {
	m.conn.Session.Close()
}

func (m *MongoDao) Conn() *bongo.Connection {
	return m.conn
}

func (m *MongoDao) SetLogger() {
	mgo.SetLoggerWithLogrus()
	mgo.SetDebug(false)
}
