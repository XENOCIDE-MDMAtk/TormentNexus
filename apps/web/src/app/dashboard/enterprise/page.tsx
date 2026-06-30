"use client";

import { useState, useEffect, useCallback } from "react";
import { Shield, Key, Users, FileText, Loader2, RotateCcw, CheckCircle, XCircle, RefreshCw } from "lucide-react";

interface LicenseInfo {
  valid: boolean;
  licensedTo?: string;
  expiresAt?: string;
  features?: string[];
  maxNodes?: number;
  tier?: string;
}

interface RBACRole {
  name: string;
  permissions: string[];
  description?: string;
}

export default function EnterprisePage() {
  const [license, setLicense] = useState<LicenseInfo | null>(null);
  const [auditLogs, setAuditLogs] = useState<any[]>([]);
  const [roles, setRoles] = useState<RBACRole[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchAll = useCallback(async () => {
    setLoading(true);
    try {
      const [licenseRes, auditRes, rolesRes] = await Promise.all([
        fetch("/api/go/api/enterprise/license").catch(() => null),
        fetch("/api/go/api/enterprise/audit?limit=20").catch(() => null),
        fetch("/api/go/api/enterprise/roles").catch(() => null),
      ]);
      if (licenseRes?.ok) {
        const d = await licenseRes.json();
        setLicense(d.data ?? d);
      }
      if (auditRes?.ok) {
        const d = await auditRes.json();
        setAuditLogs(d.data ?? []);
      }
      if (rolesRes?.ok) {
        const d = await rolesRes.json();
        setRoles(d.data ?? []);
      }
    } catch {
      // Best-effort
    }
    setLoading(false);
  }, []);

  useEffect(() => { fetchAll(); }, [fetchAll]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Shield className="w-5 h-5 text-amber-400" />
          <div>
            <h1 className="text-lg font-semibold text-white">Enterprise Security</h1>
            <p className="text-xs text-zinc-500 mt-0.5">
              License validation, RBAC permissions, and audit logging
            </p>
          </div>
        </div>
        <button
          onClick={fetchAll}
          disabled={loading}
          className="px-3 py-1.5 bg-zinc-800 rounded hover:bg-zinc-700 text-xs disabled:opacity-50 flex items-center gap-1.5"
          title="Refresh enterprise security data"
        >
          {loading ? <Loader2 className="w-3 h-3 animate-spin" /> : <RefreshCw className="w-3 h-3" />}
          Refresh
        </button>
      </div>

      {/* License Card */}
      <div className="bg-zinc-900/50 border border-zinc-800 rounded-lg p-4">
        <div className="flex items-center gap-2 mb-3">
          <Key className="w-4 h-4 text-amber-400" />
          <h2 className="text-sm font-medium text-white">License</h2>
        </div>
        {license ? (
          <div className="space-y-2 text-sm">
            <div className="flex items-center gap-2">
              {license.valid ? (
                <CheckCircle className="w-4 h-4 text-emerald-400" />
              ) : (
                <XCircle className="w-4 h-4 text-red-400" />
              )}
              <span className={license.valid ? "text-emerald-400" : "text-red-400"}>
                {license.valid ? "Valid License" : "No License Found"}
              </span>
            </div>
            {license.licensedTo && (
              <div className="text-zinc-400">
                Licensed to: <span className="text-zinc-300">{license.licensedTo}</span>
              </div>
            )}
            {license.tier && (
              <div className="text-zinc-400">
                Tier: <span className="text-zinc-300">{license.tier}</span>
              </div>
            )}
            {license.expiresAt && (
              <div className="text-zinc-400">
                Expires: <span className="text-zinc-300">{license.expiresAt}</span>
              </div>
            )}
            {license.maxNodes && (
              <div className="text-zinc-400">
                Max nodes: <span className="text-zinc-300">{license.maxNodes}</span>
              </div>
            )}
            {license.features && license.features.length > 0 && (
              <div>
                <span className="text-zinc-500 text-xs">Features</span>
                <div className="flex gap-1 mt-1 flex-wrap">
                  {license.features.map((f) => (
                    <span key={f} className="px-1.5 py-0.5 bg-zinc-800 rounded text-2xs text-zinc-400">{f}</span>
                  ))}
                </div>
              </div>
            )}
          </div>
        ) : (
          <div className="text-zinc-600 text-sm italic">
            {loading ? "Checking license..." : "No license information available"}
          </div>
        )}
      </div>

      {/* RBAC Roles */}
      <div className="bg-zinc-900/50 border border-zinc-800 rounded-lg p-4">
        <div className="flex items-center gap-2 mb-3">
          <Users className="w-4 h-4 text-blue-400" />
          <h2 className="text-sm font-medium text-white">RBAC Roles & Permissions</h2>
        </div>
        {roles.length === 0 ? (
          <div className="text-zinc-600 text-sm italic">
            {loading ? "Loading roles..." : "No roles configured"}
          </div>
        ) : (
          <div className="space-y-2">
            {roles.map((role) => (
              <div key={role.name} className="bg-zinc-950/50 rounded p-3 border border-zinc-800">
                <div className="flex items-center justify-between mb-1">
                  <span className="text-sm font-medium text-zinc-300">{role.name}</span>
                  {role.description && (
                    <span className="text-2xs text-zinc-600">{role.description}</span>
                  )}
                </div>
                <div className="flex gap-1 flex-wrap">
                  {role.permissions.map((perm) => (
                    <span key={perm} className="px-1.5 py-0.5 bg-zinc-800 rounded text-2xs text-zinc-400">
                      {perm}
                    </span>
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Audit Logs */}
      <div className="bg-zinc-900/50 border border-zinc-800 rounded-lg p-4">
        <div className="flex items-center gap-2 mb-3">
          <FileText className="w-4 h-4 text-purple-400" />
          <h2 className="text-sm font-medium text-white">Audit Log (Last 20)</h2>
        </div>
        {auditLogs.length === 0 ? (
          <div className="text-zinc-600 text-sm italic">
            {loading ? "Loading audit logs..." : "No audit entries recorded"}
          </div>
        ) : (
          <div className="space-y-1 max-h-60 overflow-y-auto">
            {auditLogs.map((log: any, i: number) => (
              <div key={i} className="text-xs flex gap-2 py-1 border-b border-zinc-800/50 last:border-0">
                <span className="text-zinc-600 font-mono shrink-0 w-16">
                  {log.timestamp?.slice(11, 19) || log.timestamp?.slice(0, 10) || "?"}
                </span>
                <span className="text-zinc-500 font-mono shrink-0 w-12">{log.action?.slice(0, 12) || "?"}</span>
                <span className="text-zinc-400 truncate">{log.detail || JSON.stringify(log).slice(0, 80)}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
