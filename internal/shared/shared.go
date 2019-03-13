package shared

import (
	"context"
	"fmt"
	loger "internal/log"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/ftloc/exception"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
)

type Session struct {
	Session *mgo.Session
}
type TaskProgress struct {
	ID     bson.ObjectId `json:"id" bson:"_id"`
	Day    string        `json:"day" bson:"day"`
	Status bool          `json:"status" bson:"status"`
}

type TaskPayload struct {
	Title string   `json:"title"`
	Days  []string `json:"days"`
	Color string   `json:"color"`
}

func NewSessesion(c chan struct{}, ctx context.Context, url string) (*Session, error) {

	loger.Info("Opening te MongoDb Connection")
	session, err := mgo.Dial(url)
	if err != nil {
		loger.Error("Unable to connect to the db check your username and password")

		exception.Throw(fmt.Errorf("%+v", errors.WithStack(err)))
		return nil, ctx.Err()

	}
	c <- struct{}{}
	return &Session{session}, err
}
func (s *Session) Copy() *Session {
	return &Session{s.Session.Copy()}
}

func (s *Session) GetCollection(db string, col string) *mgo.Collection {
	return s.Session.DB(db).C(col)
}

func (s *Session) Close() {
	if s.Session != nil {
		s.Session.Close()
	}
}
func CreateContext(seconds time.Duration) (context.Context, context.CancelFunc, chan struct{}) {
	context, cancel := context.WithTimeout(context.Background(), seconds)

	c := make(chan struct{})
	return context, cancel, c
}
func WatchContextForDBConnection(context context.Context, c chan struct{}) {
	select {
	case <-context.Done():
		if context.Err().Error() != "context canceled" {
			loger.Error("Context Failed to connect to db, Timeout Exceeds, Check your Internet connection or Connection Url")
			//os.Exit(1)
			exception.Throw(fmt.Errorf("Context Failed to connect to db, Timeout Exceeds, Check your Internet connection"))
		}
	case <-c:
		//fmt.Println("Connection Open success! Task Done")
	}
}
