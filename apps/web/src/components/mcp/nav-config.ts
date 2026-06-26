import {
	Server,
	LayoutDashboard,
	Box,
	Globe,
	Key,
	Layers,
	Shield,
	FileCode,
	FileText,
	Settings,
	Search,
	BookOpen,
	Activity,
	Zap,
	Bot,
	Wrench,
	Download,
	Rocket,
	Brain,
	FlaskConical,
	Terminal,
	FileSearch,
	Settings2,
	Workflow,
	Library,
	BookOpenText,
	BarChart3,
	Hammer,
	Users,
	Eye,
	Heart,
	BookMarked,
	Building2,
	Lightbulb,
	Cog,
	FileCode2,
	ScrollText,
	Sparkles,
	Radio,
	Network,
	ShoppingBag,
	Power,
	Stethoscope,
	BookCopy,
	Siren,
	FlaskRound,
	Timer,
	Scroll,
	GalleryThumbnails,
	Cpu,
	MonitorCheck,
} from "lucide-react";

export interface NavItem {
	title: string;
	href: string;
	icon: any;
	variant: "default" | "ghost";
}

export interface NavSection {
	title: string;
	items: NavItem[];
}

export const META_MCP_NAV: NavItem[] = [
	{
		title: "MCP Dashboard",
		href: "/dashboard/mcp",
		icon: Server,
		variant: "default",
	},
	{
		title: "Tool Catalog",
		href: "/dashboard/mcp/catalog",
		icon: Search,
		variant: "ghost",
	},
	{
		title: "Tools Inspector",
		href: "/dashboard/mcp/inspector",
		icon: Wrench,
		variant: "ghost",
	},
	{
		title: "System Audit",
		href: "/dashboard/mcp/audit",
		icon: FileText,
		variant: "ghost",
	},
	{
		title: "MCP Registry",
		href: "/dashboard/mcp/registry",
		icon: Download,
		variant: "ghost",
	},
	{
		title: "Configuration",
		href: "/dashboard/mcp/settings",
		icon: Settings,
		variant: "ghost",
	},
	{
		title: "Always-On Tools",
		href: "/dashboard/mcp/always-on",
		icon: Power,
		variant: "ghost",
	},
];

export const INTEGRATIONS_NAV: NavItem[] = [
	{
		title: "Integrations Hub",
		href: "/dashboard/integrations",
		icon: Globe,
		variant: "ghost",
	},
	{
		title: "Cloud Orchestrator",
		href: "/dashboard/cloud-orchestrator",
		icon: Rocket,
		variant: "ghost",
	},
];

export const MAIN_DASHBOARD_NAV: NavItem[] = [
	{
		title: "Dashboard Home",
		href: "/dashboard",
		icon: LayoutDashboard,
		variant: "ghost",
	},
	{
		title: "Swarm & Agents",
		href: "/dashboard/swarm",
		icon: Users,
		variant: "ghost",
	},
	{
		title: "Brain & Memory",
		href: "/dashboard/brain",
		icon: Brain,
		variant: "ghost",
	},
	{
		title: "Code Platform",
		href: "/dashboard/code",
		icon: FileCode2,
		variant: "ghost",
	},
	{
		title: "Command Console",
		href: "/dashboard/command",
		icon: Terminal,
		variant: "ghost",
	},
	{
		title: "Workflows",
		href: "/dashboard/workflows",
		icon: Workflow,
		variant: "ghost",
	},
	{
		title: "Billing & Plans",
		href: "/dashboard/billing",
		icon: Key,
		variant: "ghost",
	},
	{
		title: "Settings",
		href: "/dashboard/settings",
		icon: Settings2,
		variant: "ghost",
	},
];

export const WORKSHOP_NAV: NavItem[] = [
	{
		title: "Diagnostics & Research",
		href: "/dashboard/research",
		icon: FlaskRound,
		variant: "ghost",
	},
	{
		title: "Context & Sessions",
		href: "/dashboard/session",
		icon: Scroll,
		variant: "ghost",
	},
	{
		title: "Knowledge & Skills",
		href: "/dashboard/library",
		icon: Library,
		variant: "ghost",
	},
	{
		title: "Security Operations",
		href: "/dashboard/security",
		icon: Shield,
		variant: "ghost",
	},
];

export const SIDEBAR_SECTIONS: NavSection[] = [
	{
		title: "MCP Platform",
		items: META_MCP_NAV,
	},
	{
		title: "Core System",
		items: MAIN_DASHBOARD_NAV,
	},
	{
		title: "Diagnostics & Tools",
		items: WORKSHOP_NAV,
	},
	{
		title: "Integrations",
		items: INTEGRATIONS_NAV,
	},
];
