# Hack4Arad – date turistice Arad

Set de date deschis despre obiectivele turistice, cazarea, gastronomia și traseele din județul Arad,
pregătit pentru participanții la **[Hack4Arad](https://hack4arad.uav.ro/)** (Universitatea „Aurel Vlaicu”, Arad, 7–8 noiembrie 2026).

Fiecare înregistrare păstrează **proveniența**: linkul către fișa sursă de pe [visitarad.ro](https://www.visitarad.ro/),
data colectării și sursa.

> 🇬🇧 Open dataset of tourist attractions, accommodation, food & drink and trails in Arad County, Romania,
> collected from the public site of the county tourist information centre (CNIPT Arad) for the Hack4Arad hackathon.
> Every record carries its source URL. See [Sursă și condiții de utilizare](#sursă-și-condiții-de-utilizare).

## Ce conține

| Fișier | Conținut |
|---|---|
| [`data/locations.csv`](data/locations.csv) / [`.json`](data/locations.json) | Locații: nume, tip, categorie, localitate, adresă, coordonate GPS, descriere, site/Facebook, URL sursă |
| [`data/trails.csv`](data/trails.csv) / [`.json`](data/trails.json) | Trasee turistice: nume, categorie, descriere, imaginea hărții, URL sursă |
| [`scripts/scrape.py`](scripts/scrape.py) | Scriptul care a produs datele (doar biblioteca standard Python, reproductibil) |

Statistici și coloane: vezi [`DATA.md`](DATA.md).

## Pornire rapidă

```python
import csv
rows = list(csv.DictReader(open("data/locations.csv", encoding="utf8")))
biserici = [r for r in rows if r["category"] == "Biserici"]
print(len(rows), "locații,", len(biserici), "biserici")
```

```js
const locations = await (await fetch(
  "https://raw.githubusercontent.com/baditaflorin/hack4arad-turism-arad/main/data/locations.json"
)).json();
```

Idei pentru proiecte: hartă interactivă, itinerar generat după timp/interese, ghid audio, căutare în română/engleză,
statistici despre acoperirea turistică pe localități, verificare a calității datelor (coordonate lipsă, descrieri incomplete).

## Reproducere / actualizare

```bash
python3 scripts/scrape.py          # ~6 minute, 1 cerere/secundă, scrie în data/
```

Scriptul respectă `robots.txt` (permite tot), se identifică cu un User-Agent cu date de contact și face cache local
(`.cache/`, ignorat de git).

## Sursă și condiții de utilizare

- **Sursa:** Centrul Național de Informare și Promovare Turistică al Județului Arad (CNIPT), Consiliul Județean Arad, [visitarad.ro](https://www.visitarad.ro/).
- **Acord:** am cerut oficial (email, 29 octombrie 2026) permisiunea de a folosi și redistribui un export structurat pentru
  participanții Hack4Arad. Pe 3 noiembrie 2026 CNIPT a răspuns că site-ul nu are feed/export, dar că informațiile de pe site
  sunt publice și pot fi preluate.
- **Atribuire obligatorie:** în orice proiect care folosește datele, menționați
  *„Date: CNIPT Arad / visitarad.ro”* și păstrați `source_url` pentru fiecare înregistrare.
- **Fără licență explicită:** CNIPT nu a indicat o licență formală (ex. CC-BY). Textele, imaginile și logo-urile rămân ale
  proprietarilor lor. Repo-ul **nu** conține imaginile (doar linkuri către ele). Pentru utilizare comercială sau după hackathon,
  confirmați condițiile direct cu CNIPT.
- **Date personale:** nu includem telefoane sau adrese de email ale unităților (le găsiți pe fișa sursă).
- **Nu este un produs oficial** al CNIPT sau al Consiliului Județean Arad. Datele pot conține erori sau pot fi depășite,
  verificați pe fișa sursă.
- Codul din `scripts/` este sub licență MIT ([`LICENSE`](LICENSE)); licența nu se aplică datelor.

## Contact

Florin Bădiță, mentor voluntar Hack4Arad. Corecturi sau cereri de eliminare: deschideți un issue.
