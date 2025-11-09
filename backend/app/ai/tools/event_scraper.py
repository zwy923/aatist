import requests

API_URL = "https://www.aalto.fi/en/aalto_api/events/list?_format=json&language=en"
BASE_URL = "https://www.aalto.fi"

def fetch_aalto_events():
    res = requests.get(API_URL, headers={"User-Agent": "Mozilla/5.0"})
    res.raise_for_status()
    data = res.json()

    events_data = data.get("data", {}).get("events", [])
    events = []

    for item in events_data:
        title = item.get("name", "")
        url = item.get("url", "")
        if url.startswith("/"):
            url = BASE_URL + url

        desc = item.get("description", "")
        location = item.get("location", "")
        category_list = item.get("category", [])
        category = ", ".join(c.get("label", "") for c in category_list if isinstance(c, dict))

        image = item.get("main_image") or item.get("image", "")

        event_dates = item.get("event_dates", {})
        start = event_dates.get("start", "")
        end = event_dates.get("end", "")

        events.append({
            "title": title,
            "link": url,
            "description": desc,
            "location": location or "N/A",
            "category": category,
            "start_date": start,
            "end_date": end,
            "image": image,
        })

    return events


if __name__ == "__main__":
    events = fetch_aalto_events()
    print(f" 抓取到 {len(events)} 条活动：\n")
    for e in events[:5]:
        print(f"- {e['title']}")
        print(f"  Time: {e['start_date']} — {e['end_date']}")
        print(f"  Addr: {e['location']} | {e['category']}")
        print(f"  Link: {e['link']}\n  {e['description'][:100]}...\n")
