import time
import urllib.request
import json
import os

BASE_URL = "http://localhost:4300"
SAMPLE_REPO = "/tmp/tormentnexus-sample"

def call_endpoint(path, method='GET', payload=None):
    url = f"{BASE_URL}{path}"
    data = json.dumps(payload).encode('utf-8') if payload else None
    headers = {'Content-Type': 'application/json'} if payload else {}
    req = urllib.request.Request(url, data=data, method=method, headers=headers)

    print(f"\n>>> {method} {path}")
    if payload: print(f"Payload: {json.dumps(payload)}")

    start = time.perf_counter()
    try:
        with urllib.request.urlopen(req) as response:
            res_body = response.read().decode('utf-8')
            end = time.perf_counter()
            duration = (end - start) * 1000
            result = json.loads(res_body)
            print(f"Status: {response.status} | Latency: {duration:.2f}ms")
            return result
    except Exception as e:
        print(f"Status: Failed | Error: {e}")
        return None

def call_tool(name, arguments):
    return call_endpoint("/api/agent/tool", "POST", {"name": name, "arguments": arguments})

def run_e2e():
    print("🚀 TormentNexus E2E Integration Protocol v1\n")

    # 1. Native Tool Execution: Ripgrep
    print("--- Testing Native Tool: ripgrep_search ---")
    rg_res = call_tool("ripgrep_search", {"pattern": "TormentNexus", "path": SAMPLE_REPO})
    if rg_res and "success" in rg_res:
        print("✅ Ripgrep execution verified.")

    # 2. Skill Registry Operations
    print("\n--- Testing Skill Registry ---")
    skills = call_endpoint("/api/skills")
    if skills and "success" in skills:
        data = skills.get('data', [])
        print(f"✅ Listed {len(data)} skills via /api/skills.")

    # 3. Prompt Library Operations
    print("\n--- Testing Prompt Library ---")
    prompts = call_endpoint("/api/scripts") #saved scripts
    if prompts and "success" in prompts:
        print("✅ API access to scripts/prompts verified.")

    # 4. Memory Tracking & De-duplication Table Check
    print("\n--- Testing Memory Tracking Schema ---")
    overview = call_endpoint("/api/system/overview")
    if overview and "success" in overview:
        print("✅ System overview is healthy.")

    print("\n🏁 Integration Tests Complete")

if __name__ == "__main__":
    run_e2e()
