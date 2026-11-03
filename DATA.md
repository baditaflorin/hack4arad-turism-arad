# Descrierea datelor

Colectat: 2026-11-03T09:40:14Z de pe [visitarad.ro](https://www.visitarad.ro/ro/sitemap) (toate cele 325 locații din sitemap-ul site-ului și 30 trasee).

## `locations.csv` / `locations.json` (325 înregistrări)

| Coloană | Descriere |
|---|---|
| `id` | ID-ul din visitarad.ro (stabil) |
| `name` | Denumirea |
| `type` | Tip: Obiective turistice / Timp liber și agrement / Cazare |
| `category` | Categorie (ex. Biserici, Pensiuni, Cetăți și castele) |
| `locality` | Localitatea |
| `area` | Zona turistică (ex. Arad metropolitan, Podgoria Miniș-Măderat) |
| `address` | Adresa, așa cum e scrisă pe site |
| `latitude`, `longitude` | Coordonate WGS84 |
| `description` | Textul de prezentare (română, necorectat) |
| `website`, `facebook` | Linkuri ale obiectivului, când există |
| `photo_count` (CSV) / `photos` (JSON) | Număr / linkuri către imaginile de pe visitarad.ro (imaginile nu sunt în repo) |
| `source_url` | **Fișa sursă pe visitarad.ro** |
| `category_url` | Pagina categoriei pe visitarad.ro |
| `retrieved_at`, `source` | Proveniența |

JSON-ul are și `slug` și `marker` (tipul pinului de pe hartă).

### Tipuri
| Tip | Număr |
|---|---|
| Obiective turistice | 163 |
| Timp liber și agrement | 98 |
| Cazare | 64 |

### Categorii
| Categorie | Număr |
|---|---|
| Pensiuni | 32 |
| Clădiri reprezentative | 30 |
| Biserici | 30 |
| Hoteluri | 26 |
| Obiective turistice de pe Drumul Vinului | 24 |
| Monumente | 24 |
| Restaurant | 22 |
| Muzee și Case memoriale | 21 |
| Cafenea | 18 |
| Cofetărie | 11 |
| Cetăți și castele | 10 |
| Pizzerie | 10 |
| Crama | 10 |
| Stațiunea Moneasa | 9 |
| Fast food | 7 |
| Ștranduri | 6 |
| Agrement și relaxare | 5 |
| Camping | 5 |
| Formațiuni naturale | 4 |
| Pub | 4 |
| Vestigii arheologice | 3 |
| Cinema | 3 |
| Lacul Ghioroc | 2 |
| Săgeata Verde | 2 |
| Bistro | 2 |
| Motel | 1 |
| Clubbing | 1 |
| Parcul Natural Lunca Mureșului | 1 |
| Băile Lipova | 1 |
| Teatru | 1 |

### Localități (top 15)
| Localitate | Număr |
|---|---|
| Arad | 167 |
| Moneasa | 17 |
| Lipova | 16 |
| Ghioroc | 12 |
| Ineu | 10 |
| Șiria | 10 |
| Covăsinț | 5 |
| Păuliș | 4 |
| Macea | 4 |
| Miniș | 4 |
| Vladimirescu | 4 |
| Săvârșin | 3 |
| Pecica | 3 |
| Căsoaia | 3 |
| Pâncota | 3 |

## `trails.csv` / `trails.json` (30 înregistrări)

`id`, `name`, `category`, `description`, `map_image_url` (imagine cu harta traseului, când există), `photos` (JSON) / `photo_count` (CSV), `source_url`, `retrieved_at`, `source`.
Un traseu nu apare în nicio pagină de categorie pe site, deci are `category` gol.

## Limitări cunoscute (provin din sursă, nu au fost corectate)

- Descriere lipsă: 149 din 325 locații; adresă lipsă: 88; zonă (`area`) lipsă: 16.
- `area` nu e unitar pe site: „Munții Codru Moma” și „Munții Codru-Moma” apar ambele.
- Coordonatele sunt cele de pe site; nu au fost verificate individual (doar încadrare grosieră în județul Arad, inclusiv Nădlac, în vest).
- Unele texte nu au diacritice sau au greșeli de redactare.
- Telefoanele și adresele de email ale unităților nu sunt incluse intenționat (nici în câmpuri, nici în textul descrierilor: apar ca `[telefon pe fișa sursă]`).
- Pagina de prezentare e doar în română în acest export (site-ul are și versiuni EN/DE).
