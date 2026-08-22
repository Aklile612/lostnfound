#!/usr/bin/env python3
import json
import subprocess
import urllib.request
import uuid
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
IMG_DIR = ROOT / "scripts" / ".testdata"
IMG_DIR.mkdir(parents=True, exist_ok=True)
UA = {"User-Agent": "ReuniteCampusSeed/1.0 (educational lost-and-found demo)"}
BACKEND = "lost_and_found-backend-1"
COMPOSE = ["docker", "compose", "-f", str(ROOT / "docker-compose.yml")]

ITEMS = [
    {
        "file": "airpods.jpg",
        "url": "https://upload.wikimedia.org/wikipedia/commons/thumb/e/e2/Airpods_Pro_in_a_Case.jpg/960px-Airpods_Pro_in_a_Case.jpg",
        "category": "electronics",
        "title": "White AirPods Pro case",
        "description": "White Apple AirPods Pro charging case found on a cafeteria table. Looks new.",
        "unique_features": "white glossy case, AirPods Pro hinge, no sticker",
        "location": "cafeteria",
        "location_details": "table near the window",
        "incident_date": "2026-08-22",
        "phone": "+251911100001",
        "telegram": "campus_seed_airpods",
        "search": {
            "category": "Electronics",
            "title": "AirPods Pro case",
            "description": "I lost my white AirPods Pro case yesterday near the cafeteria. It is a white charging case.",
            "marks": "white glossy Apple case",
            "place": "Cafeteria",
            "details": "near the window",
            "date": "2026-08-21",
        },
    },
    {
        "file": "airpods.jpg",
        "url": "https://upload.wikimedia.org/wikipedia/commons/thumb/e/e2/Airpods_Pro_in_a_Case.jpg/960px-Airpods_Pro_in_a_Case.jpg",
        "category": "electronics",
        "title": "White AirPods case",
        "description": "White AirPods charging case found at the coffee shop. Looks like a regular AirPods case, not Pro.",
        "unique_features": "white case, small scuff on the front corner",
        "location": "coffee_shop",
        "location_details": "table by the door",
        "incident_date": "2026-08-22",
        "phone": "+251911100010",
        "telegram": "campus_seed_airpods_2",
        "search": {
            "category": "Electronics",
            "title": "AirPods Pro case",
            "description": "I lost my white AirPods Pro case yesterday near the cafeteria. It is a white charging case.",
            "marks": "white glossy Apple case",
            "place": "Cafeteria",
            "details": "near the window",
            "date": "2026-08-21",
        },
    },
    {
        "file": "charger.jpg",
        "url": "https://upload.wikimedia.org/wikipedia/commons/thumb/2/2d/Notebook-Computer-AC-Adapter.jpg/960px-Notebook-Computer-AC-Adapter.jpg",
        "category": "electronics",
        "title": "Laptop charger",
        "description": "Black laptop AC adapter and cable found under a lecture hall seat.",
        "unique_features": "black brick adapter, long cable, tape on the plug",
        "location": "lecture_hall",
        "location_details": "row 4, left side",
        "incident_date": "2026-08-22",
        "phone": "+251911100002",
        "telegram": "campus_seed_charger",
        "search": {
            "category": "Electronics",
            "title": "Laptop charger",
            "description": "Black laptop charger I left in the lecture hall. Brick adapter with a long cable.",
            "marks": "tape on the plug",
            "place": "Lecture hall",
            "details": "about row 4",
            "date": "2026-08-22",
        },
    },
    {
        "file": "backpack.jpg",
        "url": "https://upload.wikimedia.org/wikipedia/commons/thumb/8/8a/Arcteryx_Alpha_Fast_light_40_black_backpack.jpg/960px-Arcteryx_Alpha_Fast_light_40_black_backpack.jpg",
        "category": "bags",
        "title": "Black hiking backpack",
        "description": "Black technical backpack found near the library entrance.",
        "unique_features": "black, multiple zip pockets, hiking straps",
        "location": "library",
        "location_details": "near the entrance",
        "incident_date": "2026-08-22",
        "phone": "+251911100003",
        "telegram": "campus_seed_backpack",
        "search": {
            "category": "Bags",
            "title": "Black backpack",
            "description": "I lost my black backpack around the library. Dark daypack with zip pockets.",
            "marks": "hiking straps, several zippers",
            "place": "Library",
            "details": "near the entrance",
            "date": "2026-08-21",
        },
    },
    {
        "file": "student_id.png",
        "url": "https://upload.wikimedia.org/wikipedia/commons/thumb/0/09/Yuntech_Student_ID_Card_2017.png/960px-Yuntech_Student_ID_Card_2017.png",
        "category": "ids",
        "title": "Student ID card",
        "description": "Plastic student identity card found on a chair in the student center.",
        "unique_features": "photo ID, university logo, barcode on the back",
        "location": "student_center",
        "location_details": "lounge chairs",
        "incident_date": "2026-08-22",
        "phone": "+251911100004",
        "telegram": "campus_seed_id",
        "search": {
            "category": "IDs & Documents",
            "title": "Student ID",
            "description": "Lost my student ID card at the student center. Plastic card with my photo.",
            "marks": "barcode on the back",
            "place": "Student center",
            "details": "lounge area",
            "date": "2026-08-22",
        },
    },
    {
        "file": "cards.jpg",
        "url": "https://upload.wikimedia.org/wikipedia/commons/thumb/4/4f/Credit-cards.jpg/960px-Credit-cards.jpg",
        "category": "cards",
        "title": "Bank cards",
        "description": "Two bank cards found on a gym bench after evening training.",
        "unique_features": "chip cards, one blue one gold",
        "location": "gym",
        "location_details": "bench near the entrance",
        "incident_date": "2026-08-22",
        "phone": "+251911100005",
        "telegram": "campus_seed_cards",
        "search": {
            "category": "ATM & Bank Cards",
            "title": "Bank cards",
            "description": "I dropped my debit cards at the gym. One is blue and one is gold.",
            "marks": "chip cards",
            "place": "Gym",
            "details": "near the entrance bench",
            "date": "2026-08-22",
        },
    },
    {
        "file": "keys.jpg",
        "url": "https://upload.wikimedia.org/wikipedia/commons/thumb/b/ba/Car_keys.jpg/960px-Car_keys.jpg",
        "category": "keys",
        "title": "Car keys",
        "description": "Set of car keys found in the parking lot next to a silver sedan.",
        "unique_features": "black remote fob, metal key ring",
        "location": "parking",
        "location_details": "south lot, near the lamp post",
        "incident_date": "2026-08-22",
        "phone": "+251911100006",
        "telegram": "campus_seed_keys",
        "search": {
            "category": "Keys",
            "title": "Car keys",
            "description": "Lost my car keys in the parking lot. Black remote fob on a metal ring.",
            "marks": "black fob",
            "place": "Parking",
            "details": "south lot",
            "date": "2026-08-22",
        },
    },
    {
        "file": "hoodie.jpg",
        "url": "https://upload.wikimedia.org/wikipedia/commons/thumb/8/8c/Blue_graphic_hoodie_%2852774355054%29.jpg/960px-Blue_graphic_hoodie_%2852774355054%29.jpg",
        "category": "clothing",
        "title": "Blue graphic hoodie",
        "description": "Blue hoodie with a graphic print left in a dormitory common room.",
        "unique_features": "bright blue, graphic on the chest",
        "location": "dormitory",
        "location_details": "ground floor common room",
        "incident_date": "2026-08-22",
        "phone": "+251911100007",
        "telegram": "campus_seed_hoodie",
        "search": {
            "category": "Clothing",
            "title": "Blue hoodie",
            "description": "I left my blue graphic hoodie in the dorm common room.",
            "marks": "graphic print on the chest",
            "place": "Dormitory",
            "details": "ground floor common room",
            "date": "2026-08-21",
        },
    },
    {
        "file": "ring.jpg",
        "url": "https://upload.wikimedia.org/wikipedia/commons/thumb/1/17/Silver_ring_and_scorzoneras_by_ASQ.jpg/960px-Silver_ring_and_scorzoneras_by_ASQ.jpg",
        "category": "jewelry",
        "title": "Silver ring",
        "description": "Small silver ring found beside a coffee shop cup.",
        "unique_features": "plain silver band",
        "location": "coffee_shop",
        "location_details": "table by the door",
        "incident_date": "2026-08-22",
        "phone": "+251911100008",
        "telegram": "campus_seed_ring",
        "search": {
            "category": "Jewelry",
            "title": "Silver ring",
            "description": "I lost a plain silver ring at the coffee shop.",
            "marks": "plain silver band",
            "place": "Coffee shop",
            "details": "near the door",
            "date": "2026-08-22",
        },
    },
    {
        "file": "book.jpg",
        "url": "https://upload.wikimedia.org/wikipedia/commons/8/8f/Open_book.jpg",
        "category": "books",
        "title": "Open textbook",
        "description": "Hardcover textbook left open on a library desk.",
        "unique_features": "thick hardcover, notes in the margin",
        "location": "library",
        "location_details": "second floor quiet zone",
        "incident_date": "2026-08-22",
        "phone": "+251911100009",
        "telegram": "campus_seed_book",
        "search": {
            "category": "Books",
            "title": "Textbook",
            "description": "Lost my hardcover textbook in the library quiet zone. It has notes in the margin.",
            "marks": "handwritten notes in the margin",
            "place": "Library",
            "details": "second floor",
            "date": "2026-08-22",
        },
    },
]


def download(name: str, url: str) -> Path:
    dest = IMG_DIR / name
    if dest.exists() and dest.stat().st_size > 2000:
        print(f"using cached {name}")
        return dest
    req = urllib.request.Request(url, headers=UA)
    with urllib.request.urlopen(req, timeout=40) as res:
        data = res.read()
    if len(data) < 2000:
        raise RuntimeError(f"{name} too small ({len(data)} bytes)")
    dest.write_bytes(data)
    print(f"downloaded {name} ({len(data)} bytes)")
    return dest


def sql_str(value: str) -> str:
    return "'" + value.replace("'", "''") + "'"


def psql(sql: str) -> None:
    subprocess.run(
        COMPOSE + ["exec", "-T", "postgres", "psql", "-U", "nemma", "-d", "lost_and_found", "-v", "ON_ERROR_STOP=1"],
        input=sql.encode(),
        check=True,
    )


def main() -> None:
    psql("TRUNCATE TABLE matches, reports RESTART IDENTITY CASCADE;")
    print("database cleared\n")
    print("How to search (I lost something):\n")
    for item in ITEMS:
        local = download(item["file"], item["url"])
        ext = local.suffix.lower()
        stored = f"{uuid.uuid4()}{ext}"
        rid = str(uuid.uuid4())
        subprocess.check_call(["docker", "cp", str(local), f"{BACKEND}:/app/uploads/{stored}"])
        photos = json.dumps([stored])
        sql = f"""
INSERT INTO reports (
  id, type, category, title, description, unique_features,
  location, location_details, incident_date, photos, phone, telegram, status, created_at
) VALUES (
  {sql_str(rid)}::uuid, 'found', {sql_str(item['category'])}, {sql_str(item['title'])},
  {sql_str(item['description'])}, {sql_str(item['unique_features'])},
  {sql_str(item['location'])}, {sql_str(item['location_details'])}, DATE {sql_str(item['incident_date'])},
  {sql_str(photos)}::jsonb, {sql_str(item['phone'])}, {sql_str(item['telegram'])},
  'unclaimed', NOW()
);
"""
        psql(sql)
        s = item["search"]
        print(f"- {item['title']}")
        print(f"  Category: {s['category']}")
        print(f"  Title: {s['title']}")
        print(f"  Description: {s['description']}")
        print(f"  Unique marks: {s['marks']}")
        print(f"  Place: {s['place']} / {s['details']}")
        print(f"  Date lost: {s['date']}")
        print(f"  Finder Telegram: @{item['telegram']}")
        print()
    print("Seed complete. Open http://localhost:3000/lost")


if __name__ == "__main__":
    main()
