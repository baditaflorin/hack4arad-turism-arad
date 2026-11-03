# Hack4Arad – date turistice Arad

**🌐 Pagina web (de citit, hartă, descărcare, ghid pentru începători, info pentru presă): https://baditaflorin.github.io/hack4arad-turism-arad/**

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
| [`data/locations.csv`](data/locations.csv) / [`.json`](data/locations.json) / [`-excel.csv`](data/locations-excel.csv) | Locații: nume, tip, categorie, localitate, adresă, coordonate GPS, descriere, site/Facebook, URL sursă |
| [`data/trails.csv`](data/trails.csv) / [`.json`](data/trails.json) / [`-excel.csv`](data/trails-excel.csv) | Trasee turistice: nume, categorie, descriere, imaginea hărții, URL sursă |
| [`cmd/scrape`](cmd/scrape/main.go) | Programul în Go care a produs datele (doar biblioteca standard, reproductibil) |

`*-excel.csv` = același CSV, cu marcaj UTF-8 pentru ca Excel să afișeze corect diacriticele.
Statistici și coloane: vezi [`DATA.md`](DATA.md).

## Pornire rapidă

```js
const locations = await (await fetch(
  "https://raw.githubusercontent.com/baditaflorin/hack4arad-turism-arad/main/data/locations.json"
)).json();
console.log(locations.length, "locații");
```

```go
// go run main.go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Location struct {
	Name     string  `json:"name"`
	Category *string `json:"category"`
	Locality *string `json:"locality"`
}

func main() {
	resp, _ := http.Get("https://raw.githubusercontent.com/baditaflorin/hack4arad-turism-arad/main/data/locations.json")
	defer resp.Body.Close()
	var locs []Location
	json.NewDecoder(resp.Body).Decode(&locs)
	fmt.Println(len(locs), "locații; prima:", locs[0].Name)
}
```

Idei pentru proiecte: hartă interactivă, itinerar generat după timp/interese, ghid audio, căutare în română/engleză,
statistici despre acoperirea turistică pe localități, verificare a calității datelor (coordonate lipsă, descrieri incomplete).

## Linkuri directe (deep links)

Pagina web își păstrează starea în adresă, deci orice hartă sau filtru se poate trimite ca link:

| Link | Ce deschide |
|---|---|
| `#harta` | Harta cu toate locațiile, cu pini colorați pe categorii |
| `#harta?cat=Biserici` | Doar bisericile |
| `#harta?loc=Lipova` | Locațiile din Lipova |
| `#harta?ascunde=cazare,mancare` | Harta fără cazare și mâncare |
| `#loc/10` | O locație anume (ID din CSV/JSON), deschisă pe hartă |
| `#carti?q=vin`, `#tabel?tip=Cazare` | Vederea cărți / tabel cu filtre |
| `#trasee?q=Găina` | Trasee filtrate |
| `#incepatori`, `#presa`, `#descarca`, `#sursa` | Secțiunile paginii |

Parametri: `q`, `tip`, `cat`, `loc`, `ascunde` (grupuri de pini: `cazare`, `biserici`, `patrimoniu`, `muzee`, `natura`, `ape`, `vin`, `mancare`, `agrement`), `id`.
Pe pagină, butonul „🔗 Copiază linkul” și butoanele „🔗 Link” de pe fiecare card dau linkul gata făcut.

## Reproducere / actualizare

Scriptul de colectare este scris în **Go** (1.21+, doar biblioteca standard):

```bash
go run ./cmd/scrape          # ~6 minute, 1 cerere/secundă, scrie în data/
go run ./cmd/scrape -h       # opțiuni: -delay, -cache, -out
```

Respectă `robots.txt` (permite tot), se identifică cu un User-Agent cu date de contact și face cache local
(`.cache/`, ignorat de git), deci o a doua rulare nu mai încarcă site-ul.

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
- Codul din `cmd/` este sub licență MIT ([`LICENSE`](LICENSE)); licența nu se aplică datelor.

## Contact

Florin Bădiță, mentor voluntar Hack4Arad. Corecturi sau cereri de eliminare: deschideți un issue.
