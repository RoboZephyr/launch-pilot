import { access } from "node:fs/promises";
import { createServer } from "node:http";
import { fileURLToPath } from "node:url";
import handler from "serve-handler";

const output = new URL("../out/", import.meta.url);
await access(new URL("index.html", output)).catch(() => {
  console.error("Static export missing. Run npm run build first.");
  process.exit(1);
});

const server = createServer((request, response) => {
  handler(request, response, {
    public: fileURLToPath(output),
    directoryListing: false,
  }).catch((error) => {
    console.error(error);
    if (!response.headersSent) response.writeHead(500);
    response.end();
  });
});

server.on("error", (error) => {
  console.error(`Preview failed: ${error.message}`);
  process.exitCode = 1;
});

server.listen(4174, "127.0.0.1", () => {
  console.log("Preview: http://127.0.0.1:4174");
});
