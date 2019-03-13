package taskmodel

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type TaskModel struct {
	ID        bson.ObjectId           `json:"id" bson:"_id,omitempty"`
	Title     string                  `json:"title" from:"title" binding:"required" bson:"title"`
	Frequency []TaskProgressFrequency `json:"tpf" binding:"required" bson:"tpf"`
	Alias     string                  `json:"alias" bson:"alias"`
	Color     string                  `json:"color" bson:"color"`
}
type TaskProgressFrequency struct {
	Day    string `json:"day" bson:"day"`
	Status bool   `json:"status" bson:"status"`
}

//Export of the TaskIndex
func TaskModelIndex() mgo.Index {
	return mgo.Index{
		Key:        []string{"alias"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
}
