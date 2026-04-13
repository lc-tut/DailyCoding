import type { Metadata } from "next";
import "./globals.css";
import { UserProvider } from "@/shared/providers/UserProvider";
import Header from "@/components/layout/Header";

export const metadata: Metadata = {
  title: "DaileyCoding - Python Coding Test",
  description: "LinuxClub Python Coding Test Platform",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ja">
      <body className="flex flex-col h-screen">
        <UserProvider>
          <Header />
          <main className="flex flex-1 min-h-0 overflow-hidden">{children}</main>
        </UserProvider>
      </body>
    </html>
  );
}
