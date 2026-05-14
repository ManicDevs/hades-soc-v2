#!/usr/bin/env node

import { createServer } from 'http';
import net from 'net';

class PortManager {
  constructor() {
    this.allocatedPorts = new Set();
  }

  async findAvailablePort(startPort = 3000, maxAttempts = 10) {
    for (let i = 0; i < maxAttempts; i++) {
      const port = startPort + i;
      if (await this.isPortAvailable(port)) {
        this.allocatedPorts.add(port);
        return port;
      }
    }
    throw new Error(`No available ports found starting from ${startPort}`);
  }

  async isPortAvailable(port) {
    return new Promise((resolve) => {
      const server = createServer();
      
      server.listen(port, () => {
        const { port: actualPort } = server.address();
        server.close(() => {
          resolve(actualPort === port);
        });
      });
      
      server.on('error', () => {
        resolve(false);
      });
    });
  }

  async allocatePorts() {
    const vitePort = await this.findAvailablePort(3000);
    const hotReloadPort = await this.findAvailablePort(vitePort + 1);
    
    return {
      vitePort,
      hotReloadPort
    };
  }

  releasePort(port) {
    this.allocatedPorts.delete(port);
  }
}

// Export for use in other scripts
export default PortManager;

// CLI usage
if (import.meta.url === `file://${process.argv[1]}`) {
  const portManager = new PortManager();
  
  portManager.allocatePorts().then(ports => {
    console.log(JSON.stringify(ports));
  }).catch(error => {
    console.error('Error allocating ports:', error.message);
    process.exit(1);
  });
}
