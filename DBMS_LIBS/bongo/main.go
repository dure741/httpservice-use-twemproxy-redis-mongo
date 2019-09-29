package bongo

import (
	"errors"
	"fmt"

	//log "gitlab.10101111.com/oped/DBMS_LIBS/logrus"
	"gitlab.10101111.com/oped/DBMS_LIBS/mgo"
)

type Config struct {
	ConnectionString string
	Database         string
	DialInfo         *mgo.DialInfo
}

// var EncryptionKey [32]byte
// var EnableEncryption bool

type Connection struct {
	Config  *Config
	Session *mgo.Session
	// collection []Collection
	Context *Context
}

// Create a new connection and run Connect()
func Connect(config *Config) (*Connection, error) {
	conn := &Connection{
		Config:  config,
		Context: &Context{},
	}

	err := conn.Connect()

	return conn, err
}

// Connect to the database using the provided config
func (m *Connection) Connect() (err error) {
	defer func() {
		if r := recover(); r != nil {
			// panic(r)
			// return
			if e, ok := r.(error); ok {
				err = e
			} else if e, ok := r.(string); ok {
				err = errors.New(e)
			} else {
				err = errors.New(fmt.Sprint(r))
			}

		}
	}()

	if m.Config.DialInfo == nil {
		if m.Config.DialInfo, err = mgo.ParseURL(m.Config.ConnectionString); err != nil {
			//panic(fmt.Sprintf("cannot parse given URI %s due to error: %s", m.Config.ConnectionString, err.Error()))
			return fmt.Errorf("cannot parse given URI %s due to error: %s", m.Config.ConnectionString, err.Error())
		}
	}

	//log.Debugf("mgo.DialWithInfo...")
	session, err := mgo.DialWithInfo(m.Config.DialInfo)
	if err != nil {
		//log.Debugf("mgo.DialWithInfo err:%v", err)
		return err
	}
	//log.Debugf("mgo.DialWithInfo ok")

	m.Session = session

	//m.Session.SetMode(mgo.Monotonic, true) // read from slave, write from master
	m.Session.SetMode(mgo.Strong, true) // read or write  both from master

	return nil
}

// CollectionFromDatabase ...
func (m *Connection) CollectionFromDatabase(name string, database string) *Collection {
	// Just create a new instance - it's cheap and only has name and a database name
	return &Collection{
		Connection: m,
		Context:    m.Context,
		Database:   database,
		Name:       name,
	}
}

// Collection ...
func (m *Connection) Collection(name string) *Collection {
	return m.CollectionFromDatabase(name, m.Config.Database)
}

func (m *Connection) Refresh() {
	m.Session.Refresh()
}
