import re

main_go_path = "cmd/server/main.go"
openapi_path = "openapi.yaml"

with open(main_go_path, 'r', encoding='utf-8') as f:
    code = f.read()

# Find all api.METHOD("/path"
routes = re.findall(r'(api|adminUsers|adminAccess|adminAuth|escalationHandler)\.(GET|POST|PUT|DELETE|PATCH)\(\"([^\"]+)\"', code)

# We also need to map the groups back to /api prefixes
group_prefixes = {
    'api': '/api',
    'adminAuth': '/api/admin/auth',
    'adminUsers': '/api/admin/users',
    'adminAccess': '/api/admin/access-controls',
    'escalationHandler': '/api/escalation'
}

collected_paths = {}

for group, method, path in routes:
    full_path = group_prefixes.get(group, '') + path
    
    # Replace gin parameters /:id with OpenAPI /{id}
    full_path = re.sub(r':([a-zA-Z0-9_]+)', r'{\1}', full_path)
    
    method_lower = method.lower()
    
    if full_path not in collected_paths:
        collected_paths[full_path] = {}
        
    collected_paths[full_path][method_lower] = True

# Also add some missing ones manually if missed by regex
extra_paths = [
    ('/api/escalation/config', 'get'),
    ('/api/escalation/config/{category}', 'get'),
    ('/api/escalation/complaint/{id}', 'get'),
    ('/api/escalation/complaint/{id}/history', 'get'),
    ('/api/escalation/complaint/{id}/escalate', 'post'),
    ('/api/escalation/complaint/{id}/assign', 'post'),
    ('/api/escalation/complaint/{id}/initialize', 'post'),
    ('/api/escalation/stats', 'get'),
]

for p, m in extra_paths:
    if p not in collected_paths:
        collected_paths[p] = {}
    collected_paths[p][m] = True

# Load existing openapi.yaml
with open(openapi_path, 'r', encoding='utf-8') as f:
    existing_yaml = f.read()

# Find existing paths to avoid duplicates
existing_paths = re.findall(r'^\s\s(/\S+):', existing_yaml, re.MULTILINE)

# Append new paths
new_yaml = ""
for path, methods in collected_paths.items():
    if path in existing_paths:
        continue
        
    new_yaml += f"  {path}:\n"
    for method in methods:
        new_yaml += f"    {method}:\n"
        new_yaml += f"      summary: {method.upper()} {path}\n"
        
        # Check if there are parameters in the path
        params = re.findall(r'{([^}]+)}', path)
        if params:
            new_yaml += f"      parameters:\n"
            for p in params:
                new_yaml += f"        - name: {p}\n"
                new_yaml += f"          in: path\n"
                new_yaml += f"          required: true\n"
                new_yaml += f"          schema:\n"
                new_yaml += f"            type: string\n"
        
        new_yaml += f"      responses:\n"
        new_yaml += f"        \"200\":\n"
        new_yaml += f"          description: Success\n"

with open(openapi_path, 'a', encoding='utf-8') as f:
    f.write(new_yaml)

print("openapi.yaml appended successfully")
