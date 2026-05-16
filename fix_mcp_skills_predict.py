import sys

file_path = 'packages/core/src/MCPServer.ts'
with open(file_path, 'r') as f:
    lines = f.readlines()

new_lines = []
for line in lines:
    if 'result = await this.skillRegistry.listSkills();' in line:
        new_lines.append('                result = await this.skillRegistry.listSkills(args?.query as string);\n')
    else:
        new_lines.append(line)

with open(file_path, 'w') as f:
    f.writelines(new_lines)
