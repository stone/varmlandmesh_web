# varmlandmesh.se

Webbplats för Värmland Mesh, ett MeshCore-nät i Värmland.
Byggd med [Hugo](https://gohugo.io) och temat [Hextra](https://github.com/imfing/hextra).

## Kör lokalt

Kräver Hugo extended (0.169.0) och Go.

```sh
hugo server
```

Öppna http://localhost:1313.

## Ändra innehåll

| Vad | Var |
|---|---|
| Radioinställningar, region, kanal, Discord-länk | `data/radio.yaml` |
| Kommunkoder | `data/kommuner.yaml` |
| Sidor | `content/*.md` |
| Meny och inställningar för webbplatsen | `hugo.yaml` |

Skriv `{{< radio "frequency" >}}` på en sida för att visa ett värde från `data/radio.yaml`.

Lägg till en sida: skapa `content/<namn>.md` med en `title` i front matter.
Lägg till den i `menu.main` i `hugo.yaml` om den ska synas i menyn.

## CI och publicering

`.github/workflows/site.yaml` körs vid varje PR och vid varje push till `main`:

| Jobb | Vad det gör |
|---|---|
| `build` | Bygger med Hugo. Misslyckas vid varje varning. |
| `test` | Kör `htmltest` på den byggda webbplatsen: interna länkar, ankare, bilder, alt-texter. |
| `security` | `zizmor` kontrollerar workflow-filerna. `gitleaks` söker efter hemligheter i git-historiken. |
| `deploy` | Publicerar till GitHub Pages. Bara vid push till `main`, och bara om alla jobb ovan har gått igenom. |

Kör samma kontroller lokalt:

```sh
hugo --gc --minify --panicOnWarning
go run github.com/wjdp/htmltest@v0.17.0
uvx zizmor==1.32.0 .github/
go run github.com/zricethezav/gitleaks/v8@v8.30.1 git --redact .
```

## Uppdatera temat

```sh
hugo mod get -u github.com/imfing/hextra
hugo mod tidy
```
