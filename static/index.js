let container = document.getElementById('geojson-container')
if (!container) throw new Error('Failed to load container.')

async function load() {
  let response = await fetch('counties.geojson');
  let geojson = await response.json();

  for (let i = 0; i < geojson.features.length; i++)
    container.innerHTML += '<p>feature</p>';
}

load();
