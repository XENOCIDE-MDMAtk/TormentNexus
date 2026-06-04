import urllib.request, json
req = urllib.request.Request('http://localhost:1234/v1/chat/completions', 
    data=b'{"messages":[{"role":"user","content":"test"}]}', 
    headers={'Content-Type': 'application/json'})
try:
    urllib.request.urlopen(req)
except Exception as e:
    print(e.read().decode())
