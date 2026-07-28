import urllib.request, json

# Check latest test workflow run
url = 'https://api.github.com/repos/shahriarhaqueabir/UniversalOps/actions/workflows/test.yml/runs?per_page=1'
data = json.loads(urllib.request.urlopen(url).read())
run = data['workflow_runs'][0]
print('Latest CI Run:', run['display_title'])
print('Branch:', run['head_branch'])
print('Conclusion:', run['conclusion'])
print('Status:', run['status'])
print('URL:', run['html_url'])
print()

# Get jobs
jobs_url = run['jobs_url']
jobs_data = json.loads(urllib.request.urlopen(jobs_url).read())
for j in jobs_data.get('jobs', []):
    conclusion = j['conclusion'] or 'in_progress'
    print(f"Job: {j['name']} | Conclusion: {conclusion}")
    for s in j.get('steps', []):
        if s['conclusion'] == 'failure':
            print(f"  FAILED STEP: {s['name']}")