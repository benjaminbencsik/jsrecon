# JSRecon

A fast, single-binary JavaScript analysis tool for bug bounty hunting.

JSRecon extracts endpoints, parameters, and potential secrets from JavaScript files and web pages, making it easier to discover hidden attack surfaces.

---

## Features

* Accepts flexible input:

  * Subdomains
  * URLs
  * JavaScript files
* Automatically:

  * Detects and fetches JavaScript files
  * Extracts endpoints and routes
  * Extracts parameters
  * Identifies potential secrets (API keys, tokens)
* Concurrent processing for speed
* Clean, deduplicated output

---

## Installation

Copy and paste:

```bash
git clone https://github.com/YOUR_USERNAME/jsrecon.git
cd jsrecon

go mod tidy
go build -o jsrecon
```

---

## Usage

```bash
./jsrecon input.txt
```

### Example input types

Subdomains:

```
api.example.com
cdn.example.com
```

URLs:

```
https://example.com
https://example.com/login
```

JavaScript files:

```
https://example.com/app.js
https://cdn.example.com/main.js
```

---

## Output

* endpoints.txt → discovered endpoints and routes
* params.txt → extracted parameters
* secrets.txt → potential API keys and tokens

---

## Example Workflow

```bash
cat subdomains.txt | httpx -silent > live.txt
cat live.txt | gau | grep "\.js" | sort -u > js.txt

./jsrecon js.txt
```

---

## Notes

* Output quality depends on input quality
* Best used as part of a recon pipeline
* Some detected secrets may be false positives

---

## Roadmap

* Concurrency limiting (rate control)
* Improved endpoint detection
* Parameter classification (IDOR, SSRF, etc.)
* Integration with ffuf and nuclei
* JSON output support

---

## License

MIT

---

## Contributing

Pull requests and suggestions are welcome.
