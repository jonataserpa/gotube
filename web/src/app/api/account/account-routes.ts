const routes: Record<string, { methods: Set<string>; upstream: string }> = {
  me: { methods: new Set(["GET", "PATCH"]), upstream: "/v1/me" },
  "me/channel": { methods: new Set(["GET"]), upstream: "/v1/me/channel" },
  channels: { methods: new Set(["POST"]), upstream: "/v1/channels" },
  categories: { methods: new Set(["GET"]), upstream: "/v1/categories" },
};

const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

export function accountUpstream(method: string, path: string[]): string | null {
  const fixed = routes[path.join("/")];
  if (fixed?.methods.has(method)) return fixed.upstream;
  if (method === "GET" && path.length === 3 && path[0] === "videos" && path[2] === "processing" && uuid.test(path[1] ?? "")) {
    return `/v1/videos/${path[1]}/processing`;
  }
  if (method !== "POST") return null;
  if (path.length === 1 && path[0] === "videos") return "/v1/videos";
  if (path.length === 3 && path[0] === "videos" && path[2] === "upload-sessions" && uuid.test(path[1] ?? "")) {
    return `/v1/videos/${path[1]}/upload-sessions`;
  }
  if (path.length === 3 && path[0] === "upload-sessions" && path[2] === "complete" && uuid.test(path[1] ?? "")) {
    return `/v1/upload-sessions/${path[1]}/complete`;
  }
  if (path.length === 3 && path[0] === "videos" && path[2] === "retry-processing" && uuid.test(path[1] ?? "")) {
    return `/v1/videos/${path[1]}/retry-processing`;
  }
  return null;
}
