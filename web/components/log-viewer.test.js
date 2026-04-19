import { describe, it } from 'node:test';
import { strict as assert } from 'node:assert';
import { renderToString } from 'preact-render-to-string';
import { html } from 'htm/preact';
import { computeHasMore, LogViewerView } from './log-viewer.js';

const repeat = (n, prefix = 'x') =>
  Array.from({ length: n }, (_, i) => `${prefix}${i}`).join('\n');

describe('computeHasMore', () => {
  const cases = [
    {
      name: '(a) stdout 150 lines, requested 200 → false',
      logs: {
        stdoutAvailable: true, stdout: repeat(150),
        stderrAvailable: false, stderr: '',
      },
      requested: 200,
      want: false,
    },
    {
      name: '(b) stdout 200 lines, requested 200 → true',
      logs: {
        stdoutAvailable: true, stdout: repeat(200),
        stderrAvailable: false, stderr: '',
      },
      requested: 200,
      want: true,
    },
    {
      name: '(c) requested 10000 → false (cap reached)',
      logs: {
        stdoutAvailable: true, stdout: repeat(10000),
        stderrAvailable: false, stderr: '',
      },
      requested: 10000,
      want: false,
    },
    {
      name: '(d) stderr-only at requested → true',
      logs: {
        stdoutAvailable: false, stdout: '',
        stderrAvailable: true, stderr: repeat(200, 'e'),
      },
      requested: 200,
      want: true,
    },
    {
      name: 'both streams below requested → false',
      logs: {
        stdoutAvailable: true, stdout: repeat(100),
        stderrAvailable: true, stderr: repeat(50, 'e'),
      },
      requested: 200,
      want: false,
    },
    {
      name: 'neither stream available → false',
      logs: {
        stdoutAvailable: false, stdout: '',
        stderrAvailable: false, stderr: '',
      },
      requested: 200,
      want: false,
    },
    {
      name: 'null logs → false',
      logs: null,
      requested: 200,
      want: false,
    },
    {
      name: 'stderr stream at requested trumps stdout below',
      logs: {
        stdoutAvailable: true, stdout: repeat(100),
        stderrAvailable: true, stderr: repeat(400, 'e'),
      },
      requested: 400,
      want: true,
    },
  ];
  for (const c of cases) {
    it(c.name, () => {
      assert.equal(computeHasMore(c.logs, c.requested), c.want);
    });
  }
});

describe('LogViewerView render', () => {
  const baseLogs = {
    stdoutAvailable: true,
    stdoutPath: '/tmp/std.log',
    stdout: repeat(300, 'o'),
    stderrAvailable: true,
    stderrPath: '/tmp/err.log',
    stderr: repeat(100, 'e'),
    message: '',
  };

  it('renders terminal "Showing all 400 lines" when hasMore=false && lines=400', () => {
    const out = renderToString(html`<${LogViewerView}
      logs=${baseLogs}
      lines=${400}
      loading=${false}
      loadingMore=${false}
      error=${null}
      hasMore=${false}
      search=${''}
      onSearchInput=${() => {}}
      onLoadMore=${() => {}}
    />`);
    assert.match(out, /Showing all 400 lines/,
      `expected "Showing all 400 lines" in:\n${out}`);
    assert.match(out, /log-viewer__done/);
    assert.ok(!/Load more/.test(out),
      `should not render Load more button when hasMore=false:\n${out}`);
  });

  it('renders Load more button when hasMore=true && lines=200', () => {
    const out = renderToString(html`<${LogViewerView}
      logs=${baseLogs}
      lines=${200}
      loading=${false}
      loadingMore=${false}
      error=${null}
      hasMore=${true}
      search=${''}
      onSearchInput=${() => {}}
      onLoadMore=${() => {}}
    />`);
    assert.match(out, /Load more \(currently 200 lines\)/);
    assert.ok(!/Showing all/.test(out),
      `should not render terminal text when hasMore=true:\n${out}`);
  });

  it('does not render terminal text when lines < 200 (initial load incomplete)', () => {
    const out = renderToString(html`<${LogViewerView}
      logs=${baseLogs}
      lines=${150}
      loading=${false}
      loadingMore=${false}
      error=${null}
      hasMore=${false}
      search=${''}
      onSearchInput=${() => {}}
      onLoadMore=${() => {}}
    />`);
    assert.ok(!/Showing all/.test(out),
      `should not render terminal text when lines<200:\n${out}`);
    assert.ok(!/Load more/.test(out));
  });
});
