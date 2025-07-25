# Agent Metrics Server

A Go server for tracking agent metrics, versions, and runs using MongoDB.

## Prerequisites

- Go 1.20 or higher
- MongoDB

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/your-username/agent-metrics.git
   cd agent-metrics
   ```

2. Install dependencies:
   ```
   go mod download
   ```

## Running the Server

```
go run cmd/server/main.go --mongo-uri=mongodb://localhost:27017 --db-name=agent_metrics --port=9999
```

Command line flags:
- `--mongo-uri`: MongoDB connection URI (default: "mongodb://localhost:27017")
- `--db-name`: MongoDB database name (default: "agent_metrics")
- `--port`: HTTP server port (default: "8080")

## Running the Worker

The worker processes agent metrics data and generates aggregated statistics for the UI dashboard. It calculates metrics such as average runtime, success rate, total runs, and spend for each agent version.

```
go run cmd/worker/main.go
```

Environment variables:
- `MONGO_URL`: MongoDB connection URI (e.g., "mongodb://localhost:27017")

The worker performs the following tasks:
1. Retrieves all agents and agent versions from the database
2. For each agent version, calculates:
   - Total number of runs
   - Last seen time (most recent run)
   - Average runtime
   - Success rate (percentage of successful runs)
   - Total cost/spend
3. Stores these metrics in the `agent_version_metrics` collection for use by the UI

## API Endpoints

### Agents

- **List all agents**
  ```
  GET /api/v1/agents
  ```

- **Register a new agent**
  ```
  POST /api/v1/agents/{name}/register
  
  Request Body:
  {
    "name": "agent-name",
    "project": "project-name"
  }
  ```

### Agent Versions

- **Add a new agent version**
  ```
  POST /api/v1/agents/{agentId}/versions
  
  Request Body:
  {
    "version": "1.0.2",
    "cluster": "123",
    "tools": ["tool1", "tool2"],
    "models": ["model1", "model2"],
    "deployment": "deployment-name"
  }
  ```

- **Get all versions for an agent**
  ```
  GET /api/v1/agents/{agentId}/versions
  ```

- **Get a specific agent version**
  ```
  GET /api/v1/agents/{agentId}/versions/{version}
  ```

### Agent Runs

- **Add a new agent run**
  ```
  POST /api/v1/agents/{agentId}/versions/{version}/runs
  
  Request Body (Single Run):
  {
    "created": "2023-08-01T12:00:00Z",
    "status": "completed",
    "time_taken": "5m30s",
    "initiator": "user123",
    "tools": ["tool1", "tool2"],
    "cost": 0.1,
    "models": ["model1", "model2"],
    "id": 123,
    "task_id": 12
  }
  
  Request Body (Batch of Runs):
  {
    "runs": [
      {
        "created": "2023-08-01T12:00:00Z",
        "status": "completed",
        "time_taken": "5m30s",
        "initiator": "user123",
        "tools": ["tool1", "tool2"],
        "cost": 0.1,
        "models": ["model1", "model2"],
        "id": 123,
        "task_id": 12
      },
      {
        "created": "2023-08-01T13:00:00Z",
        "status": "failed",
        "time_taken": "2m10s",
        "initiator": "user456",
        "tools": ["tool3", "tool4"],
        "cost": 0.05,
        "models": ["model3"],
        "id": 124,
        "task_id": 13
      }
    ]
  }
  ```

- **Get all runs for a specific agent version**
  ```
  GET /api/v1/agents/{agentId}/versions/{version}/runs
  ```

- **Get all runs for an agent (across all versions)**
  ```
  GET /api/v1/agents/{agentId}/runs
  ```

### UI Endpoints

- **Get Dashboard Statistics**
  ```
  GET /api/v1/ui/stats
  
  Response:
  [
    {
      "title": "Active Agents",
      "value": "42",
      "change": "+5 from last week",
      "icon": "Bot",
      "trend": "up",
      "raw": 42
    },
    {
      "title": "Total Runs Today",
      "value": "1,234",
      "change": "+15% from yesterday",
      "icon": "Activity",
      "trend": "up",
      "raw": 1234
    },
    {
      "title": "Avg Response Time",
      "value": "3.2s",
      "change": "-0.5s from last hour",
      "icon": "Clock",
      "trend": "down",
      "raw": 3.2
    },
    {
      "title": "Total Cost Today",
      "value": "$123.45",
      "change": "+5% from yesterday",
      "icon": "DollarSign",
      "trend": "up",
      "raw": 123.45
    }
  ]
  ```

- **Get Recent Activity**
  ```
  GET /api/v1/ui/recent_activity
  
  Response:
  [
    {
      "id": 12345,
      "agent": "agent-name",
      "action": "completed run",
      "status": "completed",
      "time": "2023-08-01T12:00:00Z",
      "duration": 5.5,
      "cost": 0.1
    }
  ]
  ```

- **Get Agent Versions with Metrics**
  ```
  GET /api/v1/ui/agent_versions
  
  Response:
  [
    {
      "id": "5f8d0d55b54764429a0e36a1",
      "name": "agent-name",
      "project": "project-name",
      "status": "active",
      "lastSeen": "2023-08-01T12:00:00Z",
      "version": "1.0.2",
      "avgRuntime": 3.5,
      "successRate": 98.5,
      "totalRuns": 1234,
      "spend": 123.45,
      "tools": ["tool1", "tool2"],
      "models": ["model1", "model2"],
      "cluster": "123"
    }
  ]
  ```

## Example Usage

### Agents

#### List all agents

```bash
curl -X GET http://localhost:9999/api/v1/agents
```

#### Register a new agent

```bash
curl -X POST http://localhost:9999/api/v1/agents/my-agent/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-agent", 
    "project": "my-project"
  }'
```

### Agent Versions

#### Add a new agent version

```bash
curl -X POST http://localhost:9999/api/v1/agents/{agentId}/versions \
  -H "Content-Type: application/json" \
  -d '{
    "version": "1.0.2",
    "cluster": "123",
    "tools": ["tool1", "tool2"],
    "models": ["model1", "model2"],
    "deployment": "prod"
  }'
```

#### Get all versions for an agent

```bash
curl -X GET http://localhost:9999/api/v1/agents/{agentId}/versions
```

#### Get a specific agent version

```bash
curl -X GET http://localhost:9999/api/v1/agents/{agentId}/versions/{version}
```

### Agent Runs

#### Add a new agent run

```bash
# Single run
curl -X POST http://localhost:9999/api/v1/agents/{agentId}/versions/1.0.2/runs \
  -H "Content-Type: application/json" \
  -d '{
    "created": "2023-08-01T12:00:00Z",
    "status": "completed",
    "time_taken": "5m30s",
    "initiator": "user123",
    "tools": ["tool1", "tool2"],
    "cost": 0.1,
    "models": ["model1", "model2"],
    "id": 123,
    "task_id": 12
  }'

# Batch of runs
curl -X POST http://localhost:9999/api/v1/agents/{agentId}/versions/1.0.2/runs \
  -H "Content-Type: application/json" \
  -d '{
    "runs": [
      {
        "created": "2023-08-01T12:00:00Z",
        "status": "completed",
        "time_taken": "5m30s",
        "initiator": "user123",
        "tools": ["tool1", "tool2"],
        "cost": 0.1,
        "models": ["model1", "model2"],
        "id": 123,
        "task_id": 12
      },
      {
        "created": "2023-08-01T13:00:00Z",
        "status": "failed",
        "time_taken": "2m10s",
        "initiator": "user456",
        "tools": ["tool3", "tool4"],
        "cost": 0.05,
        "models": ["model3"],
        "id": 124,
        "task_id": 13
      }
    ]
  }'
```

#### Get all runs for a specific agent version

```bash
curl -X GET http://localhost:9999/api/v1/agents/{agentId}/versions/{version}/runs
```

#### Get all runs for an agent (across all versions)

```bash
curl -X GET http://localhost:9999/api/v1/agents/{agentId}/runs
```

### UI Endpoints

#### Get Dashboard Statistics

```bash
curl -X GET http://localhost:9999/api/v1/ui/stats
```

#### Get Recent Activity

```bash
curl -X GET http://localhost:9999/api/v1/ui/recent_activity
```

#### Get Agent Versions with Metrics

```bash
curl -X GET http://localhost:9999/api/v1/ui/agent_versions
```