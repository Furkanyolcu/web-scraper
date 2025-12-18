package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	serve := flag.Bool("serve", false, "Web arayüzünü başlat")
	addr := flag.String("addr", "127.0.0.1:8080", "Sunucu adresi")
	timeoutSec := flag.Int("timeout", 30, "İstek zaman aşımı (saniye)")
	flag.Parse()

	if *serve {
		startServer(*addr, *timeoutSec)
		return
	}

	if flag.NArg() < 1 {
		fmt.Println("Kullanım:")
		fmt.Println("  go run main.go <URL>")
		fmt.Println("  go run main.go --serve [--addr 127.0.0.1:8080] [--timeout 30]")
		fmt.Println("Örnek:")
		fmt.Println("  go run main.go https://example.com")
		fmt.Println("  go run main.go --serve")
		os.Exit(1)
	}

	if err := ensureDirs(); err != nil {
		log.Fatalf("Klasörler oluşturulamadı: %v", err)
	}

	targetURL := flag.Arg(0)
	statusCode, err := scrapeOnce(targetURL, time.Duration(*timeoutSec)*time.Second)
	if err != nil {
		log.Fatalf("Hata: %v", err)
	}
	if statusCode > 0 {
		fmt.Printf("\nislem basariyla tamamlandi HTTP %d\n", statusCode)
	} else {
		fmt.Println("\nislem basariyla tamamlandi")
	}
}

// gerekli klasorler
func ensureDirs() error {
	dirs := []string{"data", "screenshot", "urls"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// site adini urlden çikar
func getSiteName(url string) string {
	// http veya https   kaldırmak
	name := strings.TrimPrefix(url, "https://")
	name = strings.TrimPrefix(name, "http://")
	
	// www kismini da kaldirmak gerekiyor
	name = strings.TrimPrefix(name, "www.")
	
	// ozel karakterleri alt cizgi ile degistiriyoruz
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "&", "_")
	name = strings.ReplaceAll(name, "=", "_")
	
	// 50 karakterle sinirlandiriyoruz uzun olmasin diye
	if len(name) > 50 {
		name = name[:50]
	}
	
	return name
}

// status kodunu almak icin bu fonksiyonu kullaniyoruz
func getStatusCode(targetURL string) (int, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Head(targetURL)
	if err != nil {
		// head calismadiysa get ile denememiz lazim
		resp, err = client.Get(targetURL)
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()
	} else {
		defer resp.Body.Close()
	}
	return resp.StatusCode, nil
}

// ana scrape islemini yapan fonksiyon bu cli ve http icin ortak
func scrapeOnce(targetURL string, timeout time.Duration) (int, error) {
	fmt.Printf("Hedef URL: %s\n", targetURL)

	// once status kodunu alalim
	statusCode, err := getStatusCode(targetURL)
	if err != nil {
		fmt.Printf("status kodu alinamadi %v\n", err)
		statusCode = 0
	} else {
		fmt.Printf("http status %d\n", statusCode)
	}

	siteName := getSiteName(targetURL)
	htmlFile := filepath.Join("data", fmt.Sprintf("%s_data.html", siteName))
	screenshotFile := filepath.Join("screenshot", fmt.Sprintf("%s_screenshot.png", siteName))
	urlsFile := filepath.Join("urls", fmt.Sprintf("%s_urls.txt", siteName))

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	var htmlContent string
	var screenshotBuf []byte
	var links []string

	err = chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body"),
		chromedp.OuterHTML("html", &htmlContent),
		chromedp.FullScreenshot(&screenshotBuf, 90),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('a')).map(a => a.href)`, &links),
	)
	if err != nil {
		return statusCode, err
	}

	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0644); err != nil {
		return statusCode, fmt.Errorf("html kaydedilemedi %w", err)
	}
	fmt.Printf("html icerigi kaydedildi %s\n", htmlFile)

	if err := os.WriteFile(screenshotFile, screenshotBuf, 0644); err != nil {
		return statusCode, fmt.Errorf("ekran goruntusu kaydedilemedi %w", err)
	}
	fmt.Printf("ekran goruntusu kaydedildi %s\n", screenshotFile)

	if len(links) > 0 {
		urlsContent := strings.Join(links, "\n")
		if err := os.WriteFile(urlsFile, []byte(urlsContent), 0644); err != nil {
			return statusCode, fmt.Errorf("urller kaydedilemedi %w", err)
		}
		fmt.Printf("%d adet url kaydedildi %s\n", len(links), urlsFile)
	}

	return statusCode, nil
}

// web arayuzunu baslatmak icin bu fonksiyonu kullaniyoruz
func startServer(addr string, timeoutSec int) {
	if err := ensureDirs(); err != nil {
		log.Fatalf("Klasörler oluşturulamadı: %v", err)
	}

	tmpl := template.Must(template.New("index").Parse(`<!doctype html>
<html lang="tr">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Go Web Scraper</title>
  <style>
	body{font-family:system-ui,-apple-system,Segoe UI,Roboto,Arial,sans-serif;margin:40px;}
	form{display:flex;gap:8px;}
	input[type=url]{flex:1;padding:10px;font-size:16px;border:1px solid #ccc;border-radius:6px;}
	button{padding:10px 16px;font-size:16px;border:0;border-radius:6px;background:#0ea5e9;color:#fff;cursor:pointer}
	button:hover{background:#0284c7}
	.msg{margin-top:16px;padding:12px;border-radius:6px;border:1px solid #ccc}
	.ok{color: #16a34a;background:#dcfce7;border-color:#bbf7d0}
	.err{color:#dc2626;background:#fee2e2;border-color:#fecaca}
	.results{margin-top:24px}
	code{background:#f3f4f6;padding:2px 6px;border-radius:4px}
	.status{margin-top:8px;font-size:14px;color:#666}
	.status.good{color:#16a34a}
	.status.bad{color:#dc2626}
  </style>
</head>
<body>
  <h1>Go Web Scraper</h1>
  <form method="POST" action="/scrape">
	<input type="url" name="url" placeholder="https://example.com" required />
	<button type="submit">Çek ve Kaydet</button>
  </form>
  {{if .Message}}
  <div class="msg {{.Class}}">
	{{.Message}}
	{{if gt .StatusCode 0}}
	<div class="status {{if lt .StatusCode 400}}good{{else}}bad{{end}}">HTTP {{.StatusCode}}</div>
	{{end}}
  </div>
  {{end}}
  {{if .Files}}
  <div class="results">
	<h3>Oluşan dosyalar</h3>
	<ul>
	  {{range .Files}}<li><code>{{.}}</code></li>{{end}}
	</ul>
  </div>
  {{end}}
</body>
</html>`))

	type pageData struct {
		Message    string
		Class      string
		Files      []string
		StatusCode int
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, pageData{})
	})

	http.HandleFunc("/scrape", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			tmpl.Execute(w, pageData{Message: "Form parse hatası", Class: "err"})
			return
		}
		u := strings.TrimSpace(r.FormValue("url"))
		if u == "" {
			tmpl.Execute(w, pageData{Message: "Lütfen bir URL girin", Class: "err"})
			return
		}
		statusCode, err := scrapeOnce(u, time.Duration(timeoutSec)*time.Second)
		if err != nil {
			msg := fmt.Sprintf("Hata: %v", err)
			if statusCode > 0 {
				msg = fmt.Sprintf("Hata (HTTP %d): %v", statusCode, err)
			}
			tmpl.Execute(w, pageData{Message: msg, Class: "err", StatusCode: statusCode})
			return
		}
		name := getSiteName(u)
		files := []string{
			filepath.Join("data", fmt.Sprintf("%s_data.html", name)),
			filepath.Join("screenshot", fmt.Sprintf("%s_screenshot.png", name)),
			filepath.Join("urls", fmt.Sprintf("%s_urls.txt", name)),
		}
		msg := "islem tamamlandi"
		if statusCode > 0 {
			msg = fmt.Sprintf("islem tamamlandi HTTP %d", statusCode)
		}
		tmpl.Execute(w, pageData{Message: msg, Class: "ok", Files: files, StatusCode: statusCode})
	})

	log.Printf("Web arayüzü başlatıldı: http://%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
