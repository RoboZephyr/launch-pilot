import { html } from 'htm/preact';
import { searchQuery } from '../lib/state.js';

export function SearchBar() {
  const onInput = (e) => {
    searchQuery.value = e.target.value;
  };

  return html`
    <div class="search-bar">
      <input
        type="text"
        class="search-bar__input"
        placeholder="Filter by label\u2026"
        value=${searchQuery}
        onInput=${onInput}
      />
    </div>
  `;
}
