# Testing Backup & Restore in Docker

## Method 1: Using Shell Scripts (Easiest)

### Step 1: Access the running container

```bash
# Get container ID
docker ps

# Exec into container
docker exec -it <container_id> /bin/bash
```

### Step 2: Create a backup

```bash
cd /app
./scripts/backup.sh
```

Expected output:
```
FeaturePlus Database Backup
============================
Backing up database...
Source: backend/test.db
Destination: backups/featureplus_2025-10-29_13-45.db
✓ Backup successful!
Backup file: backups/featureplus_2025-10-29_13-45.db
Size: 19M
```

### Step 3: List backups

```bash
ls -lh backups/
```

### Step 4: Test restore

```bash
# Restore from a specific backup
./scripts/restore.sh backups/featureplus_2025-10-29_13-45.db
```

Expected output:
```
FeaturePlus Database Restore
============================
Creating safety backup of current database...
✓ Safety backup created: backups/featureplus_pre-restore_2025-10-29_13-46.db

Restoring database...
Source: backups/featureplus_2025-10-29_13-45.db
Destination: backend/test.db
✓ Restore successful!
Database restored: backend/test.db
Size: 19M
```

### Step 5: Verify backups on host machine

```bash
# Exit the container
exit

# Check backups on your host machine
ls -lh ./backups/
```

The backups are persisted on your host machine! 🎉

---

## Method 2: Using Go CLI Tool

### Build the backup tool

```bash
# On your host machine
cd backend
go build -o backup-tool ./cmd/backup
```

### Copy to Docker container

```bash
docker cp backup-tool <container_id>:/app/
```

### Use inside container

```bash
docker exec -it <container_id> /bin/bash

# Inside container
cd /app
./backup-tool backup          # Create backup
./backup-tool list            # List backups
./backup-tool restore <file>  # Restore from backup
```

---

## Method 3: Using Docker Exec (No Shell Access Needed)

### Create a backup

```bash
docker exec <container_id> /app/scripts/backup.sh
```

### List backups

```bash
docker exec <container_id> ls -lh /app/backups
```

### Restore from backup

```bash
docker exec <container_id> /app/scripts/restore.sh backups/featureplus_2025-10-29_13-45.db
```

---

## Method 4: Using docker-compose

If you're using docker-compose:

```bash
# Create backup
docker-compose exec featureplus /app/scripts/backup.sh

# List backups
docker-compose exec featureplus ls -lh /app/backups

# Restore
docker-compose exec featureplus /app/scripts/restore.sh backups/featureplus_2025-10-29_13-45.db
```

---

## Testing Workflow

### Complete Test Scenario

1. **Create some data** (login, create a release, etc.)
2. **Create a backup**:
   ```bash
   docker exec <container_id> /app/scripts/backup.sh
   ```

3. **Make changes** (create more data, modify something)

4. **Create another backup**:
   ```bash
   docker exec <container_id> /app/scripts/backup.sh
   ```

5. **Restore to first backup**:
   ```bash
   docker exec <container_id> /app/scripts/restore.sh backups/featureplus_2025-10-29_13-45.db
   ```

6. **Verify** the data is back to the first state

7. **Check host machine** - backups should be in `./backups/` directory

---

## Automated Backup Schedule (Bonus)

You can set up automated backups using cron inside the container or from the host:

### From Host (Recommended)

```bash
# Add to your crontab
# Backup every day at 2 AM
0 2 * * * docker exec <container_id> /app/scripts/backup.sh >> /var/log/featureplus-backup.log 2>&1
```

### Using Docker Compose

Add a backup service to docker-compose.yml:

```yaml
services:
  backup:
    image: featureplus:dev
    volumes:
      - ./backups:/app/backups
      - ./backend/test.db:/app/test.db
    entrypoint: ["/bin/sh", "-c"]
    command: ["while true; do /app/scripts/backup.sh; sleep 86400; done"]
```

---

## Troubleshooting

### Scripts not executable

```bash
docker exec <container_id> chmod +x /app/scripts/backup.sh /app/scripts/restore.sh
```

### Backups directory doesn't exist

The scripts create it automatically, but if needed:
```bash
docker exec <container_id> mkdir -p /app/backups
```

### Permission issues

```bash
docker exec <container_id> chown -R root:root /app/backups
docker exec <container_id> chmod -R 755 /app/backups
```

---

## Verification

After backup/restore, verify the database:

```bash
# Check database file size
docker exec <container_id> ls -lh /app/test.db

# Check backup files
docker exec <container_id> ls -lh /app/backups/

# View logs
docker logs <container_id> | grep -i backup
```
