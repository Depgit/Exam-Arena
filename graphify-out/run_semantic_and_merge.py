import json
from pathlib import Path

ast = json.loads(Path('graphify-out/.graphify_ast.json').read_text(encoding="utf-8"))
detect = json.loads(Path('graphify-out/.graphify_detect.json').read_text(encoding="utf-8"))

doc_files = detect.get('files', {}).get('document', [])
sem_nodes = []
sem_edges = []

for doc in doc_files:
    p = Path(doc)
    if p.exists():
        sem_nodes.append({
            "id": f"Doc::{p.name}",
            "label": p.name,
            "type": "DOCUMENT",
            "source_file": str(p),
            "source_location": f"{p}:1",
            "summary": f"Documentation file {p.name}"
        })

sem_data = {
    "nodes": sem_nodes,
    "edges": sem_edges,
    "hyperedges": [],
    "input_tokens": 0,
    "output_tokens": 0
}

Path('graphify-out/.graphify_semantic.json').write_text(json.dumps(sem_data, indent=2, ensure_ascii=False), encoding="utf-8")

# Part C Merge
seen = {n['id'] for n in ast['nodes']}
merged_nodes = list(ast['nodes'])
for n in sem_nodes:
    if n['id'] not in seen:
        merged_nodes.append(n)
        seen.add(n['id'])

merged_edges = ast['edges'] + sem_edges
merged = {
    'nodes': merged_nodes,
    'edges': merged_edges,
    'hyperedges': [],
    'input_tokens': 0,
    'output_tokens': 0,
}
Path('graphify-out/.graphify_extract.json').write_text(json.dumps(merged, indent=2, ensure_ascii=False), encoding="utf-8")
print(f"Merged extraction: {len(merged_nodes)} nodes, {len(merged_edges)} edges")
