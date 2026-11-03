#!/usr/bin/env python3
"""Colectează obiectivele turistice publice de pe visitarad.ro (CNIPT Arad).

Politicos: 1 cerere/secundă, User-Agent cu contact, cache local al HTML-ului brut.
Doar bibliotecă standard (Python 3.8+).

Utilizare:  python3 scripts/scrape.py [--cache DIR] [--delay 1.0]
Scrie:      data/locations.{json,csv}, data/trails.{json,csv}
"""
import argparse, csv, html, json, re, sys, time, urllib.request
from datetime import datetime, timezone
from pathlib import Path

BASE = "https://www.visitarad.ro"
SOURCE = "Centrul Național de Informare și Promovare Turistică al Județului Arad (CNIPT), visitarad.ro"
UA = "Hack4Arad-dataset/1.0 (+https://github.com/baditaflorin/hack4arad-turism-arad; florin@badita.org)"
ROOT = Path(__file__).resolve().parent.parent


def fetch(url, cache, delay):
    f = cache / (re.sub(r"[^A-Za-z0-9._-]", "_", url.replace(BASE, "")) + ".html")
    if f.exists():
        return f.read_text(encoding="utf8")
    for attempt in range(3):
        try:
            req = urllib.request.Request(url, headers={"User-Agent": UA})
            with urllib.request.urlopen(req, timeout=30) as r:
                body = r.read().decode("utf8")
            f.write_text(body, encoding="utf8")
            time.sleep(delay)
            return body
        except Exception as e:  # noqa: BLE001
            print(f"  retry {attempt+1} {url}: {e}", file=sys.stderr)
            time.sleep(3 * (attempt + 1))
    raise RuntimeError(f"nu am putut descărca {url}")


def text(s):
    s = re.sub(r"<br\s*/?>", "\n", s or "")
    s = re.sub(r"<[^>]+>", " ", s)
    s = html.unescape(s)
    s = re.sub(r"[ \t\r\f\v\xa0]+", " ", s)
    s = re.sub(r"\s*\n\s*", "\n", s)
    return s.strip()


PHONE = re.compile(r"(?:\+?40[ .\-]?|0)[27]\d{2}[ .\-/]?\d{3}[ .\-/]?\d{3}|(?:\+?40[ .\-]?|0)[27]\d{8}")


def redact_phones(s):
    """Numerele de telefon din texte nu sunt incluse (se găsesc pe fișa sursă)."""
    return PHONE.sub("[telefon pe fișa sursă]", s) if s else s


def attr(h, name):
    m = re.search(r'data-%s="([^"]*)"' % name, h)
    return html.unescape(m.group(1)) if m else None


def parse(h, url):
    h = re.sub(r"<!--.*?-->", "", h, flags=re.S)
    r = {"id": int(attr(h, "id") or 0), "name": attr(h, "title"), "source_url": url}
    r["slug"] = url.rsplit("/", 1)[-1]
    r["latitude"] = float(attr(h, "latitude")) if attr(h, "latitude") else None
    r["longitude"] = float(attr(h, "longitude")) if attr(h, "longitude") else None
    r["marker"] = (attr(h, "marker") or "").replace(".svg", "")
    # breadcrumb: tip › categorie
    m = re.search(r'<div class="subtitle-category">(.*?)</div>', h, re.S)
    crumbs = [text(a) for a in re.findall(r"<a [^>]*>(.*?)</a>", m.group(1), re.S)] if m else []
    r["type"] = crumbs[0] if crumbs else None
    r["category"] = crumbs[1] if len(crumbs) > 1 else None
    r["category_url"] = None
    if m:
        urls = re.findall(r'<a href="([^"]+)"', m.group(1))
        r["category_url"] = urls[1] if len(urls) > 1 else None
    m = re.search(r"Localizare:\s*(.*?)</div>", h, re.S)
    loc = [x.strip() for x in text(m.group(1)).replace("\n", " ").split(" - ")] if m else []
    r["locality"] = loc[0] if loc else None
    r["area"] = loc[1] if len(loc) > 1 else None
    # secțiuni cu <h6>Titlu</h6> ... până la următorul h6 / item-contact
    d = re.search(r'<div class="map-item__description.*?>(.*?)<div class="mx-5">', h, re.S)
    body = d.group(1) if d else ""
    secs = {}
    for t, c in re.findall(r"<h6[^>]*>(.*?)</h6>(.*?)(?=<h6|\Z)", body, re.S):
        secs[text(t)] = c
    r["description"] = redact_phones(text(secs.get("Prezentare", ""))) or None
    r["address"] = text(secs.get("Adresă", "")).replace("\n", " ") or None
    # contact: păstrăm doar site/facebook ale obiectivului; NU telefon/email (date de contact, nu sunt necesare)
    contact = secs.get("Contact", "")
    links = re.findall(r'<a\s+href="(https?://[^"]+)"[^>]*rel="nofollow"', contact)  # noqa
    r["website"] = next((u for u in links if "facebook.com" not in u), None)
    r["facebook"] = next((u for u in links if "facebook.com" in u), None)
    r["photos"] = sorted({u for u in re.findall(r'class="swipebox" href="([^"]+)"', h)})
    return r


def parse_trail(h, url, category):
    h = re.sub(r"<!--.*?-->", "", h, flags=re.S)
    t = re.search(r'<h1 class="title">(.*?)</h1>', h, re.S)
    d = re.search(r'<div class="trail-description">(.*?)</div>', h, re.S)
    mp = re.search(r'href="([^"]*/storage/maps/[^"]+)"', h)
    slug = url.rsplit("/", 1)[-1]
    return {
        "id": int(slug.split("-")[0]), "name": text(t.group(1)) if t else None, "category": category,
        "description": redact_phones(text(d.group(1))) if d else None, "map_image_url": mp.group(1) if mp else None,
        "photos": sorted({u for u in re.findall(r'class="swipebox" href="([^"]+)"', h)}),
        "source_url": url, "slug": slug,
    }


def scrape_trails(cache, delay, now, sitemap):
    idx = fetch(BASE + "/ro/trasee-turistice", cache, delay)
    cats = sorted(set(re.findall(r'href="(https://www\.visitarad\.ro/ro/trasee-turistice/[^"]+)"', idx)))
    by_url = {}
    for c in cats:
        ch = fetch(c, cache, delay)
        name = text(re.search(r'<h1 class="title">(.*?)</h1>', ch, re.S).group(1)) if re.search(r'<h1 class="title">', ch) else c.rsplit("/", 1)[-1]
        for u in set(re.findall(r'href="(https://www\.visitarad\.ro/ro/traseu-turistic/[^"]+)"', ch)):
            by_url.setdefault(u, name)
    for u in re.findall(r'href="(https://www\.visitarad\.ro/ro/traseu-turistic/[^"]+)"', sitemap):
        by_url.setdefault(u, None)  # trasee care apar doar în sitemap, fără pagină de categorie
    out = []
    for u in sorted(by_url, key=lambda u: int(re.search(r"/(\d+)-", u).group(1))):
        r = parse_trail(fetch(u, cache, delay), u, by_url[u])
        r["retrieved_at"] = now
        r["source"] = SOURCE
        out.append(r)
    (ROOT / "data/trails.json").write_text(json.dumps(out, ensure_ascii=False, indent=1) + "\n", encoding="utf8")
    cols = ["id", "name", "category", "description", "map_image_url", "photo_count", "source_url", "retrieved_at", "source"]
    with open(ROOT / "data/trails.csv", "w", newline="", encoding="utf8") as f:
        w = csv.DictWriter(f, cols, extrasaction="ignore"); w.writeheader()
        for r in out:
            w.writerow({**r, "photo_count": len(r["photos"])})
    print("trasee:", len(out))


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--cache", default=str(ROOT / ".cache"))
    ap.add_argument("--delay", type=float, default=1.0)
    a = ap.parse_args()
    cache = Path(a.cache); cache.mkdir(parents=True, exist_ok=True)
    now = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
    sm = fetch(BASE + "/ro/sitemap", cache, a.delay)
    urls = sorted(set(re.findall(r'href="(https://www\.visitarad\.ro/ro/harta-locatie/[^"]+)"', sm)),
                  key=lambda u: int(re.search(r"/(\d+)-", u).group(1)))
    print(f"{len(urls)} locații în sitemap")
    out = []
    for i, u in enumerate(urls, 1):
        rec = parse(fetch(u, cache, a.delay), u)
        rec["retrieved_at"] = now
        rec["source"] = SOURCE
        out.append(rec)
        if i % 25 == 0: print(f"  {i}/{len(urls)}")
    (ROOT / "data").mkdir(exist_ok=True)
    (ROOT / "data/locations.json").write_text(json.dumps(out, ensure_ascii=False, indent=1) + "\n", encoding="utf8")
    cols = ["id", "name", "type", "category", "locality", "area", "address", "latitude", "longitude",
            "description", "website", "facebook", "photo_count", "source_url", "category_url", "retrieved_at", "source"]
    with open(ROOT / "data/locations.csv", "w", newline="", encoding="utf8") as f:
        w = csv.DictWriter(f, cols, extrasaction="ignore"); w.writeheader()
        for r in out:
            w.writerow({**r, "photo_count": len(r["photos"])})
    print("gata:", len(out))
    scrape_trails(cache, a.delay, now, sm)


if __name__ == "__main__":
    main()
