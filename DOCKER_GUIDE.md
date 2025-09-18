# Running FeaturePlus with Docker

This guide explains how to run the FeaturePlus application using Docker on your Linux server.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) installed on your Linux server
- [Docker Compose](https://docs.docker.com/compose/install/) (optional, for easier deployment)

## Setting Up Environment Variables

1. Create or modify your `.env` file in the project root with the required variables:

   ```bash
   # Required for GitHub API access
   GITHUB_TOKEN=your_github_token_here
   
   # Optional settings
   GIN_MODE=release
   ```

## Deployment Options

### Option 1: Using Docker Directly

1. Build the Docker image:

   ```bash
   docker build -t featureplus .
   ```

2. Run the container with environment variables from your `.env` file:

   ```bash
   docker run -d \
     --name featureplus \
     -p 8080:8080 \
     -v featureplus_data:/app/data \
     --env-file .env \
     --restart unless-stopped \
     featureplus
   ```

### Option 2: Using Docker Compose

1. Simply run:

   ```bash
   docker-compose up -d
   ```

2. The application will automatically use the `.env` file in the project root

## Configuration

### Environment Variables

The following environment variables can be configured in the `docker-compose.yml` file:

- `DATA_PATH`: Path where data files (templates.json, guidance.json) are stored (default: `/app/data`)
- `GIN_MODE`: Gin framework mode (default: `release`)
- `GITHUB_TOKEN`: Your GitHub API token (required)

### Data Persistence

The application uses a Docker volume (`featureplus_data`) to persist data files. This ensures your templates, guidance, and other data are preserved between container restarts.

## Managing Data Files

### Accessing Data Files

To access or modify data files directly on your Linux server:

1. Find the volume name:

   ```bash
   docker volume ls | grep featureplus_data
   ```

2. Inspect the volume to find its location:

   ```bash
   docker volume inspect featureplus_data
   ```

3. The volume is typically located in `/var/lib/docker/volumes/featureplus_data/_data` on your Linux server. You may need root privileges to access this directory:

   ```bash
   sudo ls -la /var/lib/docker/volumes/featureplus_data/_data
   ```

### Initializing with Custom Data

To initialize the container with custom data files:

1. Place your JSON files in the `backend/data/` directory before building the Docker image.
2. Rebuild the Docker image:

   ```bash
   # Using Docker directly
   docker build -t featureplus .
   docker run -d --name featureplus -p 8080:8080 -v featureplus_data:/app/data --env-file .env featureplus
   
   # Or using Docker Compose
   docker-compose build
   docker-compose up -d
   ```

## Troubleshooting

### Checking Logs

If you encounter issues, check the container logs:

```bash
# Using Docker directly
docker logs -f featureplus

# Or using Docker Compose
docker-compose logs -f featureplus
```

### Restarting the Application

To restart the application:

```bash
# Using Docker directly
docker restart featureplus

# Or using Docker Compose
docker-compose restart
```

### Complete Reset

To completely reset the application (including data):

```bash
# Using Docker directly
docker stop featureplus
docker rm featureplus
docker volume rm featureplus_data
docker run -d --name featureplus -p 8080:8080 -v featureplus_data:/app/data --env-file .env featureplus

# Or using Docker Compose
docker-compose down -v
docker-compose up -d
```

**Warning**: This will delete all persistent data!

## Advanced Configuration

### Securing Your Application

For production deployment, consider adding a reverse proxy like Nginx in front of your application to handle SSL termination and additional security measures.

### Backup and Restore

To backup your data volume:

```bash
docker run --rm -v featureplus_data:/data -v $(pwd):/backup alpine tar -czf /backup/featureplus-data-backup.tar.gz -C /data .
```

To restore from backup:

```bash
docker run --rm -v featureplus_data:/data -v $(pwd):/backup alpine sh -c "cd /data && tar -xzf /backup/featureplus-data-backup.tar.gz"
```
