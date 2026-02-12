import * as THREE from 'three';

async function load() {
  const response = await fetch("/data");
  const buffer = await response.arrayBuffer();
  return MessagePack.decode(buffer);
}

const d2r = (d) => (d * Math.PI) / 180;

function AlbersConstant({ phi1, phi2, phi0, lam0 }) {
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
  const phir = d2r(lat);
  const lamr = d2r(lon);

  const rho = c.Rn * Math.sqrt(c.C - 2 * c.n * Math.sin(phir));
  const theta = c.n * (lamr - c.lam0r);
  const x = rho * Math.sin(theta);
  const y = c.rho0 - rho * Math.cos(theta);
  return [x, y];
}

const vertex_shader = `
  uniform float n;
  uniform float C;
  uniform float rho0;
  uniform float lam0r;
  uniform float Rn;
  uniform vec2 center;
  uniform float scale;
  uniform float zoom;
  uniform vec2 offset;

  attribute vec2 lonlat;

  vec2 Albers(float lon, float lat) {
    float phir = lat * 0.017453292519943295;
    float lamr = lon * 0.017453292519943295;

    float rho = Rn * sqrt(C - 2.0 * n * sin(phir));
    float theta = n * (lamr - lam0r);
    float x = rho * sin(theta);
    float y = rho0 - rho * cos(theta);
    return vec2(x, y);
  }

  void main() {
    vec2 projected = (Albers(lonlat.x, lonlat.y) - center) * scale * zoom + offset;
    gl_Position = projectionMatrix * modelViewMatrix * vec4(projected, 0.0, 1.0);
  }
`;

const fragment_shader = `
  uniform vec3 color;
  uniform float opacity;
  void main() {
    gl_FragColor = vec4(color, opacity);
  }
`;

async function main() {
  const canvas = document.getElementById("c") || document.querySelector("canvas");
  if (!canvas) throw new Error(`No canvas found.`);

  const renderer = new THREE.WebGLRenderer({ canvas, antialias: true, alpha: true });
  renderer.setPixelRatio(Math.min(2, window.devicePixelRatio || 1));

  const scene = new THREE.Scene();
  const camera = new THREE.OrthographicCamera(-1, 1, 1, -1, -1, 1);

  const m = await load();
  const c = AlbersConstant({ phi1: 29.5, phi2: 45.5, phi0: 23, lam0: -96 });

  const excluded = ["AK", "HI", "PR", "GU", "AS", "VI", "MP", ""];
  const conus = m.counties.filter(county => county.state && !excluded.includes(county.state));

  let minY = Infinity, minX = Infinity, maxY = -Infinity, maxX = -Infinity;

  const lonlats = [];
  const fill_indices = [];
  const line_indices = [];
  let vertex_offset = 0;

  for (const county of conus) {
    for (const part of county.coordinates) {
      if (part.length < 6) continue;

      const ring = [];
      for (let i = 0; i < part.length; i += 2) {
        const lon = part[i];
        const lat = part[i + 1];
        ring.push(new THREE.Vector2(lon, lat));
        lonlats.push(lon, lat);

        if (lon < minX) minX = lon;
        if (lon > maxX) maxX = lon;
        if (lat < minY) minY = lat;
        if (lat > maxY) maxY = lat;
      }

      if (!ring[0].equals(ring[ring.length - 1])) {
        ring.push(ring[0].clone());
        lonlats.push(ring[0].x, ring[0].y);
      }

      const vertices_count = ring.length;
      const triangles = THREE.ShapeUtils.triangulateShape(ring, []);
      for (let i = 0; i < triangles.length; i++) {
        const triangle = triangles[i];
        fill_indices.push(triangle[0] + vertex_offset, triangle[1] + vertex_offset, triangle[2] + vertex_offset);
      }

      for (let i = 0; i < vertices_count - 1; i++) {
        line_indices.push(vertex_offset + i, vertex_offset + i + 1);
      }

      vertex_offset += vertices_count;
    }
  }

  const mina = Albers(minY, minX, c);
  const maxa = Albers(maxY, maxX, c);
  minX = Math.min(mina[0], maxa[0]);
  maxX = Math.max(mina[0], maxa[0]);
  minY = Math.min(mina[1], maxa[1]);
  maxY = Math.max(mina[1], maxa[1]);

  const cx = (minX + maxX) / 2;
  const cy = (minY + maxY) / 2;
  const width = maxX - minX;
  const height = maxY - minY;
  const scale = 3 / Math.max(width, height);

  const lonlat_attr = new THREE.Float32BufferAttribute(lonlats, 2);

  let zoom = 1.0;
  let offset = new THREE.Vector2(0, 0);

  const fill_geometry = new THREE.BufferGeometry();
  fill_geometry.setAttribute('lonlat', lonlat_attr);
  fill_geometry.setIndex(fill_indices);
  const fill_material = new THREE.ShaderMaterial({
    uniforms: {
      n: { value: c.n },
      C: { value: c.C },
      rho0: { value: c.rho0 },
      lam0r: { value: c.lam0r },
      Rn: { value: c.Rn },
      center: { value: new THREE.Vector2(cx, cy) },
      scale: { value: scale },
      zoom: { value: zoom },
      offset: { value: offset },
      color: { value: new THREE.Color(0xffffff) },
      opacity: { value: 0.35 }
    },
    vertexShader: vertex_shader,
    fragmentShader: fragment_shader,
    transparent: true,
    side: THREE.DoubleSide
  });
  scene.add(new THREE.Mesh(fill_geometry, fill_material));

  const line_geometry = new THREE.BufferGeometry();
  line_geometry.setAttribute('lonlat', lonlat_attr);
  line_geometry.setIndex(line_indices);
  const line_material = new THREE.ShaderMaterial({
    uniforms: {
      n: { value: c.n },
      C: { value: c.C },
      rho0: { value: c.rho0 },
      lam0r: { value: c.lam0r },
      Rn: { value: c.Rn },
      center: { value: new THREE.Vector2(cx, cy) },
      scale: { value: scale },
      zoom: { value: zoom },
      offset: { value: offset },
      color: { value: new THREE.Color(0xffffff) },
      opacity: { value: 0.9 }
    },
    vertexShader: vertex_shader,
    fragmentShader: fragment_shader,
    transparent: true
  });
  scene.add(new THREE.LineSegments(line_geometry, line_material));

  let dragging = false;
  let previous_mouse = new THREE.Vector2();

  window.addEventListener("mousedown", (e) => {
    if (e.button === 0) {
      dragging = true;
      previous_mouse.set(e.clientX, e.clientY);
    }
  });

  window.addEventListener("mousemove", (e) => {
    if (dragging) {
      const dx = e.clientX - previous_mouse.x;
      const dy = e.clientY - previous_mouse.y;

      // Convert screen pixels to device coordinate space
      const aspect = canvas.clientWidth / canvas.clientHeight;
      offset.x += (dx / canvas.clientWidth) * 2 * aspect;
      offset.y -= (dy / canvas.clientHeight) * 2;

      fill_material.uniforms.offset.value.copy(offset);
      line_material.uniforms.offset.value.copy(offset);
      previous_mouse.set(e.clientX, e.clientY);
    }
  });

  window.addEventListener("mouseup", () => dragging = false);

  window.addEventListener("contextmenu", (e) => {
    e.preventDefault();
    zoom = 1.0;
    offset.set(0, 0);
    fill_material.uniforms.zoom.value = zoom;
    line_material.uniforms.zoom.value = zoom;
    fill_material.uniforms.offset.value.copy(offset);
    line_material.uniforms.offset.value.copy(offset);
  });

  window.addEventListener("wheel", (e) => {
    e.preventDefault();
    let factor = e.deltaY > 0 ? 0.9 : 1.1;

    // Mouse position in device coordinate space (matching camera bounds)
    const w = canvas.clientWidth || window.innerWidth;
    const h = canvas.clientHeight || window.innerHeight;
    const aspect = w / h;
    const mx = (e.clientX / w * 2 - 1) * aspect;
    const my = -(e.clientY / h * 2 - 1);

    const old = zoom;
    zoom *= factor;
    zoom = Math.min(Math.max(zoom, 0.25), 100); // Clamp zoom

    factor = zoom / old;

    // Adjust offset to keep the point under the mouse stationary
    // offset_new = S_mouse - (S_mouse - offset_old) * (zoom_new / zoom_old)
    offset.x = mx - (mx - offset.x) * factor;
    offset.y = my - (my - offset.y) * factor;

    fill_material.uniforms.zoom.value = zoom;
    line_material.uniforms.zoom.value = zoom;
    fill_material.uniforms.offset.value.copy(offset);
    line_material.uniforms.offset.value.copy(offset);
  }, { passive: false });

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
