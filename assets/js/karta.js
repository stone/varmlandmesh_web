// Karta över MeshCore-noder. Läser nodes.json från collectorn (se collector/).
/* global L */

const TYPE_LABEL = {
  repeater: "Repeater",
  room: "Room server",
  companion: "Companion",
  sensor: "Sensor",
  unknown: "Okänd",
};

const REFRESH_MS = 5 * 60 * 1000;
const DAY_H = 24;
const WEEK_H = 24 * 7;

const root = document.querySelector(".karta");
if (root && window.L) init(root);

function init(root) {
  const center = root.dataset.center.split(",").map(Number);
  const status = root.querySelector(".karta-status");
  const tbody = root.querySelector(".karta-list tbody");

  const map = L.map(root.querySelector(".karta-map"), {
    preferCanvas: true,
    scrollWheelZoom: false, // aktiveras vid klick, så sidan går att scrolla förbi kartan
  }).setView(center, Number(root.dataset.zoom) || 6);
  map.once("click focus", () => map.scrollWheelZoom.enable());

  L.tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
    maxZoom: 18,
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
  }).addTo(map);

  const linkLayer = L.layerGroup().addTo(map);
  const nodeLayer = L.layerGroup().addTo(map);

  let data = null;
  let fitted = false;

  // Färgerna definieras i karta.css (canvas förstår inte CSS-variabler).
  const css = getComputedStyle(root);
  const colors = Object.fromEntries(
    Object.keys(TYPE_LABEL).map((t) => [t, css.getPropertyValue(`--karta-${t}`).trim() || "#6b7280"]),
  );
  const typeColor = (t) => colors[t] || colors.unknown;

  root.querySelectorAll("input, select").forEach((el) => el.addEventListener("change", draw));

  load();
  setInterval(load, REFRESH_MS);

  async function load() {
    try {
      const res = await fetch(root.dataset.url, { cache: "no-cache" });
      if (!res.ok) throw new Error(`${res.status}`);
      data = await res.json();
      const now = Date.now();
      for (const n of data.nodes) n.ageH = (now - Date.parse(n.last)) / 3.6e6;
      draw();
    } catch (err) {
      if (!data) status.textContent = "Kunde inte hämta noddata. Försök igen senare.";
      console.error("karta:", err);
    }
  }

  function filters() {
    const types = new Set([...root.querySelectorAll('input[name="type"]:checked')].map((i) => i.value));
    return {
      types,
      links: root.querySelector('input[name="links"]').checked,
      maxAgeH: Number(root.querySelector('select[name="age"]').value),
    };
  }

  function draw() {
    if (!data) return;
    const f = filters();
    const nodes = data.nodes.filter((n) => f.types.has(n.type in TYPE_LABEL ? n.type : "unknown") && n.ageH <= f.maxAgeH);
    const byId = new Map(nodes.map((n) => [n.id, n]));

    linkLayer.clearLayers();
    if (f.links) {
      for (const l of data.links) {
        const a = byId.get(l.a);
        const b = byId.get(l.b);
        if (!a || !b) continue;
        L.polyline([[a.lat, a.lon], [b.lat, b.lon]], {
          color: snrColor(l.snr),
          weight: 2,
          opacity: 0.7,
        })
          .bindPopup(() => linkPopup(l, a, b))
          .addTo(linkLayer);
      }
    }

    nodeLayer.clearLayers();
    // Repeaters sist så de hamnar överst.
    const order = { repeater: 3, room: 2, sensor: 1 };
    for (const n of [...nodes].sort((x, y) => (order[x.type] || 0) - (order[y.type] || 0))) {
      L.circleMarker([n.lat, n.lon], {
        radius: n.type === "repeater" ? 7 : 5,
        color: "#fff",
        weight: 1.5,
        fillColor: typeColor(n.type),
        fillOpacity: n.ageH <= DAY_H ? 0.95 : n.ageH <= WEEK_H ? 0.6 : 0.25,
      })
        .bindPopup(() => nodePopup(n))
        .bindTooltip(n.name)
        .addTo(nodeLayer);
    }

    if (!fitted && nodes.length) {
      map.fitBounds(L.latLngBounds(nodes.map((n) => [n.lat, n.lon])), { padding: [30, 30], maxZoom: 8 });
      fitted = true;
    }

    renderStatus(nodes);
    renderTable(nodes);
  }

  function renderStatus(nodes) {
    const counts = {};
    for (const n of nodes) counts[n.type] = (counts[n.type] || 0) + 1;
    const parts = Object.entries(counts)
      .sort((a, b) => b[1] - a[1])
      .map(([t, c]) => `${c} ${(TYPE_LABEL[t] || t).toLowerCase()}${c === 1 ? "" : "s"}`.replace("sensors", "sensorer"));
    let text = nodes.length ? `${nodes.length} noder: ${parts.join(", ")}.` : "Inga noder matchar filtret.";
    const genH = (Date.now() - Date.parse(data.generated)) / 3.6e6;
    text += ` Data uppdaterad ${ago(data.generated)}.`;
    if (genH > 2) text += " Datan kan vara inaktuell.";
    status.textContent = text;
  }

  function renderTable(nodes) {
    tbody.replaceChildren(
      ...[...nodes]
        .sort((a, b) => a.ageH - b.ageH)
        .map((n) => {
          const tr = document.createElement("tr");
          for (const v of [n.name, TYPE_LABEL[n.type] || n.type, ago(n.last)]) {
            const td = document.createElement("td");
            td.textContent = v;
            tr.append(td);
          }
          return tr;
        }),
    );
  }
}

// Popups byggs med DOM-element: namnen kommer från radiotrafik och är inte betrodda.
function nodePopup(n) {
  return popup(n.name, [
    ["Typ", TYPE_LABEL[n.type] || n.type],
    ["Senast hörd", ago(n.last)],
    ["Position", `${n.lat.toFixed(5)}, ${n.lon.toFixed(5)}`],
    ["ID", n.id],
  ]);
}

function linkPopup(l, a, b) {
  return popup(`${a.name} ↔ ${b.name}`, [
    ["SNR", `${l.snr.toFixed(1).replace(".", ",")} dB`],
    ["Avstånd", `${distanceKm(a, b).toFixed(1).replace(".", ",")} km`],
    ["Senast hörd", ago(l.heard)],
    ["Rapporter", String(l.count)],
  ]);
}

function popup(title, rows) {
  const el = document.createElement("div");
  el.className = "karta-popup";
  const h = document.createElement("strong");
  h.textContent = title;
  const dl = document.createElement("dl");
  for (const [k, v] of rows) {
    const dt = document.createElement("dt");
    dt.textContent = k;
    const dd = document.createElement("dd");
    dd.textContent = v;
    dl.append(dt, dd);
  }
  el.append(h, dl);
  return el;
}

function snrColor(snr) {
  if (snr >= 5) return "#16a34a";
  if (snr >= -5) return "#ca8a04";
  return "#dc2626";
}

const rtf = new Intl.RelativeTimeFormat("sv", { numeric: "auto" });
function ago(iso) {
  const s = (Date.parse(iso) - Date.now()) / 1000;
  const abs = Math.abs(s);
  if (abs < 60) return "nyss";
  if (abs < 3600) return rtf.format(Math.round(s / 60), "minute");
  if (abs < 86400) return rtf.format(Math.round(s / 3600), "hour");
  return rtf.format(Math.round(s / 86400), "day");
}

function distanceKm(a, b) {
  const rad = Math.PI / 180;
  const dLat = (b.lat - a.lat) * rad;
  const dLon = (b.lon - a.lon) * rad;
  const h = Math.sin(dLat / 2) ** 2 + Math.cos(a.lat * rad) * Math.cos(b.lat * rad) * Math.sin(dLon / 2) ** 2;
  return 2 * 6371 * Math.asin(Math.sqrt(h));
}
