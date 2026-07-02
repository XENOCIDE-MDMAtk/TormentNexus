import { ingestAll } from '../packages/cli/dist/core/src/services/published-catalog-ingestor.js';

async function main() {
    console.log("Triggering automatic catalog ingestion...");
    try {
        const results = await ingestAll();
        console.log("Ingestion complete:");
        console.log(JSON.stringify(results, null, 2));
    } catch (e) {
        console.error("Fatal error during ingestion:", e);
        process.exit(1);
    }
}

main();
