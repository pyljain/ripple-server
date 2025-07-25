package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// Select database
	db := client.Database("agent_metrics")

	// Collections
	agentsCollection := db.Collection("agents")
	versionsCollection := db.Collection("agent_versions")
	runsCollection := db.Collection("agent_runs")

	// Clear existing data
	agentsCollection.DeleteMany(ctx, bson.M{})
	versionsCollection.DeleteMany(ctx, bson.M{})
	runsCollection.DeleteMany(ctx, bson.M{})

	// Create agents
	agentIDs := make([]primitive.ObjectID, 0)
	agentNames := []string{"search-agent", "chat-agent", "code-agent", "data-agent", "image-agent"}
	projects := []string{"search", "chat", "code", "data", "image"}

	for i, name := range agentNames {
		now := time.Now()
		agent := bson.M{
			"name":       name,
			"project":    projects[i],
			"created_at": now,
			"updated_at": now,
		}

		result, err := agentsCollection.InsertOne(ctx, agent)
		if err != nil {
			log.Fatalf("Failed to insert agent: %v", err)
		}

		agentID := result.InsertedID.(primitive.ObjectID)
		agentIDs = append(agentIDs, agentID)
		fmt.Printf("Created agent: %s with ID: %s\n", name, agentID.Hex())

		// Create versions for each agent
		versionIDs := make([]primitive.ObjectID, 0)
		versions := []string{"1.0.0", "1.1.0", "1.2.0"}
		clusters := []string{"production", "staging", "development"}
		deployments := []string{"kubernetes", "docker", "serverless"}

		for j, version := range versions {
			versionDoc := bson.M{
				"agent_id":   agentID,
				"version":    version,
				"cluster":    clusters[j%len(clusters)],
				"status":     "active",
				"tools":      []string{"tool1", "tool2", "tool3"},
				"models":     []string{"gpt-4", "claude-3", "gemini-pro"},
				"deployment": deployments[j%len(deployments)],
				"created_at": now.Add(-time.Duration(j*24) * time.Hour),
				"updated_at": now.Add(-time.Duration(j*24) * time.Hour),
			}

			versionResult, err := versionsCollection.InsertOne(ctx, versionDoc)
			if err != nil {
				log.Fatalf("Failed to insert version: %v", err)
			}

			versionID := versionResult.InsertedID.(primitive.ObjectID)
			versionIDs = append(versionIDs, versionID)
			fmt.Printf("  Created version: %s with ID: %s\n", version, versionID.Hex())
		}

		// Create runs for each agent version
		// We'll create runs for today, yesterday, last week, and within the last hour
		statuses := []string{"success", "error", "timeout"}
		initiators := []string{"user", "system", "scheduled"}

		// Create a batch of runs
		var runs []interface{}

		// Today's runs
		for j := 0; j < 50; j++ {
			versionIndex := j % len(versions)
			statusIndex := j % len(statuses)
			initiatorIndex := j % len(initiators)

			// Random time today
			createdTime := today().Add(time.Duration(rand.Intn(int(time.Since(today()).Seconds()))) * time.Second)

			// Random time taken between 0.5s and 5s
			timeTaken := 0.5 + rand.Float64()*4.5

			// Random cost between $0.01 and $0.50
			cost := 0.01 + rand.Float64()*0.49

			run := bson.M{
				"agent_id":    agentID,
				"version_id":  versionIDs[versionIndex],
				"version":     versions[versionIndex],
				"created":     createdTime,
				"status":      statuses[statusIndex],
				"time_taken":  timeTaken,
				"initiator":   initiators[initiatorIndex],
				"tools":       []string{"tool1", "tool2"},
				"cost":        cost,
				"models":      []string{"gpt-4"},
				"run_id":      int64(j + 1000),
				"task_id":     int64(j + 2000),
				"recorded_at": time.Now(),
			}

			runs = append(runs, run)
		}

		// Yesterday's runs
		for j := 0; j < 40; j++ {
			versionIndex := j % len(versions)
			statusIndex := j % len(statuses)
			initiatorIndex := j % len(initiators)

			// Random time yesterday
			createdTime := yesterday().Add(time.Duration(rand.Intn(86400)) * time.Second)

			// Random time taken between 0.5s and 5s
			timeTaken := 0.5 + rand.Float64()*4.5

			// Random cost between $0.01 and $0.50
			cost := 0.01 + rand.Float64()*0.49

			run := bson.M{
				"agent_id":    agentID,
				"version_id":  versionIDs[versionIndex],
				"version":     versions[versionIndex],
				"created":     createdTime,
				"status":      statuses[statusIndex],
				"time_taken":  fmt.Sprintf("%.2fs", timeTaken),
				"initiator":   initiators[initiatorIndex],
				"tools":       []string{"tool1", "tool2"},
				"cost":        cost,
				"models":      []string{"gpt-4"},
				"run_id":      int64(j + 3000),
				"task_id":     int64(j + 4000),
				"recorded_at": time.Now(),
			}

			runs = append(runs, run)
		}

		// Last hour runs
		for j := 0; j < 20; j++ {
			versionIndex := j % len(versions)
			statusIndex := j % len(statuses)
			initiatorIndex := j % len(initiators)

			// Random time in the last hour
			createdTime := time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second)

			// Random time taken between 0.5s and 5s
			timeTaken := 0.5 + rand.Float64()*4.5

			// Random cost between $0.01 and $0.50
			cost := 0.01 + rand.Float64()*0.49

			run := bson.M{
				"agent_id":    agentID,
				"version_id":  versionIDs[versionIndex],
				"version":     versions[versionIndex],
				"created":     createdTime,
				"status":      statuses[statusIndex],
				"time_taken":  fmt.Sprintf("%.2fs", timeTaken),
				"initiator":   initiators[initiatorIndex],
				"tools":       []string{"tool1", "tool2"},
				"cost":        cost,
				"models":      []string{"gpt-4"},
				"run_id":      int64(j + 5000),
				"task_id":     int64(j + 6000),
				"recorded_at": time.Now(),
			}

			runs = append(runs, run)
		}

		// Insert all runs in a batch
		_, err = runsCollection.InsertMany(ctx, runs)
		if err != nil {
			log.Fatalf("Failed to insert runs: %v", err)
		}

		fmt.Printf("  Created %d runs for agent %s\n", len(runs), name)
	}

	fmt.Println("Test data inserted successfully!")
}

// Helper functions to get time boundaries
func today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

func yesterday() time.Time {
	return today().AddDate(0, 0, -1)
}
