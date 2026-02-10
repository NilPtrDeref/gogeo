let container = document.getElementById('geojson-container')
if (!container) throw new Error('Failed to load container.')

// TODO: Draw a minimal map to the screen (using WebGL?). Might need to simplify first.
let Map = document.getElementById('map')
if (!Map) throw new Error('Failed to load map canvas.')

async function load() {
  const response = await fetch("/data");
  const buffer = await response.arrayBuffer();
  const m = MessagePack.decode(buffer);
  console.log(m);

  let addition = ''
  for (let i = 0; i < m.counties.length; i++) {
    addition += `<p>${m.counties[i].name}</p>`;
    for (let j = 0; j < m.counties[i].coordinates.length; j++) {
      addition += `<p>Part ${j + 1}: ${m.counties[i].coordinates[j].length} points</p>`;
    }
    addition += `<br/>`;
  }
  container.innerHTML = addition
}

load();
