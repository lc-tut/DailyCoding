import { cookies } from "next/headers";
import { notFound } from "next/navigation";
import { API_BASE } from "@/shared/config/api";

export default async function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const cookieStore = await cookies();
  const res = await fetch(`${API_BASE}/api/auth/me`, {
    cache: "no-store",
    headers: { Cookie: cookieStore.toString() },
  });

  if (!res.ok) notFound();

  const user = await res.json();
  if (user.role !== "admin") notFound();

  return <>{children}</>;
}
