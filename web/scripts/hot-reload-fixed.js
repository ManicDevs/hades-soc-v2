#!/usr/bin/env node

import { createServer } from 'http';
import { WebSocketServer } from 'ws';
import chokidar from 'chokidar';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

class FixedHotReloadServer {
  constructor() {
    this.port = 3001;
    this.wss = null;
    this.watchers = [];
  }

  async start() {
    try {
      console.log(`🔥 Starting fixed hot reload server on port ${this.port}`);
      
      // Create HTTP server for WebSocket upgrade handling
      const server = createServer();
      
      server.on('upgrade', (request, socket, head) => {
        console.log('🔄 WebSocket upgrade request received');
        
        if (request.headers['upgrade'] !== 'websocket') {
          socket.destroy();
          return;
        }

        const wss = new WebSocketServer({ noServer: true });
        wss.handleUpgrade(request, socket, head);
      });

      server.on('error', (error) => {
        console.error('HTTP server error:', error);
      });

      // Start the server
      server.listen(this.port, () => {
        console.log(`🔥 Fixed hot reload server listening on port ${this.port}`);
      });

      this.setupFileWatchers();
      
    } catch (error) {
      console.error('Failed to start hot reload server:', error.message);
      throw error;
    }
  }

  setupFileWatchers() {
    const webDir = path.resolve(__dirname, '../src');
    
    // Watch component files
    const componentWatcher = chokidar.watch(path.join(webDir, 'components/**/*.tsx'), {
      ignoreInitial: true,
      persistent: true
    });

    componentWatcher.on('change', (filePath) => {
      console.log(`🔄 Component changed: ${filePath}`);
      this.broadcastComponentUpdate(filePath);
    });

    this.watchers.push(componentWatcher);

    // Watch page files
    const pageWatcher = chokidar.watch(path.join(webDir, 'pages/**/*.tsx'), {
      ignoreInitial: true,
      persistent: true
    });

    pageWatcher.on('change', (filePath) => {
      console.log(`📄 Page changed: ${filePath}`);
      this.broadcastComponentUpdate(filePath);
    });

    this.watchers.push(pageWatcher);
  }

  broadcastComponentUpdate(filePath) {
    const componentId = this.getComponentId(filePath);
    const event = {
      type: 'component-update',
      componentId,
      data: { filePath, timestamp: Date.now() },
      timestamp: Date.now()
    };

    this.broadcast(event);
  }

  broadcast(event) {
    // Broadcast to all connected WebSocket clients
    if (this.wss) {
      this.wss.clients.forEach(client => {
        if (client.readyState === 1) { // WebSocket.OPEN
          client.send(JSON.stringify(event));
        }
      });
    }
  }

  getComponentId(filePath) {
    const parts = filePath.split('/');
    const fileName = parts[parts.length - 1];
    return fileName.replace(/\.(tsx|jsx)$/, '');
  }

  stop() {
    console.log('🔥 Stopping hot reload server');
    
    this.watchers.forEach(watcher => watcher.close());
    this.watchers = [];

    if (this.wss) {
      this.wss.close();
    }
  }
}

// Start the server if this script is run directly
if (import.meta.url === `file://${process.argv[1]}`) {
  const server = new FixedHotReloadServer();
  
  server.start();
  
  // Handle graceful shutdown
  process.on('SIGINT', () => {
    console.log('\n🔥 Shutting down hot reload server...');
    server.stop();
    process.exit(0);
  });
}

export default FixedHotReloadServer;
