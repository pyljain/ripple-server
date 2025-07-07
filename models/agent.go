package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Agent represents an agent in the system
type Agent struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name"`
	Project   string             `json:"project" bson:"project"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// AgentVersion represents a specific version of an agent
type AgentVersion struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AgentID    primitive.ObjectID `json:"agent_id" bson:"agent_id"`
	Version    string             `json:"version" bson:"version"`
	Cluster    string             `json:"cluster" bson:"cluster"`
	Tools      []string           `json:"tools" bson:"tools"`
	Models     []string           `json:"models" bson:"models"`
	Deployment string             `json:"deployment" bson:"deployment"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
}

// AgentRun represents a single run of an agent version
type AgentRun struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AgentID    primitive.ObjectID `json:"agent_id" bson:"agent_id"`
	VersionID  primitive.ObjectID `json:"version_id" bson:"version_id"`
	Version    string             `json:"version" bson:"version"`
	Created    time.Time          `json:"created" bson:"created"`
	Status     string             `json:"status" bson:"status"`
	TimeTaken  string             `json:"time_taken" bson:"time_taken"`
	Initiator  string             `json:"initiator" bson:"initiator"`
	Tools      []string           `json:"tools" bson:"tools"`
	Cost       float64            `json:"cost" bson:"cost"`
	Models     []string           `json:"models" bson:"models"`
	RunID      int64              `json:"id" bson:"run_id"`
	TaskID     int64              `json:"task_id" bson:"task_id"`
	RecordedAt time.Time          `json:"recorded_at" bson:"recorded_at"`
}

// Request and Response types

// RegisterAgentRequest represents the request to register a new agent
type RegisterAgentRequest struct {
	Name    string `json:"name"`
	Project string `json:"project"`
}

// RegisterAgentVersionRequest represents the request to register a new agent version
type RegisterAgentVersionRequest struct {
	Version    string   `json:"version"`
	Cluster    string   `json:"cluster"`
	Tools      []string `json:"tools"`
	Models     []string `json:"models"`
	Deployment string   `json:"deployment"`
}

// RegisterAgentRunRequest represents the request to register a new agent run
type RegisterAgentRunRequest struct {
	Created   string   `json:"created"`
	Status    string   `json:"status"`
	TimeTaken string   `json:"time_taken"`
	Initiator string   `json:"initiator"`
	Tools     []string `json:"tools"`
	Cost      float64  `json:"cost"`
	Models    []string `json:"models"`
	RunID     int64    `json:"id"`
	TaskID    int64    `json:"task_id"`
}

// RegisterAgentRunBatchRequest represents a batch request to register multiple agent runs
type RegisterAgentRunBatchRequest struct {
	Runs []RegisterAgentRunRequest `json:"runs"`
}
