import json
from pathlib import Path
from graphify.detect import detect

result = detect(Path('.'))
Path('graphify-out/.graphify_detect.json').write_text(
    json.dumps(result, ensure_ascii=False, indent=2), 
    encoding='utf-8'
)
print("Detected files summary:")
print(f"Total files: {result.get('total_files', 0)}")
print(f"Total words: {result.get('total_words', 0)}")
for category, files in result.get('files', {}).items():
    if files:
        print(f"  {category}: {len(files)} files")
