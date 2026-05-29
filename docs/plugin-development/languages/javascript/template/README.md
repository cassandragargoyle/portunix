# JavaScript/Node.js Plugin Template

Template for Node.js-based Portunix plugins using `@grpc/grpc-js` and `@grpc/proto-loader`.

## Quick Start

```bash
npm install
npm test
node src/main.js --config config.yaml

# Install into Portunix
portunix plugin install .
portunix plugin enable my-js-plugin
portunix plugin start my-js-plugin
```

## Structure

```text
.
├── plugin.json              # Plugin manifest consumed by Portunix
├── package.json             # npm manifest
├── config.yaml              # Default runtime configuration
├── proto/plugin.proto       # gRPC service definition (loaded at runtime)
├── src/
│   ├── main.js              # Entry point — boots gRPC server
│   ├── config.js            # YAML config loader
│   ├── handlers/plugin.js   # gRPC method handlers
│   └── services/plugin.js   # Business logic
└── test/plugin.test.js      # node --test
```

See `../getting-started.md` for the full walkthrough.
