import sys

file_path = 'go/internal/httpapi/server.go'
with open(file_path, 'r') as f:
    lines = f.readlines()

new_lines = []
for line in lines:
    new_lines.append(line)
    if 's.mux.HandleFunc("/api/skills/summary", s.handleSkillsSummary)' in line:
        new_lines.append('\ts.mux.HandleFunc("/api/skills/predict", s.handleSkillPredict)\n')

with open(file_path, 'w') as f:
    f.writelines(new_lines)
