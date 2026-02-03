# Wallet Service

A high-performance, reusable multi-asset wallet service with double-entry bookkeeping, powered by TigerBeetle distributed database.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [TigerBeetle Cluster Setup](#tigerbeetle-cluster-setup)
  - [What is TigerBeetle?](#what-is-tigerbeetle)
  - [Architecture](#architecture)
  - [Custom Docker Image](#custom-docker-image)
  - [Installation Steps](#installation-steps)
  - [Configuration](#configuration)
  - [Running the Cluster](#running-the-cluster)
  - [Health Monitoring](#health-monitoring)
  - [Verification](#verification)
  - [Troubleshooting](#troubleshooting)
- [Makefile Commands Reference](#makefile-commands-reference)
- [Advanced Operations](#advanced-operations)
- [Production Considerations](#production-considerations)

---

## Overview

This wallet service provides a robust, distributed ledger system for managing multi-asset wallets with strict double-entry bookkeeping guarantees. It leverages TigerBeetle, a purpose-built distributed financial accounting database, to ensure ACID compliance and high-performance transaction processing.

---

## Prerequisites

Before setting up the TigerBeetle cluster, ensure you have the following installed:

- **Docker**: Version 20.10 or higher
- **Docker Compose**: Version 2.0 or higher
- **Make**: GNU Make utility
- **Git**: For cloning the repository

### System Requirements

- **Memory**: Minimum 4GB RAM (8GB recommended for production)
- **Disk Space**: At least 5GB free space
- **Network**: Ports 3000, 3001, 3002 available
- **OS**: Linux, macOS, or Windows with WSL2

---

## TigerBeetle Cluster Setup

### What is TigerBeetle?

TigerBeetle is a distributed financial accounting database designed for mission-critical safety and performance. It provides:

- **ACID Guarantees**: Strict consistency for financial transactions
- **High Performance**: Millions of transactions per second
- **Fault Tolerance**: Distributed consensus with automatic failover
- **Double-Entry Bookkeeping**: Built-in support for accounting principles
- **Zero Data Loss**: Replicated storage with checksums

### Architecture

The cluster consists of **3 replicas** for high availability:

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  TigerBeetle-0  │────▶│  TigerBeetle-1  │────▶│  TigerBeetle-2  │
│  172.20.0.10    │     │  172.20.0.11    │     │  172.20.0.12    │
│  Port: 3000     │     │  Port: 3001     │     │  Port: 3002     │
│  (PRIMARY)      │     │  (BACKUP)       │     │  (BACKUP)       │
└─────────────────┘     └─────────────────┘     └─────────────────┘
         │                       │                       │
         └───────────────────────┴───────────────────────┘
                    Custom Bridge Network
                    Subnet: 172.20.0.0/24
```

**Key Features:**
- **Cluster ID**: 0 (for testing/development)
- **Replica Count**: 3
- **Consensus Protocol**: Viewstamped Replication (VSR)
- **Cache Size**: 512MiB per replica
- **Data Persistence**: Docker volumes
- **Health Monitoring**: Built-in healthchecks using netcat

---

## Custom Docker Image

This project uses a **custom TigerBeetle Docker image** that extends the official image with additional tools for health monitoring.

### Why a Custom Image?

The official TigerBeetle image is minimal and doesn't include network utilities like `netcat` (nc), which are needed for Docker healthchecks. Our custom image adds these tools while maintaining the same TigerBeetle functionality.

### Custom Image Details

**Location**: `tigerbeetle/Dockerfile`

```dockerfile
FROM ghcr.io/tigerbeetle/tigerbeetle:latest

# Install netcat-openbsd for healthchecks
RUN apk add --no-cache netcat-openbsd

LABEL maintainer="wallet-service"
LABEL description="TigerBeetle with healthcheck support"
LABEL version="0.16.70-custom"
```

**What's Added:**
- `netcat-openbsd`: Network utility for port connectivity checks
- Custom labels for image identification

**Image Name**: `tigerbeetle-custom:latest`

### Building the Custom Image

The custom image is automatically built when you run `make compose-tb-init`, but you can also build it manually:

```bash
make compose-tb-build
```

This command:
1. Pulls the latest official TigerBeetle image
2. Installs netcat-openbsd package
3. Tags the image as `tigerbeetle-custom:latest`

---

## Installation Steps

### Step 1: Clone the Repository

```bash
git clone <repository-url>
cd wallet-service
```

### Step 2: Review Environment Configuration

Check the TigerBeetle configuration file:

```bash
cat env/tigerbittle.env
```

**Default Configuration:**
```bash
# TigerBeetle Auth Token
LU_CFG_TIGERBEETLE_AUTH_TOKEN=CfDj7kdpld7BsUpXzQYiOfLdnateLkdBysPEWnvc

# Cluster Addresses
LU_CFG_TIGERBEETLE_ADDRESSES=172.20.0.10:3000,172.20.0.11:3001,172.20.0.12:3002

# Replica 0 Configuration
LU_CFG_TIGERBEETLE_0_CONTAINER_NAME=tigerbeetle-0
LU_CFG_TIGERBEETLE_0_PORT=3000
LU_CFG_TIGERBEETLE_0_IP=172.20.0.10

# Replica 1 Configuration
LU_CFG_TIGERBEETLE_1_CONTAINER_NAME=tigerbeetle-1
LU_CFG_TIGERBEETLE_1_PORT=3001
LU_CFG_TIGERBEETLE_1_IP=172.20.0.11

# Replica 2 Configuration
LU_CFG_TIGERBEETLE_2_CONTAINER_NAME=tigerbeetle-2
LU_CFG_TIGERBEETLE_2_PORT=3002
LU_CFG_TIGERBEETLE_2_IP=172.20.0.12

# Network Configuration
LU_CFG_TIGERBEETLE_SUBNET=172.20.0.0/24
```

> **Note**: You can modify these values if needed, but ensure consistency across all replicas.

### Step 3: Initialize the TigerBeetle Cluster

For first-time setup, use the initialization command:

```bash
make compose-tb-init
```

This command will:
1. **Build** the custom TigerBeetle image with healthcheck support
2. **Stop** and remove any existing TigerBeetle containers
3. **Remove** old data volumes
4. **Format** new data files for all 3 replicas
5. **Start** the cluster with healthchecks enabled

**Expected Output:**
```
Building custom TigerBeetle image...
[+] Building 4.7s (6/6) FINISHED
 => [1/2] FROM ghcr.io/tigerbeetle/tigerbeetle:latest
 => [2/2] RUN apk add --no-cache netcat-openbsd
 => exporting to image
Custom TigerBeetle image built successfully.

Formatting TigerBeetle data files...
info(io): creating "0_0.tigerbeetle"...
info(io): allocating 1.06298828125GiB...
info(main): 0: formatted: cluster=0 replica_count=3
...
TigerBeetle data files formatted successfully.
✔ Container tigerbeetle-0 Created
✔ Container tigerbeetle-1 Created
✔ Container tigerbeetle-2 Created
```

---

## Configuration

### Docker Compose Configuration

The cluster is defined in `docker-compose.tb.yaml`:

**Key Configuration Points:**

1. **Image**: `ghcr.io/tigerbeetle/tigerbeetle:latest`
2. **Command**: `start --addresses=<IP1>:3000,<IP2>:3001,<IP3>:3002 --cache-grid=512MiB /data/0_X.tigerbeetle`
3. **Volumes**: Persistent storage for each replica
4. **Network**: Custom bridge network with static IPs
5. **Security**: `seccomp=unconfined` for io_uring support (Docker 25.0.0+)

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `LU_CFG_TIGERBEETLE_AUTH_TOKEN` | Authentication token | `CfDj7kdpld7BsUpXzQYiOfLdnateLkdBysPEWnvc` |
| `LU_CFG_TIGERBEETLE_ADDRESSES` | Comma-separated cluster addresses | `172.20.0.10:3000,172.20.0.11:3001,172.20.0.12:3002` |
| `LU_CFG_TIGERBEETLE_X_IP` | IP address for replica X | `172.20.0.1X` |
| `LU_CFG_TIGERBEETLE_X_PORT` | Port for replica X | `300X` |
| `LU_CFG_TIGERBEETLE_SUBNET` | Docker network subnet | `172.20.0.0/24` |

---

## Running the Cluster

### Quick Start Commands

#### First Time Setup
```bash
# Complete initialization (format + start)
make compose-tb-init
```

#### Start Existing Cluster
```bash
# Start the cluster (data files must be formatted first)
make compose-tb-up
```

#### Stop the Cluster
```bash
# Stop all containers (preserves data)
make compose-tb-down
```

#### Reset the Cluster
```bash
# Stop and remove all data
make compose-tb-down-volumes

# Format fresh data files
make compose-tb-format

# Start the cluster
make compose-tb-up
```

### Manual Step-by-Step Process

If you prefer to run commands manually:

#### 1. Format Data Files (First Time Only)

```bash
# Format replica 0
docker compose -f docker-compose.tb.yaml --env-file ./env/tigerbittle.env run --rm \
  tigerbeetle-0 format --cluster=0 --replica=0 --replica-count=3 /data/0_0.tigerbeetle

# Format replica 1
docker compose -f docker-compose.tb.yaml --env-file ./env/tigerbittle.env run --rm \
  tigerbeetle-1 format --cluster=0 --replica=1 --replica-count=3 /data/0_1.tigerbeetle

# Format replica 2
docker compose -f docker-compose.tb.yaml --env-file ./env/tigerbittle.env run --rm \
  tigerbeetle-2 format --cluster=0 --replica=2 --replica-count=3 /data/0_2.tigerbeetle
```

#### 2. Stop Any Running Containers

```bash
docker compose -f docker-compose.tb.yaml --env-file ./env/tigerbittle.env down
```

#### 3. Start the Cluster

```bash
docker compose -f docker-compose.tb.yaml --env-file ./env/tigerbittle.env up -d
```

---

## Health Monitoring

The TigerBeetle cluster includes built-in health monitoring using Docker healthchecks. Each container is continuously monitored to ensure it's running correctly.

### Healthcheck Configuration

Each replica has the following healthcheck settings:

```yaml
healthcheck:
  test: ["CMD", "nc", "-z", "<replica-ip>", "<port>"]
  interval: 10s      # Check every 10 seconds
  timeout: 5s        # Fail if check takes longer than 5 seconds
  retries: 3         # Mark unhealthy after 3 consecutive failures
  start_period: 10s  # Grace period during container startup
```

**Replica-Specific Healthchecks:**
- **tigerbeetle-0**: Checks connectivity to `172.20.0.10:3000`
- **tigerbeetle-1**: Checks connectivity to `172.20.0.11:3001`
- **tigerbeetle-2**: Checks connectivity to `172.20.0.12:3002`

> **Note**: The healthcheck uses the container's specific IP address (not `localhost`) because TigerBeetle binds to its assigned IP address for cluster communication.

### Check Cluster Health Status

Use the dedicated health status command:

```bash
make compose-tb-status
```

**Expected Output:**
```
=== TigerBeetle Cluster Status ===
NAMES           STATUS                   PORTS
tigerbeetle-2   Up 2 minutes (healthy)   0.0.0.0:3002->3002/tcp
tigerbeetle-1   Up 2 minutes (healthy)   0.0.0.0:3001->3001/tcp
tigerbeetle-0   Up 2 minutes (healthy)   0.0.0.0:3000->3000/tcp

=== Detailed Health Information ===
--- tigerbeetle-0 ---
Status: healthy
Last Check: Connection to 172.20.0.10 3000 port [tcp/*] succeeded!

--- tigerbeetle-1 ---
Status: healthy
Last Check: Connection to 172.20.0.11 3001 port [tcp/*] succeeded!

--- tigerbeetle-2 ---
Status: healthy
Last Check: Connection to 172.20.0.12 3002 port [tcp/*] succeeded!
```

### Health Status Indicators

- **`(healthy)`**: Container is running and healthcheck is passing
- **`(health: starting)`**: Container is in the startup grace period (first 10 seconds)
- **`(unhealthy)`**: Healthcheck has failed 3 consecutive times
- **No health status**: Healthcheck is not configured (shouldn't happen with this setup)

### Manual Health Inspection

For detailed health information on a specific container:

```bash
docker inspect tigerbeetle-0 --format='{{json .State.Health}}' | jq
```

This shows:
- Current health status
- Number of consecutive failures
- Complete log of recent health checks with timestamps and outputs

---

## Verification

### Check Container Status

```bash
docker ps | grep tigerbeetle
```

**Expected Output:**
```
CONTAINER ID   IMAGE                                    STATUS          PORTS
e8e24f6801c0   ghcr.io/tigerbeetle/tigerbeetle:latest   Up 2 minutes    0.0.0.0:3000->3000/tcp
667600e32c45   ghcr.io/tigerbeetle/tigerbeetle:latest   Up 2 minutes    0.0.0.0:3001->3001/tcp
edb1f78e70dd   ghcr.io/tigerbeetle/tigerbeetle:latest   Up 2 minutes    0.0.0.0:3002->3002/tcp
```

### Check Cluster Logs

#### View All Replica Logs
```bash
# Replica 0 (Primary)
docker logs tigerbeetle-0

# Replica 1 (Backup)
docker logs tigerbeetle-1

# Replica 2 (Backup)
docker logs tigerbeetle-2
```

#### Check Connection Status
```bash
docker logs tigerbeetle-0 2>&1 | grep -E "(listening|connected|primary|backup)"
```

**Healthy Output:**
```
info(main): 0: cluster=0: listening on 172.20.0.10:3000
info(message_bus): 0: on_connect: connected to=1
info(message_bus): 0: on_connect: connected to=2
info(replica): 0N: transition_to_normal_from_view_change_status: view=3..3 primary
info(clock): 0: synchronized: accuracy=0ns
```

### Verify Network Connectivity

```bash
# Check if replicas can communicate
docker exec tigerbeetle-0 ping -c 2 172.20.0.11
docker exec tigerbeetle-0 ping -c 2 172.20.0.12
```

### Check Docker Compose Status

```bash
docker compose -f docker-compose.tb.yaml --env-file ./env/tigerbittle.env ps
```

**Expected Output:**
```
NAME            SERVICE         STATUS          PORTS
tigerbeetle-0   tigerbeetle-0   Up 5 minutes    0.0.0.0:3000->3000/tcp
tigerbeetle-1   tigerbeetle-1   Up 5 minutes    0.0.0.0:3001->3001/tcp
tigerbeetle-2   tigerbeetle-2   Up 5 minutes    0.0.0.0:3002->3002/tcp
```

---

## Troubleshooting

### Common Issues and Solutions

#### Issue 1: Containers Keep Restarting

**Symptoms:**
```bash
docker ps -a | grep tigerbeetle
# Shows: Restarting (1) X seconds ago
```

**Possible Causes & Solutions:**

1. **Data files not formatted**
   ```bash
   # Solution: Format the data files
   make compose-tb-format
   ```

2. **Invalid addresses in configuration**
   ```bash
   # Check logs for "invalid IPv4 or IPv6 address"
   docker logs tigerbeetle-0

   # Solution: Verify env/tigerbittle.env has correct IP addresses
   ```

3. **Port conflicts**
   ```bash
   # Check if ports are already in use
   sudo netstat -tulpn | grep -E '3000|3001|3002'

   # Solution: Stop conflicting services or change ports in env/tigerbittle.env
   ```

#### Issue 2: Permission Denied Errors

**Symptoms:**
```
error: PermissionDenied
```

**Solution:**
This occurs with Docker 25.0.0+ which blocks io_uring by default. The `security_opt: seccomp=unconfined` is already configured in `docker-compose.tb.yaml`.

If still occurring:
```bash
# Verify security option is present
grep -A 2 "security_opt" docker-compose.tb.yaml
```

#### Issue 3: Out of Memory (OOM) Errors

**Symptoms:**
```
exited with code 137
```

**Solutions:**

1. **Increase Docker memory limit** (Docker Desktop):
   - Settings → Resources → Memory → Increase to 4GB+

2. **Reduce cache size** (Development only):
   ```yaml
   # In docker-compose.tb.yaml, change:
   --cache-grid=512MiB
   # to:
   --cache-grid=256MiB
   ```

#### Issue 4: Replicas Not Connecting

**Symptoms:**
```
warning(message_bus): 0: on_connect: error to=1 error.ConnectionRefused
```

**Solutions:**

1. **Check network configuration**
   ```bash
   docker network inspect wallet-service_tigerbeetle-net
   ```

2. **Verify all containers are running**
   ```bash
   docker ps | grep tigerbeetle
   ```

3. **Restart the cluster**
   ```bash
   make compose-tb-down
   make compose-tb-up
   ```

#### Issue 5: Data Corruption

**Symptoms:**
```
error: data file is corrupted
```

**Solution:**
```bash
# Complete reset (WARNING: Deletes all data)
make compose-tb-down-volumes
make compose-tb-format
make compose-tb-up
```

#### Issue 6: Healthcheck Failures

**Symptoms:**
```bash
make compose-tb-status
# Shows: (unhealthy) status
```

**Possible Causes & Solutions:**

1. **TigerBeetle not listening on expected port**
   ```bash
   # Check if TigerBeetle is actually running
   docker logs tigerbeetle-0 | grep "listening on"

   # Should show: "listening on 172.20.0.10:3000"
   ```

2. **Netcat not installed in container**
   ```bash
   # Verify custom image is being used
   docker inspect tigerbeetle-0 | grep Image

   # Should show: "tigerbeetle-custom:latest"

   # If not, rebuild the image
   make compose-tb-build
   make compose-tb-down
   make compose-tb-up
   ```

3. **Wrong IP address in healthcheck**
   ```bash
   # Verify healthcheck configuration
   docker inspect tigerbeetle-0 --format='{{.Config.Healthcheck.Test}}'

   # Should show: [CMD nc -z 172.20.0.10 3000]
   ```

4. **Container still in startup period**
   ```bash
   # Wait 10-15 seconds after container starts
   # Healthcheck has a 10s start_period grace period
   sleep 15 && make compose-tb-status
   ```

### Viewing Detailed Logs

```bash
# Using Makefile (recommended)
make compose-tb-logs                    # All containers
make compose-tb-logs CONTAINER=tigerbeetle-0  # Specific container

# Using Docker directly
docker logs -f tigerbeetle-0            # Follow logs in real-time
docker logs --tail 100 tigerbeetle-0    # View last 100 lines
docker logs -t tigerbeetle-0            # View logs with timestamps

# View logs for all replicas
docker compose -f docker-compose.tb.yaml --env-file ./env/tigerbittle.env logs -f
```

---

## Makefile Commands Reference

### TigerBeetle Cluster Commands

| Command | Description |
|---------|-------------|
| `make compose-tb-init` | **Complete initialization**: Build custom image, stop, remove volumes, format, and start cluster |
| `make compose-tb-build` | **Build custom image**: Build the TigerBeetle image with healthcheck support |
| `make compose-tb-format` | **Format data files**: Initialize TigerBeetle data files for all replicas |
| `make compose-tb-up` | **Start cluster**: Launch all TigerBeetle containers |
| `make compose-tb-down` | **Stop cluster**: Stop containers (preserves data) |
| `make compose-tb-down-volumes` | **Stop and clean**: Stop containers and remove all data volumes |
| `make compose-tb-status` | **Health status**: Show cluster health status and detailed healthcheck information |
| `make compose-tb-logs` | **View logs**: Show logs for all containers (or use `CONTAINER=tigerbeetle-0` for specific container) |

### Usage Examples

```bash
# First time setup
make compose-tb-init

# Check cluster health
make compose-tb-status

# Daily operations
make compose-tb-up      # Start
make compose-tb-down    # Stop

# View logs
make compose-tb-logs                    # All containers
make compose-tb-logs CONTAINER=tigerbeetle-0  # Specific container

# Rebuild custom image (after Dockerfile changes)
make compose-tb-build

# Complete reset
make compose-tb-down-volumes
make compose-tb-init

# Format only (if volumes exist)
make compose-tb-format
```

---

## Advanced Operations

### Accessing the TigerBeetle REPL

TigerBeetle provides a REPL (Read-Eval-Print Loop) for interactive testing:

```bash
# Connect to replica 0
docker exec -it tigerbeetle-0 /tigerbeetle repl --cluster=0 --addresses=172.20.0.10:3000
```

**Example REPL Commands:**
```
# Create an account
> create_accounts id=1 code=USD ledger=1 flags=0

# Create a transfer
> create_transfers id=1 debit_account_id=1 credit_account_id=2 amount=100 ledger=1 code=1

# Lookup accounts
> lookup_accounts id=1
```

### Backup and Restore

#### Backup Data Volumes

```bash
# Create backup directory
mkdir -p backups/tigerbeetle

# Backup all volumes
docker run --rm \
  -v wallet-service_tigerbeetle-data-0:/data \
  -v $(pwd)/backups/tigerbeetle:/backup \
  alpine tar czf /backup/replica-0-$(date +%Y%m%d-%H%M%S).tar.gz -C /data .

docker run --rm \
  -v wallet-service_tigerbeetle-data-1:/data \
  -v $(pwd)/backups/tigerbeetle:/backup \
  alpine tar czf /backup/replica-1-$(date +%Y%m%d-%H%M%S).tar.gz -C /data .

docker run --rm \
  -v wallet-service_tigerbeetle-data-2:/data \
  -v $(pwd)/backups/tigerbeetle:/backup \
  alpine tar czf /backup/replica-2-$(date +%Y%m%d-%H%M%S).tar.gz -C /data .
```

#### Restore from Backup

```bash
# Stop the cluster
make compose-tb-down

# Restore replica 0
docker run --rm \
  -v wallet-service_tigerbeetle-data-0:/data \
  -v $(pwd)/backups/tigerbeetle:/backup \
  alpine sh -c "cd /data && tar xzf /backup/replica-0-TIMESTAMP.tar.gz"

# Repeat for other replicas...

# Start the cluster
make compose-tb-up
```

### Scaling Considerations

The current setup uses a **3-replica cluster** for fault tolerance:

- **Quorum**: 2 out of 3 replicas must be available
- **Fault Tolerance**: Can survive 1 replica failure
- **Performance**: Distributes read load across replicas

**To add more replicas** (requires reformatting):
1. Update `env/tigerbittle.env` with new replica configuration
2. Add new service in `docker-compose.tb.yaml`
3. Reformat all data files with new `--replica-count`
4. Restart the cluster

### Monitoring

#### Health Check Script

Create a simple health check script:

```bash
#!/bin/bash
# health-check.sh

for port in 3000 3001 3002; do
  if nc -z localhost $port 2>/dev/null; then
    echo "✓ Replica on port $port is reachable"
  else
    echo "✗ Replica on port $port is NOT reachable"
  fi
done
```

```bash
chmod +x health-check.sh
./health-check.sh
```

#### Log Monitoring

```bash
# Monitor for errors
docker compose -f docker-compose.tb.yaml --env-file ./env/tigerbittle.env logs -f | grep -i error

# Monitor for warnings
docker compose -f docker-compose.tb.yaml --env-file ./env/tigerbittle.env logs -f | grep -i warning

# Monitor cluster state changes
docker compose -f docker-compose.tb.yaml --env-file ./env/tigerbittle.env logs -f | grep -i "transition_to"
```

### Performance Tuning

#### Cache Grid Size

Adjust cache size based on available memory:

```yaml
# In docker-compose.tb.yaml
# Development: 256MiB - 512MiB
# Production: 1GiB - 4GiB
--cache-grid=512MiB
```

#### Memory Locking (Production)

For production environments, enable memory locking to prevent swapping:

```yaml
# Add to each service in docker-compose.tb.yaml
cap_add:
  - IPC_LOCK
```

Or configure Docker daemon (`/etc/docker/daemon.json`):
```json
{
  "default-ulimits": {
    "memlock": {
      "Hard": -1,
      "Name": "memlock",
      "Soft": -1
    }
  }
}
```

---

## Production Considerations

### Security

1. **Change Default Auth Token**
   ```bash
   # Generate a secure token
   openssl rand -base64 32

   # Update env/tigerbittle.env
   LU_CFG_TIGERBEETLE_AUTH_TOKEN=<your-secure-token>
   ```

2. **Use Non-Zero Cluster ID**
   ```bash
   # Cluster ID 0 is reserved for testing
   # For production, generate a random cluster ID
   # Update format commands in Makefile to use --cluster=<random-id>
   ```

3. **Network Isolation**
   - Use Docker networks to isolate TigerBeetle from public access
   - Only expose ports to application containers
   - Consider using Docker secrets for sensitive configuration

### High Availability

1. **Replica Distribution**
   - Deploy replicas across different physical hosts
   - Use Docker Swarm or Kubernetes for orchestration
   - Ensure network latency between replicas is minimal

2. **Monitoring and Alerting**
   - Monitor replica health and connectivity
   - Set up alerts for replica failures
   - Track transaction throughput and latency

3. **Backup Strategy**
   - Automated daily backups of all replicas
   - Store backups in separate location
   - Test restore procedures regularly

### Resource Allocation

**Minimum Production Requirements per Replica:**
- **CPU**: 2 cores
- **Memory**: 4GB RAM
- **Disk**: 50GB SSD (NVMe recommended)
- **Network**: 1Gbps

**Recommended Production Configuration:**
- **CPU**: 4+ cores
- **Memory**: 8GB+ RAM
- **Disk**: 100GB+ NVMe SSD
- **Network**: 10Gbps

### Disaster Recovery

1. **Regular Backups**
   ```bash
   # Automate with cron
   0 2 * * * /path/to/backup-script.sh
   ```

2. **Recovery Procedures**
   - Document step-by-step recovery process
   - Test recovery in staging environment
   - Maintain runbooks for common failure scenarios

3. **Data Retention**
   - Define retention policy for backups
   - Archive old backups to cold storage
   - Comply with regulatory requirements

---

## Quick Reference

### Essential Commands Cheat Sheet

```bash
# SETUP
make compose-tb-init                    # First time setup

# START/STOP
make compose-tb-up                      # Start cluster
make compose-tb-down                    # Stop cluster

# MAINTENANCE
make compose-tb-format                  # Format data files
make compose-tb-down-volumes            # Remove all data

# MONITORING
docker ps | grep tigerbeetle            # Check status
docker logs tigerbeetle-0               # View logs
docker stats tigerbeetle-0              # Resource usage

# DEBUGGING
docker exec -it tigerbeetle-0 sh        # Shell access
docker inspect tigerbeetle-0            # Container details
docker network inspect wallet-service_tigerbeetle-net  # Network info
```

### Port Mapping

| Replica | Internal Port | External Port | IP Address    |
|---------|---------------|---------------|---------------|
| 0       | 3000          | 3000          | 172.20.0.10   |
| 1       | 3001          | 3001          | 172.20.0.11   |
| 2       | 3002          | 3002          | 172.20.0.12   |

### File Locations

| Item | Location |
|------|----------|
| Docker Compose | `docker-compose.tb.yaml` |
| Environment Config | `env/tigerbittle.env` |
| Makefile | `Makefile` |
| Data Volumes | Docker managed volumes |
| Logs | `docker logs <container>` |

---

## Additional Resources

### Official Documentation

- **TigerBeetle Docs**: https://docs.tigerbeetle.com/
- **Docker Documentation**: https://docs.docker.com/
- **Docker Compose**: https://docs.docker.com/compose/

### TigerBeetle Resources

- **GitHub Repository**: https://github.com/tigerbeetle/tigerbeetle
- **Docker Image**: https://github.com/tigerbeetle/tigerbeetle/pkgs/container/tigerbeetle
- **Community Slack**: https://slack.tigerbeetle.com/

### Support

For issues specific to this wallet service:
1. Check the [Troubleshooting](#troubleshooting) section
2. Review TigerBeetle logs for errors
3. Consult TigerBeetle official documentation
4. Open an issue in the project repository

---

## License

[Your License Here]

---

## Contributing

[Your Contributing Guidelines Here]

---

**Last Updated**: 2026-02-03
**TigerBeetle Version**: 0.16.70
**Docker Compose Version**: 2.x
