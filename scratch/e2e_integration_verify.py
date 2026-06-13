import requests
import time
import json

BASE_URL = "http://localhost:4300"

def test_connectivity():
    print("--- Testing Service Connectivity ---")
    try:
        resp = requests.get(f"{BASE_URL}/api/service/connectivity")
        print(f"Status: {resp.status_code}")
        print(resp.json())
        return resp.status_code == 200
    except Exception as e:
        print(f"Error: {e}")
        return False

def test_ripgrep():
    print("\n--- Testing Native Tool: ripgrep_search ---")
    payload = {
        "name": "ripgrep_search",
        "arguments": {"pattern": "package", "path": "."}
    }
    try:
        start = time.time()
        resp = requests.post(f"{BASE_URL}/api/agent/tool", json=payload)
        latency = (time.time() - start) * 1000
        print(f"Status: {resp.status_code} | Latency: {latency:.2f}ms")
        if resp.status_code == 200:
            print("✅ Ripgrep execution verified.")
            return True
        else:
            print(f"❌ Ripgrep failed: {resp.text}")
            return False
    except Exception as e:
        print(f"Error: {e}")
        return False

def test_skills():
    print("\n--- Testing Skill Registry ---")
    try:
        start = time.time()
        resp = requests.get(f"{BASE_URL}/api/skills")
        latency = (time.time() - start) * 1000
        print(f"Status: {resp.status_code} | Latency: {latency:.2f}ms")
        if resp.status_code == 200:
            skills = resp.json().get("data", [])
            print(f"✅ Listed {len(skills)} skills via /api/skills.")
            return True
        else:
            print(f"❌ Skills failed: {resp.text}")
            return False
    except Exception as e:
        print(f"Error: {e}")
        return False

def test_scripts():
    print("\n--- Testing Prompt Library ---")
    try:
        start = time.time()
        resp = requests.get(f"{BASE_URL}/api/scripts")
        latency = (time.time() - start) * 1000
        print(f"Status: {resp.status_code} | Latency: {latency:.2f}ms")
        if resp.status_code == 200:
            print("✅ API access to scripts/prompts verified.")
            return True
        else:
            print(f"❌ Scripts failed: {resp.text}")
            return False
    except Exception as e:
        print(f"Error: {e}")
        return False

def test_system_overview():
    print("\n--- Testing Memory Tracking Schema ---")
    try:
        start = time.time()
        resp = requests.get(f"{BASE_URL}/api/system/overview")
        latency = (time.time() - start) * 1000
        print(f"Status: {resp.status_code} | Latency: {latency:.2f}ms")
        if resp.status_code == 200:
            print("✅ System overview is healthy.")
            return True
        else:
            print(f"❌ System overview failed: {resp.text}")
            return False
    except Exception as e:
        print(f"Error: {e}")
        return False

if __name__ == "__main__":
    print("🚀 TormentNexus E2E Integration Protocol v1\n")
    all_pass = True
    all_pass &= test_connectivity()
    all_pass &= test_ripgrep()
    all_pass &= test_skills()
    all_pass &= test_scripts()
    all_pass &= test_system_overview()

    if all_pass:
        print("\n🏁 Integration Tests Complete: ALL PASSED")
    else:
        print("\n🏁 Integration Tests Complete: FAILED")
        exit(1)
