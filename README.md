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