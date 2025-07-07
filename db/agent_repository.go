package db

import (
	"context"
	"errors"
	"time"

	"ripple/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AgentRepository handles database operations for agents
type AgentRepository struct {
	db         *MongoDB
	agents     *mongo.Collection
	versions   *mongo.Collection
	runs       *mongo.Collection
	timeoutSec int
}

// NewAgentRepository creates a new agent repository
func NewAgentRepository(db *MongoDB) *AgentRepository {
	return &AgentRepository{
		db:         db,
		agents:     db.Database.Collection("agents"),
		versions:   db.Database.Collection("agent_versions"),
		runs:       db.Database.Collection("agent_runs"),
		timeoutSec: 10,
	}
}

// CreateAgent creates a new agent in the database
func (r *AgentRepository) CreateAgent(agent *models.Agent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	// Check if agent with the same name already exists
	var existingAgent models.Agent
	err := r.agents.FindOne(ctx, bson.M{"name": agent.Name}).Decode(&existingAgent)
	if err == nil {
		return errors.New("agent with this name already exists")
	} else if err != mongo.ErrNoDocuments {
		return err
	}

	// Set timestamps
	now := time.Now()
	agent.CreatedAt = now
	agent.UpdatedAt = now

	// Insert the agent
	result, err := r.agents.InsertOne(ctx, agent)
	if err != nil {
		return err
	}

	// Set the ID from the insert result
	agent.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetAgentByID retrieves an agent by ID
func (r *AgentRepository) GetAgentByID(id primitive.ObjectID) (*models.Agent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	var agent models.Agent
	err := r.agents.FindOne(ctx, bson.M{"_id": id}).Decode(&agent)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("agent not found")
		}
		return nil, err
	}

	return &agent, nil
}

// GetAgentByName retrieves an agent by name
func (r *AgentRepository) GetAgentByName(name string) (*models.Agent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	var agent models.Agent
	err := r.agents.FindOne(ctx, bson.M{"name": name}).Decode(&agent)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("agent not found")
		}
		return nil, err
	}

	return &agent, nil
}

// ListAgents retrieves all agents
func (r *AgentRepository) ListAgents() ([]models.Agent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})
	cursor, err := r.agents.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var agents []models.Agent
	if err := cursor.All(ctx, &agents); err != nil {
		return nil, err
	}

	return agents, nil
}

// CreateAgentVersion creates a new agent version
func (r *AgentRepository) CreateAgentVersion(version *models.AgentVersion) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	// Check if agent exists
	_, err := r.GetAgentByID(version.AgentID)
	if err != nil {
		return err
	}

	// Check if version already exists for this agent
	var existingVersion models.AgentVersion
	err = r.versions.FindOne(ctx, bson.M{
		"agent_id": version.AgentID,
		"version":  version.Version,
	}).Decode(&existingVersion)
	if err == nil {
		return errors.New("version already exists for this agent")
	} else if err != mongo.ErrNoDocuments {
		return err
	}

	// Set timestamps
	now := time.Now()
	version.CreatedAt = now
	version.UpdatedAt = now

	// Insert the version
	result, err := r.versions.InsertOne(ctx, version)
	if err != nil {
		return err
	}

	// Set the ID from the insert result
	version.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetAgentVersions retrieves all versions for an agent
func (r *AgentRepository) GetAgentVersions(agentID primitive.ObjectID) ([]models.AgentVersion, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	// Check if agent exists
	_, err := r.GetAgentByID(agentID)
	if err != nil {
		return nil, err
	}

	opts := options.Find().SetSort(bson.D{{Key: "version", Value: -1}})
	cursor, err := r.versions.Find(ctx, bson.M{"agent_id": agentID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var versions []models.AgentVersion
	if err := cursor.All(ctx, &versions); err != nil {
		return nil, err
	}

	return versions, nil
}

// GetAgentVersion retrieves a specific version for an agent
func (r *AgentRepository) GetAgentVersion(agentID primitive.ObjectID, version string) (*models.AgentVersion, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	var agentVersion models.AgentVersion
	err := r.versions.FindOne(ctx, bson.M{
		"agent_id": agentID,
		"version":  version,
	}).Decode(&agentVersion)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("version not found for this agent")
		}
		return nil, err
	}

	return &agentVersion, nil
}

// CreateAgentRun creates a new agent run
func (r *AgentRepository) CreateAgentRun(run *models.AgentRun) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	// Check if agent exists
	_, err := r.GetAgentByID(run.AgentID)
	if err != nil {
		return err
	}

	// Check if version exists
	version, err := r.GetAgentVersion(run.AgentID, run.Version)
	if err != nil {
		return err
	}
	run.VersionID = version.ID

	// Set recorded timestamp
	run.RecordedAt = time.Now()

	// Insert the run
	result, err := r.runs.InsertOne(ctx, run)
	if err != nil {
		return err
	}

	// Set the ID from the insert result
	run.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// CreateAgentRunBatch creates multiple agent runs in a batch
func (r *AgentRepository) CreateAgentRunBatch(runs []*models.AgentRun) error {
	if len(runs) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeoutSec*2)*time.Second)
	defer cancel()

	// Check if agent exists (using the first run's agent ID)
	_, err := r.GetAgentByID(runs[0].AgentID)
	if err != nil {
		return err
	}

	// Process each run to set version ID and recorded timestamp
	documents := make([]interface{}, len(runs))
	now := time.Now()

	for i, run := range runs {
		// Check if version exists
		version, err := r.GetAgentVersion(run.AgentID, run.Version)
		if err != nil {
			return err
		}
		run.VersionID = version.ID
		run.RecordedAt = now
		documents[i] = run
	}

	// Insert all runs in a single batch operation
	result, err := r.runs.InsertMany(ctx, documents)
	if err != nil {
		return err
	}

	// Set the IDs from the insert result
	for i, id := range result.InsertedIDs {
		runs[i].ID = id.(primitive.ObjectID)
	}

	return nil
}

// GetAgentRuns retrieves all runs for an agent
func (r *AgentRepository) GetAgentRuns(agentID primitive.ObjectID) ([]models.AgentRun, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	// Check if agent exists
	_, err := r.GetAgentByID(agentID)
	if err != nil {
		return nil, err
	}

	opts := options.Find().SetSort(bson.D{{Key: "created", Value: -1}})
	cursor, err := r.runs.Find(ctx, bson.M{"agent_id": agentID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var runs []models.AgentRun
	if err := cursor.All(ctx, &runs); err != nil {
		return nil, err
	}

	return runs, nil
}

// GetAgentVersionRuns retrieves all runs for a specific agent version
func (r *AgentRepository) GetAgentVersionRuns(agentID primitive.ObjectID, version string) ([]models.AgentRun, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	// Check if agent exists
	_, err := r.GetAgentByID(agentID)
	if err != nil {
		return nil, err
	}

	// Check if version exists
	agentVersion, err := r.GetAgentVersion(agentID, version)
	if err != nil {
		return nil, err
	}

	opts := options.Find().SetSort(bson.D{{Key: "created", Value: -1}})
	cursor, err := r.runs.Find(ctx, bson.M{
		"agent_id":   agentID,
		"version_id": agentVersion.ID,
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var runs []models.AgentRun
	if err := cursor.All(ctx, &runs); err != nil {
		return nil, err
	}

	return runs, nil
}
