import { NextRequest, NextResponse } from 'next/server';

const GO_SIDECAR_BASE = process.env.BORG_GO_SIDECAR_URL || 'http://127.0.0.1:4300';

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ path: string[] }> },
) {
  const resolvedParams = await params;
  const pathSegments = resolvedParams.path.join('/');
  const targetURL = `${GO_SIDECAR_BASE}/api/${pathSegments}`;

  try {
    const response = await fetch(targetURL, {
      headers: {
        'accept': 'application/json',
        ...Object.fromEntries(
          Object.entries(request.headers).filter(([key]) =>
            ['authorization', 'cookie'].includes(key.toLowerCase())
          )
        ),
      },
      signal: AbortSignal.timeout(5000),
    });

    const data = await response.json();
    return NextResponse.json(data, { status: response.status });
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Go sidecar unreachable';
    return NextResponse.json(
      { success: false, error: message, sidecarURL: targetURL },
      { status: 502 },
    );
  }
}

export async function POST(
  request: NextRequest,
  { params }: { params: Promise<{ path: string[] }> },
) {
  const resolvedParams = await params;
  const pathSegments = resolvedParams.path.join('/');
  const targetURL = `${GO_SIDECAR_BASE}/api/${pathSegments}`;

  try {
    let body: string | null = null;
    const contentType = request.headers.get('content-type') || '';
    if (contentType.includes('application/json')) {
      body = await request.text();
    }

    const response = await fetch(targetURL, {
      method: 'POST',
      headers: {
        'content-type': contentType || 'application/json',
      },
      body,
      signal: AbortSignal.timeout(10000),
    });

    const data = await response.json();
    return NextResponse.json(data, { status: response.status });
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Go sidecar unreachable';
    return NextResponse.json(
      { success: false, error: message, sidecarURL: targetURL },
      { status: 502 },
    );
  }
}
