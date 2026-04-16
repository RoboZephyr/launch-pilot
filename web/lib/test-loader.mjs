/**
 * Node ESM loader — maps bare '@preact/signals' to vendored signals-core.mjs
 * so state.js can be tested without a browser/import-map.
 *
 * Usage: node --loader ./web/lib/test-loader.mjs --test web/lib/state.test.js
 */
export function resolve(specifier, context, nextResolve) {
  if (specifier === '@preact/signals') {
    const url = new URL('../vendor/signals-core.mjs', import.meta.url).href;
    return { url, shortCircuit: true };
  }
  return nextResolve(specifier, context);
}
