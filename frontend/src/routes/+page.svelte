<script lang="ts">
	import { onMount } from 'svelte';
	import * as THREE from 'three';
	import * as MessagePack from '@msgpack/msgpack';

	interface Point {
		x: number;
		y: number;
	}

	interface Rectangle {
		start: Point;
		end: Point;
	}

	interface County {
		id: String;
		name: String;
		state: String;
		intlat: number;
		intlon: number;
		minimum_bounding_rectangle: Rectangle;
		coordinates: Array<number[]>;
	}

	interface CountyMap {
		minimum_bounding_rectangle: Rectangle;
		counties: County[];
	}

	let canvas: HTMLCanvasElement | undefined = $state();

	async function load(): Promise<CountyMap> {
		const response = await fetch('/data');
		const buffer = await response.arrayBuffer();
		return MessagePack.decode<CountyMap>(buffer) as CountyMap;
	}

	// Finds if a point is in a polygon using raycasting
	function PointInPolygon(x: number, y: number, ring: number[]) {
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

		// NOTE: This application expects that the points are pre-projected
		const m = await load();
		const conus = m.counties;

		const cx = (m.minimum_bounding_rectangle.start.x + m.minimum_bounding_rectangle.end.x) / 2;
		const cy = (m.minimum_bounding_rectangle.start.y + m.minimum_bounding_rectangle.end.y) / 2;
		const width = m.minimum_bounding_rectangle.end.x - m.minimum_bounding_rectangle.start.x;
		const height = m.minimum_bounding_rectangle.end.y - m.minimum_bounding_rectangle.start.y;

		const positions = [];
		const ids = [];
		const fill_indices = [];
		const line_indices = [];
		let vertex_offset = 0;

		let id = 0;
		for (const county of conus) {
			// Translate mbr for bounds checking
			county.minimum_bounding_rectangle.start.x -= cx;
			county.minimum_bounding_rectangle.start.y -= cy;
			county.minimum_bounding_rectangle.end.x -= cx;
			county.minimum_bounding_rectangle.end.y -= cy;

			for (const part of county.coordinates) {
				if (part.length < 6) continue;

				const ring = [];
				for (let j = 0; j < part.length; j += 2) {
					const x = part[j] - cx;
					const y = part[j + 1] - cy;

					// Translate each point so that it is centered
					part[j] = x;
					part[j + 1] = y;

					ring.push(new THREE.Vector2(x, y));
					positions.push(x, y, 0);
					ids.push(id);
				}

				// Complete the loop if necessary
				if (ring.length > 0 && !ring[0].equals(ring[ring.length - 1])) {
					const first = ring[0];
					ring.push(first.clone());
					positions.push(first.x, first.y, 0);
					ids.push(id);
				}

				const vertices_count = ring.length;
				for (let j = 0; j < vertices_count - 1; j++) {
					line_indices.push(vertex_offset + j, vertex_offset + j + 1);
				}

				const triangles = THREE.ShapeUtils.triangulateShape(ring, []);
				for (let j = 0; j < triangles.length; j++) {
					const triangle = triangles[j];
					fill_indices.push(
						triangle[0] + vertex_offset,
						triangle[1] + vertex_offset,
						triangle[2] + vertex_offset
					);
				}

				vertex_offset += vertices_count;
			}
			id++;
		}

		let scale = 1.0;
		let zoom = 1.0;
		let offset = new THREE.Vector2(0, 0);

		let target_zoom = 1.0;
		let target_offset = new THREE.Vector2(0, 0);

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
			if (e.button === 2) {
				dragging = true;
				previous_mouse.set(e.clientX, e.clientY);
			} else if (e.button === 0) {
				// Smooth zoom to county
				const w = window.innerWidth;
				const h = window.innerHeight;
				const aspect = w / h;
				const mx = ((e.clientX / w) * 2 - 1) * aspect;
				const my = -((e.clientY / h) * 2 - 1);

				const lx = (mx - offset.x) / (scale * zoom);
				const ly = (my - offset.y) / (scale * zoom);

				for (const county of conus) {
					const mbr = county.minimum_bounding_rectangle;
					if (lx >= mbr.start.x && lx <= mbr.end.x && ly >= mbr.start.y && ly <= mbr.end.y) {
						let inside = false;
						for (const part of county.coordinates) {
							if (PointInPolygon(lx, ly, part)) {
								inside = true;
								break;
							}
						}
						if (inside) {
							const cw = mbr.end.x - mbr.start.x;
							const ch = mbr.end.y - mbr.start.y;
							const ctx = (mbr.start.x + mbr.end.x) / 2;
							const cty = (mbr.start.y + mbr.end.y) / 2;

							target_zoom = Math.min(1.2 / (ch * scale), (1.2 * aspect) / (cw * scale));
							target_zoom = Math.min(target_zoom, 3.0);

							target_offset.x = -ctx * scale * target_zoom;
							target_offset.y = -cty * scale * target_zoom;
							break;
						}
					}
				}
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
				target_offset.copy(offset);

				previous_mouse.set(e.clientX, e.clientY);
			}
		});

		window.addEventListener('mouseup', () => (dragging = false));

		window.addEventListener('contextmenu', (e) => {
			e.preventDefault();
		});

		window.addEventListener('keydown', (e) => {
			if (e.code === 'Escape' || e.code === 'Space') {
				target_zoom = 1.0;
				target_offset.set(0, 0);
			}
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

				target_zoom = zoom;
				target_offset.copy(offset);
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
				scale = Math.min(1.8 / height, (1.8 * aspect) / width);
				updateTransforms();
			}
		}

		window.addEventListener('resize', resize, { passive: true });
		resize();

		function render() {
			// Interpolate zoom and offset
			const lerp_factor = 0.03;
			zoom += (target_zoom - zoom) * lerp_factor;
			offset.x += (target_offset.x - offset.x) * lerp_factor;
			offset.y += (target_offset.y - offset.y) * lerp_factor;
			updateTransforms();

			// Optimized CPU picking
			const aspect = window.innerWidth / window.innerHeight;
			const mx = mouse.x * aspect;
			const my = mouse.y;

			const lx = (mx - offset.x) / (scale * zoom);
			const ly = (my - offset.y) / (scale * zoom);

			let hoveredId = -1;
			let id = 0;
			for (const county of conus) {
				if (
					lx >= county.minimum_bounding_rectangle.start.x &&
					lx <= county.minimum_bounding_rectangle.end.x &&
					ly >= county.minimum_bounding_rectangle.start.y &&
					ly <= county.minimum_bounding_rectangle.end.y
				) {
					let inside = false;
					for (const part of county.coordinates) {
						if (PointInPolygon(lx, ly, part)) {
							inside = true;
							break;
						}
					}
					if (inside) {
						hoveredId = id;
						break;
					}
				}
				id++;
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
