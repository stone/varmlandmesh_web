---
title: Kom igång
description: Från ny radio till första meddelandet i Värmland Mesh.
---

Den här guiden tar dig från en ny radio till ditt första meddelande.
Det tar ungefär 30 minuter.

{{% steps %}}

### Köp en radio

Köp en radio för **868 MHz**.
Versioner för 915 MHz eller 433 MHz fungerar inte i Sverige.

Vanliga val:

- Heltec V3 eller V4: billig, bra att börja med.
- Seeed T1000-E: liten, med batteri och GPS.
- RAK4631 (WisBlock): drar lite ström, bra för solcell.
- Heltec T114: liten och strömsnål.
- LilyGO T-Deck: har tangentbord och skärm, ingen telefon behövs.

### Installera firmware

Öppna [flasher.meshcore.io](https://flasher.meshcore.io) i Chrome eller Edge.
Anslut radion med USB-kabel.
Välj din radio och firmware **Companion** (BLE).
Klicka på flash och vänta tills det är klart.

### Installera appen

Installera MeshCore-appen på telefonen.
Den finns för Android och iOS.
Anslut till radion via Bluetooth.

### Ställ in radion

Välj förvalet **{{< radio "preset" >}}** i appen.
Kontrollera att värdena stämmer:

| Inställning | Värde |
|---|---|
| Frekvens | {{< radio "frequency" >}} MHz |
| Bandbredd | {{< radio "bandwidth" >}} kHz |
| SF | {{< radio "sf" >}} |
| CR | {{< radio "cr" >}} |

Alla radior måste ha exakt samma värden.
Annars hör de inte varandra.

### Sätt ett namn och skicka en advert

Ge din nod ett namn i appen.
Skicka sedan en **advert**.
Då ser andra noder i närheten att du finns.

### Skicka ditt första meddelande

Gå med i kanalen `{{< radio "channel" >}}` och skriv hej.
Du kan också skriva i kanalen **Public**.
Får du inget svar? Fråga på [Discord]({{< radio "discord" >}}).

{{% /steps %}}

## Vanliga frågor

{{% details title="Jag ser inga andra noder" closed="true" %}}
Kontrollera radioinställningarna.
Ett fel värde räcker för att du inte hör någon.
Det kan också vara för långt till närmaste repeater.
Se [kartan](../karta/) för var repeatrarna finns.
{{% /details %}}

{{% details title="Behöver jag en licens?" closed="true" %}}
Nej. 868 MHz är ett licensfritt band i Sverige.
Det finns gränser för sändningstid (duty cycle).
Använd de inställningar som står här så håller du dig inom gränserna.
{{% /details %}}

{{% details title="Hur gör jag nätet bättre?" closed="true" %}}
Sätt upp en repeater på en hög plats.
Se [Sätt upp en repeater](../repeater/).
{{% /details %}}
