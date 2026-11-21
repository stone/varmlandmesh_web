---
title: Kom igång
description: Från ny radio till första meddelandet i Värmland Mesh.
---

Den här guiden tar dig från en ny radio till ditt första meddelande.
Det tar ungefär 30 minuter.
Okänt ord? Se [ordlistan](../ordlista/).

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

> [!WARNING] Viktigt
> Skruva på antennen innan radion får ström.
> En radio som sänder utan antenn kan gå sönder.

### Installera firmware

Följ guiden [Flasha firmware](../flasha/) och välj **Companion (BLE)**.

### Installera appen

Installera MeshCore-appen:
[Android (Google Play)](https://play.google.com/store/apps/details?id=com.liamcottle.meshcore.android)
eller [iPhone (App Store)](https://apps.apple.com/se/app/meshcore/id6742354151).
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

Gå med i kanalen `{{< radio "channel" >}}`:

1. Öppna fliken **Channels** i appen.
2. Tryck på **+**.
3. Välj **Join a hashtag channel**.
4. Skriv `{{< radio "channel" >}}` och spara.

Skriv hej i kanalen. Du kan också skriva i kanalen **Public**.
Får du inget svar? Fråga på [Discord]({{< radio "discord" >}}).

{{% /steps %}}

## Få bättre räckvidd

- **Högre är bättre.** En radio i ett fönster på övervåningen når mycket längre än en i fickan.
- **Antennen ska stå rakt upp.**
- **Byt antenn.** Antennen som följer med är ofta dålig.
  En bra antenn för 868 MHz gör oftast mer nytta än en dyrare radio.
  Kontrollera att den är gjord för 868 MHz.
- **Metall, betong och skog dämpar signalen.**

Ungefär hur långt det når:

| Mellan | Räckvidd |
|---|---|
| Två radior i handen, i stan | några hundra meter till ett par km |
| Din radio och en repeater på hög plats | ofta 10 km eller mer |
| Två höga platser med fri sikt | 50 km eller mer |

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

{{% details title="Appen hittar inte radion" closed="true" %}}
- Kontrollera att du har flashat **Companion (BLE)**.
  Repeater-firmware går inte att ansluta till via Bluetooth.
- PIN-koden är oftast `123456` eller `0000`.
  Har radion en skärm visas koden där.
- Har du anslutit förut? Ta bort radion i telefonens Bluetooth-inställningar.
  Anslut sedan igen från appen.
- Heltec V3: den inbyggda Bluetooth-antennen når bara någon meter.
  Ha telefonen nära radion.
{{% /details %}}

{{% details title="Flashningen fungerar inte" closed="true" %}}
Se [Om något går fel](../flasha/#om-något-går-fel) i guiden för flashning.
{{% /details %}}

{{% details title="Hur vet jag att mitt meddelande kom fram?" closed="true" %}}
I en kanal får du ingen bekräftelse.
Ett direktmeddelande till en kontakt får en kvittens när det har kommit fram.
Kommer det inte fram försöker radion igen, upp till 3 gånger.
{{% /details %}}
