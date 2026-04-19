/**
 * Node ESM loader — maps bare specifiers used by the browser import-map to
 * the vendored ESM bundles so `node --test` can run component code without
 * a browser or build step.
 *
 * Usage: node --loader ./web/lib/test-loader.mjs --test web/components/job-row.test.js
 */
const BARE_TO_VENDOR = {
  '@preact/signals': '../vendor/signals-core.mjs',
  'htm/preact': '../vendor/htm-preact.mjs',
  'htm': '../vendor/htm.mjs',
  'preact/hooks': '../vendor/preact-hooks.mjs',
  'preact': '../vendor/preact.mjs',
  'preact-render-to-string': '../vendor/preact-render-to-string.mjs',
};

export function resolve(specifier, context, nextResolve) {
  const mapped = BARE_TO_VENDOR[specifier];
  if (mapped) {
    const url = new URL(mapped, import.meta.url).href;
    return { url, shortCircuit: true };
  }
  return nextResolve(specifier, context);
}
