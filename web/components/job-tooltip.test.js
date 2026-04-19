import { describe, it } from 'node:test';
import { strict as assert } from 'node:assert';
import { renderToString } from 'preact-render-to-string';
import { html } from 'htm/preact';
import { buildStatusTooltipParts, StatusTooltip } from './job-tooltip.js';
import { jobs, tooltipTarget } from '../lib/state.js';

// Reimplements the spec'd parts rules here so the assertion is an independent
// oracle — never imports the SUT to compute expected output.
const STATUSES = ['running', 'error', 'completed', 'scheduled', 'stopped', 'offline'];
const NEXT_ISO = '2026-04-19T10:00:00Z';
const LAST_ISO = '2026-04-18T10:00:00Z';
const LOG_PATH = '/tmp/x.log';
const LAST_RUN_FALLBACK_STATUSES = new Set(['completed', 'scheduled']);

function expectedParts(job) {
  const fmt = (iso) => new Date(iso).toLocaleString();
  const parts = [job.status];
  if (job.nextRunAt) parts.push(`Next run: ${fmt(job.nextRunAt)}`);
  if (job.lastRunAt) {
    parts.push(`Last run: ${fmt(job.lastRunAt)}`);
  } else if (LAST_RUN_FALLBACK_STATUSES.has(job.status)) {
    const hasLogPath = Boolean(job.standardOutPath) || Boolean(job.standardErrPath);
    parts.push(hasLogPath ? 'Last run: unknown' : 'Last run: unknown (no log path configured)');
  }
  return parts;
}

describe('buildStatusTooltipParts 48-combo', () => {
  for (const status of STATUSES) {
    for (const next of [null, NEXT_ISO]) {
      for (const last of [null, LAST_ISO]) {
        for (const log of [null, LOG_PATH]) {
          const desc = `status=${status} next=${next ? 'y' : 'n'} last=${last ? 'y' : 'n'} log=${log ? 'y' : 'n'}`;
          const job = {
            label: 'com.example.test',
            status,
            nextRunAt: next,
            lastRunAt: last,
            standardOutPath: log,
            standardErrPath: null,
          };
          it(`buildStatusTooltipParts: ${desc}`, () => {
            assert.deepEqual(buildStatusTooltipParts(job), expectedParts(job));
          });
        }
      }
    }
  }
});

describe('StatusTooltip mount (one per status)', () => {
  const mountCases = [
    {
      status: 'running',
      job: {
        label: 'com.example.run', status: 'running',
        nextRunAt: null, lastRunAt: null,
        standardOutPath: null, standardErrPath: null,
      },
    },
    {
      status: 'error',
      job: {
        label: 'com.example.err', status: 'error',
        nextRunAt: null, lastRunAt: LAST_ISO,
        standardOutPath: null, standardErrPath: null,
      },
    },
    {
      status: 'completed',
      job: {
        label: 'com.example.done', status: 'completed',
        nextRunAt: null, lastRunAt: LAST_ISO,
        standardOutPath: LOG_PATH, standardErrPath: null,
      },
    },
    {
      status: 'scheduled',
      job: {
        label: 'com.example.sched', status: 'scheduled',
        nextRunAt: NEXT_ISO, lastRunAt: null,
        standardOutPath: null, standardErrPath: null,
      },
    },
    {
      status: 'stopped',
      job: {
        label: 'com.example.stop', status: 'stopped',
        nextRunAt: null, lastRunAt: null,
        standardOutPath: null, standardErrPath: null,
      },
    },
    {
      status: 'offline',
      job: {
        label: 'com.example.off', status: 'offline',
        nextRunAt: null, lastRunAt: LAST_ISO,
        standardOutPath: null, standardErrPath: null,
      },
    },
  ];

  for (const { status, job } of mountCases) {
    it(`renders parts in order for status=${status}`, () => {
      jobs.value = [job];
      tooltipTarget.value = {
        anchor: { isConnected: true, getBoundingClientRect: () => ({}) },
        label: job.label,
        enteredVia: 'hover',
      };
      try {
        const out = renderToString(html`<${StatusTooltip} />`);
        const parts = expectedParts(job);
        let cursor = -1;
        for (const part of parts) {
          const idx = out.indexOf(part, cursor + 1);
          assert.ok(
            idx > cursor,
            `expected part ${JSON.stringify(part)} after index ${cursor} in:\n${out}`,
          );
          cursor = idx;
        }
        for (const part of parts) {
          assert.match(
            out,
            new RegExp(`<div[^>]*class="status-tooltip__line"[^>]*>[^<]*${part.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}`),
            `missing status-tooltip__line wrapping ${JSON.stringify(part)} in:\n${out}`,
          );
        }
      } finally {
        tooltipTarget.value = null;
        jobs.value = [];
      }
    });
  }
});
