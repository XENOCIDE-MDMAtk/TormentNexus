import type { Metadata } from "next";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: {
    default: "TormentNexus",
    template: "%s | TormentNexus",
  },
  description: "Local AI operations control plane for MCP routing, provider fallback, session supervision, and a unified dashboard.",
};

import { TRPCProvider } from "../utils/TRPCProvider";
import { Toaster, enterpriseRegistry } from "@tormentnexus/ui";
import { Navigation } from "../components/Navigation";
import { OidcConfig, RbacManager, AuditLogViewer } from "@tormentnexus/enterprise";

// Bind the enterprise compliance components to the runtime registry
enterpriseRegistry.OidcConfig = OidcConfig;
enterpriseRegistry.RbacManager = RbacManager;
enterpriseRegistry.AuditLogViewer = AuditLogViewer;


function getVersionLabel(): string {
  const roots = [process.cwd(), resolve(process.cwd(), '..'), resolve(process.cwd(), '..', '..')];

  for (const root of roots) {
    try {
      return readFileSync(resolve(root, 'VERSION'), 'utf8').trim();
    } catch {
      // Keep searching upward until we find the workspace VERSION file.
    }
  }

  return 'dev';
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <head>
        <script
          dangerouslySetInnerHTML={{
            __html: `
              (function() {
                const ignorePatterns = [
                  "react-devtools",
                  "React DevTools",
                  "[HMR]",
                  "[Fast Refresh]"
                ];
                function shouldIgnore(args) {
                  for (let i = 0; i < args.length; i++) {
                    const arg = args[i];
                    if (typeof arg === "string") {
                      for (const pattern of ignorePatterns) {
                        if (arg.indexOf(pattern) !== -1) {
                          return true;
                        }
                      }
                    }
                  }
                  return false;
                }
                const originalLog = console.log;
                const originalInfo = console.info;
                const originalWarn = console.warn;
                
                console.log = function(...args) {
                  if (shouldIgnore(args)) return;
                  originalLog.apply(console, args);
                };
                console.info = function(...args) {
                  if (shouldIgnore(args)) return;
                  originalInfo.apply(console, args);
                };
                console.warn = function(...args) {
                  if (shouldIgnore(args)) return;
                  originalWarn.apply(console, args);
                };
              })();
            `
          }}
        />
      </head>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <TRPCProvider>
          <div className="flex flex-col min-h-screen">
            <Navigation versionLabel={getVersionLabel()} />
            <div className="flex-1 overflow-auto min-w-0">
              {children}
            </div>
          </div>
          <Toaster />
        </TRPCProvider>
      </body>
    </html>
  );
}
