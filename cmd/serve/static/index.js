import * as THREE from 'three';

async function load() {
  const response = await fetch("/data");
  const buffer = await response.arrayBuffer();
  return MessagePack.decode(buffer);
}

function ParseLatLons(data) {
  if (!Array.isArray(data)) throw new Error("Expected an array from fetch().");

  if (data.length % 2 !== 0) throw new Error('Coordinates array must be divisible by two.');

  const len = data.length / 2;
  let pts = [];
  for (let i = 0; i < len; i++) {
    pts.push({ lat: data[i * 2], lon: data[i * 2 + 1] });
  }

  if (pts.length < 3) throw new Error("Need at least 3 points to form a polygon.");
  return pts;
}

function AlbersConstant({ phi1, phi2, phi0, lam0 }) {
  const d2r = (d) => (d * Math.PI) / 180;

  const phi1r = d2r(phi1);
  const phi2r = d2r(phi2);
  const phi0r = d2r(phi0);
  const lam0r = d2r(lam0);

  const n = 0.5 * (Math.sin(phi1r) + Math.sin(phi2r));
  const C = Math.cos(phi1r) ** 2 + 2 * n * Math.sin(phi1r);
  const Rn = 6378 / n;
  const rho0 = Rn * Math.sqrt(C - 2 * n * Math.sin(phi0r));

  return { n, C, rho0, lam0r, Rn };
}

function Albers(lat, lon, c) {
  const d2r = (d) => (d * Math.PI) / 180;
  const phir = d2r(lat);
  const lamr = d2r(lon);

  const rho = c.Rn * Math.sqrt(c.C - 2 * c.n * Math.sin(phir));
  const theta = c.n * (lamr - c.lam0r);
  const x = rho * Math.sin(theta);
  const y = c.rho0 - rho * Math.cos(theta);
  return [x, y];
}

async function main() {
  const canvas = document.getElementById("c") || document.querySelector("canvas") || undefined;
  if (!canvas) throw new Error(`No canvas found.`);

  const renderer = new THREE.WebGLRenderer({ canvas, antialias: true, alpha: true });
  renderer.setPixelRatio(Math.min(2, window.devicePixelRatio || 1));

  const scene = new THREE.Scene();
  const camera = new THREE.OrthographicCamera(-1, 1, 1, -1, -1, 1);

  const m = await load();
  const county = m.counties[150];
  const mbr = county.minimum_bounding_rectangle;
  const coordinates = county.coordinates[0];
  const latlons = ParseLatLons(coordinates);
  const c = AlbersConstant({ phi1: 29.5, phi2: 45.5, phi0: 23, lam0: 96 });
  const projected = latlons.map(({ lat, lon }) => Albers(lat, lon, c));

  const start = Albers(mbr.start.x, mbr.start.y, c);
  const end = Albers(mbr.end.x, mbr.end.y, c);
  let sv = new THREE.Vector2(start[0], start[1]);
  sv = sv.rotateAround({ x: 0, y: 0 }, 96).divideScalar(1070);
  let ev = new THREE.Vector2(end[0], end[1]);
  ev = ev.rotateAround({ x: 0, y: 0 }, 96).divideScalar(1070);
  const bounds = { minX: sv.x, maxX: ev.x, minY: sv.y, maxY: ev.y };

  // 1) Convert projected points -> Vector2s
  const pts = projected.map(([x, y]) => new THREE.Vector2(x, y));

  // 2) Compute bounds from the projected ring (more reliable than mbr math here)
  let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;
  for (const p of pts) {
    if (p.x < minX) minX = p.x;
    if (p.y < minY) minY = p.y;
    if (p.x > maxX) maxX = p.x;
    if (p.y > maxY) maxY = p.y;
  }

  const width = maxX - minX;
  const height = maxY - minY;

  // 3) Center the polygon around origin
  const cx = (minX + maxX) / 2;
  const cy = (minY + maxY) / 2;

  // 4) Scale to fit nicely in the orthographic view (leave a little margin)
  //    Our camera will be set up as [-aspect, aspect] x [-1, 1], so we fit into that.
  const margin = 0.90; // 90% of the view
  const fitScale = margin / Math.max(width, height);

  // 5) Build the THREE.Shape in normalized camera space
  //    Note: flip Y so “north is up” in screen space (common in map rendering).
  const norm = pts.map((p) => new THREE.Vector2((p.x - cx) * fitScale, (p.y - cy) * fitScale));
  for (const p of norm) p.y *= -1;

  // Ensure closed ring (Shape wants first/last to match)
  if (!norm[0].equals(norm[norm.length - 1])) norm.push(norm[0].clone());

  const shape = new THREE.Shape(norm);

  // 6) Create geometry + material
  const geom = new THREE.ShapeGeometry(shape);
  geom.computeVertexNormals();

  const fill = new THREE.MeshBasicMaterial({
    color: 0xffffff,
    transparent: true,
    opacity: 0.35,
    side: THREE.DoubleSide,
  });

  const outlineGeom = new THREE.BufferGeometry().setFromPoints(norm);
  const outlineMat = new THREE.LineBasicMaterial({ color: 0xffffff, transparent: true, opacity: 0.9 });

  const mesh = new THREE.Mesh(geom, fill);
  const outline = new THREE.Line(outlineGeom, outlineMat);

  scene.add(mesh);
  scene.add(outline);

  // 7) Resize + render loop (orthographic camera should match canvas aspect)
  function resize() {
    const w = canvas.clientWidth || window.innerWidth;
    const h = canvas.clientHeight || window.innerHeight;

    renderer.setSize(w, h, false);

    const aspect = w / h;
    camera.left = -aspect;
    camera.right = aspect;
    camera.top = 1;
    camera.bottom = -1;
    camera.near = -10;
    camera.far = 10;
    camera.position.set(0, 0, 1);
    camera.updateProjectionMatrix();
  }

  window.addEventListener("resize", resize, { passive: true });
  resize();

  function render() {
    renderer.render(scene, camera);
    requestAnimationFrame(render);
  }
  render();
}

main().catch((err) => {
  console.error(err);
  const el = document.createElement("pre");
  el.style.position = "fixed";
  el.style.left = "12px";
  el.style.top = "12px";
  el.style.padding = "12px";
  el.style.background = "rgba(0,0,0,0.75)";
  el.style.color = "white";
  el.style.font = "12px/1.4 monospace";
  el.style.zIndex = "9999";
  el.textContent = String(err?.stack || err);
  document.body.appendChild(el);
});
