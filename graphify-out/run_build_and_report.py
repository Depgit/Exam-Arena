import json
import inspect
from pathlib import Path
from graphify.build import build_from_json
from graphify.cluster import cluster, score_all
from graphify.analyze import god_nodes, surprising_connections, suggest_questions
from graphify.report import generate
from graphify.export import to_json, to_html
from graphify.diagnostics import diagnose_extraction, format_diagnostic_report
from graphify.detect import save_manifest

root_path = '.'
extraction = json.loads(Path('graphify-out/.graphify_extract.json').read_text(encoding="utf-8"))
detection  = json.loads(Path('graphify-out/.graphify_detect.json').read_text(encoding="utf-8"))

# Step 4: Build graph & cluster
G = build_from_json(extraction, root=root_path, directed=False)
if G.number_of_nodes() == 0:
    print('ERROR: Graph is empty')
    exit(1)

communities = cluster(G)
cohesion = score_all(G, communities)
tokens = {'input': extraction.get('input_tokens', 0), 'output': extraction.get('output_tokens', 0)}
gods = god_nodes(G)
surprises = surprising_connections(G, communities)

# Step 4.5 Health check
summary = diagnose_extraction(extraction, directed=False, root=root_path)
print("--- GRAPH HEALTH REPORT ---")
print(format_diagnostic_report(summary))

# Step 5: Community Labels
labels = {}
for cid, nodes in communities.items():
    sample_labels = [G.nodes[n].get('label', n) for n in nodes[:5]]
    labels[cid] = f"Community {cid}: {', '.join(sample_labels[:2])}"

questions = suggest_questions(G, communities, labels)

# Write graph.json & GRAPH_REPORT.md
to_json(G, communities, 'graphify-out/graph.json')
report = generate(G, communities, cohesion, labels, gods, surprises, detection, tokens, root_path, suggested_questions=questions)
Path('graphify-out/GRAPH_REPORT.md').write_text(report, encoding="utf-8")

analysis = {
    'communities': {str(k): v for k, v in communities.items()},
    'cohesion': {str(k): v for k, v in cohesion.items()},
    'gods': gods,
    'surprises': surprises,
    'questions': questions,
}
Path('graphify-out/.graphify_analysis.json').write_text(json.dumps(analysis, indent=2, ensure_ascii=False), encoding="utf-8")
Path('graphify-out/.graphify_labels.json').write_text(json.dumps({str(k): v for k, v in labels.items()}, ensure_ascii=False), encoding="utf-8")

print(f"Graph built: {G.number_of_nodes()} nodes, {G.number_of_edges()} edges, {len(communities)} communities")

# Step 6: HTML export via to_html
try:
    sig = inspect.signature(to_html)
    kwargs = {}
    if 'labels' in sig.parameters:
        kwargs['labels'] = labels
    to_html(G, communities, 'graphify-out/graph.html', **kwargs)
    print("Exported interactive visualization to graphify-out/graph.html")
except Exception as e:
    print(f"HTML export note: {e}")

# Step 9: Save manifest
save_manifest(detection.get('all_files') or detection['files'], root=root_path)
print("Saved manifest and report successfully.")
