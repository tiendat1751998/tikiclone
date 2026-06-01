# Script to clean and rebuild docker-compose.yml with proper ulimits
# Removing all existing ulimits/sysctls blocks then adding them once per service
import re

with open('/home/datdt/tikiclone/docker-compose.yml', 'r') as f:
    lines = f.read().split('\n')

# First pass: remove ALL ulimits and sysctls blocks
cleaned = []
i = 0
while i < len(lines):
    if lines[i].strip() == 'ulimits:' or lines[i].strip() == 'sysctls:':
        # Skip this entire block
        indent = len(lines[i]) - len(lines[i].lstrip())
        i += 1
        while i < len(lines):
            if lines[i].strip() == '':
                # Check if next non-blank line is still part of the block
                j = i + 1
                while j < len(lines) and lines[j].strip() == '':
                    j += 1
                if j < len(lines):
                    next_indent = len(lines[j]) - len(lines[j].lstrip())
                    if next_indent > indent:
                        i = j
                        continue
                break
            line_indent = len(lines[i]) - len(lines[i].lstrip())
            if line_indent <= indent and lines[i].strip() != '':
                break
            i += 1
        continue
    cleaned.append(lines[i])
    i += 1

print(f"After cleanup: {len(cleaned)} lines")

# Second pass: add ulimits/sysctls to each Go service (once per service)
skip_services = {'mysql-primary', 'redis-master', 'mongodb', 'kafka', 'zookeeper',
                 'otel-collector', 'prometheus', 'grafana', 'jaeger', 'ollama',
                 'identity-auth', 'web'}

new_lines = []
i = 0
services_added = set()

while i < len(cleaned):
    new_lines.append(cleaned[i])
    
    # Check if this line ends a depends_on block with service_healthy or service_started
    if ('condition: service_healthy' in cleaned[i] or 'condition: service_started' in cleaned[i]):
        # Find the service name by looking back
        service_name = ''
        for j in range(i-1, max(0, i-50), -1):
            m = re.match(r'^  ([\w-]+):$', cleaned[j])
            if m:
                service_name = m.group(1)
                break
        
        if service_name and service_name not in skip_services and service_name not in services_added:
            # Peek ahead to see if another condition: follows (multi-depends_on)
            j = i + 1
            while j < len(cleaned) and cleaned[j].strip() == '':
                j += 1
            
            # If another condition: follows, skip now (not last condition)
            is_last_condition = not (j < len(cleaned) and 'condition: service_healthy' in cleaned[j])
            
            if is_last_condition:
                services_added.add(service_name)
                new_lines.append('')
                new_lines.append('    ulimits:')
                new_lines.append('      nofile:')
                new_lines.append('        soft: 100000')
                new_lines.append('        hard: 100000')
                new_lines.append('    sysctls:')
                new_lines.append('      - net.core.somaxconn=65536')
                new_lines.append('      - net.ipv4.tcp_fin_timeout=15')
                new_lines.append('      - net.ipv4.tcp_tw_reuse=1')
    
    i += 1

print(f"Added ulimits to {len(services_added)} services: {sorted(services_added)}")

with open('/home/datdt/tikiclone/docker-compose.yml', 'w') as f:
    f.write('\n'.join(new_lines))

# Verify
content = '\n'.join(new_lines)
ulimits_count = len(re.findall(r'^    ulimits:$', content, re.MULTILINE))
print(f"Total ulimits blocks: {ulimits_count}")
print(f"Total lines: {len(new_lines)}")
