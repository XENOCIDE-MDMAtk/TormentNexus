"use client";

import { Sidebar } from "@/components/Sidebar";

export default function DashboardLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    return (
        <div className="flex min-h-screen bg-black">
            <Sidebar className="shrink-0" />
            <main className="flex-1 overflow-auto min-w-0">
                {children}
            </main>
        </div>
    );
}
