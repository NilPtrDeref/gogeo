<script lang="ts">
	import { onMount } from 'svelte';
	import * as THREE from 'three';
	import * as MessagePack from '@msgpack/msgpack';

	let data = $state();
	let canvas = $state();

	async function load() {
		const response = await fetch('/data');
		const buffer = await response.arrayBuffer();
		return MessagePack.decode(buffer);
	}

	const d2r = (d: number) => (d * Math.PI) / 180;

	interface AlbersParams {
		phi1: number;
		phi2: number;
		phi0: number;
		lam0: number;
	}

	function AlbersConstant(opts: AlbersParams) {
		const phi1r = d2r(opts.phi1);
		const phi2r = d2r(opts.phi2);
		const phi0r = d2r(opts.phi0);
		const lam0r = d2r(opts.lam0);

		const n = 0.5 * (Math.sin(phi1r) + Math.sin(phi2r));
		const C = Math.cos(phi1r) ** 2 + 2 * n * Math.sin(phi1r);
		const Rn = 6378 / n;
		const rho0 = Rn * Math.sqrt(C - 2 * n * Math.sin(phi0r));

		return { n, C, rho0, lam0r, Rn };
	}

	function Albers(lat: number, lon: number, c: any) {
		const phir = d2r(lat);
		const lamr = d2r(lon);

		const rho = c.Rn * Math.sqrt(c.C - 2 * c.n * Math.sin(phir));
		const theta = c.n * (lamr - c.lam0r);
		const x = rho * Math.sin(theta);
		const y = c.rho0 - rho * Math.cos(theta);
		return [x, y];
	}

	function pointInPolygon(x: number, y: number, ring: any) {
		let inside = false;
		for (let i = 0, j = ring.length - 2; i < ring.length; i += 2) {
			const xi = ring[i],
				yi = ring[i + 1];
			const xj = ring[j],
				yj = ring[j + 1];
			const intersect = yi > y !== yj > y && x < ((xj - xi) * (y - yi)) / (yj - yi) + xi;
			if (intersect) inside = !inside;
			j = i;
		}
		return inside;
	}

	const vertex_shader = `
  attribute float id;
  varying float vId;

  void main() {
    vId = id;
    gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
  }
`;

	const fragment_shader = `
  uniform vec3 color;
  uniform float opacity;
  uniform float hoveredId;
  varying float vId;

  void main() {
    vec3 c = color;
    if (abs(vId - hoveredId) < 0.5) {
      c = vec3(1.0, 0.0, 0.0);
    }
    gl_FragColor = vec4(c, opacity);
  }
`;

	onMount(async () => {
		const renderer = new THREE.WebGLRenderer({ canvas, antialias: true, alpha: true });
		renderer.setPixelRatio(Math.min(2, window.devicePixelRatio || 1));

		const scene = new THREE.Scene();
		const camera = new THREE.OrthographicCamera(-1, 1, 1, -1, -1, 1);

		const m = await load();
		const c = AlbersConstant({ phi1: 29.5, phi2: 45.5, phi0: 23, lam0: -96 });

		const excluded = ['AK', 'HI', 'PR', 'GU', 'AS', 'VI', 'MP', ''];
		const conus = m.counties.filter((county) => county.state && !excluded.includes(county.state));

		let minPX = Infinity,
			minPY = Infinity,
			maxPX = -Infinity,
			maxPY = -Infinity;

		const countiesData = conus.map((county, index) => {
			const parts = county.coordinates.map((part) => {
				const projectedPart = [];
				for (let i = 0; i < part.length; i += 2) {
					const [px, py] = Albers(part[i + 1], part[i], c);
					projectedPart.push(px, py);
					if (px < minPX) minPX = px;
					if (px > maxPX) maxPX = px;
					if (py < minPY) minPY = py;
					if (py > maxPY) maxPY = py;
				}
				return projectedPart;
			});
			return { parts, id: index, bbox: null };
		});

		const cx = (minPX + maxPX) / 2;
		const cy = (minPY + maxPY) / 2;
		const width = maxPX - minPX;
		const height = maxPY - minPY;

		const positions = [];
		const ids = [];
		const fill_indices = [];
		const line_indices = [];
		let vertex_offset = 0;

		for (const cData of countiesData) {
			let cMinX = Infinity,
				cMaxX = -Infinity,
				cMinY = Infinity,
				cMaxY = -Infinity;

			for (const part of cData.parts) {
				if (part.length < 6) continue;

				const ring = [];
				for (let j = 0; j < part.length; j += 2) {
					const x = part[j] - cx;
					const y = part[j + 1] - cy;
					part[j] = x;
					part[j + 1] = y;

					if (x < cMinX) cMinX = x;
					if (x > cMaxX) cMaxX = x;
					if (y < cMinY) cMinY = y;
					if (y > cMaxY) cMaxY = y;

					ring.push(new THREE.Vector2(x, y));
					positions.push(x, y, 0);
					ids.push(cData.id);
				}

				if (ring.length > 0 && !ring[0].equals(ring[ring.length - 1])) {
					const first = ring[0];
					ring.push(first.clone());
					positions.push(first.x, first.y, 0);
					ids.push(cData.id);
				}

				const vertices_count = ring.length;
				const triangles = THREE.ShapeUtils.triangulateShape(ring, []);
				for (let j = 0; j < triangles.length; j++) {
					const triangle = triangles[j];
					fill_indices.push(
						triangle[0] + vertex_offset,
						triangle[1] + vertex_offset,
						triangle[2] + vertex_offset
					);
				}

				for (let j = 0; j < vertices_count - 1; j++) {
					line_indices.push(vertex_offset + j, vertex_offset + j + 1);
				}

				vertex_offset += vertices_count;
			}
			cData.bbox = [cMinX, cMaxX, cMinY, cMaxY];
		}

		let scale = 1.0;
		let zoom = 1.0;
		let offset = new THREE.Vector2(0, 0);

		const fill_geometry = new THREE.BufferGeometry();
		fill_geometry.setAttribute('position', new THREE.Float32BufferAttribute(positions, 3));
		fill_geometry.setAttribute('id', new THREE.Float32BufferAttribute(ids, 1));
		fill_geometry.setIndex(fill_indices);
		const fill_material = new THREE.ShaderMaterial({
			uniforms: {
				color: { value: new THREE.Color(0xffffff) },
				opacity: { value: 0.35 },
				hoveredId: { value: -1 }
			},
			vertexShader: vertex_shader,
			fragmentShader: fragment_shader,
			transparent: true,
			side: THREE.DoubleSide
		});
		const fill_mesh = new THREE.Mesh(fill_geometry, fill_material);
		fill_mesh.frustumCulled = false;
		scene.add(fill_mesh);

		const line_geometry = new THREE.BufferGeometry();
		line_geometry.setAttribute('position', new THREE.Float32BufferAttribute(positions, 3));
		line_geometry.setAttribute('id', new THREE.Float32BufferAttribute(ids, 1));
		line_geometry.setIndex(line_indices);
		const line_material = new THREE.ShaderMaterial({
			uniforms: {
				color: { value: new THREE.Color(0xffffff) },
				opacity: { value: 0.9 },
				hoveredId: { value: -1 }
			},
			vertexShader: vertex_shader,
			fragmentShader: fragment_shader,
			transparent: true
		});
		const line_mesh = new THREE.LineSegments(line_geometry, line_material);
		line_mesh.frustumCulled = false;
		scene.add(line_mesh);

		function updateTransforms() {
			fill_mesh.scale.set(scale * zoom, scale * zoom, 1);
			fill_mesh.position.set(offset.x, offset.y, 0);
			line_mesh.scale.set(scale * zoom, scale * zoom, 1);
			line_mesh.position.set(offset.x, offset.y, 0);
		}

		let dragging = false;
		let previous_mouse = new THREE.Vector2();
		const mouse = new THREE.Vector2();

		window.addEventListener('mousedown', (e) => {
			if (e.button === 0) {
				dragging = true;
				previous_mouse.set(e.clientX, e.clientY);
			}
		});

		window.addEventListener('mousemove', (e) => {
			mouse.x = (e.clientX / window.innerWidth) * 2 - 1;
			mouse.y = -(e.clientY / window.innerHeight) * 2 + 1;

			if (dragging) {
				const w = window.innerWidth;
				const h = window.innerHeight;
				if (w <= 0 || h <= 0) return;

				const dx = e.clientX - previous_mouse.x;
				const dy = e.clientY - previous_mouse.y;

				const aspect = w / h;
				offset.x += (dx / w) * 2 * aspect;
				offset.y -= (dy / h) * 2;

				updateTransforms();
				previous_mouse.set(e.clientX, e.clientY);
			}
		});

		window.addEventListener('mouseup', () => (dragging = false));

		window.addEventListener('contextmenu', (e) => {
			e.preventDefault();
			zoom = 1.0;
			offset.set(0, 0);
			updateTransforms();
		});

		window.addEventListener(
			'wheel',
			(e) => {
				e.preventDefault();
				const w = window.innerWidth;
				const h = window.innerHeight;
				if (w <= 0 || h <= 0) return;

				let factor = e.deltaY > 0 ? 0.9 : 1.1;

				const aspect = w / h;
				const mx = ((e.clientX / w) * 2 - 1) * aspect;
				const my = -((e.clientY / h) * 2 - 1);

				const old = zoom;
				zoom *= factor;
				zoom = Math.min(Math.max(zoom, 0.25), 100);

				factor = zoom / old;

				offset.x = mx - (mx - offset.x) * factor;
				offset.y = my - (my - offset.y) * factor;

				updateTransforms();
			},
			{ passive: false }
		);

		function resize() {
			const w = window.innerWidth;
			const h = window.innerHeight;
			if (w <= 0 || h <= 0) return;

			renderer.setSize(w, h, false);
			const aspect = w / h;
			camera.left = -aspect;
			camera.right = aspect;
			camera.top = 1;
			camera.bottom = -1;
			camera.updateProjectionMatrix();

			if (width > 0 && height > 0) {
				scale = Math.min(1.9 / height, (1.9 * aspect) / width);
				updateTransforms();
			}
		}

		window.addEventListener('resize', resize, { passive: true });
		resize();

		function render() {
			// Optimized CPU picking
			const aspect = window.innerWidth / window.innerHeight;
			const mx = mouse.x * aspect;
			const my = mouse.y;

			const lx = (mx - offset.x) / (scale * zoom);
			const ly = (my - offset.y) / (scale * zoom);

			let hoveredId = -1;
			for (const cData of countiesData) {
				const bbox = cData.bbox;
				if (lx >= bbox[0] && lx <= bbox[1] && ly >= bbox[2] && ly <= bbox[3]) {
					let inside = false;
					for (const part of cData.parts) {
						if (pointInPolygon(lx, ly, part)) {
							inside = true;
							break;
						}
					}
					if (inside) {
						hoveredId = cData.id;
						break;
					}
				}
			}
			fill_material.uniforms.hoveredId.value = hoveredId;
			line_material.uniforms.hoveredId.value = hoveredId;

			renderer.render(scene, camera);
			requestAnimationFrame(render);
		}
		render();
	});
</script>

<canvas bind:this={canvas} id="c"></canvas>

<style>
	canvas {
		width: 100%;
		height: 100%;
		display: block;
	}
</style>
