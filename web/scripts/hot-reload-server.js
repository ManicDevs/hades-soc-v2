#!/usr/bin/env node

import { WebSocketServer } from 'ws';
import chokidar from 'chokidar';
import path from 'path';
import { fileURLToPath } from 'url';
import PortManager from './port-manager.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

class HotReloadServer {
  constructor() {
    this.port = null;
    this.wss = null;
    this.watchers = [];
    this.portManager = new PortManager();
  }

  async start() {
    try {
      // Use port 3002 for hot reload server in Docker setup
      this.port = 3002;
      console.log(`🔥 Starting hot reload server on port ${this.port}`);
      
      this.wss = new WebSocketServer({ 
        port: this.port,
        perMessageDeflate: false
      });
      
      this.wss.on('connection', (ws, request) => {
        console.log('🔥 Hot reload client connected from:', request.socket.remoteAddress);
        
        ws.on('message', (message) => {
          try {
            const data = JSON.parse(message.toString());
            console.log('📨 Received message:', data);
          } catch (error) {
            console.error('Failed to parse message:', error);
          }
        });

        ws.on('close', () => {
          console.log('🔥 Hot reload client disconnected');
        });

        ws.on('error', (error) => {
          console.error('Hot reload WebSocket error:', error);
        });
      });

      this.setupFileWatchers();
      console.log(`🔥 Hot reload server listening on port ${this.port}`);
      
      // Write port to file for client to read
      await this.writePortToFile();
    } catch (error) {
      console.error('Failed to start hot reload server:', error.message);
      throw error;
    }
  }

  async writePortToFile() {
    const fs = await import('fs');
    const portFile = path.join(__dirname, '..', '.hot-reload-port');
    fs.writeFileSync(portFile, this.port.toString());
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

    // Watch style files
    const styleWatcher = chokidar.watch(path.join(webDir, '**/*.css'), {
      ignoreInitial: true,
      persistent: true
    });

    styleWatcher.on('change', (filePath) => {
      console.log(`🎨 Style changed: ${filePath}`);
      this.broadcastStyleUpdate(filePath);
    });

    this.watchers.push(styleWatcher);
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

  broadcastStyleUpdate(filePath) {
    const event = {
      type: 'style-update',
      data: { filePath, timestamp: Date.now() },
      timestamp: Date.now()
    };

    this.broadcast(event);
  }

  broadcast(event) {
    if (!this.wss) return;
    
    this.wss.clients.forEach(client => {
      if (client.readyState === 1) { // WebSocket.OPEN
        client.send(JSON.stringify(event));
      }
    });
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
  const server = new HotReloadServer();
  
  server.start();
  
  // Handle graceful shutdown
  process.on('SIGINT', () => {
    console.log('\n🔥 Shutting down hot reload server...');
    server.stop();
    process.exit(0);
  });
}

export default HotReloadServer;
