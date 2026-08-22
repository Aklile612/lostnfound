#!/usr/bin/env python3
import json
import os
import subprocess
import sys
import urllib.request
from pathlib import Path

API = os.environ.get("API_URL", "http://localhost:8080")
ROOT = Path(__file__).resolve().parents[1]
IMG_DIR = ROOT / "scripts" / ".testdata"
IMG_DIR.mkdir(parents=True, exist_ok=True)

IMAGES = {
    "backpack.jpg": "https://upload.wikimedia.org/wikipedia/commons/thumb/8/8a/Arcteryx_Alpha_Fast_light_40_black_backpack.jpg/960px-Arcteryx_Alpha_Fast_light_40_black_backpack.jpg",
    "backpack_b.jpg": "https://upload.wikimedia.org/wikipedia/commons/thumb/0/02/Eastpak_Sugarbush_backpack_black.jpg/960px-Eastpak_Sugarbush_backpack_black.jpg",
    "handbag.jpg": "https://upload.wikimedia.org/wikipedia/commons/thumb/a/a3/Leather_handbag_by_Les_cuirs_d%27Agathe_%28DSC07738%29.jpg/960px-Leather_handbag_by_Les_cuirs_d%27Agathe_%28DSC07738%29.jpg",
}


def fail(msg: str) -> None:
    print(f"FAIL: {msg}", file=sys.stderr)
    sys.exit(1)


def download(name: str, url: str) -> Path:
    dest = IMG_DIR / name
    req = urllib.request.Request(
        url,
        headers={"User-Agent": "ReuniteTest/1.0 (lost-and-found matching verification)"},
    )
    with urllib.request.urlopen(req, timeout=40) as res:
        data = res.read()
    if len(data) < 2000:
        fail(f"downloaded image too small: {name} ({len(data)} bytes)")
    dest.write_bytes(data)
    print(f"downloaded {name} ({len(data)} bytes)")
    return dest


def post_report(kind: str, fields: dict[str, str], photo: Path | None) -> dict:
    args = ["curl", "-sS", "-m", "180", "-X", "POST", f"{API}/api/reports/{kind}"]
    for key, value in fields.items():
        args.extend(["-F", f"{key}={value}"])
    if photo is not None:
        args.extend(["-F", f"photos=@{photo}"])
    raw = subprocess.check_output(args)
    try:
        payload = json.loads(raw)
    except json.JSONDecodeError:
        fail(f"non-json response: {raw[:500]!r}")
    if "error" in payload:
        fail(f"{kind} report rejected: {payload['error']}")
    return payload


def main() -> None:
    health = subprocess.check_output(["curl", "-sS", "-m", "5", f"{API}/health"])
    if b'"ok"' not in health:
        fail(f"backend not healthy: {health!r}")

    backpack = download("backpack.jpg", IMAGES["backpack.jpg"])
    backpack_b = download("backpack_b.jpg", IMAGES["backpack_b.jpg"])
    handbag = download("handbag.jpg", IMAGES["handbag.jpg"])

    found_backpack = post_report(
        "found",
        {
            "category": "bags",
            "title": "Black hiking backpack",
            "description": "Black technical backpack found near the library entrance.",
            "unique_features": "black, multiple zip pockets",
            "location": "library",
            "location_details": "near the entrance",
            "incident_date": "2026-08-22",
            "phone": "+251911000001",
            "telegram": "test_finder",
        },
        backpack,
    )
    found_handbag = post_report(
        "found",
        {
            "category": "bags",
            "title": "Brown leather handbag",
            "description": "Small brown leather handbag.",
            "unique_features": "leather, structured bag",
            "location": "cafeteria",
            "incident_date": "2026-08-22",
            "phone": "+251911000002",
            "telegram": "other_finder",
        },
        handbag,
    )
    lost = post_report(
        "lost",
        {
            "category": "bags",
            "title": "Black backpack",
            "description": "I lost my black backpack around the library yesterday afternoon. Dark daypack with zip pockets.",
            "unique_features": "black backpack, zip pockets",
            "location": "library",
            "location_details": "near the entrance",
            "incident_date": "2026-08-21",
        },
        backpack_b,
    )

    matches = lost.get("matches") or []
    print(json.dumps({"lost_id": lost["report"]["id"], "match_count": len(matches), "matches": matches}, indent=2))

    if not matches:
        fail("lost backpack produced zero matches")

    top = matches[0]
    backpack_id = found_backpack["report"]["id"]
    handbag_id = found_handbag["report"]["id"]
    found_ids = [m.get("found_report", {}).get("id") for m in matches]

    if backpack_id not in found_ids:
        fail(f"expected backpack {backpack_id} in matches, got {found_ids}")

    if top["found_report"]["id"] == handbag_id:
        fail("handbag ranked above the backpack")

    if top.get("groq_score") is None and top.get("gemini_score") is None:
        fail("match used heuristic fallback; Groq/Gemini did not score")

    if "Heuristic fallback" in (top.get("reasoning") or ""):
        fail("reasoning is still heuristic")

    if top["score"] < 40:
        fail(f"combined score too low: {top['score']}")

    print("PASS")
    print(f"top found={top['found_report']['id']} score={top['score']}")
    print(f"groq={top.get('groq_score')} gemini={top.get('gemini_score')}")
    print(f"reason={top.get('reasoning')}")


if __name__ == "__main__":
    main()
