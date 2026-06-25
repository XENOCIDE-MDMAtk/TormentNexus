import { type NextRequest, NextResponse } from "next/server";
import { readFileSync, writeFileSync, existsSync } from "fs";
import { join } from "path";

const CONFIG_PATH = join(process.cwd(), "data", "always-on-tools.json");

interface AlwaysOnConfig {
  tools: Record<string, boolean>;
}

function loadConfig(): AlwaysOnConfig {
  try {
    if (existsSync(CONFIG_PATH)) {
      return JSON.parse(readFileSync(CONFIG_PATH, "utf-8"));
    }
  } catch {
    // ignore
  }
  return { tools: {} };
}

function saveConfig(config: AlwaysOnConfig) {
  const dir = join(process.cwd(), "data");
  if (!existsSync(dir)) {
    // Use temp dir as fallback
    writeFileSync(CONFIG_PATH, JSON.stringify(config, null, 2), "utf-8");
    return;
  }
  writeFileSync(CONFIG_PATH, JSON.stringify(config, null, 2), "utf-8");
}

export async function POST(request: NextRequest) {
  try {
    const { name, alwaysOn } = await request.json();

    if (!name || typeof name !== "string") {
      return NextResponse.json(
        { success: false, error: "Missing or invalid 'name'" },
        { status: 400 },
      );
    }

    const config = loadConfig();
    config.tools[name] = alwaysOn === true;
    saveConfig(config);

    return NextResponse.json({ success: true, name, alwaysOn: config.tools[name] });
  } catch (error) {
    const message = error instanceof Error ? error.message : "Unknown error";
    return NextResponse.json(
      { success: false, error: message },
      { status: 500 },
    );
  }
}
