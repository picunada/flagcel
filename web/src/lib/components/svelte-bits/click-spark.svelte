<!-- @svelte-bits
{
	"title": "Click Spark",
	"description": "Clickable container that emits radial spark lines on click.",
	"dependencies": []
}
-->
<script lang="ts">
	import { onMount, type Snippet } from 'svelte';

	type Easing = 'linear' | 'ease-in' | 'ease-out' | 'ease-in-out';
	type Spark = { x: number; y: number; angle: number; startTime: number };

	type Props = {
		children?: Snippet;
		sparkColor?: string;
		sparkSize?: number;
		sparkRadius?: number;
		sparkCount?: number;
		duration?: number;
		easing?: Easing;
		extraScale?: number;
		respectReducedMotion?: boolean;
		class?: string;
	};

	let {
		children,
		sparkColor = '#fff',
		sparkSize = 10,
		sparkRadius = 15,
		sparkCount = 8,
		duration = 400,
		easing = 'ease-out',
		extraScale = 1.0,
		respectReducedMotion = true,
		class: className = ''
	}: Props = $props();

	let canvas: HTMLCanvasElement;
	let wrapper: HTMLDivElement;
	let reducedMotion = $state(false);
	const sparks: Spark[] = [];

	function easeFunc(t: number): number {
		switch (easing) {
			case 'linear':
				return t;
			case 'ease-in':
				return t * t;
			case 'ease-in-out':
				return t < 0.5 ? 2 * t * t : -1 + (4 - 2 * t) * t;
			default:
				return t * (2 - t);
		}
	}

	function resolveCssColor(color: string) {
		if (!wrapper || !color.includes('var(')) return color;
		const property = color.match(/var\((--[^,\s)]+)/)?.[1];
		if (!property) return color;
		return getComputedStyle(wrapper).getPropertyValue(property).trim() || color;
	}

	onMount(() => {
		const media = window.matchMedia('(prefers-reduced-motion: reduce)');
		const syncReducedMotion = () => {
			reducedMotion = media.matches;
		};

		syncReducedMotion();
		media.addEventListener('change', syncReducedMotion);

		return () => {
			media.removeEventListener('change', syncReducedMotion);
		};
	});

	$effect(() => {
		if (!canvas || !wrapper || (respectReducedMotion && reducedMotion)) return;

		let resizeTimeout: ReturnType<typeof setTimeout> | undefined;
		let raf = 0;
		let cssWidth = 0;
		let cssHeight = 0;
		const ctx = canvas.getContext('2d');
		if (!ctx) return;
		const resolvedSparkColor = resolveCssColor(sparkColor);

		const resizeCanvas = () => {
			const rect = wrapper.getBoundingClientRect();
			const dpr = window.devicePixelRatio || 1;
			cssWidth = rect.width;
			cssHeight = rect.height;
			const width = Math.max(1, Math.round(cssWidth * dpr));
			const height = Math.max(1, Math.round(cssHeight * dpr));

			if (canvas.width !== width || canvas.height !== height) {
				canvas.width = width;
				canvas.height = height;
				canvas.style.width = `${cssWidth}px`;
				canvas.style.height = `${cssHeight}px`;
			}

			ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
		};

		const ro = new ResizeObserver(() => {
			clearTimeout(resizeTimeout);
			resizeTimeout = setTimeout(resizeCanvas, 100);
		});
		ro.observe(wrapper);
		resizeCanvas();

		const draw = (timestamp: number) => {
			ctx.clearRect(0, 0, cssWidth, cssHeight);
			for (let i = sparks.length - 1; i >= 0; i--) {
				const spark = sparks[i];
				const elapsed = timestamp - spark.startTime;
				if (elapsed >= duration) {
					sparks.splice(i, 1);
					continue;
				}
				const progress = elapsed / duration;
				const eased = easeFunc(progress);
				const distance = eased * sparkRadius * extraScale;
				const lineLength = sparkSize * (1 - eased);
				const x1 = spark.x + distance * Math.cos(spark.angle);
				const y1 = spark.y + distance * Math.sin(spark.angle);
				const x2 = spark.x + (distance + lineLength) * Math.cos(spark.angle);
				const y2 = spark.y + (distance + lineLength) * Math.sin(spark.angle);
				ctx.strokeStyle = resolvedSparkColor;
				ctx.lineWidth = 2;
				ctx.beginPath();
				ctx.moveTo(x1, y1);
				ctx.lineTo(x2, y2);
				ctx.stroke();
			}
			raf = requestAnimationFrame(draw);
		};
		raf = requestAnimationFrame(draw);

		return () => {
			ro.disconnect();
			clearTimeout(resizeTimeout);
			cancelAnimationFrame(raf);
			sparks.length = 0;
		};
	});

	function handleClick(e: MouseEvent) {
		if (!canvas || (respectReducedMotion && reducedMotion)) return;
		const rect = canvas.getBoundingClientRect();
		const x = e.clientX - rect.left;
		const y = e.clientY - rect.top;
		const now = performance.now();
		for (let i = 0; i < sparkCount; i++) {
			sparks.push({ x, y, angle: (2 * Math.PI * i) / sparkCount, startTime: now });
		}
	}
</script>

<div
	bind:this={wrapper}
	onclick={handleClick}
	role="presentation"
	class="relative w-full {className}"
>
	<canvas bind:this={canvas} class="pointer-events-none absolute inset-0 z-[60]"></canvas>
	{#if children}{@render children()}{/if}
</div>
