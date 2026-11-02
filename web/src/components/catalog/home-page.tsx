import Link from "next/link";
import { formatDuration } from "@/features/catalog/format-duration";
import type { CatalogState } from "@/features/catalog/load-catalog";
import type { PublicVideo } from "@/features/catalog/catalog-schema";

const chips = [
  "Tudo",
  "Programação",
  "Go",
  "Next.js",
  "Ao vivo",
  "Novidades",
] as const;

export function HomePage({ state }: { state: CatalogState }) {
  return (
    <section className="flex flex-col gap-6 py-6" aria-labelledby="catalog-title">
      <CategoryChips />
      <div className="flex flex-col gap-4">
        <h1 id="catalog-title" className="text-2xl font-semibold tracking-tight">
          Catálogo público
        </h1>
        <CatalogBody state={state} />
      </div>
    </section>
  );
}

function CategoryChips() {
  return (
    <div className="flex gap-2 overflow-x-auto pb-1" aria-label="Categorias">
      {chips.map((chip) => {
        const active = chip === "Tudo";
        return (
          <button
            key={chip}
            type="button"
            className={
              active
                ? "rounded-full bg-foreground px-4 py-1.5 text-sm font-medium text-background"
                : "rounded-full border border-line/60 bg-card px-4 py-1.5 text-sm text-muted disabled:cursor-not-allowed"
            }
            aria-pressed={active}
            disabled={!active}
            aria-label={active ? "Tudo" : `${chip}. Filtro disponível numa etapa seguinte`}
          >
            {chip}
          </button>
        );
      })}
    </div>
  );
}

function CatalogBody({ state }: { state: CatalogState }) {
  if (state.kind === "empty") {
    return (
      <p className="max-w-2xl text-base leading-7 text-muted">
        Ainda não há vídeos públicos.
      </p>
    );
  }
  if (state.kind === "error") {
    return (
      <div className="max-w-2xl rounded-2xl border border-line/70 bg-card p-6">
        <p className="text-base leading-7">Não foi possível carregar o catálogo.</p>
        <Link href={state.retryHref} className="mt-4 inline-flex text-sm font-medium text-accent">
          Tentar de novo
        </Link>
      </div>
    );
  }
  return (
    <>
      <ul className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        {state.items.map((item) => (
          <li key={item.id}>
            <VideoCard item={item} />
          </li>
        ))}
      </ul>
      {state.nextCursor ? (
        <Link
          href={`/?cursor=${encodeURIComponent(state.nextCursor)}`}
          className="text-sm font-medium text-accent"
        >
          Próxima página
        </Link>
      ) : null}
    </>
  );
}

function VideoCard({ item }: { item: PublicVideo }) {
  return (
    <article className="overflow-hidden rounded-2xl border border-line/70 bg-card">
      <Thumbnail item={item} />
      <div className="flex flex-col gap-1 p-4">
        <h2 className="text-base font-semibold">{item.title}</h2>
        <Link href={`/channels/${item.channel.handle}`} className="text-sm text-muted">
          {item.channel.name}
        </Link>
        <p className="font-mono text-sm text-muted">{formatDuration(item.duration_seconds)}</p>
      </div>
    </article>
  );
}

function Thumbnail({ item }: { item: PublicVideo }) {
  if (!item.thumbnail_url) {
    return (
      <div
        role="img"
        aria-label="Sem miniatura"
        className="flex aspect-video items-center justify-center bg-surface-high text-sm text-muted"
      >
        Sem imagem
      </div>
    );
  }
  return (
    // Thumbnails are arbitrary public URLs, so they stay outside next/image's host allowlist.
    // eslint-disable-next-line @next/next/no-img-element
    <img
      src={item.thumbnail_url}
      alt={`Miniatura de ${item.title}`}
      className="aspect-video w-full object-cover"
    />
  );
}
