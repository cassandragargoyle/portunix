# Issue #260: Plugin runtime node for JavaScript plugins

**Status**: ✅ Implemented
**Priority**: Low
**Type**: Enhancement
**Component**: portunix
**Created**: 2026-11-20
**Closed**: 2026-11-23
**GitHub**: #260
**Related**: #7, #257

## Summary

Portunix supports the plugin runtimes `native`, `java` and `python`. The
JavaScript plugin template in
`docs/plugin-development/languages/javascript/template/plugin.json` declares
`"runtime": "node"` and `"binary": "node src/main.js"`. `portunix plugin
validate` rejects it and the dispatcher could not start it, so a plugin made
from the template cannot be installed. Found while testing #257.

## Problem Description

```text
$ portunix plugin validate docs/plugin-development/languages/javascript/template
Error: validation failed: manifest validation failed: unsupported runtime: node (supported: native, java, python)
```

- `ValidateManifest` in `src/app/plugins/manifest.go` accepts only `native`,
  `java` and `python`
- `buildPluginCommand` in `src/cmd/plugin_dispatcher.go` has no branch for
  Node.js; with `runtime: native` the binary `node src/main.js` would be taken
  as one file name and fail to start
- The Java template had the same kind of defect (`runtime: jvm`); it was fixed
  in #257 because `java` is a supported runtime

## Proposed Solution

1. Add the runtime `node` to manifest validation, with an optional
   `runtime_version` (e.g. `">=18"`, as in the template `package.json`)
2. Add a `node` branch to `buildPluginCommand`: `node <install_path>/<binary>
   [args...]`, with the same working directory and environment as the other
   runtimes (`PORTUNIX_CWD`, `PORTUNIX_HELP_AI`)
3. Check the Node.js runtime during `plugin install` / `plugin check` like Java
   (missing `node` → clear message with `portunix install nodejs`)
4. Decide whether `npm install` of the plugin dependencies runs at install
   time (like the venv for Python wheels) or the plugin ships with
   `node_modules`; document the choice
5. Fix the JavaScript template: `"runtime": "node"`, `"binary": "src/main.js"`
6. Document the runtime in the plugin development guide

## Acceptance Criteria

1. `portunix plugin validate` accepts `runtime: node` and the JavaScript
   template passes
2. A plugin with `runtime: node` installs and runs through
   `portunix <plugin> <command>`
3. A missing Node.js runtime gives a clear error with an install hint
4. Existing runtimes (`native`, `java`, `python`) are unchanged
5. Unit tests cover validation and the command built for `node`

## Testing

- Unit: `ValidateManifest` with `node`, `buildPluginCommand` for `node`
- Container (`ubuntu:24.04`): `portunix install nodejs`, build the template,
  `portunix plugin install`, run its `hello` command
- Container without Node.js: install fails with the install hint
- Regression: Java and Python plugin tests in `src/cmd` and
  `src/app/plugins/manager`
