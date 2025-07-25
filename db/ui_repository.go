package db

import (
	"context"
	"fmt"
	"ripple/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UIRepository handles database operations for UI-related data
type UIRepository struct {
	db         *MongoDB
	agents     *mongo.Collection
	versions   *mongo.Collection
	runs       *mongo.Collection
	timeoutSec int
}

// NewUIRepository creates a new UI repository
func NewUIRepository(db *MongoDB) *UIRepository {
	return &UIRepository{
		db:         db,
		agents:     db.Database.Collection("agents"),
		versions:   db.Database.Collection("agent_versions"),
		runs:       db.Database.Collection("agent_runs"),
		timeoutSec: 10,
	}
}

// StatsData represents the data structure for UI stats
type StatsData struct {
	Title  string  `json:"title"`
	Value  string  `json:"value"`
	Change string  `json:"change"`
	Icon   string  `json:"icon"`
	Trend  string  `json:"trend"`
	Raw    float64 `json:"raw,omitempty"`
}

// ActivityData represents a single activity item for the UI
type ActivityData struct {
	ID       int64     `json:"id"`
	Agent    string    `json:"agent"`
	Action   string    `json:"action"`
	Status   string    `json:"status"`
	Time     time.Time `json:"time"`
	Duration float64   `json:"duration"`
	Cost     float64   `json:"cost"`
}

type AgentVersion struct {
	Id             string    `json:"id"`
	Name           string    `json:"name"`
	Project        string    `json:"project"`
	Status         string    `json:"status"`
	LastSeen       time.Time `json:"lastSeen"`
	Version        string    `json:"version"`
	AverageRunTime float32   `json:"avgRuntime"`
	SuccessRate    float32   `json:"successRate"`
	TotalRuns      int       `json:"totalRuns"`
	SpendToday     float64   `json:"costToday"`
	Tools          []string  `json:"tools"`
	Models         []string  `json:"models"`
	Cluster        string    `json:"cluster"`
}

// GetDashboardStats retrieves statistics for the dashboard
func (r *UIRepository) GetDashboardStats() ([]StatsData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	// Get current time for time-based calculations
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)
	lastWeek := today.AddDate(0, 0, -7)
	lastHour := now.Add(-1 * time.Hour)
	last48Hours := now.Add(-48 * time.Hour)

	// 1. Active Agents (agents with runs in the last 48 hours)
	activeAgentsNow, err := r.getActiveAgentsCount(ctx, last48Hours)
	if err != nil {
		return nil, fmt.Errorf("failed to get active agents count: %w", err)
	}

	activeAgentsLastWeek, err := r.getActiveAgentsCount(ctx, lastWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get last week's active agents count: %w", err)
	}

	activeAgentsDiff := activeAgentsNow - activeAgentsLastWeek
	activeAgentsTrend := "up"
	activeAgentsChangePrefix := "+"
	if activeAgentsDiff < 0 {
		activeAgentsTrend = "down"
		activeAgentsChangePrefix = ""
	} else if activeAgentsDiff == 0 {
		activeAgentsTrend = "neutral"
		activeAgentsChangePrefix = ""
	}

	// 2. Total Runs Today
	runsToday, err := r.getRunsCount(ctx, today, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get today's runs count: %w", err)
	}

	runsYesterday, err := r.getRunsCount(ctx, yesterday, today)
	if err != nil {
		return nil, fmt.Errorf("failed to get yesterday's runs count: %w", err)
	}

	runsTrend := "up"
	runsChangePrefix := "+"
	runsPercentChange := 0.0
	if runsYesterday > 0 {
		runsPercentChange = (float64(runsToday-runsYesterday) / float64(runsYesterday)) * 100
	}
	if runsPercentChange < 0 {
		runsTrend = "down"
		runsChangePrefix = ""
	} else if runsPercentChange == 0 {
		runsTrend = "neutral"
		runsChangePrefix = ""
	}

	// 3. Average Response Time
	avgResponseTimeNow, err := r.getAvgResponseTime(ctx, lastHour, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get current average response time: %w", err)
	}

	avgResponseTimePrev, err := r.getAvgResponseTime(ctx, lastHour.Add(-1*time.Hour), lastHour)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous average response time: %w", err)
	}

	responseTrend := "down" // Lower response time is better
	responseChangePrefix := "-"
	responseDiff := avgResponseTimePrev - avgResponseTimeNow
	if responseDiff < 0 {
		responseTrend = "up" // Response time increased (worse)
		responseChangePrefix = "+"
		responseDiff = -responseDiff
	} else if responseDiff == 0 {
		responseTrend = "neutral"
		responseChangePrefix = ""
	}

	// 4. Total Cost Today
	costToday, err := r.getTotalCost(ctx, today, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get today's total cost: %w", err)
	}

	costYesterday, err := r.getTotalCost(ctx, yesterday, today)
	if err != nil {
		return nil, fmt.Errorf("failed to get yesterday's total cost: %w", err)
	}

	costTrend := "up"
	costChangePrefix := "+"
	costPercentChange := 0.0
	if costYesterday > 0 {
		costPercentChange = ((costToday - costYesterday) / costYesterday) * 100
	}
	if costPercentChange < 0 {
		costTrend = "down"
		costChangePrefix = ""
	} else if costPercentChange == 0 {
		costTrend = "neutral"
		costChangePrefix = ""
	}

	// Format the stats data
	stats := []StatsData{
		{
			Title:  "Active Agents",
			Value:  fmt.Sprintf("%d", activeAgentsNow),
			Change: fmt.Sprintf("%s%d from last week", activeAgentsChangePrefix, abs(activeAgentsDiff)),
			Icon:   "Bot",
			Trend:  activeAgentsTrend,
			Raw:    float64(activeAgentsNow),
		},
		{
			Title:  "Total Runs Today",
			Value:  formatNumber(runsToday),
			Change: fmt.Sprintf("%s%d%% from yesterday", runsChangePrefix, abs(int(runsPercentChange))),
			Icon:   "Activity",
			Trend:  runsTrend,
			Raw:    float64(runsToday),
		},
		{
			Title:  "Avg Response Time",
			Value:  fmt.Sprintf("%.1fs", avgResponseTimeNow),
			Change: fmt.Sprintf("%s%.1fs from last hour", responseChangePrefix, responseDiff),
			Icon:   "Clock",
			Trend:  responseTrend,
			Raw:    avgResponseTimeNow,
		},
		{
			Title:  "Total Cost Today",
			Value:  fmt.Sprintf("$%.2f", costToday),
			Change: fmt.Sprintf("%s%d%% from yesterday", costChangePrefix, abs(int(costPercentChange))),
			Icon:   "DollarSign",
			Trend:  costTrend,
			Raw:    costToday,
		},
	}

	return stats, nil
}

func (r *UIRepository) GetAgentVersions(ctx context.Context) ([]models.AgentVersionMetrics, error) {
	cursor, err := r.db.Database.Collection("agent_version_metrics").Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	versions := []models.AgentVersionMetrics{}
	err = cursor.All(ctx, &versions)
	if err != nil {
		return nil, err
	}

	return versions, nil
}

// getActiveAgentsCount returns the count of unique agents with runs since the given time
func (r *UIRepository) getActiveAgentsCount(ctx context.Context, since time.Time) (int, error) {
	pipeline := mongo.Pipeline{
		{
			{"$match", bson.M{
				"created": bson.M{"$gte": since},
			}},
		},
		{
			{"$group", bson.M{
				"_id": "$agent_id",
			}},
		},
		{
			{"$count", "count"},
		},
	}

	cursor, err := r.runs.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return 0, err
	}

	if len(results) == 0 {
		return 0, nil
	}

	return int(results[0]["count"].(int32)), nil
}

// getRunsCount returns the count of runs between the given time range
func (r *UIRepository) getRunsCount(ctx context.Context, start, end time.Time) (int, error) {
	count, err := r.runs.CountDocuments(ctx, bson.M{
		"created": bson.M{
			"$gte": start,
			"$lt":  end,
		},
	})
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// getAvgResponseTime calculates the average response time for runs in the given time range
func (r *UIRepository) getAvgResponseTime(ctx context.Context, start, end time.Time) (float64, error) {
	// We need to convert time_taken string to seconds for calculation
	// For simplicity, we'll assume time_taken is stored in seconds as a string
	pipeline := mongo.Pipeline{
		{
			{"$match", bson.M{
				"created": bson.M{
					"$gte": start,
					"$lt":  end,
				},
			}},
		},
		{
			{"$addFields", bson.M{
				"time_taken_seconds": bson.M{
					"$toDouble": bson.M{
						"$replaceAll": bson.M{
							"input":       "$time_taken",
							"find":        "s",
							"replacement": "",
						},
					},
				},
			}},
		},
		{
			{"$group", bson.M{
				"_id":   nil,
				"avg":   bson.M{"$avg": "$time_taken_seconds"},
				"count": bson.M{"$sum": 1},
			}},
		},
	}

	cursor, err := r.runs.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return 0, err
	}

	if len(results) == 0 || results[0]["count"].(int32) == 0 {
		return 0, nil
	}

	return results[0]["avg"].(float64), nil
}

// getTotalCost calculates the total cost for runs in the given time range
func (r *UIRepository) getTotalCost(ctx context.Context, start, end time.Time) (float64, error) {
	pipeline := mongo.Pipeline{
		{
			{"$match", bson.M{
				"created": bson.M{
					"$gte": start,
					"$lt":  end,
				},
			}},
		},
		{
			{"$group", bson.M{
				"_id":   nil,
				"total": bson.M{"$sum": "$cost"},
			}},
		},
	}

	cursor, err := r.runs.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return 0, err
	}

	if len(results) == 0 {
		return 0, nil
	}

	return results[0]["total"].(float64), nil
}

// GetRecentActivity retrieves the 10 most recent agent runs
func (r *UIRepository) GetRecentActivity() ([]ActivityData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeoutSec)*time.Second)
	defer cancel()

	// Create a pipeline to get the 10 most recent runs with agent names
	pipeline := mongo.Pipeline{
		{
			{"$sort", bson.M{
				"created": -1,
			}},
		},
		{
			{"$limit", 10},
		},
		{
			{"$lookup", bson.M{
				"from":         "agents",
				"localField":   "agent_id",
				"foreignField": "_id",
				"as":           "agent_info",
			}},
		},
		{
			{"$unwind", "$agent_info"},
		},
		{
			{"$project", bson.M{
				"_id":        0,
				"id":         "$run_id",
				"agent_name": "$agent_info.name",
				"status":     1,
				"created":    1,
				"time_taken": 1,
				"cost":       1,
			}},
		},
	}

	cursor, err := r.runs.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent activity: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode recent activity: %w", err)
	}

	activities := make([]ActivityData, 0, len(results))
	for _, result := range results {
		// Determine action based on status
		action := "completed run"
		if result["status"].(string) == "error" {
			action = "failed run"
		} else if result["status"].(string) == "timeout" {
			action = "timed out"
		} else if result["status"].(string) == "running" {
			action = "started run"
		}

		// Convert MongoDB primitive.DateTime to time.Time
		var createdTime time.Time
		switch created := result["created"].(type) {
		case time.Time:
			createdTime = created
		case primitive.DateTime:
			createdTime = created.Time()
		default:
			createdTime = time.Now() // Fallback
		}

		activity := ActivityData{
			ID:       result["id"].(int64),
			Agent:    result["agent_name"].(string),
			Action:   action,
			Status:   result["status"].(string),
			Time:     createdTime,
			Duration: result["time_taken"].(float64),
			Cost:     result["cost"].(float64),
		}

		activities = append(activities, activity)
	}

	return activities, nil
}

// Helper functions

// abs returns the absolute value of an integer
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// formatNumber formats a number with commas as thousands separators
func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%d,%03d", n/1000, n%1000)
}
