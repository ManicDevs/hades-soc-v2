# VMware Pro Development Environment Playbook

Complete guide for running Hades-V2 in VMware Workstation Pro VMs.

---

## 1. Recommended VM Templates

### Base VM Specs

| Role | vCPUs | RAM | Disk | OS |
|------|-------|-----|------|-----|
| Dev | 4 | 8GB | 80GB | Ubuntu 22.04 LTS |
| Test | 4 | 6GB | 60GB | Ubuntu 22.04 LTS |
| PostgreSQL | 2 | 4GB | 50GB | Ubuntu 22.04 LTS |
| CI Runner | 2 | 4GB | 40GB | Ubuntu 22.04 LTS |

---

## 2. VM Network Configuration

### Network Types in VMware

| Network | Purpose | Use Case |
|---------|---------|----------|
| NAT (VMnet8) | Internet access, host communication | Default dev VMs |
| Bridged | Direct network access | VMs requiring external access |
| Host-Only (VMnet1) | Isolated dev network | Internal testing |
| Custom (VMnet10) | Multi-tier architecture | Hades microservices |

### Recommended Setup

```
┌─────────────────────────────────────────────────────────┐
│  Host Machine                                            │
│  ┌──────────────┐                                       │
│  │  Hades-V2    │  ← Working directory                  │
│  │  Repository  │  ← Shared via VMware Shared Folders   │
│  └──────────────┘                                       │
│         │                                               │
│         ▼ Shared Folder                                 │
│  ┌──────────────┐     ┌──────────────┐                │
│  │  Dev VM      │────▶│  DB VM       │                │
│  │  :8080       │     │  PostgreSQL  │                │
│  │  :3000       │     │  :5432       │                │
│  └──────────────┘     └──────────────┘                │
│         │                                               │
│         ▼ NAT (VMnet8)                                  │
│  ┌──────────────┐     ┌──────────────┐                │
│  │  Test VM    │─────▶│  Redis VM   │                │
│  │  :8081      │     │  :6379      │                │
│  └──────────────┘     └──────────────┘                │
└─────────────────────────────────────────────────────────┘
```

---

## 3. Quick Start

### 3.1 Create Dev VM

```bash
# 1. Download Ubuntu 22.04 LTS Server ISO
#    https://releases.ubuntu.com/22.04/

# 2. Create new VM
vmrun -t ws6 createUbuntuDev /path/to/vm.vmx \
  -m 8192 -c 4 -d 80GB

# 3. Mount ISO and install
vmrun start /path/to/vm.vmx
```

### 3.2 Install Dependencies in VM

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Go 1.21+
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

# Install Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Install PostgreSQL client
sudo apt install -y postgresql-client redis-tools

# Clone repository (via shared folder)
cd /mnt/hgfs/hades
go mod download
```

---

## 4. Shared Folders Setup

### Host Configuration (VMware Workstation)

1. **VM → Settings → Options → Shared Folders**
2. Enable "Always Enabled"
3. Add host folder: `/home/cerberus/Desktop/hades`
4. Map to: `\\vmware-host\Shared Folders\hades`

### VM Mount (Linux Guest)

```bash
# Create mount point
sudo mkdir -p /mnt/hades

# Add to /etc/fstab for auto-mount
open-vm-tools  # For automatic mounting

# Manual mount
sudo mount -t vmhgfs .host:/shared /mnt/hades

# Verify
ls -la /mnt/hades
```

---

## 5. Development Workflow

### 5.1 Edit on Host, Build in VM

```bash
# On HOST (Windows/macOS)
# Edit code in IDE → Files available instantly via shared folder

# On VM
cd /mnt/hgfs/hades
go build -o bin/hades ./cmd/hades
./bin/hades version

# Run tests
go test ./internal/... -v -short
```

### 5.2 Port Forwarding (NAT)

Configure in `~/.vmware/ui-draggable-<id>.lnx` or VMware NAT config:

```
# /etc/vmware/networking
# Add port forwards for NAT network
5005 -> 127.0.0.1:5005    # Delve debugger
8080 -> 127.0.0.1:8080    # Hades API
3000 -> 127.0.0.1:3000    # Frontend
5432 -> 127.0.0.1:5432    # PostgreSQL
```

### 5.3 SSH Access

```bash
# Generate SSH key in VM
ssh-keygen -t ed25519 -C "dev-vm"

# Add to ~/.ssh/config for easy access
Host dev-vm
    HostName 192.168.217.128
    User cerberus
    IdentityFile ~/.ssh/dev-vm
    ForwardAgent yes

# Connect
ssh dev-vm
```

---

## 6. Snapshot Strategy

### Snapshots for Hades Development

```bash
# Create snapshot before major changes
vmrun snapshot /path/to/vm.vmx "clean-install"

# After installing dependencies
vmrun snapshot /path/to/vm.vmx "dev-ready"

# Before database schema changes
vmrun snapshot /path/to/vm.vmx "pre-schema-migration"

# Restore if needed
vmrun revertToSnapshot /path/to/vm.vmx "clean-install"
```

### Snapshot Schedule

| Snapshot | Trigger | Keep For |
|----------|---------|----------|
| Base | Fresh install | Permanent |
| Dev-ready | Dependencies installed | 1 month |
| Pre-migration | Before DB changes | Until migration verified |
| Weekly | Weekly checkpoint | 4 weeks |

---

## 7. Multi-VM Architecture

### 7.1 Create Network

```bash
# In VMware Workstation:
# Edit → Virtual Network Editor → Add Network (VMnet10)
# Type: Host-only
# Subnet: 192.168.100.0/24
```

### 7.2 VM Configurations

**Dev VM** (`dev-hades`):
```yaml
Network: Host-only (VMnet10)
IP: 192.168.100.10
Services: hades-api (:8080), frontend (:3000)
```

**Database VM** (`db-postgres`):
```yaml
Network: Host-only (VMnet10)
IP: 192.168.100.20
Services: PostgreSQL (:5432)
```

**Redis VM** (`cache-redis`):
```yaml
Network: Host-only (VMnet10)
IP: 192.168.100.30
Services: Redis (:6379)
```

### 7.3 Hades Configuration

```yaml
# config.yaml for multi-VM setup
database:
  host: 192.168.100.20
  port: 5432
  username: hades
  password: ${HADES_DB_PASSWORD}
  name: hades_prod

redis:
  host: 192.168.100.30
  port: 6379

server:
  host: 0.0.0.0
  port: 8080
```

---

## 8. Performance Optimization

### VM Settings

```
# .vmx file optimizations
sched.numvcpus = "4"
sched.cpu.units = "nanosec"
sched.latencySensitivity = "medium"

# Memory
prefvmx.useRecommendedLockedMemSize = "TRUE"
prefvmx.minMemLocked = "8192"

# Disk I/O
diskLib.dataCacheMaxSize = "65536"
diskLib.dataCacheMaxReadCacheSize = "32768"
scsi0:0.deviceType = "scsi-hardDisk"
```

### Within VM

```bash
# Enable swappiness for development
echo 10 | sudo tee /proc/sys/vm/swappiness

# Increase inotify limits
echo fs.inotify.max_user_watches = 524288 | sudo tee -a /etc/sysctl.conf

# Disable transparent HugePages
echo never | sudo tee /sys/kernel/mm/transparent_hugepage/enabled
```

---

## 9. Backup & Recovery

### Backup Script

```bash
#!/bin/bash
# backup-vms.sh

VM_DIR="/path/to/vmware/vms"
BACKUP_DIR="/external/disk/vm-backups"
DATE=$(date +%Y%m%d)

for vm in dev-hades test-vm db-postgres; do
    echo "Backing up $vm..."
    vmrun stop "/path/to/$vm.vmx" soft
    rsync -avz "/path/to/$vm.vmdk" "$BACKUP_DIR/${vm}-${DATE}.vmdk"
    vmrun start "/path/to/$vm.vmx"
done
```

### Recovery

```bash
# In case of corruption
vmrun stop /path/to/vm.vmx
rm -rf /path/to/vm.vmdk
cp /path/to/backup/latest.vmdk /path/to/vm.vmdk
vmrun start /path/to/vm.vmx
```

---

## 10. Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| Shared folders not visible | Install `open-vm-tools` in VM |
| Slow build | Increase VM RAM, enable disk caching |
| Network not working | Check VMware NAT service |
| Disk space full | Expand VMDK or cleanup `/var/cache` |

### Commands

```bash
# Check VMware services (Host)
services.msc  # Windows
vmware-networks --start  # Linux

# Network reset
vmrun restartService vmware-networks

# Reinstall tools in VM
sudo vmware-toolbox-cmd timesync enable
sudo vmware-toolbox-cmd vmhgfsfs .host:/shared /mnt/hades
```

---

## 11. Security Considerations

### VM Isolation

```bash
# Use host-only network for sensitive development
# Disable NAT/Bridged unless required

# Enable firewall in VMs
sudo ufw enable
sudo ufw allow 8080/tcp   # Hades API
sudo ufw allow 5432/tcp    # PostgreSQL (from dev VM only)
sudo ufw deny 5432/tcp      # External
```

### Credentials

```bash
# Never commit .env files
echo ".env" >> ~/.gitignore_global
git config --global excludesfile ~/.gitignore_global

# Use Vault for credentials
export HADES_DB_PASSWORD=$(vault read -field=password secret/hades/db)
```

---

## 12. Quick Reference

```bash
# Common vmrun commands
vmrun start /path/to/vm.vmx                    # Start VM
vmrun stop /path/to/vm.vmx soft                 # Stop gracefully
vmrun snapshot /path/to/vm.vmx "name"            # Create snapshot
vmrun listSnapshots /path/to/vm.vmx            # List snapshots
vmrun deleteSnapshot /path/to/vm.vmx "name"     # Delete snapshot
vmrun revertToSnapshot /path/to/vm.vmx "name"   # Restore snapshot

# IP address
vmrun getGuestIP /path/to/vm.vmx

# Copy files
vmrun -h | grep -i copy                        # File operations
```

---

**Last Updated**: 2026-05-09
**HADES-V2 Version**: 2.0.0
