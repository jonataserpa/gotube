import { ResetForm } from "@/features/identity/reset-form";

export default async function ResetPasswordPage({
  searchParams,
}: {
  searchParams: Promise<{ token?: string }>;
}) {
  const params = await searchParams;
  return <ResetForm token={params.token ?? ""} />;
}
