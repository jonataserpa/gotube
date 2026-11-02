import { Suspense } from "react";
import { PublicChannelPage } from "@/components/channel/public-channel-page";
import { loadPublicChannel } from "@/features/channel/public-channel";

export default function ChannelPage({
  params,
}: {
  params: Promise<{ handle: string }>;
}) {
  return (
    <Suspense fallback={<p className="py-6 text-base text-muted">Carregando o canal.</p>}>
      <ChannelRequest params={params} />
    </Suspense>
  );
}

async function ChannelRequest({ params }: { params: Promise<{ handle: string }> }) {
  const { handle } = await params;
  const state = await loadPublicChannel(handle);
  if (state.kind === "missing") {
    return <p className="py-6 text-base text-muted">Canal não encontrado.</p>;
  }
  if (state.kind === "error") {
    return <p className="py-6 text-base text-muted">Não foi possível carregar o canal.</p>;
  }
  return <PublicChannelPage channel={state.channel} />;
}
