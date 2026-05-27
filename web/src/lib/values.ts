import type { FlagValue, ValueType } from '$lib/api';

export function defaultValueForType(type: ValueType, rule = false): FlagValue {
	if (type === 'boolean') return rule ? true : false;
	if (type === 'string') return '';
	if (type === 'number') return 0;
	return {};
}

export function formatFlagValue(value: FlagValue): string {
	if (typeof value === 'string') return value === '' ? '""' : value;
	if (typeof value === 'number' || typeof value === 'boolean') return String(value);
	if (value === null) return 'null';
	return JSON.stringify(value);
}

export function valueBadgeVariant(value: FlagValue): 'success' | 'muted' {
	return value === true ? 'success' : 'muted';
}
