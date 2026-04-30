# jsrecon

A fast, single-binary JavaScript analysis tool for bug bounty hunting.

jsrecon extracts endpoints, parameters, and potential secrets from JavaScript files and web pages.

---

## Installation

```bash
go install github.com/benjaminbencsik/jsrecon@latest
```

Make sure your Go bin directory is in PATH:

```bash
export PATH=$PATH:~/go/bin
```

---

## Usage

```bash
jsrecon input.txt
```

---

## Input Examples

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
```

---

## Output

* endpoints.txt
* params.txt
* secrets.txt

---

## Example Workflow

```bash
cat subs.txt | httpx -silent > live.txt
cat live.txt | gau | grep "\.js" > js.txt

jsrecon js.txt
```

---

## Notes

* Works best with good recon input
* Some secrets may be false positives

---

## License

MIT
