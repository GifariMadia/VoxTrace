import json
import urllib.request
from pathlib import Path


API = "http://localhost:8080/api"
OUT = Path("build/extractions/source")


def get_json(path: str):
    with urllib.request.urlopen(f"{API}{path}") as response:
        return json.load(response)


def main() -> None:
    OUT.mkdir(parents=True, exist_ok=True)
    jobs = get_json("/jobs")
    completed = [job for job in jobs if job["status"] == "completed"]
    manifest = []
    for job in completed:
        transcript = get_json(f"/jobs/{job['id']}/transcript")
        record = {**job, **transcript}
        target = OUT / f"{job['id']}.json"
        target.write_text(json.dumps(record, ensure_ascii=False, indent=2), encoding="utf-8")
        manifest.append({
            "id": job["id"],
            "filename": job["filename"],
            "duration": transcript["duration"],
            "language": transcript["language"],
            "segmentCount": len(transcript["segments"]),
            "source": str(target),
        })
    (OUT / "manifest.json").write_text(
        json.dumps(manifest, ensure_ascii=False, indent=2), encoding="utf-8"
    )
    print(json.dumps(manifest, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
