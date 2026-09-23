# Launch Pilot landing page

Next.js / React / TypeScript marketing site, exported as static HTML.

## Development

Use npm and the committed `package-lock.json` (Node.js 20.9 or newer):

```sh
npm ci
npm run dev
npm run lint
npm run build
npm start        # preview the static export at http://127.0.0.1:4174
```

`npm run build` writes the static site to `out/`, configured by `next.config.ts`.
The Cloudflare Pages output directory is `out`; the build command is `npm run build`.
The preview serves only `out/` and binds to localhost. If port 4174 is occupied,
it exits instead of silently switching ports.

## Structure

- `src/content.ts`: shared copy, metadata, links, and install command
- `src/app/`: page, root layout, and global styles
- `src/kit/`: reusable visual components and design tokens
- `src/page/`: providers and shared hooks
- `public/`: static assets

Keep generated output (`.next/` and `out/`) out of version control.
