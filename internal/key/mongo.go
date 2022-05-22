package key

import (
	"context"
	"fmt"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/retro-board/key-service/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
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
	UserID  string              `json:"user_id" bson:"user_id"`
	Created primitive.Timestamp `json:"created" bson:"created"`
	Keys    struct {
		UserService    string `json:"user_service" bson:"user_service"`
		RetroService   string `json:"retro_service" bson:"retro_service"`
		TimerService   string `json:"timer_service" bson:"timer_service"`
		CompanyService string `json:"company_service" bson:"company_service"`
		BillingService string `json:"billing_service" bson:"billing_service"`
	} `json:"keys" bson:"keys"`
}

func (m *Mongo) Get(key string) (*DataSet, error) {
	client, err := mongo.Connect(
		m.CTX,
		options.Client().ApplyURI(fmt.Sprintf(
			"mongodb+srv://%s:%s@%s/?retryWrites=true&w=majority",
			m.Config.Mongo.Username,
			m.Config.Mongo.Password,
			m.Config.Mongo.Host)),
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := client.Disconnect(m.CTX); err != nil {
			bugLog.Info(err)
		}
	}()

	var dataSet DataSet
	err = client.Database("keys").Collection("keys").FindOne(m.CTX, map[string]string{"user_id": key}).Decode(&dataSet)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	dataTime := time.Unix(int64(dataSet.Created.T), int64(dataSet.Created.I)).Unix()
	minusTime := time.Now().Add(-time.Hour * 2).Unix()
	plusTime := time.Now().Add(time.Hour * 2).Unix()

	//fmt.Printf("dataTime: %d\nMinusTime: %d\nPlusTime: %d\n", dataTime, minusTime, plusTime)
	//fmt.Printf("dataTime-MinusTime: %d\nPlusTime-dataTime: %d\n", dataTime-minusTime, plusTime-dataTime)

	if minusTime-dataTime >= 0 && dataTime-plusTime <= 0 {
		return &dataSet, nil
	}

	return nil, nil
}

func (m *Mongo) Create(data DataSet) error {
	client, err := mongo.Connect(
		m.CTX,
		options.Client().ApplyURI(fmt.Sprintf(
			"mongodb+srv://%s:%s@%s/?retryWrites=true&w=majority",
			m.Config.Mongo.Username,
			m.Config.Mongo.Password,
			m.Config.Mongo.Host)),
	)
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
		map[string]string{"user_id": data.UserID},
		bson.D{{"$set", bson.D{
			{"keys.user_service", data.Keys.UserService},
			{"keys.retro_service", data.Keys.RetroService},
			{"keys.timer_service", data.Keys.TimerService},
			{"keys.company_service", data.Keys.CompanyService},
			{"keys.billing_service", data.Keys.BillingService},
		}}},
		options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	return nil
}
