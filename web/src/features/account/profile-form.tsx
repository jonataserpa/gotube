"use client";

import Link from "next/link";
import { useEffect, useState, useSyncExternalStore, type FormEvent } from "react";
import { accountError, accountRequest } from "@/features/account/account-client";
import {
  categoryPageSchema,
  channelDocumentSchema,
  ownChannelSchema,
  profileSchema,
  type Category,
  type OwnChannel,
  type Profile,
} from "@/features/account/profile-schema";
import { authFieldClass } from "@/features/identity/auth-panel";
import { getSession, restoreSession, subscribe } from "@/features/identity/session";

export function ProfileForm() {
  const session = useSyncExternalStore(subscribe, getSession, () => null);
  const [profile, setProfile] = useState<Profile | null>(null);
  const [owned, setOwned] = useState<OwnChannel | null>(null);
  const [categories, setCategories] = useState<Category[]>([]);
  const [notice, setNotice] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    void restoreSession();
  }, []);

  useEffect(() => {
    if (!session) return;
    const token = session.accessToken;
    let cancelled = false;
    async function load() {
      const [profileResponse, channelResponse, categoryResponse] = await Promise.all([
        accountRequest("me", token),
        accountRequest("me/channel", token),
        accountRequest("categories", token),
      ]);
      if (cancelled) return;
      if (!profileResponse.ok || !channelResponse.ok || !categoryResponse.ok) {
        setError("Não foi possível carregar o perfil.");
        return;
      }
      const parsedProfile = profileSchema.safeParse(await profileResponse.json());
      const parsedChannel = ownChannelSchema.safeParse(await channelResponse.json());
      const parsedCategories = categoryPageSchema.safeParse(await categoryResponse.json());
      if (!parsedProfile.success || !parsedChannel.success || !parsedCategories.success) {
        setError("A resposta da conta não é válida.");
        return;
      }
      setProfile(parsedProfile.data);
      setOwned(parsedChannel.data);
      setCategories(parsedCategories.data.items);
      setError(null);
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, [session]);

  if (!session) {
    return (
      <section className="mx-auto flex w-full max-w-xl flex-col gap-3 py-10">
        <h1 className="text-2xl font-semibold tracking-tight">Perfil</h1>
        <Link href="/login" className="text-sm font-medium text-accent">
          Entrar
        </Link>
      </section>
    );
  }

  return (
    <section className="mx-auto flex w-full max-w-xl flex-col gap-8 py-10">
      <h1 className="text-2xl font-semibold tracking-tight">Perfil</h1>
      {error ? <p className="text-sm text-accent">{error}</p> : null}
      {notice ? <p className="text-sm text-muted">{notice}</p> : null}
      {profile ? (
        <NameForm
          profile={profile}
          token={session.accessToken}
          onSaved={(next) => {
            setProfile(next);
            setNotice("Nome atualizado.");
            setError(null);
          }}
          onError={setError}
        />
      ) : null}
      {owned ? (
        <ChannelSection
          owned={owned}
          categories={categories}
          token={session.accessToken}
          onCreated={(next) => {
            setOwned(next);
            setNotice("Canal criado.");
            setError(null);
          }}
          onError={setError}
        />
      ) : null}
    </section>
  );
}

function NameForm({
  profile,
  token,
  onSaved,
  onError,
}: {
  profile: Profile;
  token: string;
  onSaved: (profile: Profile) => void;
  onError: (message: string) => void;
}) {
  const [pending, setPending] = useState(false);

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const displayName = String(new FormData(event.currentTarget).get("display_name") ?? "");
    setPending(true);
    const response = await accountRequest("me", token, {
      method: "PATCH",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ display_name: displayName }),
    });
    setPending(false);
    if (!response.ok) {
      onError(await accountError(response, "Não foi possível salvar o nome."));
      return;
    }
    const parsed = profileSchema.safeParse(await response.json());
    if (!parsed.success) {
      onError("A resposta da conta não é válida.");
      return;
    }
    onSaved(parsed.data);
  }

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-3">
      <p className="text-sm text-muted">{profile.email}</p>
      <label className="flex flex-col gap-1 text-sm">
        Nome de exibição
        <input
          className={authFieldClass}
          name="display_name"
          defaultValue={profile.display_name ?? ""}
          maxLength={80}
          required
        />
      </label>
      <button
        type="submit"
        className="h-10 rounded-lg bg-accent text-sm font-medium text-white disabled:opacity-70"
        disabled={pending}
      >
        {pending ? "Salvando..." : "Salvar nome"}
      </button>
    </form>
  );
}

function ChannelSection({
  owned,
  categories,
  token,
  onCreated,
  onError,
}: {
  owned: OwnChannel;
  categories: Category[];
  token: string;
  onCreated: (owned: OwnChannel) => void;
  onError: (message: string) => void;
}) {
  const [pending, setPending] = useState(false);
  const channel = owned.channel;
  if (channel) {
    return (
      <div className="flex flex-col gap-2">
        <h2 className="text-lg font-semibold">{channel.name}</h2>
        <p className="text-sm text-muted">/{channel.handle}</p>
        <Link href={`/channels/${channel.handle}`} className="text-sm font-medium text-accent">
          Ver página pública
        </Link>
      </div>
    );
  }

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    setPending(true);
    const response = await accountRequest("channels", token, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        handle: String(data.get("handle") ?? ""),
        name: String(data.get("name") ?? ""),
        description: String(data.get("description") ?? ""),
        category_ids: data.getAll("category_ids").map(String),
      }),
    });
    setPending(false);
    if (!response.ok) {
      onError(await accountError(response, "Não foi possível criar o canal."));
      return;
    }
    const parsed = channelDocumentSchema.safeParse(await response.json());
    if (!parsed.success || !parsed.data) {
      onError("A resposta do canal não é válida.");
      return;
    }
    onCreated({ channel: parsed.data });
  }

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-3">
      <h2 className="text-lg font-semibold">Criar canal</h2>
      <label className="flex flex-col gap-1 text-sm">
        Endereço
        <input className={authFieldClass} name="handle" required minLength={3} maxLength={30} />
      </label>
      <label className="flex flex-col gap-1 text-sm">
        Nome do canal
        <input className={authFieldClass} name="name" required maxLength={80} />
      </label>
      <label className="flex flex-col gap-1 text-sm">
        Descrição
        <textarea
          className="min-h-24 rounded-lg border border-line/70 bg-card px-3 py-2 text-sm text-foreground"
          name="description"
          maxLength={500}
        />
      </label>
      <fieldset className="flex flex-col gap-2">
        <legend className="text-sm">Categorias</legend>
        {categories.map((category) => (
          <label key={category.id} className="flex items-center gap-2 text-sm">
            <input type="checkbox" name="category_ids" value={category.id} />
            {category.name}
          </label>
        ))}
      </fieldset>
      <button
        type="submit"
        className="h-10 rounded-lg bg-accent text-sm font-medium text-white disabled:opacity-70"
        disabled={pending}
      >
        {pending ? "Criando..." : "Criar canal"}
      </button>
    </form>
  );
}
