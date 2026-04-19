import { html, render } from 'htm/preact';
import { useEffect } from 'preact/hooks';
import { connectSSE } from './lib/sse.js';
import { SearchBar } from './components/search-bar.js';
import { FilterBar } from './components/filter-bar.js';
import { JobTable } from './components/job-table.js';
import { ToastContainer } from './components/toast.js';
import { StatusTooltip } from './components/job-tooltip.js';

function App() {
  useEffect(() => {
    const es = connectSSE();
    return () => es.close();
  }, []);

  return html`
    <header>
      <h1>Launch Pilot</h1>
      <p>macOS LaunchAgent control console</p>
    </header>
    <main>
      <${SearchBar} />
      <${FilterBar} />
      <${JobTable} />
    </main>
    <${ToastContainer} />
    <${StatusTooltip} />
  `;
}

render(html`<${App} />`, document.getElementById('app'));
