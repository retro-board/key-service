package key

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mrz1836/go-sanitize"

	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/retro-board/key-service/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	Config *config.Config
	CTX    context.Context
}

func NewMongo(c *config.Config) *Mongo {
	return &Mongo{
		Config: c,
		CTX:    context.Background(),
	}
}

type DataSet struct {
	UserID    string `json:"user_id" bson:"user_id"`
	Generated int64  `json:"generated" bson:"generated"`
	Keys      struct {
		UserService        string `json:"user_service" bson:"user_service"`
		RetroService       string `json:"retro_service" bson:"retro_service"`
		TimerService       string `json:"timer_service" bson:"timer_service"`
		CompanyService     string `json:"company_service" bson:"company_service"`
		BillingService     string `json:"billing_service" bson:"billing_service"`
		PermissionsService string `json:"permissions_service" bson:"permissions_service"`
	} `json:"keys" bson:"keys"`
}

func (m *Mongo) getConnection() (*mongo.Client, error) {
	client, err := mongo.Connect(
		m.CTX,
		options.Client().ApplyURI(fmt.Sprintf(
			"mongodb+srv://%s:%s@%s",
			m.Config.Mongo.Username,
			m.Config.Mongo.Password,
			m.Config.Mongo.Host)),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (m *Mongo) Get(key string) (*DataSet, error) {
	client, err := m.getConnection()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := client.Disconnect(m.CTX); err != nil {
			bugLog.Info(err)
		}
	}()

	var dataSet DataSet
	err = client.Database("keys").Collection("keys").FindOne(m.CTX, map[string]string{"user_id": sanitize.AlphaNumeric(key, false)}).Decode(&dataSet)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	dataTime := time.Unix(dataSet.Generated, 0).Unix()
	minusTime := time.Now().Add(-time.Hour * 2).Unix()
	plusTime := time.Now().Add(time.Hour * 2).Unix()
	if dataTime >= minusTime && dataTime <= plusTime {
		return &dataSet, nil
	}

	return nil, nil
}

func (m *Mongo) Create(data DataSet) error {
	client, err := m.getConnection()
	if err != nil {
		return err
	}
	defer func() {
		if err := client.Disconnect(m.CTX); err != nil {
			bugLog.Info(err)
		}
	}()

	_, err = client.Database("keys").Collection("keys").UpdateOne(
		m.CTX,
		map[string]string{"user_id": sanitize.AlphaNumeric(data.UserID, false)},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "generated", Value: time.Now().Unix()},
			{Key: "keys.user_service", Value: data.Keys.UserService},
			{Key: "keys.retro_service", Value: data.Keys.RetroService},
			{Key: "keys.timer_service", Value: data.Keys.TimerService},
			{Key: "keys.company_service", Value: data.Keys.CompanyService},
			{Key: "keys.billing_service", Value: data.Keys.BillingService},
			{Key: "keys.permissions_service", Value: data.Keys.PermissionsService},
		}}},
		options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	return nil
}
