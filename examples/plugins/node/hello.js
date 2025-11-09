// Example Dragonfly plugin implemented in Node.js.
// Requires: npm install @grpc/grpc-js @grpc/proto-loader

import grpc from '@grpc/grpc-js';
import protoLoader from '@grpc/proto-loader';
import { fileURLToPath } from 'url';
import { dirname, resolve } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const PROTO_PATH = resolve(__dirname, '../../../plugin/proto/plugin.proto');
const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
});
const dfplugin = grpc.loadPackageDefinition(packageDefinition).df.plugin;

const pluginId = process.env.DF_PLUGIN_ID || 'node-plugin';
const address = process.env.DF_PLUGIN_GRPC_ADDRESS || '127.0.0.1:50051';

const server = new grpc.Server();

server.addService(dfplugin.Plugin.service, {
  EventStream: streamHandler,
});

server.bindAsync(address, grpc.ServerCredentials.createInsecure(), (err) => {
  if (err) {
    console.error('Failed to bind gRPC server', err);
    process.exit(1);
  }
  console.log(`[node] plugin listening on ${address}`);
  server.start();
});

function streamHandler(call) {
  console.log('[node] host connected');

  call.on('data', (message) => {
    if (message.hello) {
      console.log('[node] host hello', message.hello);
      call.write({
        plugin_id: pluginId,
        hello: {
          name: 'example-node',
          version: '0.1.0',
          api_version: message.hello.api_version,
          commands: [
            { name: '/hello', description: 'Send a greeting from the Node plugin' },
          ],
        },
      });
      call.write({
        plugin_id: pluginId,
        subscribe: { events: ['PLAYER_JOIN', 'COMMAND', 'CHAT'] },
      });
      return;
    }

    if (message.event) {
      handleEvent(call, message.event);
    }
    if (message.shutdown) {
      console.log('[node] host shutdown:', message.shutdown.reason);
      call.end();
    }
  });

  call.on('end', () => {
    console.log('[node] stream ended');
    call.end();
  });
}

function handleEvent(call, event) {
  switch (event.type) {
    case 'PLAYER_JOIN': {
      const player = event.player_join;
      console.log(`[node] player joined ${player.name}`);
      call.write({
        plugin_id: pluginId,
        actions: {
          actions: [
            {
              send_chat: {
                target_uuid: player.player_uuid,
                message: `Welcome to the server, ${player.name}! (from Node)`,
              },
            },
          ],
        },
      });
      break;
    }
    case 'COMMAND': {
      const command = event.command;
      if (command.raw.startsWith('/hello')) {
        call.write({
          plugin_id: pluginId,
          actions: {
            actions: [
              {
                send_chat: {
                  target_uuid: command.player_uuid,
                  message: `Hello ${command.name}!`,
                },
              },
            ],
          },
        });
      }
      break;
    }
    default:
      console.log('[node] event', event.type);
  }
}
