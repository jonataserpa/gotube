export function PublicChannelPage({
  channel,
}: {
  channel: { name: string; handle: string; description: string };
}) {
  return (
    <article className="flex max-w-3xl flex-col gap-3 py-6">
      <p className="font-mono text-sm text-muted">/{channel.handle}</p>
      <h1 className="text-2xl font-semibold tracking-tight">{channel.name}</h1>
      <p className="text-base leading-7 text-foreground">{channel.description}</p>
    </article>
  );
}
