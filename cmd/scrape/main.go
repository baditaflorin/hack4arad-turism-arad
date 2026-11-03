// Comanda scrape colectează obiectivele turistice publice de pe visitarad.ro (CNIPT Arad).
//
// Politicos: 1 cerere/secundă, User-Agent cu contact, cache local al HTML-ului brut.
// Doar biblioteca standard Go.
//
//	go run ./cmd/scrape                 # scrie în data/
//	go run ./cmd/scrape -delay 2s -cache /tmp/arad-cache
package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	base   = "https://www.visitarad.ro"
	source = "Centrul Național de Informare și Promovare Turistică al Județului Arad (CNIPT), visitarad.ro"
	ua     = "Hack4Arad-dataset/1.0 (+https://github.com/baditaflorin/hack4arad-turism-arad; florin@badita.org)"
)

// Location este o locație (obiectiv, cazare, gastronomie, agrement). Ordinea câmpurilor = ordinea din JSON.
type Location struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	SourceURL   string   `json:"source_url"`
	Slug        string   `json:"slug"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
	Marker      string   `json:"marker"`
	Type        *string  `json:"type"`
	Category    *string  `json:"category"`
	CategoryURL *string  `json:"category_url"`
	Locality    *string  `json:"locality"`
	Area        *string  `json:"area"`
	Description *string  `json:"description"`
	Address     *string  `json:"address"`
	Website     *string  `json:"website"`
	Facebook    *string  `json:"facebook"`
	Photos      []string `json:"photos"`
	RetrievedAt string   `json:"retrieved_at"`
	Source      string   `json:"source"`
}

// Trail este un traseu turistic.
type Trail struct {
	ID          int      `json:"id"`
	Name        *string  `json:"name"`
	Category    *string  `json:"category"`
	Description *string  `json:"description"`
	MapImageURL *string  `json:"map_image_url"`
	Photos      []string `json:"photos"`
	SourceURL   string   `json:"source_url"`
	Slug        string   `json:"slug"`
	RetrievedAt string   `json:"retrieved_at"`
	Source      string   `json:"source"`
}

type fetcher struct {
	cache string
	delay time.Duration
	hc    *http.Client
}

var unsafe = regexp.MustCompile(`[^A-Za-z0-9._-]`)

func (f *fetcher) get(url string) (string, error) {
	path := filepath.Join(f.cache, unsafe.ReplaceAllString(strings.Replace(url, base, "", 1), "_")+".html")
	if b, err := os.ReadFile(path); err == nil {
		return string(b), nil
	}
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		body, err := f.download(url)
		if err == nil {
			if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
				return "", err
			}
			time.Sleep(f.delay)
			return body, nil
		}
		lastErr = err
		log.Printf("  reîncerc %d/3 %s: %v", attempt, url, err)
		time.Sleep(time.Duration(attempt) * 3 * time.Second)
	}
	return "", fmt.Errorf("nu am putut descărca %s: %w", url, lastErr)
}

func (f *fetcher) download(url string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", ua)
	resp, err := f.hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	return string(b), err
}

var (
	reComment  = regexp.MustCompile(`(?s)<!--.*?-->`)
	reBR       = regexp.MustCompile(`<br\s*/?>`)
	reTag      = regexp.MustCompile(`<[^>]+>`)
	reSpaces   = regexp.MustCompile(`[ \t\r\f\v\x{a0}]+`)
	reNewlines = regexp.MustCompile(`[ \t\r\f\v\x{a0}]*\n\s*`)
	reTrailNL  = regexp.MustCompile(`\s+\n`)
	rePhone    = regexp.MustCompile(`(?:\+?40[ .\-]?|0)[27]\d{2}[ .\-/]?\d{3}[ .\-/]?\d{3}|(?:\+?40[ .\-]?|0)[27]\d{8}`)
)

// text transformă un fragment HTML în text simplu, păstrând rândurile noi din <br>.
func text(s string) string {
	s = reBR.ReplaceAllString(s, "\n")
	s = reTag.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	s = reSpaces.ReplaceAllString(s, " ")
	s = reTrailNL.ReplaceAllString(s, "\n")
	s = reNewlines.ReplaceAllString(s, "\n")
	return strings.TrimSpace(s)
}

// redactPhones: numerele de telefon din texte nu sunt incluse (se găsesc pe fișa sursă).
func redactPhones(s string) string {
	return rePhone.ReplaceAllString(s, "[telefon pe fișa sursă]")
}

func nz(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func dataAttr(h, name string) string {
	m := regexp.MustCompile(`data-` + name + `="([^"]*)"`).FindStringSubmatch(h)
	if m == nil {
		return ""
	}
	return html.UnescapeString(m[1])
}

var (
	reCrumbBox   = regexp.MustCompile(`(?s)<div class="subtitle-category">(.*?)</div>`)
	reCrumbText  = regexp.MustCompile(`(?s)<a [^>]*>(.*?)</a>`)
	reCrumbHref  = regexp.MustCompile(`<a href="([^"]+)"`)
	reLocalizare = regexp.MustCompile(`(?s)Localizare:\s*(.*?)</div>`)
	reDescStart  = regexp.MustCompile(`(?s)<div class="map-item__description[^>]*>`)
	reDescEnd    = regexp.MustCompile(`<div class="mx-5">|<div id="product-photos"|</section>`)
	reH6         = regexp.MustCompile(`(?s)<h6[^>]*>(.*?)</h6>`)
	reNofollow   = regexp.MustCompile(`(?s)<a\s+href="(https?://[^"]+)"[^>]*rel="nofollow"`)
	rePhotos     = regexp.MustCompile(`class="swipebox" href="([^"]+)"`)
	reH1         = regexp.MustCompile(`(?s)<h1 class="title">(.*?)</h1>`)
	reTrailDesc  = regexp.MustCompile(`(?s)<div class="trail-description">(.*?)</div>`)
	reTrailMap   = regexp.MustCompile(`href="([^"]*/storage/maps/[^"]+)"`)
	reIDPrefix   = regexp.MustCompile(`/(\d+)-`)
)

func photos(h string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, m := range rePhotos.FindAllStringSubmatch(h, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			out = append(out, m[1])
		}
	}
	sort.Strings(out)
	return out
}

func parseLocation(h, url string) Location {
	h = reComment.ReplaceAllString(h, "")
	id, _ := strconv.Atoi(dataAttr(h, "id"))
	r := Location{ID: id, Name: dataAttr(h, "title"), SourceURL: url, Source: source}
	r.Slug = url[strings.LastIndex(url, "/")+1:]
	if v, err := strconv.ParseFloat(dataAttr(h, "latitude"), 64); err == nil {
		r.Latitude = &v
	}
	if v, err := strconv.ParseFloat(dataAttr(h, "longitude"), 64); err == nil {
		r.Longitude = &v
	}
	r.Marker = strings.Replace(dataAttr(h, "marker"), ".svg", "", 1)

	if m := reCrumbBox.FindStringSubmatch(h); m != nil {
		var crumbs []string
		for _, c := range reCrumbText.FindAllStringSubmatch(m[1], -1) {
			crumbs = append(crumbs, text(c[1]))
		}
		hrefs := reCrumbHref.FindAllStringSubmatch(m[1], -1)
		if len(crumbs) > 0 {
			r.Type = nz(crumbs[0])
		}
		if len(crumbs) > 1 {
			r.Category = nz(crumbs[1])
		}
		if len(hrefs) > 1 {
			r.CategoryURL = nz(hrefs[1][1])
		}
	}
	if m := reLocalizare.FindStringSubmatch(h); m != nil {
		parts := strings.Split(strings.ReplaceAll(text(m[1]), "\n", " "), " - ")
		r.Locality = nz(strings.TrimSpace(parts[0]))
		if len(parts) > 1 {
			r.Area = nz(strings.TrimSpace(parts[1]))
		}
	}

	// Blocul de descriere: de la container până la galeria de poze (sau sfârșitul secțiunii,
	// pentru paginile fără poze). Secțiuni: <h6>Titlu</h6> + conținut până la următorul <h6>.
	body := ""
	if loc := reDescStart.FindStringIndex(h); loc != nil {
		rest := h[loc[1]:]
		if end := reDescEnd.FindStringIndex(rest); end != nil {
			rest = rest[:end[0]]
		}
		body = rest
	}
	secs := map[string]string{}
	idx := reH6.FindAllStringSubmatchIndex(body, -1)
	for i, m := range idx {
		end := len(body)
		if i+1 < len(idx) {
			end = idx[i+1][0]
		}
		secs[text(body[m[2]:m[3]])] = body[m[1]:end]
	}
	if d := redactPhones(text(secs["Prezentare"])); d != "" {
		r.Description = &d
	}
	r.Address = nz(strings.ReplaceAll(text(secs["Adresă"]), "\n", " "))
	// Contact: păstrăm doar site/Facebook ale obiectivului; NU telefon/email.
	for _, m := range reNofollow.FindAllStringSubmatch(secs["Contact"], -1) {
		switch {
		case strings.Contains(m[1], "facebook.com"):
			if r.Facebook == nil {
				r.Facebook = nz(m[1])
			}
		case r.Website == nil:
			r.Website = nz(m[1])
		}
	}
	r.Photos = photos(h)
	return r
}

func parseTrail(h, url string, category *string) Trail {
	h = reComment.ReplaceAllString(h, "")
	slug := url[strings.LastIndex(url, "/")+1:]
	id, _ := strconv.Atoi(slug[:strings.Index(slug, "-")])
	t := Trail{ID: id, Category: category, SourceURL: url, Slug: slug, Source: source, Photos: photos(h)}
	if m := reH1.FindStringSubmatch(h); m != nil {
		t.Name = nz(text(m[1]))
	}
	if m := reTrailDesc.FindStringSubmatch(h); m != nil {
		d := redactPhones(text(m[1]))
		t.Description = nz(d)
	}
	if m := reTrailMap.FindStringSubmatch(h); m != nil {
		t.MapImageURL = nz(m[1])
	}
	return t
}

func urlID(u string) int {
	m := reIDPrefix.FindStringSubmatch(u)
	if m == nil {
		return 0
	}
	n, _ := strconv.Atoi(m[1])
	return n
}

func sortedUnique(re *regexp.Regexp, s string) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range re.FindAllStringSubmatch(s, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			out = append(out, m[1])
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if a, b := urlID(out[i]), urlID(out[j]); a != b {
			return a < b
		}
		return out[i] < out[j]
	})
	return out
}

func writeJSON(path string, v any) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", " ")
	return enc.Encode(v)
}

func s(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func f64(p *float64) string {
	if p == nil {
		return ""
	}
	return strconv.FormatFloat(*p, 'f', -1, 64)
}

// writeCSV scrie CSV-ul (UTF-8) și o copie cu BOM, ca Excel să afișeze corect diacriticele.
func writeCSV(path string, header []string, rows [][]string) error {
	var sb strings.Builder
	w := csv.NewWriter(&sb)
	if err := w.Write(header); err != nil {
		return err
	}
	if err := w.WriteAll(rows); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(sb.String()), 0o644); err != nil {
		return err
	}
	excel := strings.TrimSuffix(path, ".csv") + "-excel.csv"
	return os.WriteFile(excel, append([]byte("\xef\xbb\xbf"), sb.String()...), 0o644)
}

func main() {
	cache := flag.String("cache", ".cache", "director pentru cache-ul HTML brut")
	delay := flag.Duration("delay", time.Second, "pauza dintre cereri")
	out := flag.String("out", "data", "directorul de ieșire")
	flag.Parse()
	log.SetFlags(0)

	for _, d := range []string{*cache, *out} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			log.Fatal(err)
		}
	}
	f := &fetcher{cache: *cache, delay: *delay, hc: &http.Client{Timeout: 30 * time.Second}}
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")

	sitemap, err := f.get(base + "/ro/sitemap")
	if err != nil {
		log.Fatal(err)
	}
	reLoc := regexp.MustCompile(`href="(https://www\.visitarad\.ro/ro/harta-locatie/[^"]+)"`)
	reTrail := regexp.MustCompile(`href="(https://www\.visitarad\.ro/ro/traseu-turistic/[^"]+)"`)

	urls := sortedUnique(reLoc, sitemap)
	log.Printf("%d locații în sitemap", len(urls))
	locs := make([]Location, 0, len(urls))
	for i, u := range urls {
		h, err := f.get(u)
		if err != nil {
			log.Fatal(err)
		}
		l := parseLocation(h, u)
		l.RetrievedAt = now
		locs = append(locs, l)
		if (i+1)%25 == 0 {
			log.Printf("  %d/%d", i+1, len(urls))
		}
	}
	if err := writeJSON(filepath.Join(*out, "locations.json"), locs); err != nil {
		log.Fatal(err)
	}
	rows := make([][]string, 0, len(locs))
	for _, l := range locs {
		rows = append(rows, []string{strconv.Itoa(l.ID), l.Name, s(l.Type), s(l.Category), s(l.Locality), s(l.Area), s(l.Address),
			f64(l.Latitude), f64(l.Longitude), s(l.Description), s(l.Website), s(l.Facebook), strconv.Itoa(len(l.Photos)),
			l.SourceURL, s(l.CategoryURL), l.RetrievedAt, l.Source})
	}
	if err := writeCSV(filepath.Join(*out, "locations.csv"),
		[]string{"id", "name", "type", "category", "locality", "area", "address", "latitude", "longitude", "description",
			"website", "facebook", "photo_count", "source_url", "category_url", "retrieved_at", "source"}, rows); err != nil {
		log.Fatal(err)
	}
	log.Printf("gata: %d locații", len(locs))

	// Trasee: categoria vine din paginile de categorie; unele trasee apar doar în sitemap.
	idx, err := f.get(base + "/ro/trasee-turistice")
	if err != nil {
		log.Fatal(err)
	}
	reCat := regexp.MustCompile(`href="(https://www\.visitarad\.ro/ro/trasee-turistice/[^"]+)"`)
	byURL := map[string]*string{}
	for _, c := range sortedUnique(reCat, idx) {
		ch, err := f.get(c)
		if err != nil {
			log.Fatal(err)
		}
		name := c[strings.LastIndex(c, "/")+1:]
		if m := reH1.FindStringSubmatch(ch); m != nil {
			name = text(m[1])
		}
		for _, u := range sortedUnique(reTrail, ch) {
			if _, ok := byURL[u]; !ok {
				n := name
				byURL[u] = &n
			}
		}
	}
	for _, u := range sortedUnique(reTrail, sitemap) {
		if _, ok := byURL[u]; !ok {
			byURL[u] = nil
		}
	}
	tu := make([]string, 0, len(byURL))
	for u := range byURL {
		tu = append(tu, u)
	}
	sort.Slice(tu, func(i, j int) bool { return urlID(tu[i]) < urlID(tu[j]) })
	trails := make([]Trail, 0, len(tu))
	for _, u := range tu {
		h, err := f.get(u)
		if err != nil {
			log.Fatal(err)
		}
		t := parseTrail(h, u, byURL[u])
		t.RetrievedAt = now
		trails = append(trails, t)
	}
	if err := writeJSON(filepath.Join(*out, "trails.json"), trails); err != nil {
		log.Fatal(err)
	}
	trows := make([][]string, 0, len(trails))
	for _, t := range trails {
		trows = append(trows, []string{strconv.Itoa(t.ID), s(t.Name), s(t.Category), s(t.Description), s(t.MapImageURL),
			strconv.Itoa(len(t.Photos)), t.SourceURL, t.RetrievedAt, t.Source})
	}
	if err := writeCSV(filepath.Join(*out, "trails.csv"),
		[]string{"id", "name", "category", "description", "map_image_url", "photo_count", "source_url", "retrieved_at", "source"}, trows); err != nil {
		log.Fatal(err)
	}
	log.Printf("gata: %d trasee", len(trails))
}
