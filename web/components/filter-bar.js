import { html } from 'htm/preact';
import {
  categoryFilter,
  statusFilter,
  onlyMine,
  categoryCounts,
  statusCounts,
} from '../lib/state.js';
import { CATEGORY_LABELS, CATEGORY_KEYS, STATUS_KEYS } from '../lib/classify.js';

/** Display labels for "all" + each category/status. */
const CATEGORY_DISPLAY = { all: 'All', ...CATEGORY_LABELS };
const STATUS_DISPLAY = { all: 'All', running: 'Running', stopped: 'Stopped', error: 'Error' };

export function FilterBar() {
  const cat = categoryFilter.value;
  const st = statusFilter.value;
  const mine = onlyMine.value;
  const catCounts = categoryCounts.value;
  const stCounts = statusCounts.value;

  return html`
    <div class="filter-bar">
      <div class="filter-bar__row filter-bar__row--space-between">
        <div class="filter-bar__chips">
          ${CATEGORY_KEYS.map(key => {
            const isActive = cat === key;
            const isDisabled = mine && key !== 'mine';
            let cls = 'filter-chip';
            if (isActive) cls += ' filter-chip--active';
            if (isDisabled) cls += ' filter-chip--disabled';
            return html`
              <button
                class=${cls}
                disabled=${isDisabled}
                onClick=${() => { if (!isDisabled) categoryFilter.value = key; }}
              >
                ${CATEGORY_DISPLAY[key]}
                <span class="filter-chip__count">${catCounts[key]}</span>
              </button>
            `;
          })}
        </div>
        <label class="only-mine-toggle">
          Only Mine
          <input
            type="checkbox"
            class="only-mine-toggle__input"
            checked=${mine}
            onChange=${() => { onlyMine.value = !onlyMine.value; }}
          />
          <span class="only-mine-toggle__slider"></span>
        </label>
      </div>
      <div class="filter-bar__row">
        ${STATUS_KEYS.map(key => {
          const isActive = st === key;
          let cls = 'status-tab status-tab--' + key;
          if (isActive) cls += ' status-tab--active';
          return html`
            <button
              class=${cls}
              onClick=${() => { statusFilter.value = key; }}
            >
              ${STATUS_DISPLAY[key]} (${stCounts[key]})
            </button>
          `;
        })}
      </div>
    </div>
  `;
}
