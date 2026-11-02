import { NextResponse, type NextRequest } from "next/server";
import { accountUpstream } from "@/app/api/account/account-routes";
import { originAllowed } from "@/app/api/same-origin";

async function forward(request: NextRequest, context: { params: Promise<{ path: string[] }> }) {
  if (!originAllowed(request)) {
    return NextResponse.json(
      { error: { code: "FORBIDDEN_ORIGIN", message: "origem não permitida" } },
      { status: 403 },
    );
  }

  const { path } = await context.params;
  const upstreamPath = accountUpstream(request.method, path);
  if (!upstreamPath) {
    return NextResponse.json(
      { error: { code: "NOT_FOUND", message: "rota não encontrada" } },
      { status: 404 },
    );
  }

  const internal = process.env.GO_API_INTERNAL_URL;
  if (!internal) {
    return NextResponse.json(
      {
        error: {
          code: "DATABASE_UNAVAILABLE",
          message: "o serviço de conta não está disponível",
        },
      },
      { status: 503 },
    );
  }

  const headers = new Headers();
  const contentType = request.headers.get("content-type");
  if (contentType) headers.set("content-type", contentType);
  const authorization = request.headers.get("authorization");
  if (authorization) headers.set("authorization", authorization);

  const upstream = await fetch(new URL(upstreamPath, internal), {
    method: request.method,
    headers,
    body: request.method === "GET" ? undefined : await request.text(),
    cache: "no-store",
  });

  const responseHeaders = new Headers();
  const upstreamType = upstream.headers.get("content-type");
  if (upstreamType) responseHeaders.set("content-type", upstreamType);
  if (upstream.status === 204 || upstream.status === 205 || upstream.status === 304) {
    return new NextResponse(null, { status: upstream.status, headers: responseHeaders });
  }
  return new NextResponse(await upstream.arrayBuffer(), {
    status: upstream.status,
    headers: responseHeaders,
  });
}

export function GET(request: NextRequest, context: { params: Promise<{ path: string[] }> }) {
  return forward(request, context);
}

export function PATCH(request: NextRequest, context: { params: Promise<{ path: string[] }> }) {
  return forward(request, context);
}

export function POST(request: NextRequest, context: { params: Promise<{ path: string[] }> }) {
  return forward(request, context);
}
