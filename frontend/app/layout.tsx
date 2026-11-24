import type { Metadata } from "next";
import "./globals.css";
export const metadata: Metadata = { title: "ScholarMate AI", description: "Your university admissions companion" };
export default function RootLayout({ children }: Readonly<{children: React.ReactNode}>) { return <html lang="en"><body>{children}</body></html>; }
