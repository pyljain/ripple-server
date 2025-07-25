package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AgentVersionMetrics struct {
	Id             primitive.ObjectID `json:"id" bson:"_id"`
	Name           string             `json:"name" bson:"name"`
	Project        string             `json:"project" bson:"project"`
	Status         string             `json:"status" bson:"status"`
	LastSeen       time.Time          `json:"lastSeen" bson:"lastSeen"`
	Version        string             `json:"version" bson:"version"`
	AverageRunTime float64            `json:"avgRuntime" bson:"avgRuntime"`
	SuccessRate    float64            `json:"successRate" bson:"successRate"`
	TotalRuns      int64              `json:"totalRuns" bson:"totalRuns"`
	Spend          float64            `json:"spend" bson:"spend"`
	Tools          []string           `json:"tools" bson:"tools"`
	Models         []string           `json:"models" bson:"models"`
	Cluster        string             `json:"cluster" bson:"cluster"`
}
