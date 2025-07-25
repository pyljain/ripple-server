package main

import (
	"context"
	"log"
	"os"
	"ripple/db"
	"ripple/models"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	workerPoolSize = 10
)

func main() {
	ctx := context.Background()
	client, err := db.NewMongoDB(os.Getenv("MONGO_URL"), "agent_metrics")
	if err != nil {
		log.Printf("Unable to connect to the Mongo store to read from %s", err)
		os.Exit(-1)
	}

	// Get a list of agent names and versions
	agentCollectionCursor, err := client.Database.Collection("agents").Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Unable to fetch agents %s", err)
		os.Exit(-1)
	}

	agents := []*models.Agent{}
	err = agentCollectionCursor.All(ctx, &agents)
	if err != nil {
		log.Printf("Unable to fetch agents %s", err)
		os.Exit(-1)
	}

	agentToAgentIDLookup := make(map[string]*models.Agent, len(agents))
	for _, a := range agents {
		agentToAgentIDLookup[string(a.ID.Hex())] = a
	}

	// Get all agent versions in the collection
	agentVersionsCursor, err := client.Database.Collection("agent_versions").Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Unable to fetch agent versions %s", err)
		os.Exit(-1)
	}

	agentVersions := []*models.AgentVersion{}
	err = agentVersionsCursor.All(ctx, &agentVersions)
	if err != nil {
		log.Printf("Unable to fetch agent versions %s", err)
		os.Exit(-1)
	}

	// Get filtered runs per agent version. Use go routines, one per agent versions

	workChan := make(chan *Work)
	defer close(workChan)

	wg := sync.WaitGroup{}
	for i := 0; i < workerPoolSize; i++ {
		go worker(ctx, client, workChan, &wg)
	}

	for _, av := range agentVersions {
		w := Work{
			agent:        agentToAgentIDLookup[string(av.AgentID.Hex())],
			agentVersion: av,
		}

		wg.Add(1)
		workChan <- &w
	}

	wg.Wait()
	/* Run aggregate queries to compute all elements that need to be aggregated
	# of runs
	Last seen time
	# Average Latency
	Total errors
	Total spend for the day
	*/

	// Write into the aggregates flat collection

}

type Work struct {
	agent        *models.Agent
	agentVersion *models.AgentVersion
}

func worker(ctx context.Context, client *db.MongoDB, workChan chan *Work, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Exiting worker as the context was cancelled")
			wg.Done()
			return
		case work := <-workChan:
			agentVersion := work.agentVersion
			count, err := client.Database.Collection("agent_runs").CountDocuments(ctx, bson.M{"version_id": agentVersion.ID})
			if err != nil {
				log.Printf("Unable to fetch number of runs for the agent with ID %s and version %s. Error is %s", agentVersion.AgentID, agentVersion.Version, err)
				wg.Done()
				continue
			}

			// Last seen time
			res := client.Database.Collection("agent_runs").FindOne(ctx, bson.M{"version_id": agentVersion.ID}, &options.FindOneOptions{
				Sort: bson.M{
					"recorded_at": -1,
				},
			})
			if res.Err() != nil {
				log.Printf("Unable to fetch last seen time for the agent with ID %s and version %s. Error is %s", agentVersion.AgentID, agentVersion.Version, res.Err())
				wg.Done()
				continue
			}

			lastRecord := models.AgentRun{}
			err = res.Decode(&lastRecord)
			if err != nil {
				log.Printf("Unable to fetch last seem time for  agent with ID %s and version %s. Error is %s", agentVersion.AgentID, agentVersion.Version, err)
				wg.Done()
				continue
			}

			// Count total errors
			countErrors, err := client.Database.Collection("agent_runs").CountDocuments(ctx, bson.M{"version_id": agentVersion.ID, "status": "error"})
			if err != nil {
				log.Printf("Unable to fetch number of runs for the agent with ID %s and version %s. Error is %s", agentVersion.AgentID, agentVersion.Version, err)
				wg.Done()
				continue
			}

			//Average Time Taken and Total Cost
			pipeline := []bson.M{
				{
					"$match": bson.M{
						"version_id": agentVersion.ID,
					},
				},
				{
					"$group": bson.M{
						"_id": nil,
						"avgTimeTaken": bson.M{
							"$avg": "$time_taken",
						},
						"totalCost": bson.M{
							"$sum": "$cost",
						},
					},
				},
			}

			cursor, err := client.Database.Collection("agent_runs").Aggregate(ctx, pipeline)
			if err != nil {
				log.Printf("Unable to fetch metrics for the agent with ID %s and version %s. Error is %s", agentVersion.AgentID, agentVersion.Version, err)
				wg.Done()
				continue
			}

			var results []bson.M
			if err = cursor.All(ctx, &results); err != nil {
				log.Printf("Unable to decode metrics for the agent with ID %s and version %s. Error is %s", agentVersion.AgentID, agentVersion.Version, err)
				wg.Done()
				continue
			}

			var avgTimeTaken float64
			var totalCost float64

			if len(results) > 0 {
				if val, ok := results[0]["avgTimeTaken"].(float64); ok {
					avgTimeTaken = val
				}
				if val, ok := results[0]["totalCost"].(float64); ok {
					totalCost = val
				}
			}

			avm := models.AgentVersionMetrics{
				Id:             agentVersion.ID,
				Name:           work.agent.Name,
				Project:        work.agent.Project,
				Status:         agentVersion.Status,
				LastSeen:       lastRecord.RecordedAt,
				Version:        agentVersion.Version,
				AverageRunTime: avgTimeTaken,
				SuccessRate:    (float64(count-countErrors) / float64(count)) * 100,
				TotalRuns:      count,
				Spend:          totalCost,
				Tools:          agentVersion.Tools,
				Models:         agentVersion.Models,
				Cluster:        agentVersion.Cluster,
			}
			upsert := true
			updateDoc := bson.M{
				"$set": avm,
			}
			_, err = client.Database.Collection("agent_version_metrics").UpdateOne(ctx, bson.M{"_id": agentVersion.ID}, updateDoc, &options.UpdateOptions{
				Upsert: &upsert,
			})
			if err != nil {
				log.Printf("Unable to insert metric record for agent %s. Error is %s", work.agent.Name, err)
				wg.Done()
				continue
			}

			wg.Done()
		}
	}
}
