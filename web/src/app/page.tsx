import { Suspense } from "react";
import { HomePage } from "@/components/catalog/home-page";
import { loadCatalog } from "@/features/catalog/load-catalog";
import Loading from "./loading";

export default function Page({ searchParams }: PageProps<"/">) {
  return (
    <Suspense fallback={<Loading />}>
      <CatalogRequest searchParams={searchParams} />
    </Suspense>
  );
}

async function CatalogRequest({
  searchParams,
}: {
  searchParams: PageProps<"/">["searchParams"];
}) {
  const query = await searchParams;
  const cursor = typeof query.cursor === "string" ? query.cursor : undefined;
  const state = await loadCatalog(cursor);
  return <HomePage state={state} />;
}
