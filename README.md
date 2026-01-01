# Web Scraper

A web scraper program written in Go. It fetches the HTML content of a given website, takes a screenshot, and lists the links found on the page.

<img width="1919" height="766" alt="Screenshot 2025-12-18 104432" src="https://github.com/user-attachments/assets/48a3fdf8-34f3-411c-b433-b66db46f64ba" />

## Usage
To scrape a single website from the command line:
```powershell
go run main.go <URL>
```
To start the web interface:
```powershell
go run main.go --serve
```
Then open the following address in your browser:
http://127.0.0.1:8080

## Output Files

data/<site>_data.html – HTML content

screenshot/<site>_screenshot.png – Screenshot

urls/<site>_urls.txt – Links found on the page

## Legal & Ethical Notice
This project is for educational and research purposes only.
Use it only on websites you own or have explicit permission to scrape.
Respect robots.txt, website terms of service, and data protection laws.
