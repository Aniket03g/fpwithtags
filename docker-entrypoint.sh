#!/bin/sh
set -e

# Create data directory if it doesn't exist
mkdir -p /app/data

# Check if templates.json exists, create sample if not
if [ ! -f "/app/data/templates.json" ]; then
  echo "Creating sample templates.json file..."
  cat > /app/data/templates.json << 'EOF'
[
  {
    "id": "nodejs-express-mongodb",
    "name": "Node.js Express MongoDB",
    "stack": "Node.js",
    "description": "A template for Node.js applications with Express and MongoDB",
    "tech_stack": "Node.js",
    "feature_categories": ["API", "Auth", "Database"],
    "task_types": ["Backend", "Frontend", "Database"],
    "features": [
      {
        "name": "User Authentication",
        "description": "Implement user authentication with JWT",
        "category": "Auth",
        "tasks": [
          {
            "name": "Setup MongoDB Models",
            "description": "Create user model with authentication fields",
            "type": "Database",
            "priority": "high"
          },
          {
            "name": "Implement JWT Authentication",
            "description": "Create authentication middleware and routes",
            "type": "Backend",
            "priority": "high"
          }
        ]
      }
    ]
  },
  {
    "id": "go-postgresql",
    "name": "Go with PostgreSQL",
    "stack": "Go",
    "description": "A template for Go applications with PostgreSQL database",
    "tech_stack": "Go",
    "feature_categories": ["API", "Auth", "Database"],
    "task_types": ["Backend", "Database", "Testing"],
    "features": [
      {
        "name": "RESTful API",
        "description": "Implement RESTful API endpoints",
        "category": "API",
        "tasks": [
          {
            "name": "Setup PostgreSQL Models",
            "description": "Create database models and migrations",
            "type": "Database",
            "priority": "high"
          },
          {
            "name": "Implement API Handlers",
            "description": "Create API handlers and routes",
            "type": "Backend",
            "priority": "high"
          }
        ]
      }
    ]
  }
]
EOF
fi

# Check if guidance.json exists, create sample if not
if [ ! -f "/app/data/guidance.json" ]; then
  echo "Creating sample guidance.json file..."
  cat > /app/data/guidance.json << 'EOF'
{
  "stacks": [
    {
      "name": "Node.js",
      "task_types": {
        "Backend": "Use Express.js for routing and middleware. Consider using async/await for asynchronous operations.",
        "Frontend": "Consider using React or Vue.js for frontend development.",
        "Database": "Use Mongoose ODM for MongoDB interactions."
      }
    },
    {
      "name": "Go",
      "task_types": {
        "Backend": "Use the standard library's net/http package or a framework like Gin or Echo.",
        "Database": "Use GORM or sqlx for database interactions.",
        "Testing": "Use the testing package for unit tests and testify for assertions."
      }
    }
  ]
}
EOF
fi

# Make sure the data directory is owned by the application user
chown -R $(id -u):$(id -g) /app/data

# Print environment variables for debugging
echo "DATA_PATH=$DATA_PATH"

# Start the application
exec ./featureplus-server "$@"
