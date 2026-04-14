import { html, render } from 'htm/preact';

function App() {
  return html`
    <header>
      <h1>Launchboard</h1>
      <p>macOS LaunchAgent dashboard</p>
    </header>
    <main id="main"></main>
  `;
}

render(html`<${App} />`, document.getElementById('app'));
