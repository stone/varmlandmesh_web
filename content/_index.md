---
title: Värmland Mesh
layout: hextra-home
---

{{< hextra/hero-badge link="https://discord.com/invite/gchSzHkPGG" >}}
  Gå med på Discord
{{< /hextra/hero-badge >}}

<div class="hx:mt-6 hx:mb-6">
{{< hextra/hero-headline >}}
  Ett eget radionät&nbsp;<br class="hx:sm:block hx:hidden" />för Värmland
{{< /hextra/hero-headline >}}
</div>

<div class="hx:mb-12">
{{< hextra/hero-subtitle >}}
  Skicka textmeddelanden över radio med MeshCore.&nbsp;<br class="hx:sm:block hx:hidden" />Det fungerar utan mobilnät och utan internet.
{{< /hextra/hero-subtitle >}}
</div>

<div class="hx:mb-12">
{{< hextra/hero-button text="Kom igång" link="kom-igang" >}}
</div>

{{< hextra/feature-grid >}}
  {{< hextra/feature-card
    title="Kom igång"
    icon="light-bulb"
    link="kom-igang"
    subtitle="Köp en radio, installera appen och skicka ditt första meddelande."
  >}}
  {{< hextra/feature-card
    title="Sätt upp en repeater"
    icon="status-online"
    link="repeater"
    subtitle="En repeater på en hög plats gör nätet större för alla."
  >}}
  {{< hextra/feature-card
    title="Karta"
    icon="map"
    link="karta"
    subtitle="Se vilka noder som finns i Värmland."
  >}}
{{< /hextra/feature-grid >}}

<div class="content hx:w-full hx:mt-12">

## Om nätet

Värmland Mesh är en grupp privatpersoner som bygger ett MeshCore-nät i Värmland.
MeshCore använder LoRa-radio på 868 MHz.
Varje radio kan skicka vidare meddelanden till andra radior.
Repeatrar på höga platser gör att meddelanden når längre.
Nätet är öppet för alla och kostar inget att använda.

## Radioinställningar

Alla noder i Värmland använder samma inställningar.
I appen väljer du förvalet **{{< radio "preset" >}}**.

| Inställning | Värde |
|---|---|
| Frekvens | {{< radio "frequency" >}} MHz |
| Bandbredd | {{< radio "bandwidth" >}} kHz |
| Spreading factor (SF) | {{< radio "sf" >}} |
| Coding rate (CR) | {{< radio "cr" >}} |
| Region | `{{< radio "region" >}}` |

## Kanaler

| Kanal | Används till |
|---|---|
| Public | Den öppna kanalen. Alla MeshCore-användare ser den. |
| `{{< radio "channel" >}}` | Prat om nätet i Värmland. |

## Namn på noder

Fasta noder får namn efter formatet `SE[kommunkod]-[Plats]`, till exempel `SE1780-KarlstadREP`.
Se [Sätt upp en repeater](repeater/#namn-och-position) för alla kommunkoder.

## Gå med

Frågor, hjälp och planering sker på [Discord]({{< radio "discord" >}}).
Alla är välkomna, även om du bara är nyfiken.

</div>
