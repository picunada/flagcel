import { Select as SelectPrimitive } from 'bits-ui';

import Trigger from './select-trigger.svelte';
import Content from './select-content.svelte';
import Item from './select-item.svelte';

const Root = SelectPrimitive.Root;
const Group = SelectPrimitive.Group;
const GroupHeading = SelectPrimitive.GroupHeading;

export {
	Root,
	Group,
	GroupHeading,
	Trigger,
	Content,
	Item,
	//
	Root as Select,
	Group as SelectGroup,
	GroupHeading as SelectGroupHeading,
	Trigger as SelectTrigger,
	Content as SelectContent,
	Item as SelectItem
};
