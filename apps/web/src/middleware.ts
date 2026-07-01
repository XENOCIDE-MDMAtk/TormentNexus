import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

const MAPPINGS: Record<string, 'page-a' | 'page-b' | 'page-c' | 'page-d'> = {
  // Page A: System Recovery & Active Database Sync
  '/dashboard/runtime': 'page-a',
  '/dashboard/observability': 'page-a',
  '/dashboard/billing': 'page-a',
  '/dashboard/logs-metrics': 'page-a',
  '/dashboard/healer': 'page-a',
  '/dashboard/health': 'page-a',
  '/dashboard/mesh': 'page-a',
  '/dashboard/api-keys': 'page-a',
  '/dashboard/settings': 'page-a',
  '/dashboard/enterprise': 'page-a',

  // Page B: Native Go MCP Orchestration & Tool Control
  '/dashboard/mcp': 'page-b',
  '/dashboard/tool-console': 'page-b',
  '/dashboard/tool-karma': 'page-b',
  '/dashboard/inspector': 'page-b',
  '/dashboard/marketplace': 'page-b',
  '/dashboard/code': 'page-b',
  '/dashboard/swarm': 'page-b',
  '/dashboard/workshop': 'page-b',
  '/dashboard/director': 'page-b',
  '/dashboard/council': 'page-b',
  '/dashboard/squads': 'page-b',

  // Page C: Cognitive Memory Engines (L1 -> L4) & Skill Registries
  '/dashboard/brain': 'page-c',
  '/dashboard/memory': 'page-c',
  '/dashboard/memory-search': 'page-c',
  '/dashboard/memory-analytics': 'page-c',
  '/dashboard/cold-archive': 'page-c',
  '/dashboard/skills': 'page-c',
  '/dashboard/library': 'page-c',
  '/dashboard/sessions': 'page-c',
  '/dashboard/context': 'page-c',
  '/dashboard/imports': 'page-c',

  // Page D: Prompt Collections & Global Static Deployments
  '/dashboard/prompts': 'page-d',
  '/dashboard/workflows': 'page-d',
  '/dashboard/plans': 'page-d',
};

export function middleware(request: NextRequest) {
  const pathname = request.nextUrl.pathname;

  // Exact matches
  if (MAPPINGS[pathname]) {
    const tab = MAPPINGS[pathname];
    const url = request.nextUrl.clone();
    url.pathname = '/dashboard';
    url.searchParams.set('tab', tab);
    return NextResponse.redirect(url);
  }

  // Prefix matches for subpaths
  for (const [route, tab] of Object.entries(MAPPINGS)) {
    if (pathname.startsWith(route + '/')) {
      const url = request.nextUrl.clone();
      url.pathname = '/dashboard';
      url.searchParams.set('tab', tab);
      return NextResponse.redirect(url);
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/dashboard/:path*'],
};
