---
title: Flasha firmware
description: Så installerar du MeshCore på din radio från Windows eller Linux.
---

Firmware är programmet som körs i radion.
Du installerar det en gång, sedan bara när du vill uppdatera.
Det tar ungefär 5 minuter.

> [!WARNING] Viktigt
> Skruva på antennen innan radion får ström.
> En radio som sänder utan antenn kan gå sönder.

## Det här behöver du

- En dator med **Chrome** eller **Edge**. Firefox och Safari fungerar inte.
- En USB-kabel som kan föra över data. Många laddkablar kan bara ladda.
- Din radio, med antennen påskruvad.

## Förbered datorn

{{< tabs >}}
{{< tab name="Windows" >}}

Oftast behöver du inte göra något.

1. Anslut radion med USB-kabeln.
2. Öppna **Enhetshanteraren** och titta under **Portar (COM och LPT)**.
3. Står det en ny COM-port där? Då är du klar.

Syns ingen port? Installera en drivrutin och anslut radion igen:

- **Heltec V3**: [CP210x-drivrutin från Silicon Labs](https://www.silabs.com/developer-tools/usb-to-uart-bridge-vcp-drivers)
- Radio med chippet CH340 eller CH9102: [CH34x-drivrutin från WCH](https://www.wch-ic.com/downloads/CH341SER_EXE.html)

{{< /tab >}}
{{< tab name="Linux" >}}

Linux låter inte vanliga användare öppna USB-porten.
Ge dig själv behörighet med ett kommando:

```sh
sudo usermod -aG dialout $USER
```

Logga ut och in igen.
Det här gör du bara en gång.

På Arch Linux heter gruppen `uucp` i stället för `dialout`.

**Ubuntu:** dyker porten upp och försvinner direkt?
Då tar paketet `brltty` (för punktskriftsskärmar) över porten.
Använder du ingen punktskriftsskärm kan du ta bort det:

```sh
sudo apt remove brltty
```

{{< /tab >}}
{{< /tabs >}}

## Flasha

{{% steps %}}

### Öppna flashern

Gå till [flasher.meshcore.io](https://flasher.meshcore.io) i Chrome eller Edge.

### Välj radio och firmware

Välj din radio i listan.
Välj sedan firmware:

- **Companion (BLE)**: för en radio du använder med appen i telefonen. Välj den här om du är ny.
- **Repeater**: för en radio som ska stå fast och skicka vidare meddelanden. Se [Sätt upp en repeater](../repeater/).

### Anslut och flasha

Klicka på **Flash** och välj radions port i rutan som dyker upp.
Vänta tills det står att det är klart.
Koppla inte ur kabeln under tiden.

{{% /steps %}}

Klart! Fortsätt med [Kom igång](../kom-igang/#installera-appen).

## Om något går fel

{{% details title="Ingen port att välja i flashern" closed="true" %}}
- Byt USB-kabel. Det är det vanligaste felet.
- Windows: installera drivrutinen ovan.
- Linux: kontrollera att du har loggat ut och in efter `usermod`.
  Kör `groups` och se att `dialout` finns med.
{{% /details %}}

{{% details title="Linux: \"Failed to open serial port\"" closed="true" %}}
Du saknar behörighet till porten.
Gör stegen under fliken Linux ovan.
Vill du bara testa snabbt, utan att logga ut:

```sh
sudo setfacl -m u:$USER:rw /dev/ttyUSB0
```

Byt `ttyUSB0` mot `ttyACM0` om radion heter så.
Kör `ls /dev/tty{USB,ACM}*` för att se vilken port den har.
{{% /details %}}

{{% details title="Flashningen startar inte eller avbryts" closed="true" %}}
**LilyGO T-Deck:** sätt radion i flashläge.
Stäng av radion och anslut kabeln.
Håll in trackballen och slå på radion.
Släpp när datorn piper.

Hjälper det inte? Klicka på **Erase Flash** i flashern och flasha igen.
Då raderas alla inställningar i radion.
{{% /details %}}

{{% details title="För dig som vill använda terminalen (ESP32)" closed="true" %}}
Ladda ner filen som slutar på `-merged.bin` från flashern och kör:

```sh
uvx esptool write-flash 0x0 firmware-merged.bin
```

Det skriver över allt i radion utom inställningarna.
{{% /details %}}

Fortfarande fast? Fråga på [Discord]({{< radio "discord" >}}).
