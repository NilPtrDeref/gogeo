let container = document.getElementById('geojson-container')
if (!container) throw new Error('Failed to load container.')

async function load() {
  const response = await fetch("/data");
  const buffer = await response.arrayBuffer();
  const bytes = new Uint8Array(buffer);
  const Counties = proto.pb.Counties;
  const counties = Counties.deserializeBinary(bytes);
  console.log(counties);

  let addition = ''
  for (let i = 0; i < counties.array[0].length; i++) {
    addition += `<p>${counties.array[0][i][0]}</p>`;
    for (let j = 0; j < counties.array[0][i][3].length; j++) {
      addition += `<p>Part ${j + 1}: ${counties.array[0][i][3].length} points</p>`;
    }
    addition += `<br/>`;
  }
  container.innerHTML = addition
  console.log('Done')
}

load();
