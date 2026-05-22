console.log("DEBUG: Starting import test...");
const start = Date.now();
try {
    console.log("DEBUG: Importing @hypercode/core...");
    const core = await import('@hypercode/core');
    console.log("DEBUG: @hypercode/core loaded successfully in " + (Date.now() - start) + "ms");
    console.log("Exports:", Object.keys(core));
} catch (e) {
    console.error("DEBUG: Failed to import @hypercode/core", e);
}
