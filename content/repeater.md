---
title: Sätt upp en repeater
linkTitle: Repeater
description: Så sätter du upp en MeshCore-repeater för Värmland Mesh.
---

En repeater skickar vidare meddelanden mellan andra noder.
Den gör nätet större för alla.
Den står bäst på en hög plats med fri sikt, till exempel ett tak, en mast eller en kulle.
Okänt ord? Se [ordlistan](../ordlista/).

Instruktionerna bygger på [meshat.se](https://meshat.se/meshcore/satt-upp-egen-repeater/).

{{% steps %}}

### Installera repeater-firmware

Flasha radion med firmware **Repeater**, se [Flasha firmware](../flasha/).
Kommandona nedan skriver du i flasherns **Console**.

### Radioinställningar

{{% radio-cmd %}}

### Byt admin-lösenord

Repeatern har lösenordet `password` från början.
Med det kan vem som helst i närheten logga in och ändra repeatern.
Byt till ett eget lösenord:

```text
password DittEgetLosenord
```

Spara lösenordet. Du behöver det för att logga in på repeatern från appen.

### Namn och position

Namnet följer formatet `SE[kommunkod]-[Plats]`.
Lägg gärna till `REP` i slutet för en repeater.

```text
set name SE1780-KarlstadREP
set lat 59.3793
set lon 13.5036
```

Använd bara a–z, A–Z och siffror i namnet.
Å, ä och ö tar mer plats.
Namnet får vara högst 22 tecken om noden även delar sin position.

Positionen anges med decimaler.
Du hittar den på till exempel [latlong.net](https://www.latlong.net).
Står repeatern hemma? Flytta positionen några hundra meter.

Kommunkoder i Värmland:

{{< kommuner >}}

### Duty cycle

Kontrollera duty cycle och sätt den till 10 %, vilket gäller för {{< radio "frequency" >}} MHz:

```text
get dutycycle
set dutycycle 10
```

### Region

Sätt regionen till Värmland (`{{< radio "region" >}}`).
Tillåt även trafik utan region (`*`), så att den också kommer fram.

```text
region def eu|* europe|eu se se17
region allowf *
region save
```

Läs mer om regioner på [meshat.se](https://meshat.se/meshcore/regioner/).

### Berätta för oss

Skriv på [Discord]({{< radio "discord" >}}) var din repeater står.
Då kan vi planera nätet tillsammans.

{{% /steps %}}
