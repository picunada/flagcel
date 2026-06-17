import type { AuditAction, Flag } from '$lib/api';
import type { BadgeVariant } from '$lib/components/ui/badge.svelte';
import { formatFlagValue } from '$lib/values';

export function actionBadgeVariant(action: AuditAction): BadgeVariant {
	switch (action) {
		case 'created':
			return 'success';
		case 'deleted':
			return 'destructive';
		default:
			return 'default';
	}
}

/**
 * Describe what changed between two consecutive flag snapshots as a list of
 * human-readable lines. `prev` is the older state (null for the first version),
 * `next` is the state after the change (null for a deletion).
 */
export function describeChanges(prev: Flag | null, next: Flag | null): string[] {
	if (!next) return ['flag deleted'];
	if (!prev) return ['flag created'];

	const lines: string[] = [];

	if (prev.enabled !== next.enabled) {
		lines.push(`enabled ${prev.enabled} → ${next.enabled}`);
	}
	if (formatFlagValue(prev.default_value) !== formatFlagValue(next.default_value)) {
		lines.push(
			`default ${formatFlagValue(prev.default_value)} → ${formatFlagValue(next.default_value)}`
		);
	}
	if ((prev.description ?? '') !== (next.description ?? '')) {
		lines.push('description changed');
	}
	if ((prev.context_id ?? null) !== (next.context_id ?? null)) {
		lines.push('context changed');
	}

	lines.push(...describeRuleChanges(prev.rules, next.rules));

	return lines.length > 0 ? lines : ['no field changes'];
}

function describeRuleChanges(prev: Flag['rules'], next: Flag['rules']): string[] {
	const lines: string[] = [];
	const prevById = new Map(prev.map((r) => [r.id, r]));
	const nextById = new Map(next.map((r) => [r.id, r]));

	for (const rule of next) {
		const before = prevById.get(rule.id);
		if (!before) {
			lines.push(`rule added (${ruleLabel(rule)})`);
			continue;
		}
		const diffs = ruleFieldDiffs(before, rule);
		if (diffs.length > 0) {
			lines.push(`rule ${ruleLabel(rule)}: ${diffs.join(', ')}`);
		}
	}

	for (const rule of prev) {
		if (!nextById.has(rule.id)) {
			lines.push(`rule removed (${ruleLabel(rule)})`);
		}
	}

	// Detect reorder when the set of ids is unchanged but the order differs.
	if (
		prev.length === next.length &&
		prev.every((r) => nextById.has(r.id)) &&
		prev.some((r, i) => next[i]?.id !== r.id)
	) {
		lines.push('rules reordered');
	}

	return lines;
}

function ruleFieldDiffs(before: Flag['rules'][number], after: Flag['rules'][number]): string[] {
	const diffs: string[] = [];
	if (before.expression !== after.expression) {
		diffs.push('expression changed');
	}
	if (before.rollout.percentage !== after.rollout.percentage) {
		diffs.push(`rollout ${before.rollout.percentage}% → ${after.rollout.percentage}%`);
	}
	if ((before.rollout.bucket_by ?? '') !== (after.rollout.bucket_by ?? '')) {
		diffs.push('bucketing changed');
	}
	if (formatFlagValue(before.value) !== formatFlagValue(after.value)) {
		diffs.push(`value ${formatFlagValue(before.value)} → ${formatFlagValue(after.value)}`);
	}
	return diffs;
}

function ruleLabel(rule: Flag['rules'][number]): string {
	if (rule.description) return rule.description;
	const expr = rule.expression.trim();
	return expr.length > 40 ? `${expr.slice(0, 40)}…` : expr || rule.id;
}
