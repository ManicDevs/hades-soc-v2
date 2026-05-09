#!/bin/bash
# HADES-V2 Enterprise Port Security Configuration
# Production-ready firewall rules for enterprise deployment

# Clear existing rules
iptables -F
iptables -X
iptables -t nat -F
iptables -t nat -X
iptables -t mangle -F
iptables -t mangle -X

# Default policies - DROP everything by default
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT

# Allow loopback traffic
iptables -A INPUT -i lo -j ACCEPT
iptables -A OUTPUT -o lo -j ACCEPT

# Allow established and related connections
iptables -A INPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT

# Allow SSH (restricted to management network)
iptables -A INPUT -p tcp --dport 22 -s 192.168.0.0/24 -j ACCEPT

# HTTP/HTTPS (Load Balancer)
iptables -A INPUT -p tcp --dport 80 -j ACCEPT
iptables -A INPUT -p tcp --dport 443 -j ACCEPT

# HADES API Ports (Internal only)
iptables -A INPUT -p tcp --dport 8443 -s 127.0.0.1 -j ACCEPT
iptables -A INPUT -p tcp --dport 8443 -s 10.0.0.0/8 -j ACCEPT
iptables -A INPUT -p tcp --dport 8443 -s 172.16.0.0/12 -j ACCEPT
iptables -A INPUT -p tcp --dport 8443 -s 192.168.0.0/16 -j ACCEPT

# HADES Dashboard Port (Internal only)
iptables -A INPUT -p tcp --dport 8444 -s 127.0.0.1 -j ACCEPT
iptables -A INPUT -p tcp --dport 8444 -s 10.0.0.0/8 -j ACCEPT
iptables -A INPUT -p tcp --dport 8444 -s 172.16.0.0/12 -j ACCEPT
iptables -A INPUT -p tcp --dport 8444 -s 192.168.0.0/16 -j ACCEPT

# HADES Admin Port (Restricted to admin network)
iptables -A INPUT -p tcp --dport 8445 -s 192.168.0.0/24 -j ACCEPT

# HADES Monitoring Port (Internal only)
iptables -A INPUT -p tcp --dport 8446 -s 127.0.0.1 -j ACCEPT
iptables -A INPUT -p tcp --dport 8446 -s 10.0.0.0/8 -j ACCEPT
iptables -A INPUT -p tcp --dport 8446 -s 172.16.0.0/12 -j ACCEPT
iptables -A INPUT -p tcp --dport 8446 -s 192.168.0.0/16 -j ACCEPT

# Internal Services (Database, Redis, etc.)
iptables -A INPUT -p tcp --dport 5432 -s 127.0.0.1 -j ACCEPT
iptables -A INPUT -p tcp --dport 6379 -s 127.0.0.1 -j ACCEPT
iptables -A INPUT -p tcp --dport 2112 -s 127.0.0.1 -j ACCEPT

# HAProxy Stats (Admin only)
iptables -A INPUT -p tcp --dport 9000 -s 192.168.0.0/24 -j ACCEPT

# Development Ports (Internal only)
iptables -A INPUT -p tcp --dport 8080 -s 127.0.0.1 -j ACCEPT
iptables -A INPUT -p tcp --dport 3000 -s 127.0.0.1 -j ACCEPT

# Monitoring Stack (Internal only)
iptables -A INPUT -p tcp --dport 3002 -s 127.0.0.1 -j ACCEPT
iptables -A INPUT -p tcp --dport 9090 -s 127.0.0.1 -j ACCEPT
iptables -A INPUT -p tcp --dport 16686 -s 127.0.0.1 -j ACCEPT
iptables -A INPUT -p tcp --dport 9200 -s 127.0.0.1 -j ACCEPT
iptables -A INPUT -p tcp --dport 5601 -s 127.0.0.1 -j ACCEPT

# Rate limiting for API endpoints
iptables -A INPUT -p tcp --dport 8443 -m limit --limit 25/minute --limit-burst 100 -j ACCEPT
iptables -A INPUT -p tcp --dport 8443 -m limit --limit 25/minute --limit-burst 100 -j LOG --log-prefix "API_RATE_LIMIT: "

# SYN flood protection
iptables -A INPUT -p tcp --syn -m limit --limit 1/s --limit-burst 3 -j ACCEPT
iptables -A INPUT -p tcp --syn -j DROP

# Port scanning protection
iptables -A INPUT -m recent --name portscan --rcheck --seconds 86400 -j DROP
iptables -A INPUT -m recent --name portscan --set -j LOG --log-prefix "PORTSCAN: "
iptables -A INPUT -m recent --name portscan --update --seconds 1 --hitcount 3 -j DROP

# Logging
iptables -A INPUT -j LOG --log-prefix "INPUT_DENIED: " --log-level 4
iptables -A FORWARD -j LOG --log-prefix "FORWARD_DENIED: " --log-level 4

# Save rules
iptables-save > /etc/iptables/rules.v4

echo "HADES-V2 Enterprise Firewall Rules Applied"
echo "Port Security Configuration Complete"
