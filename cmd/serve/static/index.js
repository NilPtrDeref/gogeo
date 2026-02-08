let container = document.getElementById('geojson-container')
if (!container) throw new Error('Failed to load container.')

async function load() {
  const response = await fetch("/data");
  const buffer = await response.arrayBuffer();
  const counties = MessagePack.decode(buffer);
  console.log(counties);

  let addition = ''
  for (let i = 0; i < counties.length; i++) {
    addition += `<p>${counties[i].name}</p>`;
    for (let j = 0; j < counties[i].coordinates.length; j++) {
      addition += `<p>Part ${j + 1}: ${counties[i].coordinates[j].length} points</p>`;
    }
    addition += `<br/>`;
  }
  container.innerHTML = addition
}

load();
