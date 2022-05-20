package key

import (
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"gopkg.in/rethinkdb/rethinkdb-go.v6"
)

func (k *Key) getSession() (*rethinkdb.Session, error) {
	session, err := rethinkdb.Connect(rethinkdb.ConnectOpts{
		Address: k.Config.Rethink.Address,
	})

	if err != nil {
		return nil, err
	}

	return session, nil
}

func (k *Key) findUser(id string) (*UserKey, error) {
	session, err := k.getSession()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := session.Close(); err != nil {
			bugLog.Debug(err)
		}
	}()

	cursor, err := rethinkdb.DB("retro-board").Table("user").Get(id).Run(session)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := cursor.Close(); err != nil {
			bugLog.Debug(err)
		}
	}()

	userKey := UserKey{}
	if err := cursor.All(&userKey); err != nil {
		return nil, err
	}

	return &userKey, nil
}

func (k *Key) insertUser(key UserKey) error {
	session, err := k.getSession()
	if err != nil {
		return err
	}

	defer func() {
		if err := session.Close(); err != nil {
			bugLog.Debug(err)
		}
	}()

	_, err = rethinkdb.DB("retro-board").Table("user").Insert(key).RunWrite(session)
	if err != nil {
		return err
	}

	return nil
}
