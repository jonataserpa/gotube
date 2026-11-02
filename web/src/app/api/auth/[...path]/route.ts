import { NextResponse, type NextRequest } from "next/server";
import { originAllowed } from "@/app/api/same-origin";

const allowed = new Set([
  "register",
  "login",
  "refresh",
  "logout",
  "password/forgot",
  "password/reset",
]);

export async function POST(
  request: NextRequest,
  context: { params: Promise<{ path: string[] }> },
) {
  if (!originAllowed(request)) {
    return NextResponse.json(
      { error: { code: "FORBIDDEN_ORIGIN", message: "origem não permitida" } },
      { status: 403 },
    );
  }

  const { path } = await context.params;
  const joined = path.join("/");
  if (!allowed.has(joined)) {
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
  const cookie = request.headers.get("cookie");
  if (cookie) headers.set("cookie", cookie);

  const upstream = await fetch(new URL(`/v1/auth/${joined}`, internal), {
    method: "POST",
    headers,
    body: await request.text(),
    cache: "no-store",
  });

  const responseHeaders = new Headers();
  const upstreamType = upstream.headers.get("content-type");
  if (upstreamType) responseHeaders.set("content-type", upstreamType);
  for (const setCookie of upstream.headers.getSetCookie()) {
    responseHeaders.append("set-cookie", setCookie);
  }
  if (upstream.status === 204 || upstream.status === 205 || upstream.status === 304) {
    return new NextResponse(null, { status: upstream.status, headers: responseHeaders });
  }
  return new NextResponse(await upstream.arrayBuffer(), {
    status: upstream.status,
    headers: responseHeaders,
  });
}
