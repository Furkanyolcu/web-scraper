# Web Scraper

Go dilinde yazilmis bir web scraper programi. Verilen sitenin HTML icerigini ceker ekran goruntusu alir ve sayfadaki linkleri listeler.

<img width="1919" height="766" alt="Ekran görüntüsü 2025-12-18 104432" src="https://github.com/user-attachments/assets/48a3fdf8-34f3-411c-b433-b66db46f64ba" />

## Kullanim

komut satirindan tek site cekmek icin
```powershell
go run main.go <URL>
```

web arayuzu baslatmak icin
```powershell
go run main.go --serve
```
sonra tarayicida http://127.0.0.1:8080 adresine git

## Cikti Dosyalari

- `data/<site>_data.html` - HTML icerigi
- `screenshot/<site>_screenshot.png` - Ekran goruntusu  
- `urls/<site>_urls.txt` - Sayfa icindeki linkler

